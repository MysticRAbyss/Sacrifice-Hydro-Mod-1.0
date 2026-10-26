package main

import "math"

// Leviathan: a long, S-shaped flying sea serpent with two pairs of the Silverback's wings
// (borrowed from the Silverback's own model at load time, made smaller; the rear pair hangs off
// four added bones that flap like the front pair), on the Silverback's skeleton (stance pose).
// Bones: 1 pelvis, 8 chest, 12 neck, 13 head, 14/15 horns, 16 jaw; front wings 20-21-22 (-x) and
// 23-24-25 (+x); added rear wings 26-27-28 and 29-30-31. The legs are not used: the body and the
// whole tail follow the chest and pelvis, so the tail never folds into the body.
// the unused leg and arm bones become a 12-bone tail chain behind the pelvis that weaves
var levTail = []int{2, 3, 4, 5, 6, 7, 9, 10, 11, 17, 18, 19}

func leviathanTailRig() ([]reBone, []waveBone) {
	const length, slope = 11.2, 0.12
	step := v3{0, -1, slope}.norm().mul(length / float64(len(levTail)))
	var rb []reBone
	var wv []waveBone
	parent := 1
	for k, bn := range levTail {
		rb = append(rb, reBone{bone: bn, parent: parent, local: step})
		// like a flying Chinese dragon: a slow wave rolls up and down along the body and tail,
		// with a gentle side-to-side sway on every third bone
		ax, amp := v3{1, 0, 0}, 0.13+0.005*float64(k)
		if k%3 == 2 {
			ax, amp = v3{0, 0, 1}, 0.1
		}
		wv = append(wv, waveBone{bone: bn, axis: ax, amp: amp, phase: -0.55 * float64(k), speed: 0.5})
		parent = bn
	}
	// a level body that still lives: the root is held still, so the Leviathan flies (and rises
	// when summoned) horizontally; the pelvis rolls gently up and down, the chest and the
	// straight neck sway a little
	wv = append(wv,
		waveBone{bone: 0, axis: v3{0, 0, 1}, amp: 0},
		waveBone{bone: 1, axis: v3{1, 0, 0}, amp: 0.12, phase: 0.55, speed: 0.5}, // the body rises and dips
		waveBone{bone: 8, axis: v3{1, 0, 0}, amp: 0.1, phase: 1.1, speed: 0.5},
		waveBone{bone: 12, axis: v3{0, 0, 1}, amp: 0.05, phase: 2.4, speed: 0.5},
	)
	return rb, wv
}

var leviathanRig = struct {
	extra     []extraBone
	boneScale map[int]float64
	boneShift map[int]v3
	borrow    []borrowSpec
}{
	extra: []extraBone{
		{src: 20, parent: 1, shift: v3{0.05, -3.0, -0.14}, frameShift: 0.5},
		{src: 21, parent: 26, scale: 0.85, frameShift: 0.5},
		{src: 22, parent: 27, scale: 0.85, frameShift: 0.5},
		{src: 23, parent: 1, shift: v3{-0.05, -3.0, -0.14}, frameShift: 0.5},
		{src: 24, parent: 29, scale: 0.85, frameShift: 0.5},
		{src: 25, parent: 30, scale: 0.85, frameShift: 0.5},
	},
	boneScale: map[int]float64{21: 0.65, 22: 0.65, 24: 0.65, 25: 0.65},
	boneShift: map[int]v3{20: {0.81, 0, -1.64}, 23: {-0.81, 0, -1.64}, 12: {0, 0.9, 0.15}, 13: {0, 0.55, -0.95}, 16: {0, -0.35, 0.5}},
	borrow: []borrowSpec{
		{part: 6, scale: 0.65, rebind: map[int]int{8: 23}},
		{part: 7, scale: 0.65, rebind: map[int]int{8: 20}},
		{part: 9, scale: 0.65},
		{part: 6, scale: 0.55, rebind: map[int]int{8: 23}, boneMap: map[int]int{23: 29, 24: 30, 25: 31}},
		{part: 7, scale: 0.55, rebind: map[int]int{8: 20}, boneMap: map[int]int{20: 26, 21: 27, 22: 28}},
		{part: 9, scale: 0.55, boneMap: map[int]int{20: 26, 21: 27, 22: 28, 23: 29, 24: 30, 25: 31}},
	},
}

var levRebone, levWave = leviathanTailRig()

func leviathanDesign(b []v3) []*bodyDef {
	skin := matScales(palette{rgb{52, 86, 120}, rgb{16, 28, 48}, rgb{150, 180, 210}, rgb{150, 90, 50}}, 60, 1101)
	fin := matFin(palette{rgb{170, 90, 50}, rgb{70, 30, 16}, rgb{230, 160, 100}, rgb{60, 110, 130}}, 10, 1102)
	gums := matSkin(palette{rgb{120, 30, 36}, rgb{50, 8, 12}, rgb{180, 70, 70}, rgb{90, 20, 30}}, 6, 1104)
	horn := matStreak(palette{rgb{200, 186, 150}, rgb{90, 74, 50}, rgb{240, 230, 200}, rgb{160, 130, 90}}, 1103)
	var body, head, fins, horns []shape
	h := b[13]
	// long, stiff neck from the chest to the head (the neck bone is held still)
	nh, nc := h.add(v3{0, -0.4, -0.15}), b[8].add(v3{0, 0.2, 0.05}) // a straight neck from the chest to the head
	neck := []v3{nh, lerp3(nh, nc, 0.3), lerp3(nh, nc, 0.65), nc}
	nr := []float64{0.6, 0.68, 0.82, 0.95}
	nbn := [][2]int{{13, 8}, {8, 8}, {8, 8}}
	for i := 0; i+1 < len(neck); i++ {
		body = append(body, limb(neck[i], neck[i+1], nr[i], nr[i+1], nbn[i][0], nbn[i][1]))
	}
	// serpent body and the long weaving tail: chest -> pelvis -> 12 tail bones, each piece
	// skinned smoothly from one bone to the next
	chain := append([]int{8, 1}, levTail...)
	var tail []v3
	var rad []float64
	for i, bn := range chain {
		t := float64(i) / float64(len(chain))
		tail = append(tail, b[bn])
		rad = append(rad, 0.95*(1-t)+0.1*t+0.12*math.Sin(t*math.Pi)*(1-t))
	}
	last := b[chain[len(chain)-1]].sub(b[chain[len(chain)-2]])
	tail = append(tail, b[chain[len(chain)-1]].add(last.mul(0.8)))
	rad = append(rad, 0.09)
	chain = append(chain, chain[len(chain)-1])
	nseg := len(tail) - 1
	for i := 0; i < nseg; i++ {
		body = append(body, limb(tail[i], tail[i+1], rad[i], rad[i+1], chain[i], chain[i+1]))
	}
	// shoulders rising to the wing roots, so the borrowed wings sit on the body
	for _, w := range []struct {
		root, bone, parent int
		c               v3
	}{{20, 20, 8, b[8]}, {23, 23, 8, b[8]}, {26, 26, 1, tail[3]}, {29, 29, 1, tail[3]}} { // tail[3]: under the rear wings
		r := b[w.root]
		base := v3{r[0] * 0.3, r[1], w.c[2] + 0.3}
		body = append(body, limb(base, r.add(v3{0, 0, -0.05}), 0.55, 0.4, w.parent, w.bone)) // wing roots sit on the body
	}
	// head: a sharp, wedge-shaped dragon head - long tapering snout, angular brows, cheek spikes,
	// swept-back horns and spiny fin ears; the lower jaw is hinged at the back of the skull and
	// skinned from the head to the jaw bone, so it opens without coming apart
	snout := h.add(v3{0, 1.55, -0.22})
	// the lower jaw is its own piece on the jaw bone (hinged at the back of the mouth), so it opens
	// cleanly; inside the mouth a dark palate, gums and a tongue
	var jaw, mouth, tongue []shape
	jaw = append(jaw,
		along(h.add(v3{0, -0.15, -0.36}), snout.add(v3{0, -0.08, -0.38}), 0.3, 0.1, one(16)),
		capsl(h.add(v3{0.24, -0.2, -0.3}), h.add(v3{0.2, 0.3, -0.38}), 0.09, 0.07, one(16)), // jaw sides
		capsl(h.add(v3{-0.24, -0.2, -0.3}), h.add(v3{-0.2, 0.3, -0.38}), 0.09, 0.07, one(16)),
	)
	mouth = append(mouth,
		along(h.add(v3{0, 0.2, -0.15}), snout.add(v3{0, -0.12, -0.15}), 0.24, 0.06, one(13)),   // palate, flush with the upper jaw
	)
	// the tongue is its own piece lying on the lower jaw, so it moves with the jaw only
	tongue = append(tongue,
		along(h.add(v3{0, 0.05, -0.27}), snout.add(v3{0, -0.3, -0.3}), 0.15, 0.04, one(16)),
		capsl(snout.add(v3{0, -0.3, -0.3}), snout.add(v3{0.07, -0.1, -0.31}), 0.035, 0.012, one(16)), // forked tip
		capsl(snout.add(v3{0, -0.3, -0.3}), snout.add(v3{-0.07, -0.1, -0.31}), 0.035, 0.012, one(16)),
	)
	head = append(head,
		along(h.add(v3{0, -0.25, 0.05}), h.add(v3{0, 0.55, 0.02}), 0.48, 0.42, one(13)), // skull
		along(h.add(v3{0, 0.3, 0.02}), snout, 0.34, 0.22, one(13)),                         // upper jaw
		capsl(h.add(v3{0, 0.2, 0.32}), snout.add(v3{0, -0.05, 0.12}), 0.1, 0.05, one(13)),   // nose ridge
	)
	for _, s := range []float64{-1, 1} {
		head = append(head,
			capsl(h.add(v3{s * 0.3, 0.15, 0.36}), h.add(v3{s * 0.2, 0.95, 0.16}), 0.13, 0.03, one(13)),  // angular brow
			capsl(h.add(v3{s * 0.38, 0.0, -0.08}), h.add(v3{s * 0.32, 0.8, -0.18}), 0.14, 0.05, one(13)), // cheekbone
		)
		horns = append(horns,
			capsl(h.add(v3{s * 0.3, -0.15, 0.4}), h.add(v3{s * 0.75, -1.3, 0.95}), 0.13, 0.015, one(13)),  // swept-back horn
			capsl(h.add(v3{s * 0.4, -0.1, -0.1}), h.add(v3{s * 0.9, -0.75, -0.15}), 0.08, 0.01, one(13)),  // cheek spike
			capsl(snout.add(v3{s * 0.22, -0.15, 0.0}), snout.add(v3{s * 1.0, -0.65, -0.5}), 0.04, 0.012, one(13)), // whiskers
			capsl(snout.add(v3{s * 1.0, -0.65, -0.5}), snout.add(v3{s * 1.3, -1.4, -0.4}), 0.012, 0.006, one(13)),
		)
		// fangs along the upper and lower jaws
		for k := 0; k < 4; k++ {
			t := float64(k) / 3
			up := lerp3(h.add(v3{s * 0.24, 0.55, -0.1}), snout.add(v3{s * 0.12, -0.15, -0.1}), t)
			horns = append(horns, capsl(up.add(v3{0, 0, 0.08}), up.add(v3{0, 0.02, -0.16}), 0.045, 0.006, one(13)))
			lo := lerp3(h.add(v3{s * 0.2, 0.5, -0.38}), snout.add(v3{s * 0.08, -0.2, -0.36}), t)
			jaw = append(jaw, capsl(lo.sub(v3{0, 0, 0.06}), lo.add(v3{0, 0.02, 0.15}), 0.04, 0.006, one(16)))
		}
		// spiny fin ears
		fins = append(fins, rayFin(h.add(v3{s * 0.42, -0.15, 0.1}), []v3{
			h.add(v3{s * 0.8, -0.6, 0.75}), h.add(v3{s * 1.05, -0.9, 0.4}), h.add(v3{s * 1.1, -1.05, 0.05}), h.add(v3{s * 0.9, -0.95, -0.25})}, 0.07, 0.035, one(13))...)
	}
	// dorsal crest along the neck, back and tail, rooted inside the body, and a tail fin
	ridge := []v3{h.add(v3{0, -0.3, 0.45}), neck[1].add(v3{0, 0, 0.6}), neck[2].add(v3{0, 0, 0.75})}
	rb := [][]bw{one(13), one(8), one(8)}
	for i := 1; i < nseg-1; i++ {
		ridge = append(ridge, tail[i].add(v3{0, 0, rad[i]*0.75}))
		rb = append(rb, one(chain[i]))
	}
	for i := 0; i+1 < len(ridge); i++ {
		mid := ridge[i].add(ridge[i+1]).mul(0.5)
		if math.Abs(mid[1]-b[20][1]) < 0.9 || math.Abs(mid[1]-b[26][1]) < 0.9 {
			continue // leave room for the wing roots
		}
		hw := 0.34 * math.Min(1, 1.6-float64(i)/float64(len(ridge)))
		fins = append(fins, sheet(ridge[i], ridge[i+1], v3{0, 0, 1}, hw, 0.06, rb[i+1]))
	}
	// pointed tail fluke: spiny rays with a membrane between them
	end, tb := tail[nseg], one(chain[nseg])
	dir := end.sub(tail[nseg-1]).norm()
	for _, r := range []v3{{0, -1.1, 1.6}, {0, -1.5, 1.0}, {0, -1.7, 0.4}, {0, -1.4, -0.15}} {
		tip := end.add(r.add(dir.mul(0.3)))
		fins = append(fins, capsl(end.sub(dir.mul(0.4)), tip, 0.16, 0.015, tb))
	}
	fins = append(fins, sheet(end.sub(dir.mul(0.3)), end.add(v3{0, -1.3, 0.75}), v3{0, 0, 1}, 0.65, 0.05, tb))
	return []*bodyDef{
		boxPart("body", 4200, 0.06, 0.25, skin, body...),
		boxPart("head", 1800, 0.035, 0.15, skin, head...),
		boxPart("jaw", 700, 0.03, 0.06, skin, jaw...),
		boxPart("mouth", 300, 0.03, 0.04, gums, mouth...),
		boxPart("tongue", 200, 0.02, 0.02, gums, tongue...),
		boxPart("fins", 1800, 0.03, 0.06, fin, fins...),
		boxPart("horns", 600, 0.02, 0.02, horn, horns...),
		eyesMulti([]v3{h.add(v3{0.36, 0.52, 0.2}), h.add(v3{-0.36, 0.52, 0.2})}, 0.12, []int{13, 13}, monsterEye(rgb{250, 210, 60}, rgb{140, 50, 10}, true, 0.45)),
	}
}
