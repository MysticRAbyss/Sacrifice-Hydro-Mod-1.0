package main

import (
	"encoding/hex"
	"fmt"
)

func analyze(g *gameData, tag string) {
	var mf *modelFiles
	if c := g.cres[tag]; c != nil {
		mf = g.models[c.model]
	} else {
		mf = g.models[tag]
	}
	m, _ := parseSXMD(mf.sxmd)
	d := mf.sxmd
	fmt.Println("header:", hex.EncodeToString(d[:48]))
	fmt.Println("saxs:", hex.EncodeToString(mf.saxs[:32]), len(mf.saxs))
	for k, p := range m.parts {
		off := int(u32(d, 40+4*k))
		fmt.Printf("part %d rings %d explicit %d flags %x hdr %s\n", k, len(p.rings), p.numExplicit, p.flags, hex.EncodeToString(d[off:off+36]))
		if p.numExplicit > 0 {
			// orientation of explicit faces vs stored normals
			pos, neg := 0, 0
			r := p.rings[0]
			for _, f := range p.explicit {
				var P [3]v3
				var N v3
				for j := 0; j < 3; j++ {
					e := r.entries[f[j]]
					sp := m.positions[e.idx[0]]
					P[j] = v3{float64(sp.pos[0]), float64(sp.pos[1]), float64(sp.pos[2])}
					N = N.add(v3{float64(sp.normal[0]), float64(sp.normal[1]), float64(sp.normal[2])})
					// note: positions are bone-relative; good enough when all three share a bone
				}
				e1 := v3{P[1][0] - P[0][0], P[1][1] - P[0][1], P[1][2] - P[0][2]}
				e2 := v3{P[2][0] - P[0][0], P[2][1] - P[0][1], P[2][2] - P[0][2]}
				cr := v3{e1[1]*e2[2] - e1[2]*e2[1], e1[2]*e2[0] - e1[0]*e2[2], e1[0]*e2[1] - e1[1]*e2[0]}
				if cr[0]*N[0]+cr[1]*N[1]+cr[2]*N[2] > 0 {
					pos++
				} else {
					neg++
				}
			}
			fmt.Printf("   explicit orientation (SXMD space) cross·n >0: %d, <0: %d; ring entries %d\n", pos, neg, len(r.entries))
		}
	}
	for i, pg := range mf.pages {
		fmt.Printf("page %d: %dx%d\n", i, int(pg[12])|int(pg[13])<<8, int(pg[14])|int(pg[15])<<8)
	}
}
