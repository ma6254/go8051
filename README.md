# go8051

doc: https://pkg.go.dev/github.com/ma6254/go8051

[![made-with-Go](https://img.shields.io/badge/Made%20with-Go-1f425f.svg)](http://golang.org)
[![godoc](https://img.shields.io/badge/godoc-reference-blue.svg)](https://pkg.go.dev/github.com/ma6254/go8051/)
[![last-commit](https://img.shields.io/github/last-commit/ma6254/go8051.svg)](https://github.com/ma6254/go8051/commits)
[![Go](https://github.com/ma6254/go8051/workflows/Go/badge.svg)](https://github.com/ma6254/go8051/actions/)
[![GoReportCard](https://goreportcard.com/badge/github.com/ma6254/go8051)](https://goreportcard.com/report/github.com/ma6254/go8051)

8051 asm virtual machine by Golang

Just to learn the hardware

## Example

```go
package main

import (
	"log"

	"github.com/ma6254/go8051/asm"
)

var b = []byte{
	0x75, 0x80, 0xAA, // MOV	P0,	#0AAH
	0x00,             // NOP
	0x75, 0x80, 0x55, // MOV	P0,	#055H
	0x80, 0xF7, 0x00, // JMP #F7H
}

func main() {


	m := asm.NewMachine(asm.Frequency10Hz)
	m.ROM = b
	m.Trace(0x00, func(m *asm.Machine) {
		log.Printf("%04X P0: %02X\n", m.PC, m.DATA[asm.P0])
	})
	m.Trace(0x04, func(m *asm.Machine) {
		log.Printf("%04X P0: %02X\n", m.PC, m.DATA[asm.P0])
	})
	m.Trace(0x07, func(m *asm.Machine) {
		log.Printf("%04X R0: %02X\n", m.PC, m.DATA[asm.R0])
	})

	log.Printf("8051 Machine Running")
	m.Start()
}
```
