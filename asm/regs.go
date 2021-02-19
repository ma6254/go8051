package asm

const (
	R0 = 0x00
	R1 = 0x01
	R2 = 0x02
	R3 = 0x03
	R4 = 0x04
	R5 = 0x05
	R6 = 0x06
	R7 = 0x07

	SP  = 0x81
	DPL = 0x82
	DPH = 0x83

	// P0 : IO port 0
	P0 = 0x80
	// P1 : IO port 1
	P1 = 0x90
	// P2 : IO port 2
	P2 = 0xA0
	// P3 : IO port 3
	P3  = 0xB0
	PSW = 0xD0

	ACC = 0xE0
	B   = 0xF0
)

var regList = []Register{
	{0x81, "SP"},
	{0x82, "DPL"},
	{0x83, "DPH"},
	{0x80, "P0"},
	{0x90, "P1"},
	{0xA0, "P2"},
	{0xB0, "P3"},
	{0xD0, "PSW"},
	{0xE0, "ACC"},
	{0xF0, "B"},
}

// FindRegByName find first one register by name, if not found return nil
func FindRegByName(name string, list []Register) *Register {
	for k, r := range list {
		if r.Name == name {
			return &list[k]
		}
	}
	return nil
}

// FindRegByAddr find first one register by address, if not found return nil
func FindRegByAddr(addr uint8, list []Register) *Register {
	for k, r := range list {
		if r.Addr == addr {
			return &list[k]
		}
	}
	return nil
}

// Register 8051 machine registers
type Register struct {
	Addr uint8
	Name string
}
