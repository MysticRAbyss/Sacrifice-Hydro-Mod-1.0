package main

import "fmt"

func findExplicit(g *gameData) {
	n := 0
	for name, mf := range g.models {
		if mf.sxmd == nil {
			continue
		}
		m, err := parseSXMD(mf.sxmd)
		if err != nil {
			continue
		}
		for k, p := range m.parts {
			if p.numExplicit > 0 && n < 12 {
				n++
				pg := mf.pages[k+1]
				h := 0
				if pg != nil {
					h = int(pg[14]) | int(pg[15])<<8
				}
				fmt.Printf("%s part %d rings %d explicit %d entries %d flags %x texH %d\n", name, k, len(p.rings), p.numExplicit, len(p.rings[0].entries), p.flags, h)
			}
		}
	}
}
