package main

import "math"

// Abyss Guardian: a three-headed sea serpent. No legs: a long, thick serpent body lies on the
// ground and slithers, the chest rises from it and three long necks grow out of the chest.
// Built on the Jabberocky's skeleton (stance pose). Its twelve leg bones (2-13) are re-used as a
// chain along the ground body, which ignores the body's up-and-down motion and weaves from side
// to side; each side neck gets its own neck and head bones, so each head can bob on its own.
// Bones: 1 chest, 16 middle neck, 17 middle head; added 19/20 side necks, 21/22 side heads.
var abyssChain = []int{2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12, 13}

// the ground body: a gentle S from under the chest to the tip of the tail
func abyssPath() ([]v3, []float64) {
	var pts []v3
	var rad []float64
	n := len(abyssChain)
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		r := 1.05*(1-t) + 0.12*t
		pts = append(pts, v3{0, 0.3 - 8.4*t, r * 0.9}) // straight at rest: the slither swings it evenly to both sides
		rad = append(rad, r)
	}
	return pts, rad
}

var abyssRebone, abyssWave = func() ([]reBone, []waveBone) {
	pts, _ := abyssPath()
	var rb []reBone
	var wv []waveBone
	prev, parent := v3{}, 0
	var mv []waveBone
	for k, bn := range abyssChain {
		rb = append(rb, reBone{bone: bn, parent: parent, local: pts[k].sub(prev)})
		prev, parent = pts[k], bn
		// the slither: a wave running down the body (the first bone is held straight)
		amp := 0.15 + 0.006*float64(k) // a constant, clearly visible side-to-side slither
		if k == 0 {
			amp = 0
		}
		wv = append(wv, waveBone{bone: bn, axis: v3{0, 0, 1}, amp: amp, phase: -0.75 * float64(k), speed: 0.7})
		// when moving: a strong serpentine wave runs down the whole body and tail
		mamp := 0.24 + 0.012*float64(k)
		if k == 0 {
			mamp = 0.12
		}
		mv = append(mv, waveBone{bone: bn, axis: v3{0, 0, 1}, amp: mamp, phase: -0.75 * float64(k), speed: 1.6, moveOnly: true})
	}
	wv = append(wv,
		waveBone{bone: 0, axis: v3{0, 0, 1}, amp: 0}, // the whole creature stays upright
		// necks sway slowly from side to side, heads nod a little, in every animation...
		waveBone{bone: 16, axis: v3{0, 0, 1}, amp: 0.07},
		waveBone{bone: 19, axis: v3{0, 0, 1}, amp: 0.07, phase: 2},
		waveBone{bone: 20, axis: v3{0, 0, 1}, amp: 0.07, phase: 4},
		waveBone{bone: 17, axis: v3{1, 0, 0}, amp: 0.06, phase: 1.3},
		waveBone{bone: 21, axis: v3{1, 0, 0}, amp: 0.06, phase: 2.4},
		waveBone{bone: 22, axis: v3{1, 0, 0}, amp: 0.06, phase: 0.4},
		// ...and when standing, each head in turn slowly dips down to the middle of the body and
		// back, one about every four seconds (a cycle of about twelve to fourteen seconds)
		waveBone{bone: 16, axis: v3{1, 0, 0}, amp: -0.85, pow: 4, idleOnly: true, period: 13.7, phase: 0},
		waveBone{bone: 19, axis: v3{1, 0, 0}, amp: -0.85, pow: 4, idleOnly: true, period: 13.7, phase: -2 * math.Pi / 3},
		waveBone{bone: 20, axis: v3{1, 0, 0}, amp: -0.85, pow: 4, idleOnly: true, period: 13.7, phase: -4 * math.Pi / 3},
	)
	return rb, append(wv, mv...)
}()

// side necks and heads: copies of the middle neck (16) and head (17)
var abyssExtra = []extraBone{{src: 16, parent: 1}, {src: 16, parent: 1}, {src: 17, parent: 19}, {src: 17, parent: 20}}

func abyssDesign(b []v3) []*bodyDef {
	base := matReptile(palette{rgb{58, 70, 46}, rgb{18, 22, 14}, rgb{130, 150, 100}, rgb{90, 96, 60}}, 40, 60, 1001)
	skin := withStripes(base, rgb{196, 206, 150}, 9, 1005)
	frill := matFin(palette{rgb{120, 132, 96}, rgb{40, 46, 30}, rgb{200, 210, 160}, rgb{90, 80, 60}}, 9, 1002)
	spine := matStreak(palette{rgb{60, 62, 50}, rgb{20, 20, 16}, rgb{150, 150, 120}, rgb{80, 70, 50}}, 1006)
	gums := matSkin(palette{rgb{120, 30, 36}, rgb{50, 8, 12}, rgb{180, 70, 70}, rgb{90, 20, 30}}, 6, 1004)
	tooth := matStreak(palette{rgb{226, 216, 190}, rgb{130, 110, 80}, rgb{250, 245, 230}, rgb{180, 160, 120}}, 1003)
	var body, heads, crests, spines, teeth []shape
	// the ground body, each piece skinned smoothly from one chain bone to the next
	_, rad := abyssPath()
	chain := abyssChain
	for i := 0; i+1 < len(chain); i++ {
		b0, b1 := chain[i], chain[i+1]
		body = append(body, limb(b[b0], b[b1], rad[i], rad[i+1], b0, b1))
	}
	tail := b[chain[len(chain)-1]]
	body = append(body, limb(tail, tail.add(tail.sub(b[chain[len(chain)-2]]).mul(0.8)), rad[len(rad)-1], 0.03, chain[len(chain)-1], chain[len(chain)-1]))
	// the chest rising out of the ground body
	body = append(body,
		// a lean, upright chest leaning forward to the necks (no hump at the back)
		ell(v3{0, 0.85, 2.75}, v3{0.95, 0.72, 0.85}, one(1)),
		limb(v3{0, 0.75, 2.3}, b[chain[0]].add(v3{0, 0.15, 0.1}), 0.95, 1.0, 1, chain[0]),
	)
	// back spines along the chest and the ground body
	for i := 1; i < len(chain)-1; i += 2 {
		p := b[chain[i]]
		dir := b[chain[i+1]].sub(b[chain[i-1]]).norm()
		r := rad[i]
		spines = append(spines, rayFin(p.add(v3{0, 0, r*0.6}), []v3{p.add(v3{0, 0, r + 0.45*r}).sub(dir.mul(0.3 * r)), p.add(v3{0, 0, r + 0.25*r}).sub(dir.mul(0.9 * r))}, 0.11*r+0.03, 0.035, one(chain[i]))...)
	}
	// three necks and heads, growing out of the chest
	type hd struct {
		base, mid, head v3
		nb, hb          int
	}
	hs := []hd{
		{v3{0, 1.1, 3.0}, b[16].add(v3{0, 0.2, 0.7}), v3{0, 2.5, 5.7}, 16, 17},
		{v3{0.75, 1.0, 2.9}, v3{1.75, 1.6, 4.6}, v3{2.15, 2.3, 5.1}, 19, 21},
		{v3{-0.75, 1.0, 2.9}, v3{-1.75, 1.6, 4.6}, v3{-2.15, 2.3, 5.1}, 20, 22},
	}
	var eyeC []v3
	var eyeB []int
	var jaws, mouth []shape
	for i, h := range hs {
		s := 1.0
		if i == 0 {
			s = 1.15
		}
		mid := h.base.add(h.mid.sub(h.base).mul(1.5))
		h0 := h.base.add(h.head.sub(h.base).mul(1.5))
		hb := []bw{{h.hb, 1}}
		at := func(x, y, z float64) v3 { return h0.add(v3{x * s, y * s, z * s}) }
		heads = append(heads,
			limb(h.base, mid, 0.62*s, 0.44*s, 1, h.nb),
			limb(mid, h0, 0.44*s, 0.38*s, h.nb, h.hb),
			ell(at(0, 0, 0.05), v3{0.5 * s, 0.55 * s, 0.42 * s}, hb),        // skull
			along(at(0, 0.2, 0), at(0, 1.25, -0.12), 0.38*s, 0.26*s, hb),     // upper jaw / snout
			ell(at(0.14, 1.22, 0.02), v3{0.08 * s, 0.07 * s, 0.07 * s}, hb), // nostrils
			ell(at(-0.14, 1.22, 0.02), v3{0.08 * s, 0.07 * s, 0.07 * s}, hb),
		)
		for _, sx := range []float64{-1, 1} {
			heads = append(heads,
				along(at(0.3*sx, 0.12, 0.3), at(0.27*sx, 0.78, 0.2), 0.15*s, 0.11*s, hb), // heavy brow
				ell(at(0.33*sx, 0.15, -0.15), v3{0.17 * s, 0.3 * s, 0.2 * s}, hb),       // cheek
			)
			teeth = append(teeth, capsl(at(0.3*sx, -0.1, 0.38), at(0.6*sx, -0.95, 0.8), 0.12*s, 0.02*s, hb)) // horn
			eyeC = append(eyeC, at(0.34*sx, 0.52, 0.12))
			eyeB = append(eyeB, h.hb)
			// sharp fin "ears": spiny rays fanning back from the side of the head
			crests = append(crests, rayFin(at(0.4*sx, -0.05, 0.05), []v3{
				at(0.75*sx, -0.55, 0.75), at(0.95*sx, -0.85, 0.4), at(1.0*sx, -1.0, 0.05), at(0.85*sx, -0.95, -0.3)}, 0.07*s, 0.035, hb)...)
			// rows of teeth, rooted deep inside the jaws (positions follow the jaw surfaces)
			for k := 0; k < 5; k++ {
				t := float64(k) / 4
				y := 0.42 + 0.66*t
				u := (y - 0.725) / 0.525
				q := math.Sqrt(math.Max(0.05, 1-u*u))
				cz := -0.06 * (y - 0.2) / 0.525
				bot := cz - 0.26*q
				up := at(0.38*q*0.72*sx, y, bot+0.09)
				teeth = append(teeth, capsl(up, at(0.38*q*0.72*sx, y+0.02, bot-0.13), 0.05*s, 0.008*s, hb))
				y2 := 0.35 + 0.6*t
				u2 := (y2 - 0.625) / 0.525
				q2 := math.Sqrt(math.Max(0.05, 1-u2*u2))
				cz2 := -0.35 - 0.37*(y2-0.1)/1.05
				top := cz2 + 0.11*q2
				lo := at(0.32*q2*0.7*sx, y2, top-0.07)
				teeth = append(teeth, capsl(lo, at(0.32*q2*0.7*sx, y2+0.02, top+0.12), 0.045*s, 0.008*s, hb))
			}
		}
		// spiky crest over the head and down the back of the neck
		crests = append(crests, rayFin(at(0, -0.25, 0.3), []v3{at(0, 0.3, 0.95), at(0, -0.15, 1.05), at(0, -0.6, 0.9), at(0, -0.95, 0.6)}, 0.07*s, 0.035, hb)...)
		nd := mid.sub(h.base).norm()
		back := nd.cross(v3{1, 0, 0}).norm().mul(-1)
		if back[1] > 0 {
			back = back.mul(-1)
		}
		for k := 1; k <= 3; k++ {
			p := lerp3(h.base, mid, 0.25*float64(k)+0.1)
			spines = append(spines, rayFin(p.add(back.mul(0.25*s)), []v3{p.add(back.mul(0.85 * s)).add(nd.mul(0.35 * s)), p.add(back.mul(0.75 * s)).sub(nd.mul(0.1 * s))}, 0.08*s, 0.03, []bw{{1, 0.3}, {h.nb, 0.7}})...)
		}
		// open lower jaw, with the dark mouth inside
		jaws = append(jaws, along(at(0, 0.1, -0.35), at(0, 1.15, -0.72), 0.32*s, 0.11*s, hb))
		mouth = append(mouth, along(at(0, 0.25, -0.3), at(0, 1.05, -0.44), 0.25*s, 0.12*s, hb))
	}
	return []*bodyDef{
		boxPart("body", 3600, 0.07, 0.3, skin, body...),
		boxPart("heads", 3400, 0.04, 0.18, skin, heads...),
		boxPart("jaws", 900, 0.035, 0.05, skin, jaws...),
		boxPart("mouth", 400, 0.035, 0.05, gums, mouth...),
		boxPart("crests", 1400, 0.025, 0.03, frill, crests...),
		boxPart("spines", 900, 0.03, 0.02, spine, spines...),
		boxPart("teeth", 1400, 0.022, 0, tooth, teeth...),
		eyesMulti(eyeC, 0.16, eyeB, monsterEye(rgb{250, 190, 30}, rgb{130, 40, 10}, true, 0.4)),
	}
}
