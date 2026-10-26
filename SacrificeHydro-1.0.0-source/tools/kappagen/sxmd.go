package main

// Sacrifice creature meshes (SXMD + SAXS + SXTX pages).
// Format knowledge and the mesh-assembly algorithm follow sacengine by tg
// (https://github.com/tg-2/sacengine, GPL-3.0): sxmd.d / saxs.d / saxs_.d.

import (
	"encoding/binary"
	"fmt"
	"math"
)

type v3 [3]float64

func (a v3) add(b v3) v3      { return v3{a[0] + b[0], a[1] + b[1], a[2] + b[2]} }
func (a v3) mul(s float64) v3 { return v3{a[0] * s, a[1] * s, a[2] * s} }

// fromSXMD converts model coordinates to sacengine's world space (z up), as in util.d.
func fromSXMD(x, y, z float64) v3 { return v3{-x, -z, y} }

type ringEntry struct {
	idx   [3]uint16
	u, v  uint8
	flags uint16
	ptr   int // global id of this entry (for "emitted" bookkeeping)
}

type ring struct {
	numEntries uint8
	unknown    uint8
	texture    uint16
	entries    []ringEntry
	entryOff   int    // file offset of the first entry
	post       uint16 // flags of the entry right after the ring (read past the end, as the original does)
}

type seamRec struct {
	bodyPart, ring, vertex uint16
	u, v                   uint8
}

type bodyPart struct {
	numExplicit uint16
	flags       uint32
	rings       []ring
	explicit    [][3]uint16
	hasStrip    bool
	strip0      [2]uint8
	strips      []seamRec
}

type sxBone struct {
	pos    v3
	parent int
}

type sxPos struct {
	pos    [3]int16
	bone   int16
	weight uint16
	normal [3]int16
}

type sxModel struct {
	zfactor   float64
	bones     []sxBone // without the implicit root
	parts     []bodyPart
	positions []sxPos
	raw       []byte
}

func u16(b []byte, o int) uint16 { return binary.LittleEndian.Uint16(b[o:]) }
func u32(b []byte, o int) uint32 { return binary.LittleEndian.Uint32(b[o:]) }
func f32(b []byte, o int) float64 {
	return float64(math.Float32frombits(binary.LittleEndian.Uint32(b[o:])))
}

func parseSXMD(d []byte) (*sxModel, error) {
	if len(d) < 64 {
		return nil, fmt.Errorf("SXMD too short")
	}
	m := &sxModel{zfactor: f32(d, 8), raw: d}
	nParts := int(d[0])
	nBones := int(u32(d, 28)) - 1
	boneOff := int(u32(d, 32)) + 140
	vertOff := int(u32(d, 36))
	for i := 0; i < nBones; i++ {
		o := boneOff + i*140
		m.bones = append(m.bones, sxBone{pos: v3{f32(d, o), f32(d, o+4), f32(d, o+8)}, parent: int(u16(d, o+12))})
	}
	numPos := 0
	for k := 0; k < nParts; k++ {
		off := int(u32(d, 40+4*k))
		var bp bodyPart
		nRings := int(d[off])
		add := int(d[off+1])
		bp.numExplicit = u16(d, off+2)
		bp.flags = u32(d, off+20)
		expOff := int(u32(d, off+24))
		stripOff := int(u32(d, off+28))
		for r := 0; r < nRings; r++ {
			ro := int(u32(d, off+36+4*r))
			rg := ring{numEntries: d[ro], unknown: d[ro+1], texture: u16(d, ro+2)}
			eo := int(u32(d, ro+4))
			rg.entryOff = eo
			n := int(rg.numEntries) + 256*add
			for e := 0; e < n; e++ {
				p := eo + 10*e
				re := ringEntry{idx: [3]uint16{u16(d, p), u16(d, p+2), u16(d, p+4)}, u: d[p+6], v: d[p+7], flags: u16(d, p+8), ptr: p}
				for _, ix := range re.idx {
					if ix != 0xffff && int(ix) >= numPos {
						numPos = int(ix) + 1
					}
				}
				rg.entries = append(rg.entries, re)
			}
			if p := eo + 10*n + 8; p+2 <= len(d) {
				rg.post = u16(d, p)
			}
			bp.rings = append(bp.rings, rg)
		}
		for i := 0; i < int(bp.numExplicit); i++ {
			p := expOff + 6*i
			bp.explicit = append(bp.explicit, [3]uint16{u16(d, p), u16(d, p+2), u16(d, p+4)})
		}
		if stripOff != 0 {
			bp.hasStrip = true
			bp.strip0 = [2]uint8{d[stripOff], d[stripOff+1]}
			ns := int(u16(d, stripOff+2))
			for i := 0; i < ns; i++ {
				p := stripOff + 8 + 8*i
				bp.strips = append(bp.strips, seamRec{u16(d, p), u16(d, p+2), u16(d, p+4), d[p+6], d[p+7]})
			}
		}
		m.parts = append(m.parts, bp)
	}
	for i := 0; i < numPos; i++ {
		p := vertOff + 16*i
		if p+16 > len(d) {
			return nil, fmt.Errorf("SXMD position table truncated")
		}
		m.positions = append(m.positions, sxPos{
			pos:    [3]int16{int16(u16(d, p)), int16(u16(d, p+2)), int16(u16(d, p+4))},
			bone:   int16(u16(d, p+6)),
			weight: u16(d, p+8),
			normal: [3]int16{int16(u16(d, p+10)), int16(u16(d, p+12)), int16(u16(d, p+14))},
		})
	}
	return m, nil
}

// prepareUVs maps the stored texel coordinates onto each part's own texture page.
func prepareUVs(m *sxModel, heights []int) {
	for bi := range m.parts {
		bp := &m.parts[bi]
		for ri := range bp.rings {
			rg := &bp.rings[ri]
			if bp.numExplicit != 0 {
				vScale := 1.0 / ((float64(heights[bi])-1.0)/256.0 + 0.001)
				for e := range rg.entries {
					rg.entries[e].v = uint8(uint32(float64(rg.entries[e].v) * vScale))
				}
			} else {
				var row uint32
				if ri+1 == len(bp.rings) {
					row = 255
				} else {
					row = (255*(uint32(rg.texture)+1) + 8192) / 16384
				}
				for e := range rg.entries {
					rg.entries[e].v = uint8(row)
				}
			}
			rg.texture = 256
		}
		if bp.hasStrip {
			bp.strip0[1] = 255
		}
	}
}

// ---- mesh assembly (port of saxs.d / sxmd.d) ----

type ringVertex struct {
	flags  uint16
	u      uint8
	active bool
}

func buildRingFaces(oldRing, newRing []ringVertex, oldPost, newPost uint16) [][3]int {
	count := func(r []ringVertex) int {
		n := 0
		for _, v := range r {
			if v.active {
				n++
			}
		}
		return n
	}
	oldActive, newActive := count(oldRing), count(newRing)
	var out [][3]int
	if oldActive == 0 || newActive == 0 {
		return out
	}
	oldBase := 0
	newBase := oldActive + 1
	var oldCursor, newCursor int
	var oldOrd, newOrd int
	oldIndex, newIndex, previous := -1, -1, -1
	havePrev, oldWrapped, newWrapped := false, false, false
	fetchedNew := true
	specialOld := func() bool { return oldRing[oldCursor].flags&0x60 != 0 }
	specialNew := func() bool { return newRing[newCursor].flags&0x60 != 0 }
	nextOldSpecial := func() bool {
		f := oldPost
		if oldCursor+1 < len(oldRing) {
			f = oldRing[oldCursor+1].flags
		}
		return f&0x60 != 0
	}
	nextNewSpecial := func() bool {
		f := newPost
		if newCursor+1 < len(newRing) {
			f = newRing[newCursor+1].flags
		}
		return f&0x60 != 0
	}
	fetchOld := func() {
		for !oldRing[oldCursor].active {
			oldCursor++
			if oldCursor == len(oldRing) {
				oldCursor = 0
				oldWrapped = true
			}
		}
		oldIndex = oldBase + oldOrd
		oldOrd++
		fetchedNew = false
	}
	fetchNew := func() {
		for !newRing[newCursor].active {
			newCursor++
			if newCursor == len(newRing) {
				newCursor = 0
				newWrapped = true
			}
		}
		newIndex = newBase + newOrd
		newOrd++
		fetchedNew = true
	}
	for guard := 0; guard < 100000; guard++ {
		if oldIndex < 0 {
			fetchOld()
		}
		if newIndex < 0 {
			fetchNew()
		}
		if havePrev {
			if fetchedNew {
				out = append(out, [3]int{previous, oldIndex, newIndex})
			} else {
				out = append(out, [3]int{newIndex, previous, oldIndex})
			}
		}
		consumeOld := (oldRing[oldCursor].u < newRing[newCursor].u || newWrapped) && !oldWrapped
		os, ns := specialOld(), specialNew()
		if os && consumeOld && nextNewSpecial() {
			consumeOld = false
		} else if !os && ns && !consumeOld && nextOldSpecial() {
			consumeOld = true
		}
		if consumeOld {
			previous = oldIndex
			havePrev = true
			oldIndex = -1
			oldCursor++
			if oldCursor == len(oldRing) {
				if newWrapped {
					out = append(out, [3]int{newIndex, previous, oldBase + oldOrd})
					return out
				}
				oldCursor = 0
				oldIndex = oldBase + oldOrd
				oldOrd = 0
				fetchedNew = !fetchedNew
				oldWrapped = true
			}
		} else {
			previous = newIndex
			havePrev = true
			newIndex = -1
			newCursor++
			if newCursor == len(newRing) {
				if oldWrapped {
					out = append(out, [3]int{previous, oldIndex, newBase + newOrd})
					return out
				}
				newCursor = 0
				newIndex = newBase + newOrd
				newOrd = 0
				fetchedNew = !fetchedNew
				newWrapped = true
			}
		}
		if oldWrapped && newWrapped {
			return out
		}
	}
	return out
}

type seamVertex struct {
	idx    [3]uint16
	u, v   uint8
	active bool
}
type seamRing struct {
	verts    []seamVertex
	active   bool
	selFlags uint8
}
type seamBody struct {
	rings     []seamRing
	records   []seamRec
	flags     uint32
	otherPart uint8
	wrapU     uint8
}

func prepareSeam(b *seamBody, owner int) {
	if b.flags&0x80 == 0 || len(b.records) == 0 {
		return
	}
	for a, r := range b.records {
		if int(r.bodyPart) != owner || r.u != 0 {
			continue
		}
		b.records = append(append([]seamRec{}, b.records[a:]...), b.records[:a]...)
		if b.records[0].ring == 0 {
			b.flags |= 0x100
		} else {
			b.flags |= 0x200
		}
		return
	}
}

func advanceRing(actorFlags uint32, lodStep int, b *seamBody, selector, boundarySelector int) int {
	if actorFlags&0x100 != 0 {
		return selector + lodStep
	}
	result := selector + lodStep
	scan := (selector >> 16) + 1
	target := result >> 16
	boundary := int(int16(boundarySelector >> 16))
	for scan < target && scan < boundary {
		if b.rings[scan].selFlags&1 != 0 {
			result = scan << 16
			target = scan
		} else {
			b.rings[scan].selFlags &^= 2
		}
		scan++
	}
	return result
}

func selectRings(b *seamBody, actorFlags uint32, lodStep int) []int {
	ringCount := len(b.rings)
	boundary := ringCount << 16
	selector := 0
	if b.flags&0x80 != 0 && actorFlags&0x1400 == 0x1400 {
		if b.flags&0x100 != 0 {
			selector = advanceRing(actorFlags, lodStep, b, 0, boundary)
		} else if b.flags&0x200 != 0 {
			c := boundary - lodStep
			if c >= 0x20000 {
				boundary = c
			} else {
				boundary = ringCount << 16
			}
		}
	}
	var out []int
	for guard := 0; guard < 100000; guard++ {
		if selector >= boundary {
			if boundary != ringCount<<16 {
				break
			}
			if (selector-lodStep)>>16 == ringCount-1 {
				break
			}
			selector = (ringCount - 1) << 16
		}
		out = append(out, selector>>16)
		selector = advanceRing(actorFlags, lodStep, b, selector, boundary)
	}
	return out
}

func selectVertices(flags []uint16, vertexStep int, explicit bool, actorFlags uint32) []int {
	end := len(flags) << 16
	step := vertexStep
	if explicit {
		step = 0x10000
	}
	if end/step < 3 {
		step = end/3 + 1
	}
	var out []int
	selector := 0
	for selector < end {
		cur := selector >> 16
		flags[cur] |= 0x100
		out = append(out, cur)
		next := selector + step
		skippedFixed := selector + 0x10000
		for skipped := cur + 1; skipped < next>>16 && skippedFixed < end; skipped, skippedFixed = skipped+1, skippedFixed+0x10000 {
			if actorFlags&0x4000 == 0 && flags[skipped]&0x80 != 0 {
				flags[skipped] |= 0x100
				out = append(out, skipped)
			} else {
				flags[skipped] &= 0xfeff
			}
		}
		selector = next
	}
	return out
}

func buildCap(actorFlags, bodyFlags uint32, first bool, base, emitted int) [][3]int {
	var faces [][3]int
	bit := uint32(2)
	if first {
		bit = 1
	}
	if actorFlags&0x402 == 0 || actorFlags&0x800 == 0 || bodyFlags&bit == 0 || emitted < 3 {
		return faces
	}
	for k := 1; k < emitted-1; k++ {
		if first {
			faces = append(faces, [3]int{base, base + k + 1, base + k})
		} else {
			faces = append(faces, [3]int{base, base + k, base + k + 1})
		}
	}
	return faces
}

type outVertex struct {
	refs []int // position indices (1..3)
	u, v float64
}

type outPart struct {
	verts []outVertex
	faces [][3]int
}

func buildSeam(bodies []seamBody, owner, firstRing, lastRing int) ([]outVertex, [][3]int) {
	body := bodies[owner]
	recs := body.records
	var verts []outVertex
	var faces [][3]int
	if len(recs) == 0 {
		return nil, nil
	}
	boundary := body.flags&0x80 != 0
	selRing := lastRing
	if body.flags&0x100 != 0 {
		selRing = firstRing
	}
	if boundary && selRing >= len(body.rings) {
		return nil, nil
	}
	var a, b int
	var wrappedA, wrappedB, haveA, haveB, havePrev bool
	var curA, curB, prev int
	scan := func(cursor *int, wrapped *bool, isOwner bool) (seamVertex, bool) {
		crossed := false
		for guard := 0; guard < 1000000; guard++ {
			r := recs[*cursor]
			part := int(body.otherPart)
			if isOwner {
				part = owner
			}
			if int(r.bodyPart) == part && part < len(bodies) {
				rg := int(r.ring)
				if isOwner && boundary {
					rg = selRing
				}
				if rg < len(bodies[part].rings) {
					src := bodies[part].rings[rg]
					if src.active && int(r.vertex) < len(src.verts) && src.verts[r.vertex].active {
						return src.verts[r.vertex], true
					}
				}
			}
			*cursor++
			if *cursor == len(recs) {
				*cursor = 0
				*wrapped = true
				if crossed {
					return seamVertex{}, false
				}
				crossed = true
			}
		}
		return seamVertex{}, false
	}
	refs := func(idx [3]uint16) []int {
		var r []int
		for _, i := range idx {
			if i != 0xffff {
				r = append(r, int(i))
			}
		}
		return r
	}
	for guard := 0; guard < 1000000; guard++ {
		if !haveA {
			src, ok := scan(&a, &wrappedA, true)
			if !ok {
				return verts, faces
			}
			curA = len(verts)
			u := float64(src.u)
			if wrappedA {
				u = float64(body.wrapU)
			}
			verts = append(verts, outVertex{refs(src.idx), u / 256, float64(src.v) / 256})
			haveA = true
		}
		if !haveB {
			src, ok := scan(&b, &wrappedB, false)
			if !ok {
				return verts, faces
			}
			curB = len(verts)
			r := recs[b]
			v := r.v
			if boundary {
				v = bodies[owner].rings[recs[a].ring].verts[recs[a].vertex].v
			}
			u := float64(r.u)
			if wrappedB {
				u = float64(body.wrapU)
			}
			verts = append(verts, outVertex{refs(src.idx), u / 256, float64(v) / 256})
			haveB = true
		}
		if havePrev {
			faces = append(faces, [3]int{curA, prev, curB})
			if wrappedA && wrappedB {
				return verts, faces
			}
		}
		if (a < b && !wrappedA) || wrappedB {
			prev = curA
			haveA = false
			a++
			if a == len(recs) {
				a = 0
				wrappedA = true
			}
		} else {
			prev = curB
			haveB = false
			b++
			if b == len(recs) {
				b = 0
				wrappedB = true
			}
		}
		havePrev = true
	}
	return verts, faces
}

// buildMesh assembles the triangle meshes of every body part at full detail.
func buildMesh(m *sxModel) []outPart {
	const actorFlags = 0x1c10
	const vertexStep, ringStep = 0x10000, 0x10000
	n := len(m.parts)
	bodies := make([]seamBody, n)
	for i, bp := range m.parts {
		var other, wrap uint8
		if bp.hasStrip {
			other, wrap = bp.strip0[0], bp.strip0[1]
		}
		bodies[i] = seamBody{records: append([]seamRec{}, bp.strips...), flags: bp.flags, otherPart: other, wrapU: wrap}
		prepareSeam(&bodies[i], i)
	}
	selRings := make([][]int, n)
	for i, bp := range m.parts {
		var rs []seamRing
		for _, rg := range bp.rings {
			rs = append(rs, seamRing{active: true, selFlags: rg.unknown})
		}
		bodies[i].rings = rs
		selRings[i] = selectRings(&bodies[i], actorFlags, ringStep)
	}
	selVerts := make([][][]int, n)
	emitted := map[int]bool{}
	for i, bp := range m.parts {
		selVerts[i] = make([][]int, len(bp.rings))
		for _, sr := range selRings[i] {
			ents := bp.rings[sr].entries
			fl := make([]uint16, len(ents))
			for k, e := range ents {
				fl[k] = e.flags & 0xfeff
			}
			sel := selectVertices(fl, vertexStep, bp.numExplicit != 0, actorFlags)
			selVerts[i][sr] = sel
			for _, k := range sel {
				emitted[ents[k].ptr] = true
			}
		}
	}
	// seam rings see every entry the seam records reference, even past the ring's end (as the original)
	seamLen := make([][]int, n)
	for i, bp := range m.parts {
		seamLen[i] = make([]int, len(bp.rings))
		for j, rg := range bp.rings {
			seamLen[i][j] = len(rg.entries)
		}
	}
	for owner, body := range bodies {
		for _, r := range body.records {
			if int(r.bodyPart) == owner && len(selRings[owner]) > 0 {
				rgi := int(r.ring)
				if body.flags&0x80 != 0 {
					if body.flags&0x100 != 0 {
						rgi = selRings[owner][0]
					} else {
						rgi = selRings[owner][len(selRings[owner])-1]
					}
				}
				if rgi >= 0 && rgi < len(m.parts[owner].rings) && int(r.vertex)+1 > seamLen[owner][rgi] {
					seamLen[owner][rgi] = int(r.vertex) + 1
				}
			}
			op := int(body.otherPart)
			if op < n && int(r.bodyPart) == op && int(r.ring) < len(m.parts[op].rings) && int(r.vertex)+1 > seamLen[op][r.ring] {
				seamLen[op][r.ring] = int(r.vertex) + 1
			}
		}
	}
	for i, bp := range m.parts {
		var rs []seamRing
		for j, rg := range bp.rings {
			var src []seamVertex
			for k := 0; k < seamLen[i][j]; k++ {
				var e ringEntry
				if k < len(rg.entries) {
					e = rg.entries[k]
				} else {
					p := rg.entryOff + 10*k
					if p+10 > len(m.raw) {
						break
					}
					d := m.raw
					e = ringEntry{idx: [3]uint16{u16(d, p), u16(d, p+2), u16(d, p+4)}, u: d[p+6], v: d[p+7], flags: u16(d, p+8), ptr: p}
				}
				act, ok := emitted[e.ptr]
				if !ok {
					act = e.flags&0x100 != 0
				}
				src = append(src, seamVertex{e.idx, e.u, e.v, act})
			}
			active := false
			for _, s := range selRings[i] {
				if s == j {
					active = true
				}
			}
			rs = append(rs, seamRing{src, active, rg.unknown})
		}
		bodies[i].rings = rs
	}
	refs := func(idx [3]uint16) []int {
		var r []int
		for _, i := range idx {
			if i != 0xffff {
				r = append(r, int(i))
			}
		}
		return r
	}
	parts := make([]outPart, n)
	for i, bp := range m.parts {
		var verts []outVertex
		vrt := make([][]int, len(bp.rings))
		for _, j := range selRings[i] {
			rg := bp.rings[j]
			sel := selVerts[i][j]
			vrt[j] = make([]int, len(sel)+1)
			for o, k := range sel {
				e := rg.entries[k]
				vrt[j][o] = len(verts)
				verts = append(verts, outVertex{refs(e.idx), float64(e.u) / 256, float64(e.v) / 256})
			}
			vrt[j][len(sel)] = len(verts)
			if len(sel) > 0 {
				c := verts[vrt[j][0]]
				c.u = float64(rg.texture) / 256
				verts = append(verts, c)
			}
		}
		var faces [][3]int
		for k := 0; k+1 < len(selRings[i]); k++ {
			j, nx := selRings[i][k], selRings[i][k+1]
			oe, ne := bp.rings[j].entries, bp.rings[nx].entries
			oldR := make([]ringVertex, len(oe))
			newR := make([]ringVertex, len(ne))
			for q, e := range oe {
				oldR[q] = ringVertex{e.flags, e.u, false}
			}
			for q, e := range ne {
				newR[q] = ringVertex{e.flags, e.u, false}
			}
			for _, q := range selVerts[i][j] {
				oldR[q].active = true
			}
			for _, q := range selVerts[i][nx] {
				newR[q].active = true
			}
			rf := buildRingFaces(oldR, newR, bp.rings[j].post, bp.rings[nx].post)
			base := vrt[j][0]
			for _, f := range rf {
				faces = append(faces, [3]int{f[0] + base, f[2] + base, f[1] + base})
			}
		}
		for _, first := range []bool{true, false} {
			if len(selRings[i]) == 0 {
				break
			}
			j := selRings[i][len(selRings[i])-1]
			if first {
				j = selRings[i][0]
			}
			if (first && j != 0) || (!first && j != len(bp.rings)-1) {
				continue
			}
			cnt := len(selVerts[i][j])
			if !first {
				cnt++
			}
			for _, f := range buildCap(actorFlags, bp.flags, first, vrt[j][0], cnt) {
				faces = append(faces, [3]int{f[0], f[2], f[1]})
			}
		}
		if len(bp.explicit) > 0 && len(vrt) == 1 {
			for _, ef := range bp.explicit {
				faces = append(faces, [3]int{vrt[0][ef[0]], vrt[0][ef[1]], vrt[0][ef[2]]})
			}
		}
		parts[i] = outPart{verts, faces}
	}
	for i := range bodies {
		if len(bodies[i].records) == 0 || len(selRings[i]) == 0 {
			continue
		}
		sv, sf := buildSeam(bodies, i, selRings[i][0], selRings[i][len(selRings[i])-1])
		base := len(parts[i].verts)
		parts[i].verts = append(parts[i].verts, sv...)
		for _, f := range sf {
			parts[i].faces = append(parts[i].faces, [3]int{f[0] + base, f[2] + base, f[1] + base})
		}
	}
	return parts
}
