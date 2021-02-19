package asm

import (
	"fmt"
	"time"
)

const (
	// Frequency1Hz machine main 1Hz frequency
	Frequency1Hz = 1 * time.Second
	// Frequency10Hz machine main 1Hz frequency
	Frequency10Hz = time.Duration(int(float32(Frequency1Hz / 10)))
	// Frequency800kHz machine main 1MHz frequency
	Frequency800kHz = time.Duration(int(float32(Frequency1Hz / 800000)))
	// Frequency1MHz machine main 1MHz frequency
	Frequency1MHz = time.Duration(int(float32(Frequency1Hz / 1000000)))
)

// BreakePoint Trace break point
type BreakePoint struct {
	Addr uint
	Func func(m *Machine)
}

// Machine 8051 microchip
type Machine struct {
	mainTick *time.Ticker
	exitCh   chan int
	isrCh    chan int

	DATA         [0x100]byte   // RAM: DATA Range
	XDATA        [0x10000]byte // RAM: XDATA Range
	ROM          []byte        // ROM: CODE Range
	PC           uint          // PC: program counter
	brakepoints  map[uint][]func(m *Machine)
	regDefines   []Register
	insHookDATAR map[uint8][]func(m *Machine, val uint8)
	insHookDATAW map[uint8][]func(m *Machine, old uint8, new uint8)
	Frequency    time.Duration
}

// NewMachine Create 8051 machine
func NewMachine(f time.Duration) *Machine {
	m := &Machine{}
	m.brakepoints = make(map[uint][]func(m *Machine))
	m.insHookDATAR = make(map[uint8][]func(m *Machine, val uint8))
	m.insHookDATAW = make(map[uint8][]func(m *Machine, old uint8, val uint8))
	m.Frequency = f
	m.regDefines = regList

	m.insideHookDATAWrite(FindRegByName("P1", regList).Addr, func(m *Machine, old uint8, new uint8) {
		fmt.Printf("DATA W %02X : old(%02X) new(%02X)\n", P1, old, new)
	})
	m.insideHookDATAWrite(FindRegByName("P0", regList).Addr, func(m *Machine, old uint8, new uint8) {
		fmt.Printf("DATA W %02X : old(%02X) new(%02X)\n", P0, old, new)
	})
	// m.insideHookDATAWrite(R0, func(m *Machine, old uint8, new uint8) {
	// 	log.Printf("PC(%04X) DATA W %02X : old(%02X) new(%02X)\n", m.PC, R0, old, new)
	// })
	return m
}

// DumpFakeCode dump asm fakecode
func (m Machine) DumpFakeCode() (string, error) {
	var (
		err  error
		pc   uint   = 0
		code string = ""
		ins  *INS
	)

	for ; pc < uint(len(m.ROM)); pc += uint(ins.Bytes) {
		ins, err = FindINS(m.ROM[pc])
		if err != nil {
			return code, err
		}
		bbb := ""
		for i := uint(0); i < uint(ins.Bytes); i++ {
			bbb += fmt.Sprintf("%02X", m.ROM[pc+i])
		}
		code += fmt.Sprintf("%04X\t%s\t%s", pc, bbb, ins.Mnemonic)
		if ins.FakeCode != nil {
			code += fmt.Sprintf("\t%s", ins.FakeCode(m, pc))
		}
		code += fmt.Sprintf("\n")
	}
	return code, nil
}

// Start 8051 machine
func (m *Machine) Start() {
	m.mainTick = time.NewTicker(m.Frequency)
	m.exitCh = make(chan int, 1)
	m.isrCh = make(chan int, 1)
	defer m.mainTick.Stop()
	defer close(m.exitCh)
	defer close(m.isrCh)
	for {
		select {
		case <-m.isrCh:
		case <-m.mainTick.C:
			m.Single()
		case <-m.exitCh:
			return
		}
	}
}

// Stop 8051 machine
func (m *Machine) Stop() {
	m.exitCh <- 1
}

// Single 8051 machine
func (m *Machine) Single() {
	i, err := FindINS(m.ROM[m.PC])
	if err != nil {
		fmt.Printf("FindINS: PC:%X %s\n", m.PC, err)
		m.Stop()
		return
	}
	fns, ok := m.brakepoints[m.PC]
	if ok && (len(fns) != 0) {
		for i := range fns {
			fns[i](m)
		}
	}
	if i.Func != nil {
		i.Func(m)
	}
}

// ReadDATA read mechine DATA range
func (m *Machine) ReadDATA(addr uint8) uint8 {
	val := m.DATA[addr]
	if hooks, ok := m.insHookDATAR[addr]; ok {
		for _, hook := range hooks {
			hook(m, val)
		}
	}
	return val
}

// WriteDATA write mechine DATA range
func (m *Machine) WriteDATA(addr uint8, val uint8) {
	if hooks, ok := m.insHookDATAW[addr]; ok {
		for _, hook := range hooks {
			hook(m, m.DATA[addr], val)
		}
	}
	m.DATA[addr] = val
}

// Trace breakepoint
func (m *Machine) Trace(addr uint, fn func(m *Machine)) {
	if m.brakepoints[addr] == nil {
		m.brakepoints[addr] = make([]func(m *Machine), 0)
	}
	m.brakepoints[addr] = append(m.brakepoints[addr], fn)
}

func (m *Machine) insideHookDATARead(addr uint8, fn func(m *Machine, val uint8)) {
	if m.insHookDATAR[addr] == nil {
		m.insHookDATAR[addr] = make([]func(m *Machine, val uint8), 0)
	}
	m.insHookDATAR[addr] = append(m.insHookDATAR[addr], fn)
}

func (m *Machine) insideHookDATAWrite(addr uint8, fn func(m *Machine, old uint8, new uint8)) {
	if m.insHookDATAW[addr] == nil {
		m.insHookDATAW[addr] = make([]func(m *Machine, old uint8, new uint8), 0)
	}
	m.insHookDATAW[addr] = append(m.insHookDATAW[addr], fn)
}

// GetBankSelect get Register bank select
func (m *Machine) GetBankSelect() int {
	/*
		Register bank select:
		RS1 RS0 Working Register Bank and Address
		0 0 Bank0 (D:0x00 - D:0x07)
		0 1 Bank1 (D:0x08 - D:0x0F)
		1 0 Bank2 (D:0x10 - D:0x17)
		1 1 Bank3 (D:0x18H - D:0x1F)
	*/
	var (
		banksel = 0
	)

	if (m.DATA[PSW] & uint8(1<<3)) != 0 {
		banksel |= (1 << 0)
	}
	if (m.DATA[PSW] & uint8(1<<4)) != 0 {
		banksel |= (1 << 1)
	}
	return banksel
}

// GetBankRx get Register bank select R0~R7
func (m *Machine) GetBankRx() []byte {
	addr := m.GetBankSelect() * 8
	return m.DATA[addr : addr+0x08]
}

// ReadRx Read R0~R7
func (m *Machine) ReadRx(x uint8) uint8 {
	return m.ReadDATA(uint8(m.GetBankSelect()*8) + x)
}

// WriteRx Write R0~R7
func (m *Machine) WriteRx(x uint8, val uint8) {
	m.WriteDATA(uint8(m.GetBankSelect()*8)+x, val)
}
