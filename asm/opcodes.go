package asm

import "fmt"

// INS : 8051 Instructions Code
type INS struct {
	Code     byte
	Bytes    byte
	Mnemonic string
	Func     func(*Machine)
	FakeCode func(Machine, uint) string
}

func genNOPn(x int) func(m *Machine) {
	return func(m *Machine) {
		m.PC += uint(x)
	}
}

// genMOVRxImmed, "MOV Rx #immed" ,ins code 78~7F
func genMOVRxImmed(x uint8) func(m *Machine) {
	return func(m *Machine) {
		m.WriteRx(x, m.ROM[m.PC+1])
		m.PC += 2
	}
}

func genMOVRxImmedFakeCode(x uint8) func(m Machine, pc uint) string {
	return func(m Machine, pc uint) string {
		return fmt.Sprintf("R%d #0x%02X", x, m.ROM[pc+1])
	}
}

// genMOVdirectRx, "MOV direct Rx" ,ins code 88~8F
func genMOVdirectRx(x uint8) func(m *Machine) {
	return func(m *Machine) {
		m.WriteDATA(m.ROM[m.PC+1], m.ReadRx(x))
		m.PC += 2
	}
}

// genMOVARx, "MOV A, Rx", ins code E8~EF
func genMOVARx(x uint8) func(m *Machine) {
	return func(m *Machine) {
		m.WriteDATA(ACC, m.ReadRx(x))
		m.PC++
	}
}

// genMOVRxA, "MOV Rx, A", ins code F8~FF
func genMOVRxA(x uint8) func(m *Machine) {
	return func(m *Machine) {
		m.WriteRx(x, m.ReadDATA(ACC))
		m.PC++
	}
}

// genMOVRxAFakeCode, "MOV Rx, A", ins code F8~FF
func genMOVRxAFakeCode(x uint8) func(m Machine, pc uint) string {
	return func(m Machine, pc uint) string {
		return fmt.Sprintf("R%d A", x)
	}
}

// genDJNZRxOffset, "DJNZ Rx, offset", ins code D8~DF
func genDJNZRxOffset(x uint8) func(m *Machine) {
	return func(m *Machine) {
		offset := int8(m.ROM[m.PC+1])
		m.WriteRx(x, m.ReadRx(x)-1)
		if (m.ReadRx(x)) != 0 {
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

func genDJNZRxOffsetFakeCode(x uint8) func(m Machine, pc uint) string {
	return func(m Machine, pc uint) string {
		offset := int8(m.ROM[pc+1])
		pc += 2
		if offset > 0 {
			pc += uint(offset)
		} else {
			pc -= uint(-offset)

		}
		return fmt.Sprintf("C:%d(%04X)", offset, pc)
	}
}

// Instructions : The following table lists the 8051 instructions by HEX code.
var Instructions = [0xFF]INS{
	{Code: 0x00, Bytes: 1, Mnemonic: "NOP", Func: genNOPn(1)},
	{Code: 0x02, Bytes: 3, Mnemonic: "LJMP", Func: func(m *Machine) {
		addrH := uint(m.ROM[m.PC+1])
		addrL := uint(m.ROM[m.PC+2])
		m.PC = (addrH << 8) | addrL
	}, FakeCode: func(m Machine, pc uint) string {
		return fmt.Sprintf("C:%04X", (uint(m.ROM[pc+1])<<8)|uint(m.ROM[pc+2]))
	}},
	{Code: 0x12, Bytes: 3, Mnemonic: "LCALL", Func: func(m *Machine) {
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

		m.WriteDATA(SP, m.ReadDATA(SP)+1)
		m.WriteDATA(m.ReadDATA(SP), uint8(m.PC))
		m.WriteDATA(SP, m.ReadDATA(SP)+1)
		m.WriteDATA(m.ReadDATA(SP), uint8(m.PC>>8))
		m.PC = (addrH << 8) | addrL
	}, FakeCode: func(m Machine, pc uint) string {
		return fmt.Sprintf("C:0x%02X%02X", m.ROM[pc+1], m.ROM[pc+2])
	}},
	{Code: 0x22, Bytes: 1, Mnemonic: "RET", Func: func(m *Machine) {
		/*
			PC15-8 = (SP)
			SP = SP - 1
			PC7-0 = (SP)
			SP = SP - 1
		*/
		addrH := m.ReadDATA(m.ReadDATA(SP))
		m.WriteDATA(SP, m.ReadDATA(SP)-1)
		addrL := m.ReadDATA(m.ReadDATA(SP))
		m.WriteDATA(SP, m.ReadDATA(SP)-1)
		m.PC = (uint(addrH) << 8) | uint(addrL)
	}},
	{Code: 0x75, Bytes: 0x03, Mnemonic: "MOV", Func: func(m *Machine) {
		m.WriteDATA(m.ROM[m.PC+1], m.ROM[m.PC+2])
		m.PC += 3
	}, FakeCode: func(m Machine, pc uint) string {
		s := ""
		if r := FindRegByAddr(m.ROM[pc+1], m.regDefines); r != nil {
			s += fmt.Sprintf("%s(0x%02X)", r.Name, r.Addr)
		} else {
			s += fmt.Sprintf("0x%02X", m.ROM[pc+1])
		}
		s += fmt.Sprintf(" #0x%02X", m.ROM[pc+2])
		return s
	}},
	{Code: 0x78, Bytes: 2, Mnemonic: "MOV", Func: genMOVRxImmed(0), FakeCode: genMOVRxImmedFakeCode(0)},
	{Code: 0x79, Bytes: 2, Mnemonic: "MOV", Func: genMOVRxImmed(1), FakeCode: genMOVRxImmedFakeCode(1)},
	{Code: 0x7A, Bytes: 2, Mnemonic: "MOV", Func: genMOVRxImmed(2), FakeCode: genMOVRxImmedFakeCode(2)},
	{Code: 0x7B, Bytes: 2, Mnemonic: "MOV", Func: genMOVRxImmed(3), FakeCode: genMOVRxImmedFakeCode(3)},
	{Code: 0x7C, Bytes: 2, Mnemonic: "MOV", Func: genMOVRxImmed(4), FakeCode: genMOVRxImmedFakeCode(4)},
	{Code: 0x7D, Bytes: 2, Mnemonic: "MOV", Func: genMOVRxImmed(5), FakeCode: genMOVRxImmedFakeCode(5)},
	{Code: 0x7E, Bytes: 2, Mnemonic: "MOV", Func: genMOVRxImmed(6), FakeCode: genMOVRxImmedFakeCode(6)},
	{Code: 0x7F, Bytes: 2, Mnemonic: "MOV", Func: genMOVRxImmed(7), FakeCode: genMOVRxImmedFakeCode(7)},
	{Code: 0x80, Bytes: 2, Mnemonic: "SJMP", Func: func(m *Machine) {
		offset := int8(m.ROM[m.PC+1])
		m.PC += 2
		if offset > 0 {
			m.PC += uint(offset)
		} else {
			m.PC -= uint(-offset)
		}
	}, FakeCode: func(m Machine, pc uint) string {
		offset := int8(m.ROM[pc+1])
		pc += 2
		if offset > 0 {
			pc += uint(offset)
		} else {
			pc -= uint(-offset)
		}
		return fmt.Sprintf("C:%d(%04X)", offset, pc)
	}},

	{Code: 0x88, Bytes: 2, Mnemonic: "MOV", Func: genMOVdirectRx(0)},
	{Code: 0x89, Bytes: 2, Mnemonic: "MOV", Func: genMOVdirectRx(1)},
	{Code: 0x8A, Bytes: 2, Mnemonic: "MOV", Func: genMOVdirectRx(2)},
	{Code: 0x8B, Bytes: 2, Mnemonic: "MOV", Func: genMOVdirectRx(3)},
	{Code: 0x8C, Bytes: 2, Mnemonic: "MOV", Func: genMOVdirectRx(4)},
	{Code: 0x8D, Bytes: 2, Mnemonic: "MOV", Func: genMOVdirectRx(5)},
	{Code: 0x8E, Bytes: 2, Mnemonic: "MOV", Func: genMOVdirectRx(6)},
	{Code: 0x8F, Bytes: 2, Mnemonic: "MOV", Func: genMOVdirectRx(7)},
	{Code: 0x90, Bytes: 3, Mnemonic: "MOV", Func: func(m *Machine) {
		// MOV	DPTR, #immed
		m.WriteDATA(DPH, m.ROM[m.PC+1])
		m.WriteDATA(DPL, m.ROM[m.PC+2])
		m.PC += 3
	}, FakeCode: func(m Machine, pc uint) string {
		return fmt.Sprintf("DPTR #0x%02X%02X", m.ROM[pc+1], m.ROM[pc+2])
	}},

	{Code: 0xC0, Bytes: 2, Mnemonic: "PUSH", Func: func(m *Machine) {
		// SP = SP + 1
		// (SP) = (direct)
		m.WriteDATA(SP, m.ReadDATA(SP)+1)
		m.WriteDATA(m.ReadDATA(SP), m.ROM[m.PC+1])
		m.PC += 2
	}, FakeCode: func(m Machine, pc uint) string {
		if r := FindRegByAddr(m.ROM[pc+1], m.regDefines); r != nil {
			return fmt.Sprintf("%s(0x%02X)", r.Name, r.Addr)
		}
		return fmt.Sprintf("0x%02X", m.ROM[pc+1])
	}},

	{Code: 0xD0, Bytes: 2, Mnemonic: "POP", Func: func(m *Machine) {
		// (direct) = (SP)
		// SP = SP - 1
		m.WriteDATA(m.ROM[m.PC+1], m.ReadDATA(m.ReadDATA(SP)))
		m.WriteDATA(SP, m.ReadDATA(SP)-1)
		m.PC += 2
	}, FakeCode: func(m Machine, pc uint) string {
		if r := FindRegByAddr(m.ROM[pc+1], m.regDefines); r != nil {
			return fmt.Sprintf("%s(0x%02X)", r.Name, r.Addr)
		}
		return fmt.Sprintf("0x%02X", m.ROM[pc+1])
	}},

	{Code: 0xD8, Bytes: 2, Mnemonic: "DJNZ", Func: genDJNZRxOffset(0), FakeCode: genDJNZRxOffsetFakeCode(0)},
	{Code: 0xD9, Bytes: 2, Mnemonic: "DJNZ", Func: genDJNZRxOffset(1), FakeCode: genDJNZRxOffsetFakeCode(1)},
	{Code: 0xDA, Bytes: 2, Mnemonic: "DJNZ", Func: genDJNZRxOffset(2), FakeCode: genDJNZRxOffsetFakeCode(2)},
	{Code: 0xDB, Bytes: 2, Mnemonic: "DJNZ", Func: genDJNZRxOffset(3), FakeCode: genDJNZRxOffsetFakeCode(3)},
	{Code: 0xDC, Bytes: 2, Mnemonic: "DJNZ", Func: genDJNZRxOffset(4), FakeCode: genDJNZRxOffsetFakeCode(4)},
	{Code: 0xDD, Bytes: 2, Mnemonic: "DJNZ", Func: genDJNZRxOffset(5), FakeCode: genDJNZRxOffsetFakeCode(5)},
	{Code: 0xDE, Bytes: 2, Mnemonic: "DJNZ", Func: genDJNZRxOffset(6), FakeCode: genDJNZRxOffsetFakeCode(6)},
	{Code: 0xDF, Bytes: 2, Mnemonic: "DJNZ", Func: genDJNZRxOffset(7), FakeCode: genDJNZRxOffsetFakeCode(7)},
	{Code: 0xE0, Bytes: 1, Mnemonic: "MOVX", Func: func(m *Machine) {
		// MOVX	A, @DPTR
		dptrAddr := uint16(m.ReadDATA(DPH))>>8 | uint16(m.ReadDATA(DPL))
		m.WriteDATA(ACC, m.XDATA[dptrAddr])
		m.PC++
	}, FakeCode: func(m Machine, pc uint) string { return fmt.Sprintf("A @DPTR") }},
	{0xE4, 1, "CLR", func(m *Machine) {
		m.WriteDATA(ACC, 0)
		m.PC += 2
	}, func(m Machine, pc uint) string { return "A" }},

	{Code: 0xE8, Bytes: 1, Mnemonic: "MOV", Func: genMOVARx(0)},
	{Code: 0xE9, Bytes: 1, Mnemonic: "MOV", Func: genMOVARx(1)},
	{Code: 0xEA, Bytes: 1, Mnemonic: "MOV", Func: genMOVARx(2)},
	{Code: 0xEB, Bytes: 1, Mnemonic: "MOV", Func: genMOVARx(3)},
	{Code: 0xEC, Bytes: 1, Mnemonic: "MOV", Func: genMOVARx(4)},
	{Code: 0xED, Bytes: 1, Mnemonic: "MOV", Func: genMOVARx(5)},
	{Code: 0xEE, Bytes: 1, Mnemonic: "MOV", Func: genMOVARx(6)},
	{Code: 0xEF, Bytes: 1, Mnemonic: "MOV", Func: genMOVARx(7)},
	{Code: 0xF0, Bytes: 1, Mnemonic: "MOV", Func: func(m *Machine) {
		// MOV @DPTR, A
		addr := uint16(m.DATA[DPH])<<8 | uint16(m.DATA[DPL])
		m.WriteDATA(uint8(addr), ACC)
		m.PC++
	}, FakeCode: func(m Machine, pc uint) string { return fmt.Sprintf("@DPTR A") }},
	{Code: 0xF6, Bytes: 1, Mnemonic: "MOV", Func: func(m *Machine) {
		// 	MOV	@R0, A
		m.WriteDATA(m.ReadRx(0), ACC)
		m.PC++
	}, FakeCode: func(m Machine, pc uint) string { return fmt.Sprintf("@R0, A") }},
	{Code: 0xF7, Bytes: 1, Mnemonic: "MOV", Func: func(m *Machine) {
		// MOV	@R1, A
		m.WriteDATA(m.ReadRx(1), ACC)
		m.PC++
	}, FakeCode: func(m Machine, pc uint) string { return fmt.Sprintf("@R1, A") }},
	{Code: 0xF8, Bytes: 1, Mnemonic: "MOV", Func: genMOVRxA(0), FakeCode: genMOVRxAFakeCode(0)},
	{Code: 0xF9, Bytes: 1, Mnemonic: "MOV", Func: genMOVRxA(1), FakeCode: genMOVRxAFakeCode(1)},
	{Code: 0xFA, Bytes: 1, Mnemonic: "MOV", Func: genMOVRxA(2), FakeCode: genMOVRxAFakeCode(2)},
	{Code: 0xFB, Bytes: 1, Mnemonic: "MOV", Func: genMOVRxA(3), FakeCode: genMOVRxAFakeCode(3)},
	{Code: 0xFC, Bytes: 1, Mnemonic: "MOV", Func: genMOVRxA(4), FakeCode: genMOVRxAFakeCode(4)},
	{Code: 0xFD, Bytes: 1, Mnemonic: "MOV", Func: genMOVRxA(5), FakeCode: genMOVRxAFakeCode(5)},
	{Code: 0xFE, Bytes: 1, Mnemonic: "MOV", Func: genMOVRxA(6), FakeCode: genMOVRxAFakeCode(6)},
	{Code: 0xFF, Bytes: 1, Mnemonic: "MOV", Func: genMOVRxA(7), FakeCode: genMOVRxAFakeCode(7)},
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
