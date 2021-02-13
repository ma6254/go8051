package main

import (
	"log"
	"os"

	"github.com/marcinbor85/gohex"

	"github.com/ma6254/go8051/asm"
)

var b = []byte{
	0x75, 0x80, 0xAA, // MOV	P0,	#0AAH
	0x00,             // NOP
	0x75, 0x80, 0x55, // MOV	P0,	#055H
	0x80, 0xF7, 0x00, // JMP #F7H
}

func main() {

	file, err := os.Open("release/MDK51/Objects/a.hex")
	if err != nil {
		panic(err)
	}
	defer file.Close()
	mem := gohex.NewMemory()
	err = mem.ParseIntelHex(file)
	if err != nil {
		panic(err)
	}

	m := asm.NewMachine(asm.Frequency10Hz)
	m.ROM = mem.ToBinary(0x00, 0xFFFF, 0x00)
	m.Trace(0x0F, func(m *asm.Machine) {
		log.Printf("%04X P0: %02X\n", m.PC, m.DATA[asm.P0])
	})
	m.Trace(0x12, func(m *asm.Machine) {
		log.Printf("%04X P0: %02X\n", m.PC, m.DATA[asm.P0])
	})
	m.Trace(0x06, func(m *asm.Machine) {
		log.Printf("%04X R0: %02X\n", m.PC, m.DATA[asm.R0])
	})

	log.Printf("8051 Machine Running")
	m.Start()
}
