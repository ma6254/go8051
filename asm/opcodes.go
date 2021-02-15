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
		m.GetBankRx()[x] = val
		m.PC += 2
	}
}

// genMOVdirectRx, "MOV direct Rx" ,ins code 88~8F
func genMOVdirectRx(x uint8) func(m *Machine) {
	return func(m *Machine) {
		addr := m.ROM[m.PC+1]
		m.DATA[addr] = m.GetBankRx()[x]
		m.PC += 2
	}
}

// genMOVARx, "MOV A, Rx", ins code E8~EF
func genMOVARx(x int) func(m *Machine) {
	return func(m *Machine) {
		m.DATA[ACC] = m.GetBankRx()[x]
		m.PC++
	}
}

// genMOVRxA, "MOV Rx, A", ins code F8~FF
func genMOVRxA(x int) func(m *Machine) {
	return func(m *Machine) {
		m.GetBankRx()[x] = m.DATA[ACC]
		m.PC++
	}
}

// genDJNZRxOffset, "DJNZ Rx, offset", ins code D8~DF
func genDJNZRxOffset(x int) func(m *Machine) {
	return func(m *Machine) {
		offset := int8(m.ROM[m.PC+1])
		m.GetBankRx()[x]--
		if m.GetBankRx()[x] != 0 {
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
	{0x00, 1, "NOP", genNOPn(1)},
	{0x02, 3, "LJMP", func(m *Machine) {
		addrH := uint(m.ROM[m.PC+1])
		addrL := uint(m.ROM[m.PC+2])
		m.PC = (addrH << 8) | addrL
	}},
	{0x12, 3, "LCALL", func(m *Machine) {
		/*
			PC = PC + 3
			SP = SP + 1
			(SP) = PC[7-0]
			SP = SP + 1
			(SP) = PC[15-8]
			PC = addr16
		*/
		addrH := uint(m.ROM[m.PC+1])
		addrL := uint(m.ROM[m.PC+2])
		m.PC += 3

		m.DATA[SP]++
		m.DATA[m.DATA[SP]] = uint8(m.PC)
		m.DATA[SP]++
		m.DATA[m.DATA[SP]] = uint8(m.PC >> 8)
		m.PC = (addrH << 8) | addrL
	}},
	{0x22, 1, "RET", func(m *Machine) {
		/*
			PC15-8 = (SP)
			SP = SP - 1
			PC7-0 = (SP)
			SP = SP - 1
		*/
		addrH := m.DATA[m.DATA[SP]]
		m.DATA[SP]--
		addrL := m.DATA[m.DATA[SP]]
		m.DATA[SP]--
		m.PC = (uint(addrH) << 8) | uint(addrL)
	}},
	{0x75, 0x03, "MOV", func(m *Machine) {
		addr := m.ROM[m.PC+1]
		val := m.ROM[m.PC+2]
		m.DATA[addr] = val
		m.PC += 3
	}},
	{0x78, 2, "MOV", genMOVRxImmed(0)},
	{0x79, 2, "MOV", genMOVRxImmed(1)},
	{0x7A, 2, "MOV", genMOVRxImmed(2)},
	{0x7B, 2, "MOV", genMOVRxImmed(3)},
	{0x7C, 2, "MOV", genMOVRxImmed(4)},
	{0x7D, 2, "MOV", genMOVRxImmed(5)},
	{0x7E, 2, "MOV", genMOVRxImmed(6)},
	{0x7F, 2, "MOV", genMOVRxImmed(7)},
	{0x80, 2, "SJMP", func(m *Machine) {
		offset := int8(m.ROM[m.PC+1])
		m.PC += 2
		if offset > 0 {
			m.PC += uint(offset)
		} else {
			m.PC -= uint(-offset)
		}
	}},

	{0x88, 2, "MOV", genMOVdirectRx(0)},
	{0x89, 2, "MOV", genMOVdirectRx(1)},
	{0x8A, 2, "MOV", genMOVdirectRx(2)},
	{0x8B, 2, "MOV", genMOVdirectRx(3)},
	{0x8C, 2, "MOV", genMOVdirectRx(4)},
	{0x8D, 2, "MOV", genMOVdirectRx(5)},
	{0x8E, 2, "MOV", genMOVdirectRx(6)},
	{0x8F, 2, "MOV", genMOVdirectRx(7)},
	{0x90, 3, "MOV", func(m *Machine) {
		// MOV	DPTR, #immed
		m.DATA[DPH] = m.ROM[m.PC+1]
		m.DATA[DPL] = m.ROM[m.PC+2]
		m.PC += 3
	}},

	{0xC0, 2, "PUSH", func(m *Machine) {
		// SP = SP + 1
		// (SP) = (direct)
		val := m.DATA[m.PC+1]
		m.DATA[SP]++
		m.DATA[m.DATA[SP]] = val
		m.PC += 2
	}},

	{0xD0, 2, "POP", func(m *Machine) {
		// (direct) = (SP)
		// SP = SP - 1
		addr := m.DATA[m.PC+1]
		m.DATA[addr] = m.DATA[m.DATA[SP]]
		m.DATA[SP]--
		m.PC += 2
	}},

	{0xD8, 2, "DJNZ", genDJNZRxOffset(0)},
	{0xD9, 2, "DJNZ", genDJNZRxOffset(1)},
	{0xDA, 2, "DJNZ", genDJNZRxOffset(2)},
	{0xDB, 2, "DJNZ", genDJNZRxOffset(3)},
	{0xDC, 2, "DJNZ", genDJNZRxOffset(4)},
	{0xDD, 2, "DJNZ", genDJNZRxOffset(5)},
	{0xDE, 2, "DJNZ", genDJNZRxOffset(6)},
	{0xDF, 2, "DJNZ", genDJNZRxOffset(7)},
	{0xE0, 1, "MOVX", func(m *Machine) {
		// MOVX	A, @DPTR
		dptrAddr := uint(m.DATA[DPH])>>8 | uint(m.DATA[DPL])
		m.DATA[ACC] = m.XDATA[dptrAddr]
		m.PC++
	}},
	{0xE4, 2, "CLR", func(m *Machine) {
		m.DATA[ACC] = 0
		m.PC += 2
	}},

	{0xE8, 1, "DJNZ", genMOVARx(0)},
	{0xE9, 1, "DJNZ", genMOVARx(1)},
	{0xEA, 1, "DJNZ", genMOVARx(2)},
	{0xEB, 1, "DJNZ", genMOVARx(3)},
	{0xEC, 1, "DJNZ", genMOVARx(4)},
	{0xED, 1, "DJNZ", genMOVARx(5)},
	{0xEE, 1, "DJNZ", genMOVARx(6)},
	{0xEF, 1, "DJNZ", genMOVARx(7)},
	{0xF0, 1, "MOV", func(m *Machine) {
		// MOV @DPTR, A
		addr := uint16(m.DATA[DPH])<<8 | uint16(m.DATA[DPL])
		m.DATA[addr] = m.DATA[ACC]
		m.PC++
	}},
	{0xF6, 1, "MOV", func(m *Machine) {
		// 	MOV	@R0, A
		addr := m.GetBankRx()[0]
		val := m.ROM[m.PC+2]
		m.DATA[addr] = val
		m.PC++
	}},
	{0xF7, 1, "MOV", func(m *Machine) {
		// MOV	@R1, A
		addr := m.GetBankRx()[1]
		val := m.ROM[m.PC+2]
		m.DATA[addr] = val
		m.PC++
	}},
	{0xF8, 1, "MOV", genMOVRxA(0)},
	{0xF9, 1, "MOV", genMOVRxA(1)},
	{0xFA, 1, "MOV", genMOVRxA(2)},
	{0xFB, 1, "MOV", genMOVRxA(3)},
	{0xFC, 1, "MOV", genMOVRxA(4)},
	{0xFD, 1, "MOV", genMOVRxA(5)},
	{0xFE, 1, "MOV", genMOVRxA(6)},
	{0xFF, 1, "MOV", genMOVRxA(7)},
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
