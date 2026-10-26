package main

import "fmt"

func headBob(cr *creature) {
	a := findAnim(cr, "stance")
	for f := 0; f < len(a.frames); f += 6 {
		pos, rot := boneTransforms(cr, &a.frames[f])
		s := ""
		for _, b := range []int{17, 21, 22} {
			tip := pos[b].add(qrot(rot[b], v3{0, 1.0, 0}))
			s += fmt.Sprintf(" %5.2f", tip[2])
		}
		fmt.Println("frame", f, s)
	}
}
