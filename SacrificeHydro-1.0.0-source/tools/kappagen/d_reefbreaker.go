package main

// Reefbreaker: a hermit crab living in a big coral shell with a coral cannon, on the Bombard's
// skeleton (stance pose). Bones: 9 body, 10-11-12 neck/head, 13-14 jaw, 15 top, 16-17-18 tail,
// front legs 1-2-3-4 (-x) and 5-6-7-8 (+x).
func reefbreakerDesign(b []v3) []*bodyDef {
	coral := matCoral(palette{rgb{206, 120, 108}, rgb{100, 44, 50}, rgb{246, 196, 160}, rgb{170, 150, 110}}, 901)
	crab := matChitin(palette{rgb{186, 84, 40}, rgb{80, 30, 14}, rgb{240, 160, 90}, rgb{120, 90, 50}}, 7, 902)
	bone := matCoral(palette{rgb{220, 208, 184}, rgb{120, 104, 84}, rgb{246, 240, 224}, rgb{190, 170, 130}}, 903)
	dark := plainTex(rgb{30, 20, 24}, rgb{10, 8, 10}, 904)
	var shell, body, legs, barrel []shape
	// spiral shell on the back, getting smaller towards the tail
	shell = append(shell,
		ell(v3{0, -0.2, 2.2}, v3{1.05, 1.0, 0.85}, []bw{{9, 0.7}, {16, 0.3}}),
		ell(v3{0, -1.05, 2.05}, v3{0.75, 0.6, 0.65}, one(16)),
		ell(v3{0, -1.65, 1.75}, v3{0.45, 0.4, 0.42}, one(17)),
		ell(v3{0, -2.0, 1.45}, v3{0.22, 0.22, 0.22}, one(18)),
	)
	// coral knobs on the shell
	for i, p := range []v3{{0.6, -0.4, 2.85}, {-0.7, 0.0, 2.7}, {0.2, -0.9, 2.65}, {-0.4, -1.4, 2.25}, {0.75, 0.3, 2.5}, {-0.2, 0.4, 2.95}} {
		bn := 9
		if i == 3 {
			bn = 16
		}
		shell = append(shell, capsl(p.mul(0.95), p.add(v3{p[0] * 0.15, 0, 0.3}), 0.14, 0.09, one(bn)))
	}
	// crab body and head poking out of the shell
	body = append(body,
		ell(v3{0, 0.65, 1.55}, v3{0.75, 0.65, 0.55}, []bw{{9, 0.5}, {10, 0.5}}),
		ell(b[11].add(v3{0, 0.05, 0}), v3{0.42, 0.35, 0.32}, one(11)),
		ell(b[13].add(v3{0, 0.1, 0}), v3{0.3, 0.25, 0.15}, one(13)), // jaw
	)
	eye := v3{0.28, 1.35, 2.35}
	body = append(body,
		capsl(b[11].add(v3{0.15, 0.1, 0.2}), eye.add(v3{0, 0, -0.1}), 0.07, 0.06, one(11)),
		capsl(b[11].add(v3{-0.15, 0.1, 0.2}), mirrorX(eye).add(v3{0, 0, -0.1}), 0.07, 0.06, one(11)),
	)
	// front legs ending in big claws
	for _, ch := range [][4]int{{1, 2, 3, 4}, {5, 6, 7, 8}} {
		p1, p2, p3, p4 := b[ch[0]], b[ch[1]], b[ch[2]], b[ch[3]]
		legs = append(legs,
			limb(v3{p1[0] * 0.35, 0.45, 1.6}, p1, 0.32, 0.3, 9, ch[0]), // shoulder from inside the body
			limb(p1, p2, 0.3, 0.24, ch[0], ch[1]),
			capsl(p2, p3, 0.24, 0.2, one(ch[1])),
			ell(p3.mul(0.5).add(p4.mul(0.5)).add(v3{0, 0.2, 0.05}), v3{0.32, 0.45, 0.3}, one(ch[2])),
			along(p4.add(v3{0, 0.25, 0.15}), p4.add(v3{0, 0.85, 0.25}), 0.14, 0.12, one(ch[3])),
			along(p4.add(v3{0, 0.25, -0.05}), p4.add(v3{0, 0.75, -0.08}), 0.12, 0.1, one(ch[3])),
		)
	}
	// coral cannon on top of the shell
	// the cannon grows out of a coral mound on the shell and moves with it
	gun0, gun1 := v3{0, -0.1, 2.75}, v3{0, 1.35, 3.25}
	shellB := []bw{{9, 0.7}, {16, 0.3}}
	shell = append(shell, ell(v3{0, -0.05, 2.85}, v3{0.55, 0.6, 0.4}, shellB))
	// a hollow barrel with a flared, ringed muzzle; the bore is dark inside
	dir := gun1.sub(gun0).norm()
	bore := []shape{{kind: capsule, a: gun1.sub(dir.mul(0.75)), b: gun1.add(dir.mul(0.6)), ra: 0.17, rb: 0.17}}
	barrel = append(barrel,
		shape{kind: capsule, a: gun0, b: gun1, ra: 0.3, rb: 0.27, bones: shellB, sub: bore},
		shape{kind: capsule, a: gun1.sub(dir.mul(0.16)), b: gun1.sub(dir.mul(0.02)), ra: 0.38, rb: 0.38, bones: shellB, sub: bore},
		shape{kind: capsule, a: gun1.sub(dir.mul(0.55)), b: gun1.sub(dir.mul(0.45)), ra: 0.33, rb: 0.33, bones: shellB, sub: bore},
	)
	muzzle := shape{kind: capsule, a: gun1.sub(dir.mul(0.7)), b: gun1.sub(dir.mul(0.62)), ra: 0.19, rb: 0.19, bones: shellB}
	return []*bodyDef{
		boxPart("shell", 1800, 0.04, 0.15, coral, shell...),
		boxPart("body", 1100, 0.035, 0.1, crab, body...),
		boxPart("legs", 1300, 0.035, 0.1, crab, legs...),
		boxPart("cannon", 1100, 0.018, 0.03, bone, barrel...),
		boxPart("muzzle", 120, 0.02, 0, dark, muzzle),
		eyesMulti([]v3{eye, mirrorX(eye)}, 0.12, []int{11, 11}, monsterEye(rgb{70, 170, 200}, rgb{10, 40, 60}, true, 0.3)),
	}
}
