package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
	"path/filepath"
	"strings"
)

func baseRig(g *gameData, tag string) (*modelFiles, []v3, float64) {
	c := g.cres[tag]
	mf := g.models[c.model]
	m, _ := parseSXMD(mf.sxmd)
	sc := f32(mf.saxs, 0)
	bw := make([]v3, len(m.bones)+1)
	for i, b := range m.bones {
		bw[i+1] = bw[b.parent].add(fromSXMD(b.pos[0], b.pos[1], b.pos[2]).mul(sc))
	}
	return mf, bw, sc
}

// groundBone is the index of the engine's static ground bone for the creature being built: vertices
// bound to it stay put relative to the creature's position, whatever the animation does.
var groundBone int

// designPose gives the bone positions and rotations of the creature being built in a frame of one
// of its animations (t = 0..1), for designs that must look right in a particular pose.
var designPose func(anim string, t float64) ([]v3, [][4]float64)

func generate(g *gameData, d *design, out string) error {
	baseTag, size, model := d.base, d.size, d.model
	mf, bw, sc := baseRig(g, baseTag)
	nb0 := len(bw)
	bm, _ := parseSXMD(mf.sxmd)
	baseLocals, baseParents := make([]v3, nb0), make([]int, nb0)
	for i, b := range bm.bones {
		baseLocals[i+1], baseParents[i+1] = bw[i+1].sub(bw[b.parent]), b.parent
	}
	L, P := rigLocals(d, baseLocals, baseParents)
	nb := len(L)
	bw = make([]v3, nb)
	for i := 1; i < nb; i++ {
		bw[i] = bw[P[i]].add(L[i])
	}
	var bindRot [][4]float64
	if d.bindAnim != "" { // design in the base creature's pose instead of its rest pose
		cr, err := loadOriginal(g, baseTag)
		if err != nil {
			return err
		}
		applyRig(cr, d)
		a := findAnim(cr, d.bindAnim)
		if a == nil || len(a.frames) == 0 {
			return fmt.Errorf("no %s animation", d.bindAnim)
		}
		bw, bindRot = boneTransforms(cr, &a.frames[0])
		bw, bindRot = bw[:nb], bindRot[:nb]
		designPose = func(anim string, t float64) ([]v3, [][4]float64) {
			a := findAnim(cr, anim)
			if a == nil || len(a.frames) == 0 {
				return nil, nil
			}
			return boneTransforms(cr, &a.frames[int(t*float64(len(a.frames)-1))])
		}
	}
	groundBone = nb // the engine's static ground bone comes right after the skeleton
	defs := d.build(bw)
	bw = append(bw, v3{})
	if bindRot != nil {
		bindRot = append(bindRot, [4]float64{0, 0, 0, 1})
	}
	parts, recs := buildParts(defs, bw, bindRot)
	fmt.Printf("  %d position records\n", len(recs))
	os.MkdirAll(out, 0o755)
	if err := writeRig(out, d, sc); err != nil {
		return err
	}
	path := filepath.Join(out, model+".SXMD")
	if err := writeSXMD(path, mf.sxmd, parts, recs, sc, nb); err != nil {
		return err
	}
	if err := writeSAXS(filepath.Join(out, model+".SAXS"), sc*size); err != nil {
		return err
	}
	return writeTextures(out, parts)
}

// readBack loads the written files the way the engine does and returns a posable creature.
func readBack(g *gameData, d *design, dir string) (*creature, error) {
	baseTag := d.base
	sxmd, err := os.ReadFile(filepath.Join(dir, d.model+".SXMD"))
	if err != nil {
		return nil, err
	}
	saxs, _ := os.ReadFile(filepath.Join(dir, d.model+".SAXS"))
	pages := map[int][]byte{}
	files, _ := filepath.Glob(filepath.Join(dir, ".*.SXTX"))
	for _, f := range files {
		var n int
		fmt.Sscanf(filepath.Base(f), ".%03d.SXTX", &n)
		pages[n], _ = os.ReadFile(f)
	}
	base := g.cres[baseTag]
	bm := g.models[base.model]
	sxmd, saxs = withBaseSkeleton(sxmd, saxs, bm.sxmd, bm.saxs)
	c := &cre8{tag: d.model, model: d.model, displayName: d.name, anims: base.anims}
	animDir := strings.Split(strings.TrimPrefix(g.models[base.model].path, "/"), "/")[1]
	rawUV = true
	anims := g.anims[animDir]
	if len(d.extra) > 0 {
		var copies []int
		var shifts []float64
		for _, e := range d.extra {
			copies = append(copies, e.src)
			shifts = append(shifts, e.frameShift)
		}
		ext := map[string][2][]byte{}
		for k, v := range anims {
			if len(v[0]) >= 10 && int(u16(v[0], 8)) == int(u32(bm.sxmd, 28)) {
				v[0] = addBonesSXSK(v[0], copies, shifts)
			}
			ext[k] = v
		}
		anims = ext
	}
	cr, err := buildCreature(c, saxs, sxmd, pages, anims)
	rawUV = false
	if err == nil && d.hasRigChanges() { // the skeleton as the engine builds it from rig.json
		_, bw0, _ := baseRig(g, baseTag)
		pm, _ := parseSXMD(bm.sxmd)
		bl, bp := make([]v3, len(bw0)), make([]int, len(bw0))
		for i, b := range pm.bones {
			bl[i+1], bp[i+1] = bw0[i+1].sub(bw0[b.parent]), b.parent
		}
		L, P := rigLocals(d, bl, bp)
		for i := 1; i < len(L) && i < len(cr.bones); i++ {
			cr.bones[i] = bone{local: L[i].mul(d.size), parent: P[i]}
		}
		applyWaves(cr, d)
		setNoDisp(cr, d)
	}
	if err != nil || (len(d.borrow) == 0 && len(d.borrowX) == 0) {
		return cr, err
	}
	orig, err := loadOriginal(g, baseTag)
	if err != nil {
		return nil, err
	}
	f := d.size
	specs := d.borrowX
	for _, i := range d.borrow {
		specs = append(specs, borrowSpec{part: i, scale: 1})
	}
	rest := make([]v3, len(orig.bones))
	for i := 1; i < len(orig.bones); i++ {
		rest[i] = rest[orig.bones[i].parent].add(orig.bones[i].local)
	}
	for _, b := range specs {
		p := orig.parts[b.part]
		np := part{faces: p.faces, tex: p.tex, keyed: p.keyed}
		for _, v := range p.verts {
			for k := range v.offs {
				if rb, ok := b.rebind[v.joints[k]]; ok && v.weights[k] > 0 {
					v.offs[k] = v.offs[k].add(rest[v.joints[k]]).sub(rest[rb])
					v.joints[k] = rb
				}
				v.offs[k] = v.offs[k].mul(f * b.scale)
				if nb, ok := b.boneMap[v.joints[k]]; ok && v.weights[k] > 0 {
					v.joints[k] = nb
				}
			}
			np.verts = append(np.verts, v)
		}
		cr.parts = append(cr.parts, np)
	}
	return cr, nil
}

func findAnim(cr *creature, name string) *anim {
	for i := range cr.anims {
		if cr.anims[i].name == name {
			return &cr.anims[i]
		}
	}
	return nil
}

func bounds(cr *creature, fr *animFrame) (v3, v3) {
	lo, hi := v3{99, 99, 99}, v3{-99, -99, -99}
	for _, pp := range pose(cr, fr) {
		for _, p := range pp {
			for k := 0; k < 3; k++ {
				lo[k] = math.Min(lo[k], p[k])
				hi[k] = math.Max(hi[k], p[k])
			}
		}
	}
	return lo, hi
}

func previews(cr *creature, d *design, path string) error {
	bg := color.NRGBA{58, 70, 86, 255}
	type view struct {
		anim  string
		frame float64
		yaw   float64
	}
	views := []view{{"", 0, 0.5}, {"", 0, 1.6}, {"", 0, 3.6}, {"stance", 0, 0.7}, {"run", 0.3, 1.2}, {"attack0", 0.5, 0.6}}
	if d.flying {
		views[4].anim = "fly"
	}
	if d.bindAnim != "" {
		for i := 0; i < 3; i++ {
			views[i].anim = d.bindAnim
		}
	}
	const W, H = 360, 420
	sheet := image.NewNRGBA(image.Rect(0, 0, W*len(views), H))
	var fr0 *animFrame
	if a := findAnim(cr, "stance"); a != nil && len(a.frames) > 0 {
		fr0 = &a.frames[0]
	}
	lo0, hi0 := bounds(cr, fr0)
	size := math.Max(hi0[0]-lo0[0], math.Max(hi0[1]-lo0[1], hi0[2]-lo0[2]))
	for i, v := range views {
		var fr *animFrame
		if v.anim != "" {
			if a := findAnim(cr, v.anim); a != nil && len(a.frames) > 0 {
				fr = &a.frames[int(v.frame*float64(len(a.frames)-1))]
			}
		}
		lo, hi := bounds(cr, fr)
		img := renderCreature(cr, fr, camera{yaw: v.yaw, pitch: 0.2, center: lo.add(hi).mul(0.5), scale: 0.8 * W / size}, W, H, bg)
		draw.Draw(sheet, image.Rect(W*i, 0, W*(i+1), H), img, image.Point{}, draw.Src)
	}
	return writePNG(path, sheet)
}

func icon(cr *creature, d *design, path string, size int) error {
	bg := color.NRGBA{24, 74, 110, 255}
	var fr *animFrame
	if a := findAnim(cr, "stance"); a != nil && len(a.frames) > 0 {
		fr = &a.frames[0]
	}
	lo, hi := bounds(cr, fr)
	ext := math.Max(hi[0]-lo[0], math.Max(hi[1]-lo[1], hi[2]-lo[2]))
	c := lo.add(hi).mul(0.5)
	sc := 256 / ext
	if d.iconZoom > 0 { // frame the upper part (the face) more closely
		e := hi.sub(lo)
		c = lo.add(v3{e[0] * d.iconFocus[0], e[1] * d.iconFocus[1], e[2] * d.iconFocus[2]})
		sc *= d.iconZoom
	}
	big := renderCreature(cr, fr, camera{yaw: d.iconYaw, pitch: 0.12, center: c, scale: sc}, 256, 256, bg)
	small := image.NewNRGBA(image.Rect(0, 0, size, size))
	f := 256 / size
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			var r, gg, b int
			for dy := 0; dy < f; dy++ {
				for dx := 0; dx < f; dx++ {
					c := big.NRGBAAt(x*f+dx, y*f+dy)
					r, gg, b = r+int(c.R), gg+int(c.G), b+int(c.B)
				}
			}
			n := f * f
			small.SetNRGBA(x, y, color.NRGBA{uint8(r / n), uint8(gg / n), uint8(b / n), 255})
		}
	}
	return writePNG(path, small)
}

// iconRawRender renders the whole creature (stance) for a round spellbook icon, twice: on black
// and on white, so that the icon tool (tools/sfx/make_creature_icons.py) can recover its outline.
func iconRawRender(cr *creature, d *design, base string, yaw, pitch float64) error {
	var fr *animFrame
	if a := findAnim(cr, "stance"); a != nil && len(a.frames) > 0 {
		fr = &a.frames[0]
	}
	lo, hi := bounds(cr, fr)
	ext := math.Max(hi[0]-lo[0], math.Max(hi[1]-lo[1], hi[2]-lo[2]))
	c := lo.add(hi).mul(0.5)
	cam := camera{yaw: yaw, pitch: pitch, center: c, scale: 300 / ext}
	for _, v := range []struct {
		suffix string
		bg     color.NRGBA
	}{{"_k.png", color.NRGBA{0, 0, 0, 255}}, {"_w.png", color.NRGBA{255, 255, 255, 255}}} {
		img := renderCreature(cr, fr, cam, 384, 384, v.bg)
		if err := writePNG(base+v.suffix, img); err != nil {
			return err
		}
	}
	return nil
}

func closeup(cr *creature, path string, yaw float64) error {
	var fr *animFrame
	if a := findAnim(cr, "stance"); a != nil && len(a.frames) > 0 {
		fr = &a.frames[0]
	}
	lo, hi := bounds(cr, fr)
	ext := math.Max(hi[0]-lo[0], math.Max(hi[1]-lo[1], hi[2]-lo[2]))
	img := renderCreature(cr, fr, camera{yaw: yaw, pitch: 0.15, center: lo.add(hi).mul(0.5), scale: 560 / ext}, 640, 640, color.NRGBA{58, 70, 86, 255})
	return writePNG(path, img)
}
