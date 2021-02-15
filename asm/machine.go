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

type BreakePoint struct {
	Addr uint
	Func func(m *Machine)
}

// Machine 8051 microchip
type Machine struct {
	mainTick *time.Ticker
	exitCh   chan int
	isrCh    chan int

	DATA        [0x100]byte   // RAM: DATA Range
	XDATA       [0x10000]byte // RAM: XDATA Range
	ROM         []byte        // ROM: CODE Range
	PC          uint          // PC: program counter
	brakepoints map[uint][]func(m *Machine)
	Frequency   time.Duration
}

// NewMachine Create 8051 machine
func NewMachine(f time.Duration) *Machine {
	m := &Machine{}
	m.brakepoints = make(map[uint][]func(m *Machine))
	m.Frequency = f
	return m
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

// Trace breakepoint
func (m *Machine) Trace(addr uint, fn func(m *Machine)) {
	if m.brakepoints[addr] == nil {
		m.brakepoints[addr] = make([]func(m *Machine), 0)
	}
	m.brakepoints[addr] = append(m.brakepoints[addr], fn)
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
	addr := m.GetBankSelect() * 3
	return m.DATA[addr : addr+0x08]
}
