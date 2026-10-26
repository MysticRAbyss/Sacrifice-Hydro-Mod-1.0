package main

// Reefstalker: a naga archer - a broad-shouldered scaly torso with a reptile head, fin "ears"
// and a long serpent tail instead of legs, in gold armour (shoulder guards, breastplate with a
// gem, bracers, belt), with a coral bow. Built on the Ranger's skeleton (stance pose); the tail
// follows the average of both legs, so it sways as he moves.
// Human bones: legs 1-2-3 / 18-19-20, spine 4, chest 5, neck 6, head 7,
// left arm 8-9-10-11, right arm 12-13-14-15, bow 16-17.
func reefstalkerDesign(b []v3) []*bodyDef {
	skin := matScales(palette{rgb{52, 124, 72}, rgb{14, 46, 28}, rgb{150, 200, 110}, rgb{170, 150, 60}}, 56, 1301)
	belly := matChitin(palette{rgb{196, 186, 120}, rgb{90, 80, 40}, rgb{236, 226, 170}, rgb{150, 160, 90}}, 18, 1304)
	fin := matFin(palette{rgb{220, 140, 50}, rgb{110, 50, 16}, rgb{250, 210, 120}, rgb{80, 130, 70}}, 9, 1302)
	gold := matChitin(palette{rgb{200, 160, 60}, rgb{90, 60, 16}, rgb{250, 220, 130}, rgb{150, 110, 40}}, 4, 1305)
	gem := matStreak(palette{rgb{200, 30, 40}, rgb{80, 6, 12}, rgb{255, 140, 140}, rgb{160, 20, 60}}, 1306)
	horn := matStreak(palette{rgb{40, 40, 36}, rgb{10, 10, 10}, rgb{120, 120, 100}, rgb{60, 50, 40}}, 1307)
	avg := func(a, c int) []bw { return []bw{{a, 0.5}, {c, 0.5}} }
	var body, plates, fins, armour, gems, horns []shape
	head := b[7]
	hd := func(x, y, z float64) v3 { return head.add(v3{x, y, z}) }
	body = append(body,
		ell(v3{0, 0.0, 1.0}, v3{0.15, 0.13, 0.13}, avg(1, 18)), // hips
		ell(v3{0, 0.02, 1.18}, v3{0.14, 0.11, 0.15}, one(4)),   // waist
		ell(v3{0, 0.04, 1.36}, v3{0.21, 0.13, 0.14}, one(5)),   // broad chest
		ell(v3{0.09, 0.1, 1.38}, v3{0.09, 0.06, 0.07}, one(5)),  // pectorals
		ell(v3{-0.09, 0.1, 1.38}, v3{0.09, 0.06, 0.07}, one(5)),
		limb(v3{0, 0.06, 1.44}, hd(0, -0.08, -0.08), 0.07, 0.06, 5, 7), // thick neck
		// reptile head: skull, long snout, heavy brows, jaw
		ell(hd(0, -0.02, 0.02), v3{0.08, 0.1, 0.085}, one(7)),
		along(hd(0, 0.03, 0.0), hd(0, 0.17, -0.03), 0.06, 0.05, one(7)),
		along(hd(0.04, 0.02, 0.05), hd(0.035, 0.12, 0.03), 0.028, 0.02, one(7)),
		along(hd(-0.04, 0.02, 0.05), hd(-0.035, 0.12, 0.03), 0.028, 0.02, one(7)),
		along(hd(0, 0.0, -0.06), hd(0, 0.15, -0.105), 0.048, 0.02, one(7)), // open lower jaw
		// tail: hips -> knees -> ankles of both legs averaged, then along the ground
		shape{kind: capsule, a: v3{0, 0.0, 0.98}, b: v3{0, 0.12, 0.6}, ra: 0.16, rb: 0.13, bones: avg(1, 18)},
		shape{kind: capsule, a: v3{0, 0.12, 0.6}, b: v3{0, -0.05, 0.17}, ra: 0.13, rb: 0.09, bones: avg(2, 19)},
		shape{kind: capsule, a: v3{0, -0.05, 0.17}, b: v3{0.08, -0.42, 0.08}, ra: 0.09, rb: 0.06, bones: avg(3, 20)},
		shape{kind: capsule, a: v3{0.08, -0.42, 0.08}, b: v3{-0.05, -0.75, 0.06}, ra: 0.06, rb: 0.02, bones: avg(3, 20)},
	)
	// pale belly plates down the front of the tail
	plates = append(plates,
		ell(v3{0, 0.11, 1.0}, v3{0.1, 0.05, 0.12}, avg(1, 18)),
	)
	// muscular arms with clawed hands; gold bracers on the forearms
	for _, a := range [][4]int{{8, 9, 10, 11}, {12, 13, 14, 15}} {
		sh, el, wr, hn := b[a[0]], b[a[1]], b[a[2]], b[a[3]]
		sx := 1.0
		if sh[0] < 0 {
			sx = -1
		}
		body = append(body,
			limb(v3{sx * 0.1, 0.04, 1.42}, el, 0.07, 0.055, 5, a[1]),
			limb(el, wr, 0.055, 0.042, a[1], a[2]),
			limb(wr, hn, 0.04, 0.034, a[2], a[3]),
			ell(hn, v3{0.04, 0.04, 0.035}, one(a[3])),
		)
		fa := wr.sub(el)
		armour = append(armour,
			limb(el.add(fa.mul(0.35)), el.add(fa.mul(0.85)), 0.058, 0.05, a[1], a[2]), // bracer
			// layered shoulder guard over the shoulder joint
			ell(el.mul(0.25).add(v3{sx * 0.16, 0.04, 1.47}.mul(0.75)), v3{0.1, 0.1, 0.07}, []bw{{5, 0.4}, {a[1], 0.6}}),
			ell(el.mul(0.25).add(v3{sx * 0.17, 0.04, 1.43}.mul(0.75)), v3{0.105, 0.1, 0.05}, []bw{{5, 0.4}, {a[1], 0.6}}),
		)
		horns = append(horns, capsl(el.mul(0.25).add(v3{sx * 0.17, 0.04, 1.5}.mul(0.75)), el.mul(0.25).add(v3{sx * 0.24, 0.0, 1.62}.mul(0.75)), 0.022, 0.004, []bw{{5, 0.4}, {a[1], 0.6}}))
		for k := 0; k < 3; k++ { // claws
			c := hn.add(v3{sx * 0.01 * float64(k-1), 0.03, -0.02})
			horns = append(horns, capsl(c, c.add(v3{0, 0.04, -0.04}), 0.01, 0.002, one(a[3])))
		}
		// spiny fin "ears" and a fin on each forearm
		fins = append(fins, rayFin(hd(sx*0.06, -0.03, 0.02), []v3{hd(sx*0.14, -0.06, 0.16), hd(sx*0.2, -0.12, 0.09), hd(sx*0.22, -0.16, 0.01), hd(sx*0.18, -0.17, -0.07)}, 0.012, 0.005, one(7))...)
		fins = append(fins,
			sheet(wr.add(el).mul(0.5), el.add(v3{sx * 0.03, -0.1, 0.05}), v3{0, -1, 0}, 0.04, 0.008, one(a[1])),
		)
		horns = append(horns, // horns sweeping back from the brows
			capsl(hd(sx*0.04, 0.0, 0.08), hd(sx*0.08, -0.16, 0.16), 0.02, 0.004, one(7)),
		)
	}
	// gold breastplate with a red gem, gold belt
	armour = append(armour,
		ell(v3{0, 0.125, 1.35}, v3{0.15, 0.055, 0.1}, one(5)),
		capsl(v3{-0.11, 0.14, 1.43}, v3{0.11, 0.14, 1.43}, 0.018, 0.018, one(5)),
		shape{kind: torusZ, c: v3{0, 0.02, 1.1}, r: v3{0.145, 0.025, 0}, bones: one(4)},
	)
	gems = append(gems, along(hd(0, 0.02, -0.055), hd(0, 0.14, -0.07), 0.035, 0.018, one(7))) // dark red mouth
	gems = append(gems, ell(v3{0, 0.18, 1.37}, v3{0.035, 0.02, 0.035}, one(5)), ell(v3{0, 0.17, 1.1}, v3{0.025, 0.015, 0.025}, one(4)))
	// crest fin from the head down the back, tail fin
	fins = append(fins,
		sheet(v3{0, -0.12, 1.4}, v3{0, -0.08, 0.8}, v3{0, -1, 0}, 0.07, 0.012, []bw{{4, 0.5}, {1, 0.25}, {18, 0.25}}),
	)
	// wide, sharp tail fluke: two pointed lobes of spiny rays spreading out sideways
	for _, sx := range []float64{-1, 1} {
		root := v3{-0.05, -0.7, 0.06}
		fins = append(fins, rayFin(root, []v3{
			{-0.05 + sx*0.12, -0.86, 0.07}, {-0.05 + sx*0.24, -0.98, 0.07}, {-0.05 + sx*0.36, -1.12, 0.08}, {-0.05 + sx*0.3, -1.0, 0.075}}, 0.022, 0.008, avg(3, 20))...)
	}
	// spiny crest over the head
	fins = append(fins, rayFin(hd(0, -0.02, 0.06), []v3{hd(0, 0.08, 0.2), hd(0, -0.04, 0.24), hd(0, -0.16, 0.2), hd(0, -0.26, 0.1)}, 0.013, 0.005, one(7))...)
	// mouth: a row of small teeth
	for k := 0; k < 4; k++ {
		y := 0.06 + 0.03*float64(k)
		for _, sx := range []float64{-1, 1} {
			horns = append(horns, capsl(hd(sx*0.035, y, -0.04), hd(sx*0.035, y+0.005, -0.064), 0.007, 0.002, one(7)))
			horns = append(horns, capsl(hd(sx*0.03, y-0.01, -0.085-0.01*float64(k)), hd(sx*0.03, y-0.005, -0.065-0.01*float64(k)), 0.006, 0.002, one(7)))
		}
	}
	_ = head
	// her own coral bow, in one piece, held rigidly in her right hand (bone 15). It is laid out in
	// the middle of her shooting animation, upright, with the grip as its far point and the string
	// between her and the bow, and carried back to the stance pose the model is built in.
	G := b[15].add(v3{0, 0.02, -0.08})
	up, fwd := v3{0, 0, 1}, v3{0, 1, 0}
	toStance := func(p v3) v3 { return p }
	if designPose != nil {
		ps, rs := designPose("stance", 0)
		ph, rh := designPose("shoot0", 0.5)
		if ph != nil {
			inv := func(q [4]float64) [4]float64 { return [4]float64{-q[0], -q[1], -q[2], q[3]} }
			toShoot := func(p v3) v3 { return ph[15].add(qrot(rh[15], qrot(inv(rs[15]), p.sub(ps[15])))) }
			toStance = func(p v3) v3 { return ps[15].add(qrot(rs[15], qrot(inv(rh[15]), p.sub(ph[15])))) }
			G = toShoot(G)
			fwd = G.sub(ph[5])
			fwd[2] = 0
			fwd = fwd.norm()
		}
	}
	U, L := G.add(up.mul(0.85)).sub(fwd.mul(0.32)), G.sub(up.mul(0.8)).sub(fwd.mul(0.32))
	A, B := U.add(G).mul(0.5).add(fwd.mul(0.07)), G.add(L).mul(0.5).add(fwd.mul(0.07))
	tU, tL := U.sub(fwd.mul(0.08)).add(up.mul(0.06)), L.sub(fwd.mul(0.08)).sub(up.mul(0.06)) // recurved tips
	pts := []v3{tU, U, A, G, B, L, tL}
	for i := range pts {
		pts[i] = toStance(pts[i])
	}
	gs := toStance(G)
	gd := toStance(G.add(up)).sub(gs).norm()
	rads := []float64{0.012, 0.018, 0.028, 0.038, 0.028, 0.018, 0.012}
	var bowShapes []shape
	for i := 0; i+1 < len(pts); i++ {
		bowShapes = append(bowShapes, capsl(pts[i], pts[i+1], rads[i], rads[i+1], one(15)))
	}
	bowShapes = append(bowShapes,
		capsl(gs.add(gd.mul(0.09)), gs.sub(gd.mul(0.09)), 0.045, 0.045, one(15)), // wrapped grip
		capsl(pts[1], pts[5], 0.007, 0.007, one(15)),                             // string
	)
	bowTex := matStreak(palette{rgb{150, 92, 70}, rgb{60, 30, 28}, rgb{220, 170, 140}, rgb{90, 120, 120}}, 1303)
	return []*bodyDef{
		boxPart("bow", 900, 0.005, 0.01, bowTex, bowShapes...),
		boxPart("body", 3200, 0.011, 0.04, skin, body...),
		boxPart("belly", 300, 0.01, 0.03, belly, plates...),
		boxPart("armour", 1600, 0.007, 0.004, gold, armour...),
		boxPart("gems", 120, 0.006, 0.01, gem, gems...),
		boxPart("horns", 500, 0.004, 0, horn, horns...),
		boxPart("fins", 900, 0.006, 0.012, fin, fins...),
		eyesMulti([]v3{hd(0.052, 0.075, 0.035), hd(-0.052, 0.075, 0.035)}, 0.026, []int{7, 7}, monsterEye(rgb{250, 170, 40}, rgb{140, 50, 10}, true, 0.4)),
	}
}
