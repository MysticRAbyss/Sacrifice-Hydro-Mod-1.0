package main

// Brineclaw: a chunky spiked crab with oversized claws, on the Tickferno's skeleton (stance pose).
// Tickferno stance bones: 1 body, 17 top, 8 rear, 15-16 head; claw arms 2-3 (+x), 13-14 (-x);
// legs 4-5, 6-7 (+x), 11-12, 9-10 (-x).
func brineclawDesign(b []v3) []*bodyDef {
	shell := matShell(palette{rgb{54, 106, 104}, rgb{18, 42, 46}, rgb{120, 170, 150}, rgb{90, 80, 50}}, 6, 601)
	red := matChitin(palette{rgb{176, 64, 38}, rgb{70, 20, 14}, rgb{236, 150, 90}, rgb{110, 80, 40}}, 6, 602)
	var body, spikes, legs, claws []shape
	body = append(body,
		ell(v3{0, 0.05, 0.50}, v3{0.62, 0.52, 0.26}, []bw{{1, 0.7}, {17, 0.3}}),
		ell(v3{0, 0.45, 0.48}, v3{0.38, 0.22, 0.18}, one(15)), // front rim / face
		ell(v3{0, -0.35, 0.50}, v3{0.42, 0.25, 0.2}, one(8)),
	)
	// spikes on the shell, a ring and a crest
	for i := 0; i < 9; i++ {
		a := float64(i)/9*2*3.14159 + 0.3
		x, y := 0.48*cosf(a), 0.38*sinf(a)
		base := v3{x * 0.85, y*0.85 + 0.05, 0.56}
		spikes = append(spikes, capsl(base, base.add(v3{x * 0.45, y * 0.45, 0.36}), 0.075, 0.008, []bw{{1, 0.7}, {17, 0.3}}))
	}
	for _, p := range []v3{{0, 0.05, 0.74}, {0.2, -0.1, 0.72}, {-0.2, -0.1, 0.72}, {0, -0.25, 0.68}} {
		spikes = append(spikes, capsl(p.add(v3{0, 0, -0.1}), p.add(v3{0, 0, 0.3}), 0.08, 0.01, []bw{{1, 0.7}, {17, 0.3}}))
	}
	// walking legs: two segments plus a pointed foot
	for _, c := range [][2]int{{4, 5}, {6, 7}, {11, 12}, {9, 10}} {
		p0, p1 := b[c[0]], b[c[1]]
		foot := p1.add(p1.sub(p0).mul(0.35)).add(v3{0, 0, -0.45})
		legs = append(legs,
			limb(p0.mul(0.45).add(v3{0, 0, 0.1}), p0, 0.09, 0.07, 1, c[0]),
			capsl(p0, p1, 0.07, 0.06, one(c[0])),
			capsl(p1, foot, 0.06, 0.012, one(c[1])),
		)
	}
	// big claws
	for _, c := range [][2]int{{2, 3}, {13, 14}} {
		p0, p1 := b[c[0]], b[c[1]]
		s := 1.0
		if p1[0] < 0 {
			s = -1
		}
		palm := p1.add(v3{s * 0.05, 0.12, 0.05})
		claws = append(claws,
			limb(p0.mul(0.45).add(v3{0, 0, 0.05}), p0, 0.1, 0.08, 1, c[0]),
			capsl(p0, p1, 0.08, 0.09, one(c[0])),
			ell(palm, v3{0.17, 0.2, 0.15}, one(c[1])),
			along(palm.add(v3{0, 0.12, 0.06}), palm.add(v3{-s * 0.06, 0.48, 0.1}), 0.07, 0.06, one(c[1])),
			along(palm.add(v3{0, 0.12, -0.06}), palm.add(v3{-s * 0.04, 0.42, -0.08}), 0.06, 0.05, one(c[1])),
		)
	}
	// eye stalks
	eyeC := v3{0.13, 0.62, 0.86}
	body = append(body,
		capsl(v3{0.1, 0.5, 0.55}, eyeC.add(v3{0, 0, -0.05}), 0.035, 0.03, one(15)),
		capsl(v3{-0.1, 0.5, 0.55}, mirrorX(eyeC).add(v3{0, 0, -0.05}), 0.035, 0.03, one(15)),
	)
	return []*bodyDef{
		boxPart("shell", 1600, 0.025, 0.08, shell, body...),
		boxPart("spikes", 700, 0.014, 0.02, red, spikes...),
		boxPart("legs", 900, 0.018, 0.03, red, legs...),
		boxPart("claws", 1200, 0.018, 0.05, red, claws...),
		eyesMulti([]v3{eyeC, mirrorX(eyeC)}, 0.055, []int{15, 15}, monsterEye(rgb{230, 120, 30}, rgb{90, 20, 10}, true, 0.3)),
	}
}
