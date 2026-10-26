package main

import "math"

// Puffy: a grumpy, spiky pufferfish that floats along, on a hovering creature's skeleton
// (stance pose). Bones: 1 body, 2 mouth, 9 top; fins 3-4-5 (+x), 6-7-8 (-x); tail 10-11-12.
func puffyDesign(b []v3) []*bodyDef {
	skin := matSkin(palette{rgb{206, 186, 144}, rgb{120, 92, 60}, rgb{236, 224, 190}, rgb{150, 110, 70}}, 4, 801)
	back := matStreak(palette{rgb{150, 96, 56}, rgb{60, 36, 20}, rgb{210, 170, 120}, rgb{120, 90, 60}}, 802)
	fin := matFin(palette{rgb{220, 150, 80}, rgb{120, 60, 24}, rgb{250, 210, 150}, rgb{180, 120, 70}}, 9, 803)
	lip := matSkin(palette{rgb{190, 110, 96}, rgb{100, 50, 46}, rgb{230, 160, 140}, rgb{150, 90, 80}}, 6, 804)
	c := b[1].add(v3{0, 0.05, 0.0})
	R := 0.62
	var body, spikes, fins []shape
	body = append(body, ell(c, v3{R, R * 1.05, R * 0.95}, []bw{{1, 0.8}, {9, 0.2}}))
	// spikes all over (golden-angle spiral)
	n := 46
	for i := 0; i < n; i++ {
		z := 1 - 2*(float64(i)+0.5)/float64(n)
		r := math.Sqrt(1 - z*z)
		a := float64(i) * 2.39996
		d := v3{r * math.Cos(a), r * math.Sin(a), z}
		if d[1] > 0.75 { // keep the face clear
			continue
		}
		p := c.add(v3{d[0] * R, d[1] * R * 1.05, d[2] * R * 0.95})
		spikes = append(spikes, capsl(p.sub(d.mul(0.03)), p.add(d.mul(0.2)), 0.04, 0.006, one(1)))
	}
	// sharp fish fins: spiny rays with a membrane between them, rooted inside the body
	for _, ch := range [][3]int{{3, 4, 5}, {6, 7, 8}} {
		p0, p1 := b[ch[0]], b[ch[1]]
		root := c.add(p0.sub(c).norm().mul(R * 0.82))
		out := p1.add(p1.sub(p0).mul(0.7)).sub(root)
		L := out.len()
		o := out.norm()
		var tips []v3
		for k := 0; k < 5; k++ {
			a := (float64(k) - 2) * 0.3
			d := o.mul(math.Cos(a)).add(v3{0, 0, 1}.mul(math.Sin(a))).norm()
			tips = append(tips, root.add(d.mul(L*(1.15-0.12*math.Abs(float64(k)-2)))).add(v3{0, -0.12 * L, 0}))
		}
		fins = append(fins, rayFin(root, tips, 0.045, 0.02, []bw{{1, 0.5}, {ch[1], 0.5}})...)
	}
	t0, t1 := c.add(v3{0, -R * 0.95, 0}), b[11]
	body = append(body, capsl(t0, t1, 0.16, 0.09, one(10)))
	// forked, pointed tail fin
	var tt []v3
	for k := 0; k < 6; k++ {
		a := (float64(k) - 2.5) / 2.5
		l := 0.55 + 0.25*math.Abs(a)
		tt = append(tt, t1.add(v3{0, -l * (1 - 0.3*math.Abs(a)), 0.45 * a}))
	}
	fins = append(fins, rayFin(t1.add(v3{0, 0.08, 0}), tt, 0.05, 0.022, one(11))...)
	// spiny dorsal fin
	dr := c.add(v3{0, -0.25, R * 0.82})
	fins = append(fins, rayFin(dr, []v3{dr.add(v3{0, 0.05, 0.3}), dr.add(v3{0, -0.15, 0.32}), dr.add(v3{0, -0.33, 0.24}), dr.add(v3{0, -0.45, 0.1})}, 0.035, 0.016, one(9))...)
	// heavy brows and cheeks around the eyes, so the eyes sit in sockets
	eyeL, eyeR := c.add(v3{0.27, R * 0.8, 0.18}), c.add(v3{-0.27, R * 0.8, 0.18})
	for _, e := range []v3{eyeL, eyeR} {
		body = append(body,
			ell(e.add(v3{0, -0.04, 0.1}), v3{0.17, 0.1, 0.07}, []bw{{1, 0.8}, {9, 0.2}}),
			ell(e.add(v3{0, -0.05, -0.08}), v3{0.15, 0.09, 0.06}, []bw{{1, 0.8}, {9, 0.2}}),
		)
	}
	mouth := c.add(v3{0, R * 0.98, -0.14})
	lips := []shape{
		shape{kind: torusZ, c: mouth, r: v3{0.09, 0.05, 0}, bones: one(2)},
	}
	// rotate the lip ring to face forward: approximate with two lip ellipsoids instead
	lips = []shape{
		ell(mouth.add(v3{0, 0, 0.045}), v3{0.15, 0.1, 0.055}, []bw{{2, 0.6}, {1, 0.4}}),
		ell(mouth.add(v3{0, 0, -0.045}), v3{0.14, 0.1, 0.055}, []bw{{2, 0.6}, {1, 0.4}}),
	}
	return []*bodyDef{
		boxPart("body", 1600, 0.03, 0.08, skin, body...),
		boxPart("spikes", 1200, 0.016, 0, back, spikes...),
		boxPart("fins", 700, 0.012, 0.02, fin, fins...),
		boxPart("lips", 200, 0.012, 0.03, lip, lips...),
		eyesMulti([]v3{eyeL, eyeR}, 0.13, []int{1, 1}, monsterEye(rgb{60, 150, 120}, rgb{10, 40, 30}, false, 0.2)),
	}
}
