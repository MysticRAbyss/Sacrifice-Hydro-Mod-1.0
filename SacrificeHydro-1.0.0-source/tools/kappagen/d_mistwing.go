package main

// Mistwing: a glowing manta ray, on the Shrike's skeleton (rest pose). One flat, diamond-shaped
// body with broad swept-back wings in a single piece; the wings are skinned smoothly from the
// body out along the wing bones, so they bend with the wingbeat and never come apart. Cephalic
// fins in front, two ribbons trailing behind.
// Shrike bones: 1 hips, 2 chest, 3 neck, 4 jaw, 5 head; wings 6-7-8 (+x) and 9-10-11 (-x).
// a manta's wingbeat replaces the Shrike's bird-like folding in every animation: each wing
// bends in a wave from the root to the tip, and the body stays level
var mistwingWave = func() []waveBone {
	var w []waveBone
	for _, bn := range []int{0, 1, 2, 3} {
		w = append(w, waveBone{bone: bn, axis: v3{0, 0, 1}, amp: 0})
	}
	amps := []float64{0.42, 0.3, 0.3}
	for k, bn := range []int{6, 7, 8} {
		w = append(w, waveBone{bone: bn, axis: v3{0, 1, 0}, amp: amps[k], phase: -0.7 * float64(k)})
		w = append(w, waveBone{bone: bn + 3, axis: v3{0, -1, 0}, amp: amps[k], phase: -0.7 * float64(k)})
	}
	return w
}()

func mistwingDesign(b []v3) []*bodyDef {
	top := withGlowSpots(matSkin(palette{rgb{38, 62, 120}, rgb{12, 20, 52}, rgb{90, 120, 180}, rgb{50, 40, 90}}, 8, 301), rgb{90, 230, 255}, 14, 0.22, 305)
	z := 0.86
	var body, tendrils []shape
	// flat body: broad in front, narrowing to the tail
	body = append(body,
		ell(v3{0, 0.0, z}, v3{0.36, 0.62, 0.16}, []bw{{2, 0.7}, {3, 0.3}}),
		ell(v3{0, 0.36, z + 0.02}, v3{0.32, 0.26, 0.12}, one(3)),
		ell(v3{0, -0.5, z - 0.01}, v3{0.16, 0.3, 0.1}, one(2)),
	)
	for _, s := range []float64{1, -1} {
		keys := []spanKey{{0.18, 2}, {0.5, 6}, {0.92, 7}, {1.4, 8}}
		if s < 0 {
			keys = []spanKey{{0.18, 2}, {0.5, 9}, {0.92, 10}, {1.4, 11}}
		}
		// wing: chordwise slices from the leading edge (swept back) to the trailing edge
		const n = 12
		for i := 0; i <= n; i++ {
			t := float64(i) / n
			x := 0.15 + 2.0*t
			lead := v3{s * x, 0.42 - 0.85*t*t - 0.1*t, z + 0.06*t}
			trail := v3{s * x * 0.98, -0.48 - 0.05*t + 0.2*t*t, z + 0.04*t}
			if t > 0.75 { // pointed tip, curving back
				u := (t - 0.75) / 0.25
				trail = trail.add(lead.sub(trail).mul(0.85 * u))
			}
			th := 0.13*(1-t) + 0.025
			w := 0.11
			sh := along(lead, trail, w, th, nil)
			sh.span = keys
			body = append(body, sh)
		}
		// cephalic fins: rolled-up horns at the front
		body = append(body, along(v3{s * 0.2, 0.5, z + 0.01}, v3{s * 0.24, 0.8, z + 0.06}, 0.05, 0.04, one(5)))
	}
	// two long ribbons trailing behind the body
	for _, s := range []float64{-1, 1} {
		x := s * 0.1
		up := v3{0, 0, 1}
		tendrils = append(tendrils,
			sheet(v3{x, -0.6, z - 0.02}, v3{x * 1.4, -1.35, z}, up, 0.06, 0.022, one(2)),
			sheet(v3{x * 1.4, -1.25, z}, v3{x * 2.0, -2.05, z - 0.1}, up, 0.05, 0.02, []bw{{2, 0.6}, {1, 0.4}}),
			sheet(v3{x * 2.0, -1.95, z - 0.1}, v3{x * 2.4, -2.8, z - 0.05}, up, 0.04, 0.018, []bw{{2, 0.4}, {1, 0.3}, {12, 0.3}}),
		)
	}
	return []*bodyDef{
		boxPart("body", 2600, 0.02, 0.1, top, body...),
		boxPart("ribbons", 900, 0.014, 0.06, matFin(palette{rgb{80, 170, 220}, rgb{30, 60, 120}, rgb{170, 240, 255}, rgb{90, 70, 160}}, 5, 302), tendrils...),
		eyesMulti([]v3{{0.27, 0.42, z + 0.07}, {-0.27, 0.42, z + 0.07}}, 0.05, []int{3, 3}, monsterEye(rgb{120, 250, 255}, rgb{20, 80, 140}, false, 0.6)),
	}
}
