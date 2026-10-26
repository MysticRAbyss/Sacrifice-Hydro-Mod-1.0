package main

import (
	"fmt"
	"image"
	"image/color"
	"math"
	"strings"
)

// Animation slot names of the CRE8 table (order as in sacengine's animations.d).
var animSlots = func() []string {
	var s []string
	add := func(name string, n int) {
		if n == 1 {
			s = append(s, name)
			return
		}
		for i := 0; i < n; i++ {
			s = append(s, fmt.Sprintf("%s%d", name, i))
		}
	}
	add("stance", 1)
	add("idle", 4)
	for _, n := range []string{"tumble", "run", "falling", "hitFloor", "knocked2Floor", "getUp"} {
		add(n, 1)
	}
	add("attack", 3)
	for _, n := range []string{"damageFront", "damageRight", "damageBack", "damageLeft", "damageTop"} {
		add(n, 1)
	}
	add("death", 3)
	add("shoot", 2)
	for _, n := range []string{"walk", "thrash", "spellcastStart", "spellcast", "spellcastEnd", "runSpellcastStart",
		"runSpellcast", "runSpellcastEnd", "takeoff", "fly", "land", "flyDamage", "flyDeath", "flyAttack", "hover",
		"pickUp", "badLanding", "carry", "carried", "flyShoot", "notify", "stance2", "rise", "corpse", "float",
		"float2Thrash", "sorrow", "doubletake", "ambivalence", "disgust", "bow", "laugh", "disoriented", "corpseRise",
		"floatStatic", "floatMove", "float2Stance", "talk"} {
		add(n, 1)
	}
	add("pulling", 2)
	return s
}()

type cre8 struct {
	path        string
	tag         string // e.g. "fflp"
	model       string // model folder name, e.g. "fflp"
	displayName string
	anims       []string // 64 animation codes (readable), "" if unused
}

func rd4(b []byte) string { // stored reversed in the data
	if b[0] == 0 && b[1] == 0 && b[2] == 0 && b[3] == 0 {
		return ""
	}
	return string([]byte{b[3], b[2], b[1], b[0]})
}

func parseCre8(path string, d []byte) *cre8 {
	if len(d) != 528 {
		return nil
	}
	c := &cre8{path: path, tag: rd4(d[8:12]), model: rd4(d[148:152])}
	for i := 0; i < 64; i++ {
		c.anims = append(c.anims, rd4(d[176+4*i:180+4*i]))
	}
	return c
}

var rawUV = false // set when reading back our own models

type bone struct {
	local  v3 // translation relative to the parent (sac space)
	parent int
}

type vtx struct {
	pos     v3
	uv      [2]float64
	joints  [4]int
	weights [4]float64
	offs    [4]v3 // per-influence offset from the bone (as the engine stores it)
}

type part struct {
	verts []vtx
	faces [][3]int
	tex   *image.NRGBA
	keyed bool
}

type animFrame struct {
	rot  [][4]float64 // per bone quaternion (x,y,z,w) in sac space
	disp v3
}

type anim struct {
	name   string
	code   string
	fps    float64
	frames []animFrame
}

type creature struct {
	noDisp    map[int]bool // bones that ignore the root displacement (stay on the ground)
	tag, name string
	bones     []bone
	parts     []part
	anims     []anim
	height    float64
}

func tgaImage(d []byte, key int) (*image.NRGBA, error) {
	if len(d) < 18 || d[2] != 2 {
		return nil, fmt.Errorf("unsupported TGA")
	}
	idl := int(d[0])
	w, h := int(u16(d, 12)), int(u16(d, 14))
	bpp := int(d[16])
	desc := d[17]
	if bpp != 24 && bpp != 32 {
		return nil, fmt.Errorf("TGA %d bpp", bpp)
	}
	bp := bpp / 8
	px := d[18+idl:]
	if len(px) < w*h*bp {
		return nil, fmt.Errorf("TGA truncated")
	}
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		sy := h - 1 - y // stored bottom-up unless the top-left origin bit is set
		if desc&0x20 != 0 {
			sy = y
		}
		for x := 0; x < w; x++ {
			o := (sy*w + x) * bp
			b, g, r := px[o], px[o+1], px[o+2]
			a := uint8(255)
			if key >= 0 && (int(r)<<16|int(g)<<8|int(b)) == key {
				a = 0
			}
			img.SetNRGBA(x, y, color.NRGBA{r, g, b, a})
		}
	}
	return img, nil
}

// buildCreature assembles mesh, skeleton and animations of one creature (sac world space, z up).
func buildCreature(c *cre8, saxsData, sxmdData []byte, pages map[int][]byte, anims map[string][2][]byte) (*creature, error) {
	if len(saxsData) < 224 {
		return nil, fmt.Errorf("SAXS too short")
	}
	scaling := f32(saxsData, 0)
	m, err := parseSXMD(sxmdData)
	if err != nil {
		return nil, err
	}
	cr := &creature{tag: c.tag, name: c.displayName}
	var heights []int
	for i := range m.parts {
		lo := 0x58 + 6*i
		key := -1
		if i < 16 && saxsData[lo]&1 != 0 {
			key = int(u32(saxsData, lo+2) & 0xffffff)
		}
		pg, ok := pages[i+1]
		if !ok {
			return nil, fmt.Errorf("missing texture page %d", i+1)
		}
		img, err := tgaImage(pg, key)
		if err != nil {
			return nil, fmt.Errorf("page %d: %w", i+1, err)
		}
		heights = append(heights, img.Bounds().Dy())
		cr.parts = append(cr.parts, part{tex: img, keyed: key >= 0})
	}
	if !rawUV {
		prepareUVs(m, heights)
	}
	// skeleton: implicit root + model bones
	cr.bones = append(cr.bones, bone{local: v3{}, parent: -1})
	for _, b := range m.bones {
		cr.bones = append(cr.bones, bone{local: fromSXMD(b.pos[0], b.pos[1], b.pos[2]).mul(scaling), parent: b.parent})
	}
	ap := make([]v3, len(cr.bones)+1) // +1: the static ground bone
	for i := 1; i < len(cr.bones); i++ {
		if cr.bones[i].parent >= i {
			return nil, fmt.Errorf("bone order")
		}
		ap[i] = ap[cr.bones[i].parent].add(cr.bones[i].local)
	}
	type pinfo struct {
		bone int
		off  v3
		w    float64
	}
	pos := make([]pinfo, len(m.positions))
	for i, p := range m.positions {
		pos[i] = pinfo{int(p.bone), fromSXMD(float64(p.pos[0]), float64(p.pos[1]), float64(p.pos[2])).mul(scaling), float64(p.weight) / 16384}
	}
	minZ, maxZ := math.Inf(1), math.Inf(-1)
	for pi, op := range buildMesh(m) {
		pt := &cr.parts[pi]
		pt.faces = op.faces
		for _, ov := range op.verts {
			var v vtx
			v.uv = [2]float64{ov.u, ov.v}
			tw := 0.0
			for k, r := range ov.refs {
				if k >= 3 || r >= len(pos) {
					continue
				}
				p := pos[r]
				v.pos = v.pos.add(ap[p.bone].add(p.off).mul(p.w))
				v.joints[k] = p.bone
				v.offs[k] = p.off
				v.weights[k] = p.w
				tw += p.w
			}
			if tw > 0 {
				for k := range v.weights {
					v.weights[k] /= tw
				}
			} else {
				v.weights[0] = 1
			}
			minZ, maxZ = math.Min(minZ, v.pos[2]), math.Max(maxZ, v.pos[2])
			pt.verts = append(pt.verts, v)
		}
	}
	cr.height = maxZ - minZ
	// animations
	seen := map[string]bool{}
	for slot, code := range c.anims {
		if code == "" || seen[code] {
			continue
		}
		pair, ok := anims[code]
		if !ok {
			continue
		}
		seen[code] = true
		a, err := parseAnim(pair[0], pair[1], scaling, m.zfactor, len(cr.bones))
		if err != nil {
			continue
		}
		a.name = animSlots[slot]
		a.code = code
		cr.anims = append(cr.anims, a)
	}
	return cr, nil
}

func parseAnim(sxsk, skel []byte, scaling, zfactor float64, nBones int) (anim, error) {
	var a anim
	if len(sxsk) < 12 {
		return a, fmt.Errorf("short SXSK")
	}
	nFrames := int(u16(sxsk, 2))
	refHeight := f32(sxsk, 4)
	nb := int(u16(sxsk, 8))
	if nb != nBones {
		return a, fmt.Errorf("bone count %d != %d", nb, nBones)
	}
	a.fps = 30
	if len(skel) >= 8 {
		if fr := u32(skel, 4); fr > 0 && fr < 200 {
			a.fps = float64(fr)
		}
	}
	ratio := 1.0
	if refHeight != 0 {
		ratio = zfactor / refHeight
	}
	for i := 0; i < nFrames; i++ {
		h := 12 + 12*i
		if h+12 > len(sxsk) {
			return a, fmt.Errorf("truncated")
		}
		dx, dy, dz := float64(int16(u16(sxsk, h+2))), float64(int16(u16(sxsk, h+4))), float64(int16(u16(sxsk, h+6)))
		off := int(u32(sxsk, h+8))
		if off+nb*8 > len(sxsk) {
			return a, fmt.Errorf("truncated frame")
		}
		var fr animFrame
		// frameOffset: 10*scale*fromSXMD(disp*0.1*ratio)
		fr.disp = fromSXMD(dx*0.1*ratio, dy*0.1*ratio, dz*0.1*ratio).mul(10 * scaling)
		for b := 0; b < nb; b++ {
			o := off + 8*b
			x, y, z, w := float64(int16(u16(sxsk, o))), float64(int16(u16(sxsk, o+2))), float64(int16(u16(sxsk, o+4))), float64(int16(u16(sxsk, o+6)))
			q := fromSXMD(x, y, z)
			n := math.Sqrt(q[0]*q[0] + q[1]*q[1] + q[2]*q[2] + w*w)
			if n == 0 {
				fr.rot = append(fr.rot, [4]float64{0, 0, 0, 1})
				continue
			}
			fr.rot = append(fr.rot, [4]float64{q[0] / n, q[1] / n, q[2] / n, w / n})
		}
		a.frames = append(a.frames, fr)
	}
	return a, nil
}

func safeName(s string) string {
	return strings.Map(func(r rune) rune {
		if r >= 'a' && r <= 'z' || r >= 'A' && r <= 'Z' || r >= '0' && r <= '9' || r == '_' || r == '-' {
			return r
		}
		return '_'
	}, s)
}
