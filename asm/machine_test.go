package asm_test

import (
	"strings"
	"testing"

	"github.com/ma6254/go8051/asm"

	"github.com/marcinbor85/gohex"
)

func Test_Base(t *testing.T) {
	var (
		OK_0_cnt uint8
		OK_0     bool
		OK_1     bool
		OK_2     bool
		OK_3     bool
	)

	// Flip IO
	/*
		#include "REG51.h"
		void main(void){
			while(1)
			{
				P0 = 0x55;
				P0 = 0xAA;
			}
		}
	*/
	ROM := `:08000F007580557580AA80F888
:03000000020003F8
:0C000300787FE4F6D8FD75810702000F3D
:00000001FF`
	mem := gohex.NewMemory()
	err := mem.ParseIntelHex(strings.NewReader(ROM))
	if err != nil {
		t.Errorf("ParseIntelHex %s", err)
	}
	m := asm.NewMachine(asm.Frequency1MHz)
	m.ROM = mem.ToBinary(0x00, 0xFFFF, 0x00)

	m.Trace(0x06, func(m *asm.Machine) {
		OK_0_cnt++
	})

	m.Trace(0x0F, func(m *asm.Machine) {
		t.Logf("%04X P0: %02X\n", m.PC, m.DATA[asm.P0])
		if m.DATA[asm.P0] == 0xAA {
			OK_1 = true
		}
	})
	m.Trace(0x12, func(m *asm.Machine) {
		t.Logf("%04X P0: %02X\n", m.PC, m.DATA[asm.P0])
		if m.DATA[asm.P0] == 0x55 {
			OK_2 = true
		}
	})
	m.Trace(0x15, func(m *asm.Machine) {
		t.Logf("%04X LoopJump\n", m.PC)
		if OK_1 && OK_2 {
			OK_3 = true
			m.Stop()
		}
	})

	t.Logf("8051 Machine Running")
	m.Start()

	if m.DATA[asm.R0] == 0 {
		if OK_0_cnt == (0x7F - 1) {
			OK_0 = true
		} else {
			OK_0 = false
		}
	}

	if OK_0 && OK_1 && OK_2 && OK_3 {
		return
	} else {
		if !OK_0 {
			t.Logf("OK_0_cnt %02X", OK_0_cnt)
		}
		t.Errorf("Fail 0:%t 1:%t 2:%t 3:%t", OK_0, OK_1, OK_2, OK_3)
	}
}

func Test_Call_Func(t *testing.T) {
	/*
		#include "REG51.h"
		void a(unsigned char x);
		void main(void)
		{
			a(0x88); // CODE 0x03
			while (1)
			{
				P0 = 0x55; // CODE 0x08
				P0 = 0xAA;
			}
		}
		void a(unsigned char x)
		{
			P1 = x; // CODE 0x1C
		}
	*/
	// 	ROM := `:0D0003007F8812001C7580557580AA80F85A
	// :03001C008F9022A0
	// :03000000020010EB
	// :0C001000787FE4F6D8FD7581070200033C
	// :00000001FF`
	ROM := `:10000300C0D075D008900000EFF0900000E0FF8FA3
:0400130090D0D02297
:0E0017007F881200037580557580AA80F8223C
:03000000020025D6
:0C002500787FE4F6D8FD75810F0200170B
:00000001FF`

	mem := gohex.NewMemory()
	err := mem.ParseIntelHex(strings.NewReader(ROM))
	if err != nil {
		t.Errorf("ParseIntelHex %s", err)
	}
	m := asm.NewMachine(asm.Frequency1MHz)
	m.ROM = []byte{
		0x2, 0x0, 0x25, // C:0x0000: LJMP C:0025
		0xc0, 0xd0, //  C:0x0003 PUSH PSW(0xD0) function aaa begin
		0x75, 0xd0, 0x8, // C:0x0005 MOV PSW(0xD0), 0x08
		0x90, 0x0, 0x0, // C:0x0008 MOV DPTR, #0x0000
		0xef,           // C:0x000B MOV A, R7
		0xf0,           // C:0x000C MOVX DPTR, A
		0x90, 0x0, 0x0, // C:0x000D MOV DPTR, #0x0000
		0xe0,       // C:0x0010 MOVX A, @DPTR
		0xff,       // C:0x0011
		0x8f, 0x90, // C:0x0012
		0xd0, 0xd0, // C:0x0014
		0x22,       // C:0x0016 RET function aaa end
		0x7f, 0x88, // C:0x0017 MOV R7, #0x88
		0x12, 0x0, 0x3, // C:0x0019
		0x75, 0x80, 0x55, // C:0x001C
		0x75, 0x80, 0xaa, // C:0x001F
		0x80, 0xf8, // C:0x0022
		0x22,       // C:0x0024
		0x78, 0x7f, // C:0x0025 MOV R0, #0x7F
		0xe4,       // C:0x0027 CLR A
		0xf6,       // C:0x0028 MOV @R0, A
		0xd8, 0xfd, //  C:0x0029 DJNZ R0, C:0028
		0x75, 0x81, 0xf, //  C:0x002B MOV SP(0x81), #0x0F
		0x2, 0x0, 0x17, // C:0x002E LJMP main(C:0017)
	}

	var (
		sp0 uint8
		sp1 uint8
		sp2 uint8
	)

	m.Trace(0x17, func(m *asm.Machine) {
		sp0 = m.DATA[asm.SP]
		t.Logf("%04X SP: %02X P1: %02X\n", m.PC, m.DATA[asm.SP], m.DATA[asm.P1])
	})

	m.Trace(0x1C, func(m *asm.Machine) {
		sp2 = m.DATA[asm.SP]
		t.Logf("%04X SP: %02X P1: %02X\n", m.PC, m.DATA[asm.SP], m.DATA[asm.P1])
		m.Stop()
	})

	m.Trace(0x0D, func(m *asm.Machine) {
		sp1 = m.DATA[asm.SP]
		t.Logf("%04X SP: %02X P1: %02X\n", m.PC, m.DATA[asm.SP], m.DATA[asm.P1])
	})

	t.Logf("8051 Machine Running")
	m.Start()

	if sp0 != sp2 {
		t.Logf("sp restore fail %02X %02X", sp0, sp2)
	}
	if (sp0 + 2) == sp1 {
		t.Logf("sp inc fail")
	}

}
