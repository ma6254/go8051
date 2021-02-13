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

	DATA        [0xFF]byte // RAM: DATA Range
	ROM         []byte     // ROM: CODE Range
	PC          uint       // PC: program counter
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
	defer m.mainTick.Stop()
	defer close(m.exitCh)
	for {
		select {
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
