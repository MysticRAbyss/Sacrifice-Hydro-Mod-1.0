package main

// Mantis Shrimp: a colourful, bug-eyed shrimp with folded punching claws, on the Vortick's
// six-legged skeleton (designed in its stance pose).
// Vortick stance bones: 1 body, 2-3 head, 4 back; legs 5-6-7, 8-9-10, 11-12-13 (-x), 14-22 (+x).
func mantisDesign(b []v3) []*bodyDef {
	shell := matChitin(palette{rgb{44, 122, 82}, rgb{14, 50, 40}, rgb{120, 190, 120}, rgb{150, 70, 50}}, 9, 501)
	limbTex := matChitin(palette{rgb{190, 70, 44}, rgb{80, 24, 20}, rgb{240, 160, 90}, rgb{120, 90, 50}}, 6, 502)
	var body, legs, claws []shape
	// segmented body from tail (y -0.8) to head (y 0.75)
	segs := []struct {
		y, r, rz float64
		bone     int
	}{{-0.72, 0.13, 0.08, 1}, {-0.55, 0.16, 0.11, 1}, {-0.37, 0.19, 0.14, 1}, {-0.18, 0.21, 0.16, 1}, {0.02, 0.22, 0.17, 1}, {0.22, 0.21, 0.16, 2}, {0.42, 0.19, 0.15, 2}, {0.62, 0.15, 0.13, 3}}
	for i, s := range segs {
		z := 0.36 + 0.02*float64(i)
		if s.y < -0.3 {
			z = 0.36 + (s.y+0.3)*0.25
		}
		body = append(body, ell(v3{0, s.y, z}, v3{s.r, 0.12, s.rz}, one(s.bone)))
	}
	// tail fan
	body = append(body,
		along(v3{0, -0.78, 0.25}, v3{-0.22, -0.98, 0.22}, 0.07, 0.02, one(1)),
		along(v3{0, -0.78, 0.25}, v3{0.22, -0.98, 0.22}, 0.07, 0.02, one(1)),
		along(v3{0, -0.78, 0.25}, v3{0, -1.02, 0.22}, 0.07, 0.02, one(1)),
	)
	// antennae and eye stalks
	for _, s := range []float64{-1, 1} {
		body = append(body,
			capsl(v3{s * 0.06, 0.70, 0.40}, v3{s * 0.25, 1.15, 0.75}, 0.018, 0.008, one(3)),
			capsl(v3{s * 0.06, 0.72, 0.42}, v3{s * 0.12, 0.78, 0.60}, 0.03, 0.025, one(3)),
		)
	}
	// six thin walking legs: bone chain, then a foot down to the ground
	chains := [][3]int{{5, 6, 7}, {8, 9, 10}, {11, 12, 13}, {14, 15, 16}, {17, 18, 19}, {20, 21, 22}}
	for _, c := range chains {
		p0, p1, p2 := b[c[0]], b[c[1]], b[c[2]]
		out := 1.0
		if p2[0] < 0 {
			out = -1
		}
		foot := p2.add(v3{out * 0.12, (p2[1] - p1[1]) * 0.3, -0.5})
		legs = append(legs,
			limb(v3{p0[0] * 0.25, p0[1], 0.38}, p1, 0.05, 0.035, 1, c[0]),
			capsl(p1, p2, 0.035, 0.03, one(c[1])),
			capsl(p2, foot, 0.03, 0.012, one(c[2])),
		)
	}
	// folded raptorial claws ("punchers") under the head
	for _, s := range []float64{-1, 1} {
		claws = append(claws,
			capsl(v3{s * 0.05, 0.46, 0.4}, v3{s * 0.14, 0.62, 0.22}, 0.07, 0.06, one(2)), // shoulder from inside the body
			capsl(v3{s * 0.14, 0.62, 0.22}, v3{s * 0.2, 0.86, 0.2}, 0.06, 0.055, one(2)),
			along(v3{s * 0.2, 0.86, 0.2}, v3{s * 0.18, 0.76, 0.03}, 0.08, 0.09, one(2)), // the club
		)
	}
	return []*bodyDef{
		boxPart("body", 1800, 0.018, 0.05, shell, body...),
		boxPart("legs", 900, 0.012, 0.02, limbTex, legs...),
		boxPart("claws", 700, 0.012, 0.06, limbTex, claws...),
		eyesMulti([]v3{{0.12, 0.78, 0.63}, {-0.12, 0.78, 0.63}}, 0.045, []int{3, 3}, monsterEye(rgb{120, 230, 140}, rgb{20, 70, 40}, false, 0.3)),
	}
}
