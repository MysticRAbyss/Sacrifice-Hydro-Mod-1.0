package main

// Writes a creature in Sacrifice's own model format (SXMD/SAXS/SXTX), the way the game's
// explicit-face body parts are stored: one ring per part holding every vertex, plus a face list.

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math"
	"os"
	"path/filepath"
)

// textures of the new creatures; the engine reads their UVs directly ("rawUV" in creature.json)
var texSize = 512

type outPartData struct {
	def     *bodyDef
	mesh    *mesh
	entries []entry
	faces   [][3]int // into entries; SXMD winding
	tex     [][3]uint8
}

type entry struct {
	refs  []int // position records
	u, v  float64
	vert  int // mesh vertex
}

type posRec struct {
	off    v3 // engine space, relative to bone
	bone   int
	w      float64
	normal v3
}

// toSX converts engine space to SXMD axes: engine = (-x, -z, y)
func toSX(e v3) v3 { return v3{-e[0], e[2], -e[1]} }

func buildParts(defs []*bodyDef, boneWorld []v3, bindRot [][4]float64) ([]*outPartData, []posRec) {
	var parts []*outPartData
	var recs []posRec
	for _, d := range defs {
		m := d.polygonize()
		raw := len(m.verts)
		for t := d.target; ; t *= 2 { // small, separate pieces can collapse away: give them more vertices
			mm := &mesh{verts: append([]meshVert{}, m.verts...), faces: append([][3]int{}, m.faces...)}
			decimate(mm, t, d)
			if len(mm.faces) >= len(mm.verts) || t > raw {
				m = mm
				break
			}
		}
		if len(m.faces) == 0 {
			fmt.Printf("  part %s is empty, skipped\n", d.name)
			continue
		}
		d.skin(m)
		op := &outPartData{def: d, mesh: m}
		// position records per mesh vertex
		vrefs := make([][]int, len(m.verts))
		for vi, v := range m.verts {
			for _, b := range v.bones {
				off, n := v.p.sub(boneWorld[b.bone]), v.n
				if bindRot != nil { // designed in an animation pose: undo that bone's rotation
					inv := bindRot[b.bone]
					inv = [4]float64{-inv[0], -inv[1], -inv[2], inv[3]}
					off, n = qrot(inv, off), qrot(inv, n)
				}
				recs = append(recs, posRec{off: off, bone: b.bone, w: b.w, normal: n})
				vrefs[vi] = append(vrefs[vi], len(recs)-1)
			}
		}
		// UV groups: one entry per (vertex, group)
		key := map[[2]int]int{}
		for _, f := range m.faces {
			pa, pb, pc := m.verts[f[0]].p, m.verts[f[1]].p, m.verts[f[2]].p
			fn := pb.sub(pa).cross(pc.sub(pa)).norm()
			_, _, g := d.uv(pa.add(pb).add(pc).mul(1.0/3), fn)
			var tri [3]int
			for k, vi := range f {
				kk := [2]int{vi, g}
				ei, ok := key[kk]
				if !ok {
					u, v, _ := d.uv(m.verts[vi].p, fn)
					op.entries = append(op.entries, entry{refs: vrefs[vi], u: u, v: v, vert: vi})
					ei = len(op.entries) - 1
					key[kk] = ei
				}
				tri[k] = ei
			}
			// keep engine-space outward CCW: the game's own models have positive volume this way,
			// and the engine culls back faces
			op.faces = append(op.faces, tri)
		}
		op.tex = d.tex(texSize, texSize)
		parts = append(parts, op)
		fmt.Printf("  part %-10s %6d -> %5d vertices %6d entries %6d faces\n", d.name, raw, len(m.verts), len(op.entries), len(op.faces))
	}
	return parts, recs
}

func le32(b *bytes.Buffer, v uint32) { binary.Write(b, binary.LittleEndian, v) }
func le16(b *bytes.Buffer, v uint16) { binary.Write(b, binary.LittleEndian, v) }

func writeSXMD(path string, base []byte, parts []*outPartData, recs []posRec, unit float64, nb1 int) error {
	if len(recs) >= 0xffff {
		return fmt.Errorf("too many position records (%d)", len(recs))
	}
	baseBoneStart := int(u32(base, 32))
	if nb1 == 0 {
		nb1 = int(u32(base, 28)) // bones + 1
	}
	_ = baseBoneStart
	// The skeleton is not stored: the engine copies it from the base creature at load time,
	// so no data from the game's own files ends up in the mod.
	boneBlock := make([]byte, nb1*140)
	n := len(parts)
	hdrLen := 40 + 4*n
	for hdrLen%4 != 0 {
		hdrLen++
	}
	var body bytes.Buffer // everything after the header; offsets are absolute
	off := func() int { return hdrLen + body.Len() }
	boneStart := off()
	body.Write(boneBlock)
	partOffs := make([]int, n)
	for i, p := range parts {
		ne := len(p.entries)
		if ne > 0xffff || len(p.faces) > 0xffff {
			return fmt.Errorf("part %s too large", p.def.name)
		}
		partOffs[i] = off()
		hdr := make([]byte, 36)
		hdr[0] = 1
		hdr[1] = byte(ne / 256)
		binary.LittleEndian.PutUint16(hdr[2:], uint16(len(p.faces)))
		ringOff := partOffs[i] + 36 + 4
		entOff := ringOff + 8
		faceOff := entOff + 10*ne + 10
		binary.LittleEndian.PutUint32(hdr[24:], uint32(faceOff))
		body.Write(hdr)
		le32(&body, uint32(ringOff))
		body.Write([]byte{byte(ne % 256), 0, 0, 0})
		le32(&body, uint32(entOff))
		vScale := (float64(texSize)-1)/256 + 0.001 // undone by the engine for explicit-face parts
		for _, e := range p.entries {
			var idx [3]uint16
			for k := 0; k < 3; k++ {
				idx[k] = 0xffff
				if k < len(e.refs) {
					idx[k] = uint16(e.refs[k])
				}
			}
			for _, x := range idx {
				le16(&body, x)
			}
			u := clamp(math.Floor(e.u*256), 0, 255)
			v := clamp(math.Floor(e.v*256), 0, 255)
			_ = vScale
			body.Write([]byte{byte(u), byte(v)})
			le16(&body, 0)
		}
		body.Write(make([]byte, 10))
		for _, f := range p.faces {
			for _, x := range f {
				le16(&body, uint16(x))
			}
		}
		for body.Len()%4 != 0 {
			body.WriteByte(0)
		}
	}
	posStart := off()
	for _, r := range recs {
		s := toSX(r.off).mul(1 / unit)
		for k := 0; k < 3; k++ {
			if math.Abs(s[k]) > 32767 {
				return fmt.Errorf("position out of range")
			}
			le16(&body, uint16(int16(math.Round(s[k]))))
		}
		le16(&body, uint16(r.bone))
		le16(&body, uint16(math.Round(r.w*16384)))
		nn := toSX(r.normal)
		for k := 0; k < 3; k++ {
			le16(&body, uint16(int16(math.Round(clamp(nn[k], -1, 1)*32767))))
		}
	}
	// header
	h := make([]byte, hdrLen)
	h[0] = byte(n)
	binary.LittleEndian.PutUint32(h[8:], math.Float32bits(float32(f32(base, 8)))) // z factor of the base rig
	for i := 0; i < n && 12+i < 28; i++ {
		h[13+i] = byte(i) // per-part list (unused by sacengine)
	}
	binary.LittleEndian.PutUint32(h[28:], uint32(nb1))
	binary.LittleEndian.PutUint32(h[32:], uint32(boneStart))
	binary.LittleEndian.PutUint32(h[36:], uint32(posStart))
	for i, o := range partOffs {
		binary.LittleEndian.PutUint32(h[40+4*i:], uint32(o))
	}
	return os.WriteFile(path, append(h, body.Bytes()...), 0o644)
}

// writeSAXS writes the model settings. Reference points and hitbox bones are filled in by the
// engine from the base creature at load time.
func writeSAXS(path string, scaling float64) error {
	b := make([]byte, 224)
	binary.LittleEndian.PutUint32(b[0:], math.Float32bits(float32(scaling)))
	binary.LittleEndian.PutUint32(b[8:], math.Float32bits(2.0))  // vertex quality (explicit faces ignore it)
	binary.LittleEndian.PutUint32(b[12:], math.Float32bits(2.0)) // ring quality
	return os.WriteFile(path, b, 0o644) // limbs all zero: no colour keys, every part opaque
}

// withBaseSkeleton returns copies of the mod's SXMD/SAXS with the base creature's skeleton,
// reference points and hitbox bones filled in (what the engine does when loading).
func withBaseSkeleton(sxmd, saxs, baseSXMD, baseSAXS []byte) ([]byte, []byte) {
	m := append([]byte{}, sxmd...)
	s := append([]byte{}, saxs...)
	mb, bb, n := int(u32(m, 32)), int(u32(baseSXMD, 32)), int(u32(baseSXMD, 28))
	copy(m[mb:mb+n*140], baseSXMD[bb:bb+n*140])
	copy(s[0x18:0x58], baseSAXS[0x18:0x58])
	copy(s[0xB8:0xB8+40], baseSAXS[0xB8:0xB8+40])
	return m, s
}

func writeTextures(dir string, parts []*outPartData) error {
	for i, p := range parts {
		if err := writeTGA(filepath.Join(dir, fmt.Sprintf(".%03d.SXTX", i+1)), texSize, texSize, p.tex); err != nil {
			return err
		}
	}
	return nil
}
