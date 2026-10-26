package main

import (
	"encoding/binary"
	"fmt"
)

var statNames = []string{"minCastingTime", "castingTime", "spellOrder", "unknown9", "meleeAccuracy", "meleeEvasion", "aggressiveness", "runningSpeed", "flyingSpeed", "health", "regeneration", "drain", "rangedAccuracy", "meleeResistance", "directSpellResistance", "splashSpellResistance", "directRangedResistance", "splashRangedResistance", "unknown15", "unknown16", "sightRange", "sightFov", "god", "souls"}

func dumpStats(g *gameData, tag string) {
	b := g.raw[tag]
	fmt.Printf("%s %-12s cool %d mana %d ", tag, g.cres[tag].model, binary.LittleEndian.Uint32(b[20:]), binary.LittleEndian.Uint32(b[24:]))
	for i, n := range statNames {
		fmt.Printf("%s=%d ", n, binary.LittleEndian.Uint16(b[36+2*i:]))
	}
	// meleeStrength, mana after ability fields: compute offsets from struct
	fmt.Printf("meleeStrength=%d manaStat=%d\n", binary.LittleEndian.Uint32(b[132:]), binary.LittleEndian.Uint32(b[136:]))
}
