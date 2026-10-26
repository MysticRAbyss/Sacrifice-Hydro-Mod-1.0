package main

import "fmt"

func dbgPart(g *gameData, tag string, pi int) {
	cr, _ := loadOriginal(g, tag)
	var fr *animFrame
	if a := findAnim(cr, "stance"); a != nil {
		fr = &a.frames[0]
	}
	posed := pose(cr, fr)
	cnt := map[int]int{}
	sum := map[int]v3{}
	for i, v := range cr.parts[pi].verts {
		b := v.joints[0]
		if v.weights[1] > v.weights[0] {
			b = v.joints[1]
		}
		cnt[b]++
		sum[b] = sum[b].add(posed[pi][i])
	}
	pos, _ := boneTransforms(cr, fr)
	far := map[int]v3{}
	for i, v := range cr.parts[pi].verts {
		b := v.joints[0]
		if v.weights[1] > v.weights[0] {
			b = v.joints[1]
		}
		if f, ok := far[b]; !ok || posed[pi][i].sub(pos[b]).len() > f.sub(pos[b]).len() {
			far[b] = posed[pi][i]
		}
	}
	for b, n := range cnt {
		fmt.Println("   far", b, far[b])
		fmt.Printf("  part %d bone %d: %d verts, mean %v\n", pi, b, n, sum[b].mul(1/float64(n)))
	}
}
