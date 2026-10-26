package main

import "math"

// Tidecaller: a sea witch caster - a human upper body (tan skin, long flowing teal hair, a
// sea-steel armour) on a mass of eight teal octopus tentacles with pale blue suckers, holding a
// pearl staff. Built on the Sylph's skeleton (stance pose). The two leg chains run inside the
// two front tentacles and sway, so those tentacles curl and wriggle; the others follow the hips.
// Sylph bones: 1-2-3 / 18-19-20 legs, 4 spine, 5 chest, 6 neck, 7 head, 9-10-11 left arm,
// 13-14-15 right arm, 16-17 the weapon held in the right hand.
// the leg bones hang off the root, which the Sylph's animations turn a quarter turn about x:
// in that frame (0,-1,0) is "up", so the tentacles wriggle sideways over the ground
var tidecallerWave = []waveBone{
	{bone: 1, axis: v3{0, -1, 0}, amp: 0.15, phase: 0}, {bone: 2, axis: v3{0, -1, 0}, amp: 0.2, phase: -0.9}, {bone: 3, axis: v3{0, -1, 0}, amp: 0.2, phase: -1.8},
	{bone: 18, axis: v3{0, -1, 0}, amp: 0.15, phase: 2.5}, {bone: 19, axis: v3{0, -1, 0}, amp: 0.2, phase: 1.6}, {bone: 20, axis: v3{0, -1, 0}, amp: 0.2, phase: 0.7},
}

// the tentacles stay on the ground while the upper body bobs
var tidecallerNoDisp = []int{0, 1, 2, 3, 18, 19, 20}

// tentacles 1..6 each get their own bone (added after the skeleton), swaying from side to side
var tidecallerExtra = func() []extraBone {
	var e []extraBone
	for k := 1; k <= 6; k++ {
		e = append(e, extraBone{src: 4, parent: 0})
		tidecallerWave = append(tidecallerWave, waveBone{bone: 20 + k, axis: v3{0, -1, 0}, amp: 0.13 + 0.03*float64(k%3), phase: 1.1 * float64(k)})
		tidecallerNoDisp = append(tidecallerNoDisp, 20+k)
	}
	return e
}()

// while shooting, the Tidecaller raises its staff high (the pose at the end of the Sylph's "rise"),
// and the water bolt leaves from the top of the staff (creature.json "shootFrom")
var tidecallerPoses = []posePose{{anims: []string{"shoot0", "shoot1"}, from: "rise", frame: 1.0,
	bones: []int{4, 5, 9, 10, 11, 13, 14, 15, 16, 17}, in: 0.25, out: 0.8}}

func tidecallerDesign(b []v3) []*bodyDef {
	// smooth, even sea-green skin
	skin := gritty(wrinkleRelief(2, 701), palette{rgb{138, 176, 140}, rgb{110, 146, 116}, rgb{170, 204, 168}, rgb{128, 168, 140}}, 0.15, 0.3, 701)
	tent := matSkin(palette{rgb{34, 156, 146}, rgb{14, 84, 84}, rgb{110, 214, 200}, rgb{40, 130, 160}}, 4, 702)
	shellTex := matChitin(palette{rgb{60, 170, 160}, rgb{20, 90, 90}, rgb{150, 230, 210}, rgb{80, 150, 170}}, 9, 705)
	goldTex := matStreak(palette{rgb{196, 156, 70}, rgb{100, 70, 24}, rgb{240, 210, 130}, rgb{120, 140, 100}}, 708)
	var body, arms, hair, tentacles, suckers, staff []shape
	H := v3{0, 0.06, 1.6} // centre of the head
	hd := func(x, y, z float64) v3 { return H.add(v3{x, y, z}) }
	// a slim, athletic upper body: flat belly, narrow waist, moderate hips, a small bust that
	// the clam shells cover
	const W = 1.15
	body = append(body,
		ell(v3{0, 0.0, 0.95}, v3{0.1 * W, 0.08, 0.12}, one(4)),                         // hips (under the gown)
		ell(v3{0, 0.02, 1.1}, v3{0.08 * W, 0.058, 0.1}, one(4)),                        // waist (under the gown)
		ell(v3{0, 0.02, 1.26}, v3{0.08 * W, 0.058, 0.09}, one(5)),                      // rib cage (under the gown)
		ell(v3{0, 0.01, 1.36}, v3{0.14 * W, 0.07, 0.055}, one(5)),                      // upper chest and back
		capsl(v3{-0.13 * W, 0.02, 1.4}, v3{0.13 * W, 0.02, 1.4}, 0.045, 0.045, one(5)), // shoulders
		limb(v3{0, 0.02, 1.4}, hd(0, -0.02, -0.08), 0.038, 0.034, 5, 7),                // slender neck
	)
	// a tentacle through the given points, curling at the end, with small suckers underneath
	tentHair := func(pts []v3, r0 float64, bones []int, curl float64) {
		// a curl at the end: a spiral continuing from the last point
		last, prev := pts[len(pts)-1], pts[len(pts)-2]
		d := last.sub(prev).norm()
		side := d.cross(v3{0, 0, 1})
		if side.len() < 0.1 {
			side = v3{1, 0, 0}
		}
		side = side.norm()
		bend := side.cross(d).norm().mul(curl)
		seg := last.sub(prev).len() * 0.6
		for i := 0; i < 6; i++ {
			a := float64(i+1) * 0.9
			last = last.add(d.mul(seg * math.Cos(a)).add(bend.mul(seg * math.Sin(a))))
			seg *= 0.82
			pts = append(pts, last)
		}
		n := len(pts) - 1
		for i := 0; i < n; i++ {
			t0, t1 := float64(i)/float64(n), float64(i+1)/float64(n)
			b0, b1 := bones[min(i, len(bones)-1)], bones[min(i+1, len(bones)-1)]
			hair = append(hair, limb(pts[i], pts[i+1], r0*(1-0.8*t0), r0*(1-0.8*t1), b0, b1))
			if i >= 2 && i%2 == 1 {
				r := r0 * (1 - 0.8*t0)
				suckers = append(suckers, ell(lerp3(pts[i], pts[i+1], 0.5).add(v3{0, -r * 0.8, 0}), v3{r * 0.35, r * 0.35, r * 0.35}, one(b0)))
			}
		}
	}
	// an octopus head, like an old sea god: a bulbous mantle swelling up and back, a knobbly
	// crown, heavy brow ridges over deep-set red eyes, finned frills at the sides and a beard of
	// curling face tentacles hanging over the chest
	headTex := gritty(wrinkleRelief(3, 711), palette{rgb{96, 126, 82}, rgb{52, 74, 50}, rgb{150, 170, 110}, rgb{130, 110, 64}}, 0.45, 0.35, 711)
	var head, frills []shape
	head = append(head,
		ell(hd(0, 0.0, 0.03), v3{0.074, 0.078, 0.08}, one(7)),    // cranium
		ell(hd(0, 0.04, -0.015), v3{0.056, 0.048, 0.05}, one(7)), // face
		ell(hd(0, 0.055, -0.05), v3{0.05, 0.04, 0.03}, one(7)),   // mouth base the tentacles hang from
	)
	for i := 0; i <= 10; i++ { // a smooth mantle swelling up and back to a rounded tip
		t := float64(i) / 10
		r := 0.074 - 0.026*t*t
		head = append(head, ell(lerp3(hd(0, -0.01, 0.05), hd(0, -0.12, 0.18), t), v3{r * 0.95, r, r}, one(7)))
	}
	for _, sx := range []float64{-1, 1} {
		head = append(head,
			capsl(hd(sx*0.008, 0.08, 0.026), hd(sx*0.058, 0.052, 0.05), 0.013, 0.008, one(7)), // heavy brow, angled down to the centre
			ell(hd(sx*0.042, 0.055, -0.005), v3{0.02, 0.024, 0.018}, one(7)),                  // cheek
		)
		// finned frills at the sides of the head
		frills = append(frills, rayFin(hd(sx*0.065, 0.0, 0.0), []v3{
			hd(sx*0.13, -0.02, 0.07), hd(sx*0.145, -0.04, 0.02), hd(sx*0.14, -0.05, -0.03), hd(sx*0.115, -0.04, -0.07)}, 0.008, 0.004, one(7))...)
	}
	// knobbly crown: rows of bumps from the forehead over the dome
	for _, x := range []float64{-0.032, 0, 0.032} {
		for k := 0; k < 4; k++ {
			t := float64(k) / 3
			c := lerp3(hd(x*0.7, 0.075, 0.085), hd(x, -0.11, 0.245), t)
			r := 0.011 - 0.002*t
			if x == 0 {
				r += 0.002
			}
			head = append(head, ell(c, v3{r, r, r * 1.2}, one(7)))
		}
	}
	// curling face tentacles hanging from the mouth over the chest, outer ones longer
	for i, x := range []float64{-0.044, -0.026, -0.009, 0.009, 0.026, 0.044} {
		L := 0.85 + 0.5*math.Abs(x)/0.044
		sp := 1.6 + 0.6*math.Abs(x)/0.044 // fanning out as they hang
		pts := []v3{hd(x, 0.075, -0.05), hd(x*1.3, 0.1, -0.1*L), hd(x*sp, 0.13, -0.17*L), hd(x*sp*1.15, 0.15, -0.24*L)}
		curl := 1.0
		if i%2 == 1 {
			curl = -1
		}
		tentHair(pts, 0.012, []int{7, 7, 7, 6}, curl)
	}
	// two shorter ones from the corners of the mouth
	for _, sx := range []float64{-1, 1} {
		tentHair([]v3{hd(sx*0.04, 0.06, -0.03), hd(sx*0.07, 0.08, -0.08), hd(sx*0.09, 0.09, -0.14)}, 0.012, []int{7, 7, 7}, -sx)
	}
	// arms, each in their own part so the hands never stick to the body
	for _, c := range [][3]int{{9, 10, 11}, {13, 14, 15}} {
		p0, p1, p2 := b[c[0]], b[c[1]], b[c[2]]
		arms = append(arms,
			limb(v3{p0[0] * 0.5, 0.02, 1.38}, p0, 0.045, 0.04, 5, c[0]),
			limb(p0, p1, 0.04, 0.032, c[0], c[1]),
			limb(p1, p2, 0.032, 0.024, c[1], c[2]),
			ell(p2, v3{0.03, 0.034, 0.038}, one(c[2])),
		)
	}
	// full armour over the upper body, in dark sea-steel with tarnished bronze edges: a sculpted
	// cuirass with a keel down the front, segmented plates over the belly, a gorget round the
	// neck, layered scallop-shell pauldrons, plated arms with elbow cops, gauntlets, and
	// overlapping faulds over the hips, the tentacles coming out from under them
	var dress, trim, gaunt, snail, plates []shape
	const e = 0.016                                    // the armour sits this far outside the skin
	ring := func(z, rx, ry, cy, r float64, bone int) { // a bronze edge round the body
		for k := 0; k < 32; k++ {
			a0, a1 := float64(k)/32*2*math.Pi, float64(k+1)/32*2*math.Pi
			trim = append(trim, capsl(v3{rx * math.Sin(a0), cy + ry*math.Cos(a0), z}, v3{rx * math.Sin(a1), cy + ry*math.Cos(a1), z}, r, r, one(bone)))
		}
	}
	torso := func(t float64) (rx, ry float64) {
		if t < 0.4 {
			u := t / 0.4
			return 0.13*W*(1-u) + 0.112*W*u, 0.105*(1-u) + 0.08*u
		}
		u := (t - 0.4) / 0.6
		return 0.112*W*(1-u) + 0.14*W*u, 0.08*(1-u) + 0.07*u
	}
	for i := 3; i <= 12; i++ { // the cuirass, from the waist up to the chest
		t := float64(i) / 12
		z := 0.95 + 0.38*t
		rx, ry := torso(t)
		bn := 4
		if z > 1.17 {
			bn = 5
		}
		dress = append(dress, ell(v3{0, 0.01, z}, v3{rx + e, ry + e, 0.06}, one(bn)))
	}
	dress = append(dress,
		ell(v3{0, 0.04, 1.29}, v3{0.115, 0.07, 0.075}, one(5)),                             // sculpted chest
		ell(v3{0, 0.01, 1.36}, v3{0.14*W + e, 0.07 + e, 0.055 + e}, one(5)),                // upper chest and back
		capsl(v3{-0.13 * W, 0.02, 1.4}, v3{0.13 * W, 0.02, 1.4}, 0.045+e, 0.045+e, one(5)), // over the shoulders
		limb(v3{0, 0.02, 1.39}, v3{0, 0.02, 1.47}, 0.05+e, 0.04+e, 5, 5),                   // gorget
	)
	ring(1.465, 0.05, 0.05, 0.02, 0.006, 5)
	trim = append(trim, capsl(v3{0, 0.112, 1.13}, v3{0, 0.112, 1.27}, 0.007, 0.007, one(5))) // bronze keel down the front

	for k := 0; k < 3; k++ { // segmented belly plates, each a little proud of the one above
		z := 1.0 + 0.05*float64(k)
		rx, ry := torso((z - 0.95) / 0.38)
		dress = append(dress, ell(v3{0, 0.01, z}, v3{rx + e + 0.006, ry + e + 0.006, 0.026}, one(4)))
	}
	// hip armour, open at the sides like a warrior's tassets: a belt, wide plates front and back
	// narrowing to a point, and on each hip three overlapping shield-shaped
	// plates, each a little further out and lower than the one above, riveted in bronze
	dress = append(dress, ell(v3{0, 0.01, 0.965}, v3{0.13*W + e + 0.006, 0.105 + e + 0.006, 0.026}, one(4)))
	ring(0.945, 0.13*W+e+0.004, 0.105+e+0.004, 0.01, 0.006, 4)
	// front and back: a wide plate curving round from hip to hip, narrowing to a sharp point low down,
	// edged in bronze along its two slanted sides
	const fTop, fTip = 0.965, 0.7
	for _, fy := range []float64{1, -1} {
		fPtR := func(z, ang, dr float64) v3 {
			R := 0.13*W + e + 0.012 + (fTop-z)*0.3 + dr
			return v3{R * math.Sin(ang), 0.01 + fy*R*0.9*math.Cos(ang), z}
		}
		fPt := func(z, ang float64) v3 { return fPtR(z, ang, 0) }
		for j := 0; j <= 16; j++ {
			f := float64(j) / 16
			z := fTop - (fTop-fTip)*f
			amax := 1.45 * math.Pow(1-f, 1.15) // from hip to hip at the top, to a point at the bottom
			if j == 16 {
				plates = append(plates, ell(fPt(z, 0), v3{0.009, 0.009, 0.009}, one(4))) // the point
				break
			}
			const ns = 7
			for i := 0; i < ns; i++ {
				a0 := -amax + 2*amax*float64(i)/ns
				a1 := -amax + 2*amax*float64(i+1)/ns
				plates = append(plates, capsl(fPt(z, a0), fPt(z, a1), 0.011, 0.011, one(4)))
			}
			if j%2 == 0 && j < 16 {
				f2 := float64(j+2) / 16
				z2 := fTop - (fTop-fTip)*f2
				a2 := 1.45 * math.Pow(1-f2, 1.15)
				for _, sg := range []float64{-1, 1} {
					trim = append(trim, capsl(fPtR(z, sg*amax, 0.006), fPtR(z2, sg*a2, 0.006), 0.007, 0.007, one(4)))
				}
			}
		}

		trim = append(trim, ell(fPt(0.93, 0).add(v3{0, fy * 0.008, 0}), v3{0.014, 0.008, 0.014}, one(4))) // bronze boss
	}
	for _, sx := range []float64{-1, 1} {
		for k := 0; k < 3; k++ {
			kf := float64(k)
			top := v3{sx * (0.155 + 0.014*kf), 0.01, 0.955 - 0.055*kf}
			bot := top.add(v3{sx * (0.06 + 0.01*kf), 0, -0.15})
			dress = append(dress, sheet(top, bot, v3{0, 1, 0}, 0.085-0.008*kf, 0.011, one(4)))
			trim = append(trim, ell(top.add(v3{sx * 0.012, 0, -0.02}), v3{0.008, 0.008, 0.008}, one(4))) // rivet
		}
	}
	for _, c := range [][3]int{{9, 10, 11}, {13, 14, 15}} { // plated arms, skinned like the arms
		p0, p1, p2 := b[c[0]], b[c[1]], b[c[2]]
		sx := math.Copysign(1, p0[0])
		dress = append(dress,
			limb(v3{p0[0] * 0.5, 0.02, 1.38}, p0, 0.045+e, 0.04+e, 5, c[0]),
			limb(p0, p1, 0.04+e, 0.032+e, c[0], c[1]),                 // upper arm
			ell(p1, v3{0.04, 0.04, 0.04}.add(v3{e, e, e}), one(c[1])), // elbow cop
			limb(p1, p2, 0.032+e, 0.024+e, c[1], c[2]),                // vambrace
		)
		// dark gauntlets with sharp claws, and spikes along the outside of the forearm
		hd2 := p2.sub(p1).norm()
		side := hd2.cross(v3{0, 0, 1}).norm()
		gaunt = append(gaunt,
			ell(p2, v3{0.032 + e, 0.036 + e, 0.04 + e}, one(c[2])),
			limb(lerp3(p1, p2, 0.6), p2, 0.024+e+0.005, 0.03+e, c[2], c[2]), // flared cuff
		)
		for k := 0; k < 4; k++ {
			o := side.mul((float64(k) - 1.5) * 0.017)
			base := p2.add(hd2.mul(0.035)).add(o)
			mid := base.add(hd2.mul(0.045)).add(o.mul(0.3))
			tip := mid.add(hd2.mul(0.04)).add(v3{0, 0, -0.025})
			gaunt = append(gaunt, capsl(base, mid, 0.01, 0.006, one(c[2])), capsl(mid, tip, 0.006, 0.0012, one(c[2])))
		}
		out := v3{sx, 0, 0.35}.norm()
		for i, t := range []float64{0.3, 0.52, 0.74} {
			bn := c[1]
			if i == 2 {
				bn = c[2]
			}
			r := 0.032*(1-t) + 0.024*t + e
			base := lerp3(p1, p2, t).add(out.mul(r * 0.7))
			gaunt = append(gaunt, capsl(base, base.add(out.mul(0.045)).sub(hd2.mul(0.035)), 0.012, 0.0015, one(bn)))
		}
		top := p0.add(v3{sx * 0.012, 0, 0.03})
		if c[0] == 9 {
			// left shoulder: a big spiral sea-snail shell, its spire pointing up and out
			ax := v3{sx * 0.55, -0.1, 0.83}.norm()
			u := ax.cross(v3{0, 1, 0}).norm()
			v := ax.cross(u).norm()
			const nw = 34
			for i := 0; i < nw; i++ {
				t := float64(i) / (nw - 1)
				a := t * 5.5 * math.Pi
				r := 0.084*math.Pow(1-t, 1.6) + 0.005
				cc := top.add(ax.mul(-0.01 + 0.2*math.Pow(t, 0.9))).add(u.mul(math.Cos(a) * r * 0.55)).add(v.mul(math.Sin(a) * r * 0.55))
				snail = append(snail, ell(cc, v3{r, r, r}, one(c[0])))
			}
			// the flared lip of the opening, facing forward
			snail = append(snail, sheet(top.add(v3{sx * 0.02, 0.07, -0.03}), top.add(v3{sx * 0.07, 0.07, 0.03}), v3{0, 0, 1}, 0.045, 0.01, one(c[0])))
			continue
		}
		// right shoulder: a layered scallop-shell pauldron, a little smaller, with bronze ribs
		const ps = 0.85
		R := v3{0.092, 0.088, 0.058}.mul(ps)
		dress = append(dress,
			ell(top, R, one(c[0])),
			ell(lerp3(p0, p1, 0.2).add(v3{sx * 0.01, 0, 0.004}), v3{0.074, 0.074, 0.032}.mul(ps), one(c[0])),
			ell(lerp3(p0, p1, 0.34).add(v3{sx * 0.008, 0, -0.004}), v3{0.062, 0.062, 0.028}.mul(ps), one(c[0])),
		)
		for k := 0; k < 7; k++ {
			a := (float64(k) - 3) * 0.38
			var pts []v3
			for j := 0; j <= 4; j++ {
				el := 1.45 - 1.5*float64(j)/4
				uu := v3{sx * math.Cos(a) * math.Cos(el), math.Sin(a) * math.Cos(el), math.Sin(el)}
				pts = append(pts, top.add(v3{R[0] * uu[0], R[1] * uu[1], R[2] * uu[2]}.mul(1.02)))
			}
			for j := 0; j+1 < len(pts); j++ {
				trim = append(trim, capsl(pts[j], pts[j+1], 0.006, 0.006, one(c[0])))
			}
		}
	}
	// a low crest of fins rising behind the head from the back of the gorget
	for k := -2; k <= 2; k++ {
		x := float64(k) * 0.028
		dress = append(dress, capsl(v3{x, -0.05, 1.42}, v3{x * 2.2, -0.095 - 0.01*math.Abs(float64(k)), 1.58 - 0.03*math.Abs(float64(k))}, 0.014, 0.003, one(5)))
	}
	// shell brooch between the breasts
	var brooch []shape
	bc := v3{0, 0.118, 1.3}
	brooch = append(brooch, ell(bc, v3{0.028, 0.012, 0.024}, one(5)))
	for k := 0; k < 5; k++ {
		a := (30 + 30*float64(k)) * math.Pi / 180
		brooch = append(brooch, capsl(bc.add(v3{0, 0.008, -0.018}), bc.add(v3{0.028 * math.Cos(a), 0.006, -0.018 + 0.04*math.Sin(a)}), 0.005, 0.003, one(5)))
	}
	// the octopus half: a mantle at the hips and eight tentacles that reach out over the ground
	// and curl up at the tips
	tentacles = append(tentacles, ell(v3{0, 0.0, 0.84}, v3{0.15, 0.14, 0.15}, one(4)))
	for k := 0; k < 8; k++ {
		a := (float64(k) + 0.5) / 8 * 2 * math.Pi // around, starting front-right
		d := v3{math.Sin(a), math.Cos(a), 0}
		perp := v3{d[1], -d[0], 0}
		if k%2 == 1 {
			perp = perp.mul(-1)
		}
		R := 0.85 + 0.12*float64(k%3)
		pts := []v3{
			d.mul(0.06).add(v3{0, 0, 0.84}),
			d.mul(0.2).add(v3{0, 0, 0.58}),
			d.mul(0.36).add(v3{0, 0, 0.26}),
			d.mul(0.55 * R).add(v3{0, 0, 0.08}),
			d.mul(0.8 * R).add(perp.mul(0.05)).add(v3{0, 0, 0.07}),
			d.mul(0.98 * R).add(perp.mul(0.14)).add(v3{0, 0, 0.13}),
			d.mul(1.0 * R).add(perp.mul(0.24)).add(v3{0, 0, 0.25}),
			d.mul(0.92 * R).add(perp.mul(0.28)).add(v3{0, 0, 0.34}),
			d.mul(0.85 * R).add(perp.mul(0.23)).add(v3{0, 0, 0.36}),
		}
		rad := []float64{0.075, 0.068, 0.058, 0.048, 0.04, 0.032, 0.025, 0.019, 0.014}
		// the two front tentacles carry the leg chains: they wriggle
		var chain []int
		if k == 0 {
			chain = []int{4, 1, 2, 3, 3}
		} else if k == 7 {
			chain = []int{4, 18, 19, 20, 20}
		}
		bone := func(i int) int {
			if chain == nil { // the hips, the root, then the tentacle's own swaying bone
				switch i {
				case 0:
					return 4
				case 1:
					return 0
				}
				return 20 + k // added bones 21..26 for tentacles 1..6
			}
			return chain[min(i, len(chain)-1)]
		}
		for i := 0; i+1 < len(pts); i++ {
			tentacles = append(tentacles, limb(pts[i], pts[i+1], rad[i], rad[i+1], bone(i), bone(i+1)))
		}
		// leaf-shaped tip
		tip := pts[len(pts)-1]
		td := tip.sub(pts[len(pts)-2]).norm()
		tentacles = append(tentacles, sheet(tip, tip.add(td.mul(0.1)), v3{0, 0, 1}, 0.04, 0.008, one(bone(len(pts)))))
		// suckers on the outer side of the reaching part
		for i := 2; i < 6; i++ {
			for _, f := range []float64{0.3, 0.7} {
				p := lerp3(pts[i], pts[i+1], f)
				r := rad[i]*(1-f) + rad[i+1]*f
				out := perp.mul(-r * 0.85).add(v3{0, 0, -r * 0.2})
				suckers = append(suckers, ell(p.add(out), v3{r * 0.42, r * 0.42, r * 0.42}, one(bone(i))))
			}
		}
	}
	// staff: a twisted coral shaft with gold bands and a gold butt cap, ending in a cup of
	// shell prongs that holds the pearl
	hand := b[15]
	dir := b[16].sub(b[17]).norm()
	top, bot := hand.add(dir.mul(0.62)), hand.sub(dir.mul(0.62))
	side := dir.cross(v3{0, 0, 1}).norm()
	up2 := side.cross(dir).norm()
	var gold []shape
	for i := 0; i < 24; i++ { // two strands twisted around each other
		t0, t1 := float64(i)/24, float64(i+1)/24
		for _, ph := range []float64{0, math.Pi} {
			p := func(t float64) v3 {
				a := t*7*math.Pi + ph
				return lerp3(bot, top, t).add(side.mul(0.014 * math.Cos(a))).add(up2.mul(0.014 * math.Sin(a)))
			}
			staff = append(staff, capsl(p(t0), p(t1), 0.02, 0.02, one(15)))
		}
	}
	for _, t := range []float64{0.06, 0.45, 0.62, 0.9} {
		c := lerp3(bot, top, t)
		gold = append(gold, capsl(c.sub(dir.mul(0.012)), c.add(dir.mul(0.012)), 0.033, 0.033, one(15)))
	}
	gold = append(gold, capsl(bot.sub(dir.mul(0.04)), bot.add(dir.mul(0.02)), 0.022, 0.03, one(15)))
	// a spiral conch shell on top: whorls winding to a pointed spire, knobs on the shoulder of the
	// big whorl and a flared lip, set in a gold collar
	var shell []shape
	cup := top.add(dir.mul(0.02))
	gold = append(gold, capsl(cup.sub(dir.mul(0.02)), cup.add(dir.mul(0.01)), 0.03, 0.036, one(15)))
	const nw = 28
	for i := 0; i < nw; i++ {
		t := float64(i) / (nw - 1)
		a := t * 4.5 * math.Pi
		r := 0.056*math.Pow(1-t, 1.3) + 0.005
		o := side.mul(math.Cos(a)).add(up2.mul(math.Sin(a))).mul(r * 0.45)
		c := cup.add(dir.mul(0.05 + 0.2*math.Pow(t, 0.8))).add(o)
		shell = append(shell, ell(c, v3{r, r, r}, one(15)))
		if i < 12 && i%2 == 0 { // knobs on the big whorl
			k := side.mul(math.Cos(a + 1.5)).add(up2.mul(math.Sin(a + 1.5)))
			shell = append(shell, capsl(c.add(k.mul(r*0.7)), c.add(k.mul(r*1.35)).add(dir.mul(0.02)), r*0.25, 0.004, one(15)))
		}
	}
	// flared lip of the opening
	shell = append(shell, sheet(cup.add(dir.mul(0.02)).add(side.mul(0.06)), cup.add(dir.mul(0.14)).add(side.mul(0.07)), up2, 0.035, 0.008, one(15)))
	staff = append(staff, gold...)
	conch := matStreak(palette{rgb{238, 212, 184}, rgb{176, 112, 92}, rgb{255, 242, 226}, rgb{226, 158, 136}}, 712)
	armourTex := gritty(armourRelief(8, 40, 713), palette{rgb{172, 178, 186}, rgb{92, 98, 108}, rgb{232, 236, 240}, rgb{142, 150, 162}}, 0.9, 0.8, 713)
	return []*bodyDef{
		boxPart("body", 2600, 0.006, 0.03, skin, body...),
		boxPart("head", 1400, 0.004, 0.012, headTex, head...),
		boxPart("frills", 400, 0.003, 0.004, tent, frills...),
		boxPart("arms", 700, 0.008, 0.02, skin, arms...),
		boxPart("facetentacles", 1600, 0.004, 0.01, headTex, hair...),
		boxPart("armour", 2800, 0.005, 0.012, armourTex, dress...),
		boxPart("hipplates", 900, 0.004, 0.006, armourTex, plates...),
		boxPart("trim", 900, 0.003, 0, goldTex, trim...),
		boxPart("gauntlets", 700, 0.003, 0.004, gritty(wrinkleRelief(5, 715), palette{rgb{40, 46, 60}, rgb{12, 14, 20}, rgb{110, 128, 156}, rgb{34, 50, 72}}, 0.12, 0.95, 715), gaunt...),
		boxPart("snail", 900, 0.004, 0.006, matStreak(palette{rgb{226, 206, 170}, rgb{120, 84, 56}, rgb{250, 236, 210}, rgb{180, 128, 88}}, 716), snail...),
		boxPart("brooch", 200, 0.003, 0.002, shellTex, brooch...),
		boxPart("tentacles", 3200, 0.012, 0.04, headTex, tentacles...),
		boxPart("suckers", 900, 0.008, 0, matSkin(palette{rgb{176, 186, 140}, rgb{110, 124, 86}, rgb{214, 220, 180}, rgb{160, 150, 110}}, 4, 718), suckers...),
		boxPart("staff", 1400, 0.004, 0.006, goldTex, staff...),
		boxPart("shell", 900, 0.004, 0.006, conch, shell...),
		eyesMulti([]v3{hd(0.03, 0.066, 0.018), hd(-0.03, 0.066, 0.018)}, 0.015, []int{7, 7}, monsterEye(rgb{255, 70, 20}, rgb{120, 10, 0}, false, 0.6)),
	}
}
