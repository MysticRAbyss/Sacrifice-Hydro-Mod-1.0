package main

import (
	"fmt"
	"math"
)

// meshCheck reports, per body part, the orientation (signed volume) and topology defects:
// boundary edges (holes) and edges used twice in the same direction (flipped faces).
func meshCheck(cr *creature, label string) {
	var fr *animFrame
	if a := findAnim(cr, "stance"); a != nil && len(a.frames) > 0 {
		fr = &a.frames[0]
	}
	posed := pose(cr, fr)
	fmt.Println("check", label)
	for pi, pt := range cr.parts {
		key := func(p v3) [3]int64 {
			return [3]int64{int64(math.Round(p[0] * 2e4)), int64(math.Round(p[1] * 2e4)), int64(math.Round(p[2] * 2e4))}
		}
		ids := map[[3]int64]int{}
		vid := make([]int, len(pt.verts))
		for i := range pt.verts {
			k := key(posed[pi][i])
			id, ok := ids[k]
			if !ok {
				id = len(ids)
				ids[k] = id
			}
			vid[i] = id
		}
		edges := map[[2]int]int{}
		vol := 0.0
		for _, f := range pt.faces {
			a, b, c := posed[pi][f[0]], posed[pi][f[1]], posed[pi][f[2]]
			vol += a.dot(b.cross(c)) / 6
			for k := 0; k < 3; k++ {
				edges[[2]int{vid[f[k]], vid[f[(k+1)%3]]}]++
			}
		}
		boundary, same := 0, 0
		for e, n := range edges {
			if n > 1 {
				same++
			}
			if edges[[2]int{e[1], e[0]}] == 0 {
				boundary++
			}
		}
		fmt.Printf("  part %2d faces %5d  volume %+9.4f  boundary %4d  sameDir %4d\n", pi, len(pt.faces), vol, boundary, same)
	}
}
