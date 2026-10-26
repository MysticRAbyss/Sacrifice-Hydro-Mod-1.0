package main

// Tidekin: a kappa for the Hydro god. Built on the skeleton of a humanoid Sacrifice creature
// (the Fallen/Earthfling rig) so that it can use that rig's animations.

import "math"

// bone indices of the base rig (Fallen/Earthfling)
const (
	bPelvis = 1
	bSpine  = 2
	bChest  = 3
	bHead   = 4
	bJawA   = 5
	bJawB   = 6
	bClavR  = 7
	bUArmR  = 8
	bFArmR  = 9
	bClavL  = 10
	bUArmL  = 11
	bFArmL  = 12
	bThighR = 13
	bShinR  = 14
	bFootR  = 15
	bThighL = 16
	bShinL  = 17
	bFootL  = 18
)

func one(b int) []bw { return []bw{{b, 1}} }

var kappaSkin = matSkin(palette{rgb{72, 110, 62}, rgb{30, 52, 34}, rgb{140, 168, 100}, rgb{96, 84, 46}}, 10, 1200)

func kappaDesign(bw0 []v3) []*bodyDef {
	B := func(i int) v3 { return bw0[i] }
	sides := []struct {
		s                        float64
		clav, uarm, farm, th, sh, ft int
	}{{-1, bClavR, bUArmR, bFArmR, bThighR, bShinR, bFootR}, {1, bClavL, bUArmL, bFArmL, bThighL, bShinL, bFootL}}

	// --- body skin: torso, limbs, webbed hands and feet -----------------------------------
	var body, arms []shape
	body = append(body,
		shape{kind: ellipsoid, c: v3{0, 0.07, 0.07}, r: v3{0.19, 0.16, 0.15}, bones: one(bPelvis)},
		shape{kind: ellipsoid, c: v3{0, 0.07, 0.38}, r: v3{0.17, 0.15, 0.29}, bones: []bw{{bSpine, 0}, {bChest, 0}}, zBlend: &[3]float64{0.25, 0.55}},
		shape{kind: ellipsoid, c: v3{0, 0.07, 0.64}, r: v3{0.17, 0.13, 0.12}, bones: one(bChest)},
		shape{kind: capsule, a: v3{0, 0.09, 0.66}, b: v3{0, 0.14, 0.86}, ra: 0.10, rb: 0.09, bones: one(bChest)},
	)
	for _, sd := range sides {
		s := sd.s
		sh := v3{s * 0.21, 0.05, 0.47}
		elbow := v3{s * 0.24, 0.05, 0.21}
		wrist := v3{s * 0.24, 0.08, -0.04}
		arms = append(arms,
			shape{kind: ellipsoid, c: sh, r: v3{0.065, 0.065, 0.065}, bones: one(sd.clav)},
			shape{kind: capsule, a: sh, b: elbow, ra: 0.052, rb: 0.043, bones: one(sd.uarm)},
			shape{kind: capsule, a: elbow, b: wrist, ra: 0.043, rb: 0.036, bones: one(sd.farm)},
			// webbed paddle hand with three fingers
			shape{kind: ellipsoid, c: v3{s * 0.245, 0.10, -0.11}, r: v3{0.022, 0.065, 0.07}, bones: one(sd.farm)},
		)
		for f := -1; f <= 1; f++ {
			tip := v3{s * 0.245, 0.10 + float64(f)*0.06, -0.20 + math.Abs(float64(f))*0.02}
			arms = append(arms, shape{kind: capsule, a: v3{s * 0.245, 0.10 + float64(f)*0.025, -0.10}, b: tip, ra: 0.016, rb: 0.009, bones: one(sd.farm)})
		}
		hip := v3{s * 0.14, 0.07, -0.02}
		knee := v3{s * 0.21, 0.10, -0.31}
		ankle := v3{s * 0.22, 0.07, -0.60}
		body = append(body,
			shape{kind: capsule, a: hip, b: knee, ra: 0.09, rb: 0.062, bones: one(sd.th)},
			shape{kind: capsule, a: knee, b: ankle, ra: 0.06, rb: 0.044, bones: one(sd.sh)},
			// webbed foot with three toes
			shape{kind: ellipsoid, c: v3{s * 0.22, 0.14, -0.665}, r: v3{0.075, 0.13, 0.03}, bones: one(sd.ft)},
		)
		for f := -1; f <= 1; f++ {
			body = append(body, shape{kind: capsule, a: v3{s*0.22 + float64(f)*0.03, 0.18, -0.67}, b: v3{s*0.22 + float64(f)*0.07, 0.31, -0.685}, ra: 0.02, rb: 0.012, bones: one(sd.ft)})
		}
	}
	bodyLo, bodyHi := v3{-0.35, -0.25, -0.75}, v3{0.35, 0.40, 0.90}
	span := math.Max(bodyHi[0]-bodyLo[0], math.Max(bodyHi[1]-bodyLo[1], bodyHi[2]-bodyLo[2]))

	// --- head: round skull with brow ridge -------------------------------------------------
	head := []shape{
		{kind: ellipsoid, c: v3{0, 0.17, 0.93}, r: v3{0.145, 0.15, 0.135}, bones: one(bHead)},
		{kind: ellipsoid, c: v3{0, 0.25, 0.995}, r: v3{0.125, 0.07, 0.045}, bones: one(bHead)}, // brow
		{kind: ellipsoid, c: v3{0, 0.22, 0.86}, r: v3{0.11, 0.10, 0.07}, bones: one(bHead)}, // jaw/cheeks
	}
	headLo, headHi := v3{-0.2, 0.0, 0.75}, v3{0.2, 0.35, 1.10}

	// --- shell (carapace) on the back --------------------------------------------------------
	shell := []shape{
		{kind: ellipsoid, c: v3{0, 0.02, 0.38}, r: v3{0.29, 0.21, 0.37}, clip: true, clipN: v3{0, 1, 0}, clipD: 0.02,
			bones: []bw{{bSpine, 0}, {bChest, 0}}, zBlend: &[3]float64{0.25, 0.55}},
		{kind: ellipsoid, c: v3{0, 0.00, 0.12}, r: v3{0.30, 0.17, 0.10}, clip: true, clipN: v3{0, 1, 0}, clipD: 0.02, // rim flare
			bones: one(bSpine)},
	}
	shellLo, shellHi := v3{-0.32, -0.22, -0.02}, v3{0.32, 0.03, 0.76}

	// --- plastron (belly plate) --------------------------------------------------------------
	plastron := []shape{
		{kind: ellipsoid, c: v3{0, 0.145, 0.36}, r: v3{0.14, 0.095, 0.25}, clip: true, clipN: v3{0, -1, 0}, clipD: -0.15,
			bones: []bw{{bSpine, 0}, {bChest, 0}}, zBlend: &[3]float64{0.25, 0.55}},
	}
	plLo, plHi := v3{-0.15, 0.15, 0.10}, v3{0.15, 0.25, 0.62}

	// --- beak --------------------------------------------------------------------------------
	beak := []shape{
		{kind: ellipsoid, c: v3{0, 0.30, 0.895}, r: v3{0.075, 0.10, 0.035}, bones: one(bHead)},
		{kind: capsule, a: v3{0, 0.31, 0.90}, b: v3{0, 0.395, 0.875}, ra: 0.03, rb: 0.012, bones: one(bHead)},
		{kind: ellipsoid, c: v3{0, 0.29, 0.855}, r: v3{0.065, 0.085, 0.025}, bones: one(bHead)},
	}
	beakLo, beakHi := v3{-0.1, 0.18, 0.80}, v3{0.1, 0.42, 0.95}

	// --- sara: the water dish on top of the head ---------------------------------------------
	dish := []shape{
		{kind: discZ, c: v3{0, 0.15, 1.055}, r: v3{0.085, 0.016, 0}, bones: one(bHead),
			sub: []shape{{kind: ellipsoid, c: v3{0, 0.15, 1.105}, r: v3{0.075, 0.075, 0.045}}}},
	}
	dishLo, dishHi := v3{-0.09, 0.06, 1.0}, v3{0.09, 0.24, 1.08}

	// --- hair fringe around the dish -----------------------------------------------------------
	hair := []shape{
		{kind: torusZ, c: v3{0, 0.15, 1.035}, r: v3{0.105, 0.032, 0}, noise: 0.012, nfreq: 60, bones: one(bHead)},
	}
	hairLo, hairHi := v3{-0.15, 0.0, 0.98}, v3{0.15, 0.30, 1.08}

	// --- eyes ------------------------------------------------------------------------------------
	eyeR := 0.034
	eyes := []shape{
		{kind: ellipsoid, c: v3{-0.07, 0.298, 0.955}, r: v3{eyeR, eyeR * 0.8, eyeR}, bones: one(bHead)},
		{kind: ellipsoid, c: v3{0.07, 0.298, 0.955}, r: v3{eyeR, eyeR * 0.8, eyeR}, bones: one(bHead)},
	}
	eyeUV := func(p, n v3) (float64, float64, int) {
		cx := -0.07
		if p[0] > 0 {
			cx = 0.07
		}
		u := (p[0]-cx)/(2*eyeR) + 0.5
		v := 0.5 - (p[2]-0.955)/(2*eyeR)
		return clamp(u, 0, 0.999), clamp(v, 0, 0.999), 0
	}

	_, _, _, _ = B, headHi, beakHi, hairHi
	return []*bodyDef{
		{name: "body", target: 2200, shapes: body, blend: 0.05, cell: 0.02, uv: boxUV(bodyLo, span), tex: kappaSkin},
		{name: "arms", target: 1000, shapes: arms, blend: 0.03, cell: 0.012, uv: boxUV(bodyLo, span), tex: kappaSkin},
		{name: "head", target: 1100, shapes: head, blend: 0.04, cell: 0.014, uv: boxUV(headLo, 0.36), tex: kappaSkin},
		{name: "shell", target: 800, shapes: shell, blend: 0.03, cell: 0.018, uv: planarUV(shellLo, shellHi, 0, 2, false), tex: matShell(palette{rgb{92, 76, 42}, rgb{38, 30, 16}, rgb{150, 128, 74}, rgb{70, 84, 40}}, 4, 1201)},
		{name: "plastron", target: 350, shapes: plastron, blend: 0, cell: 0.018, uv: planarUV(plLo, plHi, 0, 2, false), tex: matChitin(palette{rgb{196, 172, 104}, rgb{110, 88, 50}, rgb{236, 214, 150}, rgb{150, 140, 80}}, 5, 1202)},
		{name: "beak", target: 400, shapes: beak, blend: 0.025, cell: 0.01, uv: boxUV(beakLo, 0.25), tex: matStreak(palette{rgb{214, 160, 50}, rgb{120, 70, 20}, rgb{250, 210, 110}, rgb{160, 110, 40}}, 1203)},
		{name: "dish", target: 300, shapes: dish, blend: 0, cell: 0.01, uv: planarUV(dishLo, dishHi, 0, 1, false), tex: dishTex},
		{name: "hair", target: 800, shapes: hair, blend: 0, cell: 0.012, uv: boxUV(hairLo, 0.31), tex: matStreak(palette{rgb{34, 50, 34}, rgb{10, 16, 12}, rgb{70, 90, 60}, rgb{30, 40, 50}}, 1204)},
		{name: "eyes", target: 240, shapes: eyes, blend: 0, cell: 0.008, uv: eyeUV, tex: monsterEye(rgb{236, 196, 40}, rgb{110, 70, 10}, true, 0.25)},
	}
}
