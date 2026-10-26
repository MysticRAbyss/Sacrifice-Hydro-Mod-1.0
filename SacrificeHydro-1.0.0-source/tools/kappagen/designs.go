package main

import "math"

// design describes one Hydro creature: which original creature lends its skeleton and
// animations, how big it is, and the shapes and textures of its body parts.
type design struct {
	name, dir string // display name, folder name
	model     string // 4-letter model/tag (readable form)
	base      string // base creature tag (readable form)
	size      float64
	flying    bool
	bindAnim  string // design in this animation's first frame ("" = rest pose)
	borrow    []int  // body parts taken from the base creature's own model at load time
	borrowX   []borrowSpec
	extra     []extraBone
	boneScale map[int]float64
	boneShift map[int]v3
	rebone    []reBone
	wave      []waveBone
	poses     []posePose // bones held in a pose borrowed from another animation (rig.json "poses")
	slowBob   []string // animations whose root up-and-down motion is played at half speed
	noDisp    []int // bones that stay on the ground: they ignore the animation's root displacement
	iconYaw   float64
	iconFocus v3 // point of interest inside the bounding box (0..1 per axis)
	iconZoom  float64
	build     func(bw []v3) []*bodyDef
}

func designs() []*design {
	return []*design{
		{name: "Mistwing", dir: "mistwing", model: "mstw", base: "hsbt", size: 0.8, flying: true, wave: mistwingWave, iconYaw: 0.5, iconFocus: v3{0.5, 0.7, 0.6}, iconZoom: 2.2, build: mistwingDesign},
		{name: "Reefstalker", dir: "reefstalker", model: "rfsk", base: "huma", size: 1.3, bindAnim: "stance", iconYaw: 0.45, iconFocus: v3{0.5, 0.5, 0.8}, iconZoom: 2.2, build: reefstalkerDesign},
		{name: "Mantis Shrimp", dir: "mantisshrimp", model: "mshr", base: "aarc", size: 1.35, bindAnim: "stance", iconYaw: 0.6, iconFocus: v3{0.5, 0.8, 0.6}, iconZoom: 2.0, build: mantisDesign},
		{name: "Brineclaw", dir: "brineclaw", model: "brcl", base: "farc", size: 3.0, bindAnim: "stance", iconYaw: 0.5, iconFocus: v3{0.5, 0.7, 0.55}, iconZoom: 1.5, build: brineclawDesign},
		{name: "Tidecaller", dir: "tidecaller", model: "tdcl", base: "rcha", size: 1.7, bindAnim: "stance", wave: tidecallerWave, noDisp: tidecallerNoDisp, extra: tidecallerExtra, iconYaw: 0.5, iconFocus: v3{0.5, 0.5, 0.78}, iconZoom: 2.2, poses: tidecallerPoses, build: tidecallerDesign},
		{name: "Puffy", dir: "puffy", model: "pufy", base: "sbob", size: 2.55, flying: true, bindAnim: "stance", iconYaw: 0.4, iconFocus: v3{0.5, 0.6, 0.5}, iconZoom: 1.3, build: puffyDesign},
		{name: "Reefbreaker", dir: "reefbreaker", model: "rfbr", base: "fclw", size: 0.9, bindAnim: "stance", iconYaw: 0.6, iconFocus: v3{0.5, 0.7, 0.6}, iconZoom: 1.3, build: reefbreakerDesign},
		{name: "Abyss Guardian", dir: "abyssguardian", model: "abgd", base: "estm", size: 0.8, bindAnim: "stance", wave: abyssWave, extra: abyssExtra, rebone: abyssRebone, noDisp: abyssChain, iconYaw: 0.5, iconFocus: v3{0.5, 0.75, 0.8}, iconZoom: 1.6, build: abyssDesign},
		{name: "Leviathan", dir: "leviathan", model: "lvtn", base: "bdrg", size: 0.95, flying: true, bindAnim: "stance", borrowX: leviathanRig.borrow, extra: leviathanRig.extra, boneScale: leviathanRig.boneScale, boneShift: leviathanRig.boneShift, rebone: levRebone, wave: levWave, slowBob: []string{"fly", "hover"}, iconYaw: 0.5, iconFocus: v3{0.5, 0.75, 0.6}, iconZoom: 3.0, build: leviathanDesign},
		{name: "Tidekin", dir: "tidekin", model: "tidk", base: "klap", size: 1.35, iconYaw: 0.45, iconFocus: v3{0.5, 0.5, 0.72}, iconZoom: 1.9, build: kappaDesign},
	}
}

// --- shape helpers ---------------------------------------------------------------------------

// oriented ellipsoid: long axis from a to b (half length = |b-a|/2), cross radii rx (side) and rz (up)
func along(a, b v3, rx, rz float64, bones []bw) shape {
	c := a.add(b).mul(0.5)
	l := b.sub(a)
	ax := l.norm()
	up := v3{0, 0, 1}
	if math.Abs(ax.dot(up)) > 0.95 {
		up = v3{0, 1, 0}
	}
	side := up.cross(ax).norm()
	up2 := ax.cross(side).norm()
	rot := [3]v3{side, up2, ax}
	return shape{kind: ellipsoid, c: c, r: v3{rx, rz, l.len() / 2}, rot: &rot, bones: bones}
}

func ell(c v3, r v3, bones []bw) shape { return shape{kind: ellipsoid, c: c, r: r, bones: bones} }
func capsl(a, b v3, ra, rb float64, bones []bw) shape {
	return shape{kind: capsule, a: a, b: b, ra: ra, rb: rb, bones: bones}
}
func mirrorX(v v3) v3 { return v3{-v[0], v[1], v[2]} }

// two eyes as a part: centres given for the +x eye
func eyesPart(c v3, r float64, bone int, tex func(w, h int) [][3]uint8) *bodyDef {
	shapes := []shape{ell(c, v3{r, r * 0.85, r}, one(bone)), ell(mirrorX(c), v3{r, r * 0.85, r}, one(bone))}
	uv := func(p, n v3) (float64, float64, int) {
		cx := c[0]
		if p[0] < 0 {
			cx = -c[0]
		}
		return clamp((p[0]-cx)/(2*r)+0.5, 0, 0.999), clamp(0.5-(p[2]-c[2])/(2*r), 0, 0.999), 0
	}
	cell := math.Max(r/4, 0.004)
	return &bodyDef{name: "eyes", target: 240, shapes: shapes, cell: cell, uv: uv, tex: tex}
}

// part with box-projected UVs over its own bounds
func boxPart(name string, target int, cell, blend float64, tex func(w, h int) [][3]uint8, shapes ...shape) *bodyDef {
	d := &bodyDef{name: name, target: target, shapes: shapes, cell: cell, blend: blend, tex: tex}
	lo, hi := d.bounds()
	span := math.Max(hi[0]-lo[0], math.Max(hi[1]-lo[1], hi[2]-lo[2]))
	d.uv = boxUV(lo, span)
	return d
}

// part with a planar projection along one axis (ax1, ax2 are the texture axes)
func planarPart(name string, target int, cell, blend float64, tex func(w, h int) [][3]uint8, ax1, ax2 int, shapes ...shape) *bodyDef {
	d := &bodyDef{name: name, target: target, shapes: shapes, cell: cell, blend: blend, tex: tex}
	lo, hi := d.bounds()
	d.uv = planarUV(lo, hi, ax1, ax2, false)
	return d
}

func cosf(a float64) float64 { return math.Cos(a) }
func sinf(a float64) float64 { return math.Sin(a) }

// any number of eyes (each with its own bone), looking along +y
func eyesMulti(cs []v3, r float64, bones []int, tex func(w, h int) [][3]uint8) *bodyDef {
	var shapes []shape
	for i, c := range cs {
		shapes = append(shapes, ell(c, v3{r, r * 0.85, r}, one(bones[i])))
	}
	uv := func(p, n v3) (float64, float64, int) {
		best, bd := 0, 1e9
		for i, c := range cs {
			if d := p.sub(c).len(); d < bd {
				best, bd = i, d
			}
		}
		c := cs[best]
		return clamp((p[0]-c[0])/(2*r)+0.5, 0, 0.999), clamp(0.5-(p[2]-c[2])/(2*r), 0, 0.999), 0
	}
	return &bodyDef{name: "eyes", target: 120 * len(cs), shapes: shapes, cell: math.Max(r/4, 0.004), uv: uv, tex: tex}
}

// flat sheet (fin, wing membrane): long from a to b, wide towards dir, given half width and thickness
func sheet(a, b, dir v3, halfWidth, thick float64, bones []bw) shape {
	ax := b.sub(a).norm()
	w := dir.sub(ax.mul(dir.dot(ax))).norm()
	n := ax.cross(w).norm()
	rot := [3]v3{w, n, ax}
	return shape{kind: ellipsoid, c: a.add(b).mul(0.5), r: v3{halfWidth, thick, b.sub(a).len() / 2}, rot: &rot, bones: bones}
}

// capsule whose skinning blends from bone ba at a to bone bb at b (keeps joints attached)
func limb(a, b v3, ra, rb float64, ba, bb int) shape {
	if ba == bb {
		return capsl(a, b, ra, rb, one(ba))
	}
	return shape{kind: capsule, a: a, b: b, ra: ra, rb: rb, bones: []bw{{ba, 1}, {bb, 1}}, grad: true}
}

// rayFin: spiny fin rays fanning out from root to each tip, with a thin membrane between
// neighbouring rays
func rayFin(root v3, tips []v3, r0, thick float64, bones []bw) []shape {
	var out []shape
	for i, t := range tips {
		out = append(out, capsl(root, t, r0, r0*0.12, bones))
		if i+1 < len(tips) {
			m := t.add(tips[i+1]).mul(0.5)
			w := tips[i+1].sub(t)
			out = append(out, sheet(root, lerp3(root, m, 0.92), w, w.len()*0.45, thick, bones))
		}
	}
	return out
}
