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
		endCh    chan int = make(chan int)
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
			endCh <- 1
		}
	})

	t.Logf("8051 Machine Running")
	go m.Start()

	<-endCh

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
