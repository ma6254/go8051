package asm

import "fmt"

// INS : 8051 Instructions Code
type INS struct {
	Code     byte
	Bytes    byte
	Mnemonic string
	Func     func(*Machine)
}

func genNOPn(x int) func(m *Machine) {
	return func(m *Machine) {
		m.PC += uint(x)
	}
}

// genMOVRxImmed, "MOV Rx #immed" ,ins code 78~7F
func genMOVRxImmed(x uint8) func(m *Machine) {
	return func(m *Machine) {
		val := m.ROM[m.PC+1]
		m.DATA[R0+x] = val
		m.PC += 2
	}
}

// genDJNZRxOffset, "DJNZ Rx, offset", ins code D8~DF
func genDJNZRxOffset(x int) func(m *Machine) {
	return func(m *Machine) {
		offset := int8(m.ROM[m.PC+1])
		m.DATA[R0+x]--
		if m.DATA[R0+x] != 0 {
			if offset > 0 {
				m.PC += uint(offset)
			} else {
				m.PC -= uint(-offset)
			}
		} else {
			m.PC += 2
		}
	}
}

// Instructions : The following table lists the 8051 instructions by HEX code.
var Instructions = [0xFF]INS{
	{0x00, 0x01, "NOP", genNOPn(1)},
	{0x02, 0x03, "LJMP", func(m *Machine) {
		addrH := uint(m.ROM[m.PC+1])
		addrL := uint(m.ROM[m.PC+2])
		m.PC = (addrH << 8) | addrL
	}},
	{0x75, 0x03, "MOV", func(m *Machine) {
		addr := m.ROM[m.PC+1]
		val := m.ROM[m.PC+2]
		m.DATA[addr] = val
		m.PC += 3
	}},
	{0x78, 0x02, "MOV", genMOVRxImmed(0)},
	{0x79, 0x02, "MOV", genMOVRxImmed(1)},
	{0x7A, 0x02, "MOV", genMOVRxImmed(2)},
	{0x7B, 0x02, "MOV", genMOVRxImmed(3)},
	{0x7C, 0x02, "MOV", genMOVRxImmed(4)},
	{0x7D, 0x02, "MOV", genMOVRxImmed(5)},
	{0x7E, 0x02, "MOV", genMOVRxImmed(6)},
	{0x7F, 0x02, "MOV", genMOVRxImmed(7)},
	{0x80, 0x02, "SJMP", func(m *Machine) {
		offset := int8(m.ROM[m.PC+1])
		m.PC += 2
		if offset > 0 {
			m.PC += uint(offset)
		} else {
			m.PC -= uint(-offset)
		}
	}},

	{0xD8, 0x02, "DJNZ", genDJNZRxOffset(0)},
	{0xD9, 0x02, "DJNZ", genDJNZRxOffset(1)},
	{0xDA, 0x02, "DJNZ", genDJNZRxOffset(2)},
	{0xDB, 0x02, "DJNZ", genDJNZRxOffset(3)},
	{0xDC, 0x02, "DJNZ", genDJNZRxOffset(4)},
	{0xDD, 0x02, "DJNZ", genDJNZRxOffset(5)},
	{0xDE, 0x02, "DJNZ", genDJNZRxOffset(6)},
	{0xDF, 0x02, "DJNZ", genDJNZRxOffset(7)},
	{0xE4, 0x02, "CLR", func(m *Machine) {
		m.DATA[ACC] = 0
		m.PC += 2
	}},
	{0xF6, 0x01, "MOV", func(m *Machine) {
		addr := m.DATA[R0]
		val := m.ROM[m.PC+2]
		m.DATA[addr] = val
		m.PC++
	}},
	{0xF7, 0x01, "MOV", func(m *Machine) {
		addr := m.DATA[R1]
		val := m.ROM[m.PC+2]
		m.DATA[addr] = val
		m.PC++
	}},
}

// FindINS find Instructions
func FindINS(a byte) (i *INS, err error) {
	for i := 0; i < 0xFF; i++ {
		if Instructions[i].Code == a {
			return &Instructions[i], nil
		}
	}
	return nil, fmt.Errorf("Unsupport instructions %02X", a)
}
