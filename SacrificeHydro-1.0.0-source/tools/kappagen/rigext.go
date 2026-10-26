package main

// Skeleton changes a design can make on top of its base creature's skeleton:
//   - boneScale / boneShift: move original bones (e.g. smaller wings);
//   - extra: add bones after the base skeleton, each animated like a source bone (e.g. a second
//     pair of wings further back);
//   - borrowX: borrow body parts of the base creature's model, re-bound to other bones and resized.
// The engine applies the same changes from rig.json at load time (the skeleton itself is never
// shipped: only the moved bones' new positions and the added bones).

import (
	"slices"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strings"
)

type extraBone struct {
	src, parent int
	shift       v3      // for a bone hanging off an original bone: offset from the source's position
	scale       float64 // for a bone hanging off an added bone: its source's local offset, scaled
	frameShift  float64 // takes its rotation from this far ahead in the animation (0.5 = counter-time)
}

// reBone moves a base bone to a new parent and local offset (engine axes, base scaling)
type reBone struct {
	bone, parent int
	local        v3
}

// waveBone replaces a bone's rotation in every frame by a sway about axis (engine axes, local)
type waveBone struct {
	bone       int
	axis       v3
	amp, phase float64
	pow        float64 // >0: a short bump, max(0,sin)^pow
	idleOnly   bool    // only in the standing animations (stance, idles)
	period     float64 // seconds: a slow, single cycle over a standing animation played several times
	speed      float64 // 0 or 1: normal; 0.5: half as many cycles
	moveOnly   bool    // only in the run and walk animations
}

func waveCycles(frames int) float64 { return math.Max(1, math.Round(float64(frames)/30)) }

func waveQuat(w waveBone, f, frames int) [4]float64 {
	cycles := waveCycles(frames)
	if w.speed > 0 {
		cycles = math.Max(1, math.Round(float64(frames)/30*w.speed))
	}
	if w.idleOnly && w.period > 0 {
		cycles = 1
	}
	sv := math.Sin(2*math.Pi*cycles*float64(f)/float64(frames) + w.phase)
	a := w.amp * sv
	if w.pow > 0 {
		a = w.amp * math.Pow(math.Max(0, sv), w.pow)
	}
	ax := w.axis.norm().mul(math.Sin(a / 2))
	return [4]float64{ax[0], ax[1], ax[2], math.Cos(a / 2)}
}

// applyWaves sets the swaying bones' rotations in all animations
func applyWaves(cr *creature, d *design) {
	for ai := range cr.anims {
		name0 := cr.anims[ai].name
		if slices.Contains(d.slowBob, name0) && len(cr.anims[ai].frames) > 1 { // as the engine: body bob at half speed
			o := cr.anims[ai].frames
			n0 := len(o)
			var fr []animFrame
			for f := 0; f < 2*n0; f++ {
				x := float64(f) / 2
				i0 := int(x) % n0
				i1 := (i0 + 1) % n0
				t := x - math.Floor(x)
				nf := o[f%n0]
				nf.rot = append([][4]float64{}, nf.rot...)
				nf.disp = o[i0].disp.add(o[i1].disp.sub(o[i0].disp).mul(t))
				fr = append(fr, nf)
			}
			cr.anims[ai].frames = fr
		}
		if name0 == "stance" || strings.HasPrefix(name0, "idle") { // as the engine: repeat for slow waves
			n0 := len(cr.anims[ai].frames)
			reps := 1
			for _, w := range d.wave {
				if w.idleOnly && w.period > 0 && n0 > 0 {
					reps = max(reps, int(math.Round(w.period*30/float64(n0))))
				}
			}
			var fr []animFrame
			for r := 0; r < reps; r++ {
				for _, f := range cr.anims[ai].frames {
					f.rot = append([][4]float64{}, f.rot...)
					fr = append(fr, f)
				}
			}
			cr.anims[ai].frames = fr
		}
		n := len(cr.anims[ai].frames)
		for fi := range cr.anims[ai].frames {
			fr := &cr.anims[ai].frames[fi]
			name := cr.anims[ai].name
			idle := name == "stance" || strings.HasPrefix(name, "idle")
			for _, w := range d.wave {
				if w.idleOnly && !idle {
					continue
				}
				if w.moveOnly && name != "run" && name != "walk" {
					continue
				}
				if w.bone < len(fr.rot) {
					fr.rot[w.bone] = waveQuat(w, fi, n)
				}
			}
		}
	}
}

type borrowSpec struct {
	part    int
	boneMap map[int]int
	scale   float64
	rebind  map[int]int
}

func (d *design) hasRigChanges() bool {
	return len(d.poses) > 0 || len(d.slowBob) > 0 || len(d.noDisp) > 0 || len(d.extra) > 0 || len(d.boneScale) > 0 || len(d.boneShift) > 0 || len(d.rebone) > 0 || len(d.wave) > 0
}

// rigLocals returns local bone offsets (engine axes, base scaling) and parents of the changed
// skeleton, given the base skeleton's locals and parents (index 0 = root).
func rigLocals(d *design, locals []v3, parents []int) ([]v3, []int) {
	n := len(locals)
	L := append([]v3{}, locals...)
	P := append([]int{}, parents...)
	for i := 1; i < n; i++ {
		if s, ok := d.boneScale[i]; ok {
			L[i] = L[i].mul(s)
		}
		if s, ok := d.boneShift[i]; ok {
			L[i] = L[i].add(s)
		}
	}
	for _, r := range d.rebone {
		L[r.bone], P[r.bone] = r.local, r.parent
	}
	world := func(i int) v3 {
		var w v3
		for ; i > 0; i = P[i] {
			w = w.add(L[i])
		}
		return w
	}
	for _, e := range d.extra {
		var l v3
		if e.parent >= n {
			s := e.scale
			if s == 0 {
				s = 1
			}
			l = L[e.src].mul(s)
		} else {
			l = world(e.src).sub(world(e.parent)).add(e.shift)
		}
		L = append(L, l)
		P = append(P, e.parent)
	}
	return L, P
}

// applyRig changes a loaded creature's skeleton and animations like the engine does.
func applyRig(cr *creature, d *design) {
	if !d.hasRigChanges() {
		return
	}
	n := len(cr.bones)
	locals := make([]v3, n)
	parents := make([]int, n)
	for i, b := range cr.bones {
		locals[i], parents[i] = b.local, b.parent
	}
	L, P := rigLocals(d, locals, parents)
	cr.bones = cr.bones[:0]
	for i := range L {
		cr.bones = append(cr.bones, bone{local: L[i], parent: P[i]})
	}
	cr.bones[0].parent = -1
	for ai := range cr.anims {
		frs := cr.anims[ai].frames
		orig := make([][][4]float64, len(frs))
		for fi := range frs {
			orig[fi] = append([][4]float64{}, frs[fi].rot...)
		}
		for fi := range frs {
			fr := &frs[fi]
			if len(fr.rot) != n {
				continue
			}
			for _, e := range d.extra {
				g := (fi + int(math.Round(e.frameShift*float64(len(frs))))) % len(frs)
				fr.rot = append(fr.rot, orig[g][e.src])
			}
		}
	}
	applyWaves(cr, d)
	setNoDisp(cr, d)
}

func setNoDisp(cr *creature, d *design) {
	if len(d.noDisp) > 0 {
		cr.noDisp = map[int]bool{}
		for _, b := range d.noDisp {
			cr.noDisp[b] = true
		}
	}
}

// addBonesSXSK mirrors the engine: every frame gets rotations for the added bones.
func addBonesSXSK(d []byte, copies []int, shifts []float64) []byte {
	nf := int(u16(d, 2))
	nb := int(u16(d, 8))
	he := 12 + 12*nf
	r := append([]byte{}, d[:he]...)
	binary.LittleEndian.PutUint16(r[8:], uint16(nb+len(copies)))
	for f := 0; f < nf; f++ {
		off := int(u32(d, 12+12*f+8))
		binary.LittleEndian.PutUint32(r[12+12*f+8:], uint32(len(r)))
		r = append(r, d[off:off+8*nb]...)
		for k, c := range copies {
			g := (f + int(math.Round(shifts[k]*float64(nf)))) % nf
			og := int(u32(d, 12+12*g+8))
			r = append(r, d[og+8*c:og+8*c+8]...)
		}
	}
	return r
}

// writeRig writes rig.json, read by the engine, which applies the changes to the base skeleton
// from the player's own game files (SXMD units: engine axes converted, base scaling undone).
func writeRig(dir string, d *design, sc float64) error {
	rig := map[string]any{}
	sx := func(e v3) [3]float64 { p := toSX(e).mul(1 / sc); return [3]float64{p[0], p[1], p[2]} }
	if len(d.boneScale) > 0 {
		m := map[string]float64{}
		for k, v := range d.boneScale {
			m[fmt.Sprint(k)] = v
		}
		rig["boneScale"] = m
	}
	if len(d.boneShift) > 0 {
		m := map[string][3]float64{}
		for k, v := range d.boneShift {
			m[fmt.Sprint(k)] = sx(v)
		}
		rig["boneShift"] = m
	}
	if len(d.extra) > 0 {
		var es []any
		for _, e := range d.extra {
			m := map[string]any{"copy": e.src, "parent": e.parent}
			if e.frameShift != 0 {
				m["frameShift"] = e.frameShift
			}
			if e.scale != 0 {
				m["scale"] = e.scale
			}
			if e.shift != (v3{}) {
				m["shift"] = sx(e.shift)
			}
			es = append(es, m)
		}
		rig["extraBones"] = es
	}
	if len(d.rebone) > 0 {
		var rs []any
		for _, r := range d.rebone {
			rs = append(rs, map[string]any{"bone": r.bone, "parent": r.parent, "pos": sx(r.local)})
		}
		rig["rebone"] = rs
	}
	if len(d.noDisp) > 0 {
		rig["noDisplacement"] = d.noDisp
	}
	if len(d.slowBob) > 0 {
		var sl []int
		for _, n := range d.slowBob {
			sl = append(sl, slices.Index(animSlots, n))
		}
		rig["slowBob"] = sl
	}
	if len(d.wave) > 0 {
		var ws []any
		for _, w := range d.wave {
			a := toSX(w.axis.norm())
			m := map[string]any{"bone": w.bone, "axis": [3]float64{a[0], a[1], a[2]}, "amp": w.amp, "phase": w.phase}
			if w.pow > 0 {
				m["pow"] = w.pow
			}
			if w.idleOnly {
				m["idleOnly"] = true
			}
			if w.period > 0 {
				m["period"] = w.period
			}
			if w.moveOnly {
				m["moveOnly"] = true
			}
			if w.speed > 0 && w.speed != 1 {
				m["speed"] = w.speed
			}
			ws = append(ws, m)
		}
		rig["wave"] = ws
	}
	if len(d.poses) > 0 {
		var ps []any
		for _, p := range d.poses {
			var slots []int
			for _, n := range p.anims {
				slots = append(slots, slices.Index(animSlots, n))
			}
			ps = append(ps, map[string]any{"slots": slots, "from": slices.Index(animSlots, p.from), "frame": p.frame, "bones": p.bones, "in": p.in, "out": p.out})
		}
		rig["poses"] = ps
	}
	if len(d.borrowX) > 0 {
		var bs []any
		for _, b := range d.borrowX {
			m := map[string]int{}
			for k, v := range b.boneMap {
				m[fmt.Sprint(k)] = v
			}
			rb := map[string]int{}
			for k, v := range b.rebind {
				rb[fmt.Sprint(k)] = v
			}
			bs = append(bs, map[string]any{"part": b.part, "bones": m, "scale": b.scale, "rebind": rb})
		}
		rig["borrowParts"] = bs
	}
	path := filepath.Join(dir, "rig.json")
	if len(rig) == 0 {
		os.Remove(path)
		return nil
	}
	js, _ := json.MarshalIndent(rig, "", "  ")
	return os.WriteFile(path, js, 0o644)
}
