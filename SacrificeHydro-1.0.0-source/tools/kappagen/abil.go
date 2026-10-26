package main

import (
	"encoding/binary"
	"fmt"
)

func dumpAbil(g *gameData, tag string) {
	b := g.raw[tag]
	r := func(o int) string { return rd4(b[o : o+4]) }
	// offsets: ability0 92, ability1 96, unknown22 100..124, ability2 124, unknown23 128, meleeStrength 132, mana 136, fsm 140, type 144, model 148, behaviorFlags 152
	fmt.Printf("%s abil0=%q abil1=%q abil2=%q passive=%q passive2=%q flags=%x health=%d souls=%d melee=%d ranged=%d speed=%d fly=%d mana=%d\n", tag, r(92), r(96), r(124), r(176+256+8), r(176+256+20),
		binary.LittleEndian.Uint32(b[152:]), binary.LittleEndian.Uint16(b[54:]), binary.LittleEndian.Uint16(b[82:]), binary.LittleEndian.Uint32(b[132:]), binary.LittleEndian.Uint16(b[60:]), binary.LittleEndian.Uint16(b[50:]), binary.LittleEndian.Uint16(b[52:]), binary.LittleEndian.Uint32(b[24:]))
}
