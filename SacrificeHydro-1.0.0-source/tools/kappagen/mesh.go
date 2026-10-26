package main

// Signed-distance modelling and meshing for new Sacrifice creatures.
// Shapes are described in engine space (z up, +y forward), meshed with marching tetrahedra,
// skinned to the base creature's bones and UV-mapped per body part.

import (
	"math"
	"sort"
)

func (a v3) sub(b v3) v3     { return v3{a[0] - b[0], a[1] - b[1], a[2] - b[2]} }
func (a v3) dot(b v3) float64 { return a[0]*b[0] + a[1]*b[1] + a[2]*b[2] }
func (a v3) cross(b v3) v3 {
	return v3{a[1]*b[2] - a[2]*b[1], a[2]*b[0] - a[0]*b[2], a[0]*b[1] - a[1]*b[0]}
}
func (a v3) len() float64 { return math.Sqrt(a.dot(a)) }
func (a v3) norm() v3 {
	l := a.len()
	if l == 0 {
		return v3{0, 0, 1}
	}
	return a.mul(1 / l)
}
func lerp3(a, b v3, t float64) v3 { return a.add(b.sub(a).mul(t)) }
func clamp(x, a, b float64) float64 {
	return math.Max(a, math.Min(b, x))
}

type bw struct {
	bone int
	w    float64
}

type shapeKind int

const (
	capsule shapeKind = iota
	ellipsoid
	torusZ
	discZ
)

// shape is one primitive. Its bones say which skeleton bones move it.
type shape struct {
	kind   shapeKind
	a, b   v3      // capsule ends
	ra, rb float64 // capsule radii at a and b
	c      v3      // ellipsoid / torus / disc centre
	r      v3      // ellipsoid radii; torus: r[0]=major, r[1]=minor; disc: r[0]=radius, r[1]=half height
	rot    *[3]v3  // optional local axes (rows) for the ellipsoid
	clipN  v3      // optional half-space: keep dot(clipN,p) <= clipD
	clipD  float64
	clip   bool
	sub    []shape // carved out of this shape
	noise  float64 // surface roughness amplitude
	nfreq  float64
	bones  []bw
	zBlend *[3]float64 // optional: bones[0] below z0, bones[1] above z1, blended between
	grad   bool        // capsule: bones[0] at a, bones[1] at b, blended along the length
	span   []spanKey   // skinned by |x| (distance from the middle): blends from key to key
}

func (s *shape) sd(p v3) float64 {
	var d float64
	switch s.kind {
	case capsule:
		ab := s.b.sub(s.a)
		t := clamp(p.sub(s.a).dot(ab)/ab.dot(ab), 0, 1)
		q := lerp3(s.a, s.b, t)
		d = p.sub(q).len() - (s.ra + (s.rb-s.ra)*t)
	case ellipsoid:
		q := p.sub(s.c)
		if s.rot != nil {
			q = v3{s.rot[0].dot(q), s.rot[1].dot(q), s.rot[2].dot(q)}
		}
		k0 := v3{q[0] / s.r[0], q[1] / s.r[1], q[2] / s.r[2]}.len()
		k1 := v3{q[0] / (s.r[0] * s.r[0]), q[1] / (s.r[1] * s.r[1]), q[2] / (s.r[2] * s.r[2])}.len()
		if k1 == 0 {
			d = -math.Min(s.r[0], math.Min(s.r[1], s.r[2]))
		} else {
			d = k0 * (k0 - 1) / k1
		}
	case torusZ:
		q := p.sub(s.c)
		x := math.Hypot(q[0], q[1]) - s.r[0]
		d = math.Hypot(x, q[2]) - s.r[1]
	case discZ:
		q := p.sub(s.c)
		dx := math.Hypot(q[0], q[1]) - s.r[0]
		dz := math.Abs(q[2]) - s.r[1]
		d = math.Min(math.Max(dx, dz), 0) + math.Hypot(math.Max(dx, 0), math.Max(dz, 0))
	}
	if s.clip {
		d = math.Max(d, s.clipN.dot(p)-s.clipD)
	}
	for i := range s.sub {
		d = math.Max(d, -s.sub[i].sd(p))
	}
	if s.noise != 0 {
		d += s.noise * vnoise(p.mul(s.nfreq))
	}
	return d
}

type spanKey struct {
	x    float64
	bone int
}

func (s *shape) boneWeights(p v3) []bw {
	if len(s.span) > 0 {
		x := math.Abs(p[0])
		k := s.span
		if x <= k[0].x {
			return []bw{{k[0].bone, 1}}
		}
		for i := 0; i+1 < len(k); i++ {
			if x <= k[i+1].x {
				t := (x - k[i].x) / (k[i+1].x - k[i].x)
				t = t * t * (3 - 2*t)
				return []bw{{k[i].bone, 1 - t}, {k[i+1].bone, t}}
			}
		}
		return []bw{{k[len(k)-1].bone, 1}}
	}
	if s.grad && s.kind == capsule && len(s.bones) == 2 {
		ab := s.b.sub(s.a)
		t := clamp(p.sub(s.a).dot(ab)/ab.dot(ab), 0, 1)
		t = t * t * (3 - 2*t)
		return []bw{{s.bones[0].bone, 1 - t}, {s.bones[1].bone, t}}
	}
	if s.zBlend == nil {
		return s.bones
	}
	z0, z1 := s.zBlend[0], s.zBlend[1]
	t := clamp((p[2]-z0)/(z1-z0), 0, 1)
	t = t * t * (3 - 2*t)
	return []bw{{s.bones[0].bone, 1 - t}, {s.bones[1].bone, t}}
}

func smin(a, b, k float64) float64 {
	if k <= 0 {
		return math.Min(a, b)
	}
	h := clamp(0.5+0.5*(b-a)/k, 0, 1)
	return b + (a-b)*h - k*h*(1-h)
}

// --- value noise -----------------------------------------------------------------------------

func hash3(x, y, z int) float64 {
	h := uint32(x)*374761393 + uint32(y)*668265263 + uint32(z)*2147483647
	h = (h ^ (h >> 13)) * 1274126177
	h ^= h >> 16
	return float64(h&0xffffff)/float64(0xffffff)*2 - 1
}

func vnoise(p v3) float64 {
	x0, y0, z0 := math.Floor(p[0]), math.Floor(p[1]), math.Floor(p[2])
	fx, fy, fz := p[0]-x0, p[1]-y0, p[2]-z0
	sx, sy, sz := fx*fx*(3-2*fx), fy*fy*(3-2*fy), fz*fz*(3-2*fz)
	ix, iy, iz := int(x0), int(y0), int(z0)
	l := func(a, b, t float64) float64 { return a + (b-a)*t }
	c := func(dx, dy, dz int) float64 { return hash3(ix+dx, iy+dy, iz+dz) }
	return l(l(l(c(0, 0, 0), c(1, 0, 0), sx), l(c(0, 1, 0), c(1, 1, 0), sx), sy),
		l(l(c(0, 0, 1), c(1, 0, 1), sx), l(c(0, 1, 1), c(1, 1, 1), sx), sy), sz)
}

// --- body parts ------------------------------------------------------------------------------

type uvFunc func(p, n v3) (u, v float64, group int)

type bodyDef struct {
	name   string
	shapes []shape
	blend  float64 // smooth-union radius
	cell   float64 // grid cell size
	target int     // vertex budget after simplification
	uv     uvFunc
	tex    func(w, h int) [][3]uint8
}

func (b *bodyDef) sd(p v3) float64 {
	d := b.shapes[0].sd(p)
	for i := 1; i < len(b.shapes); i++ {
		d = smin(d, b.shapes[i].sd(p), b.blend)
	}
	return d
}

func (b *bodyDef) grad(p v3) v3 {
	const e = 1e-4
	return v3{
		b.sd(p.add(v3{e, 0, 0})) - b.sd(p.sub(v3{e, 0, 0})),
		b.sd(p.add(v3{0, e, 0})) - b.sd(p.sub(v3{0, e, 0})),
		b.sd(p.add(v3{0, 0, e})) - b.sd(p.sub(v3{0, 0, e})),
	}.mul(1 / (2 * e))
}

func (b *bodyDef) bounds() (v3, v3) {
	lo, hi := v3{1e9, 1e9, 1e9}, v3{-1e9, -1e9, -1e9}
	grow := func(p v3, r float64) {
		for k := 0; k < 3; k++ {
			lo[k] = math.Min(lo[k], p[k]-r)
			hi[k] = math.Max(hi[k], p[k]+r)
		}
	}
	for _, s := range b.shapes {
		switch s.kind {
		case capsule:
			grow(s.a, s.ra)
			grow(s.b, s.rb)
		case ellipsoid:
			grow(s.c, math.Max(s.r[0], math.Max(s.r[1], s.r[2])))
		case torusZ:
			grow(s.c, s.r[0]+s.r[1])
		case discZ:
			grow(s.c, math.Max(s.r[0], s.r[1]))
		}
	}
	m := b.blend + 3*b.cell
	return lo.sub(v3{m, m, m}), hi.add(v3{m, m, m})
}

type meshVert struct {
	p, n  v3
	bones []bw
}

type mesh struct {
	verts []meshVert
	faces [][3]int // outward winding: (b-a)x(c-a) points along the outside normal (engine space)
}

// polygonize meshes the zero level set of the part with marching tetrahedra.
func (b *bodyDef) polygonize() *mesh {
	lo, hi := b.bounds()
	nx := int(math.Ceil((hi[0]-lo[0])/b.cell)) + 1
	ny := int(math.Ceil((hi[1]-lo[1])/b.cell)) + 1
	nz := int(math.Ceil((hi[2]-lo[2])/b.cell)) + 1
	idx := func(i, j, k int) int { return (k*ny+j)*nx + i }
	pos := func(i, j, k int) v3 {
		return v3{lo[0] + float64(i)*b.cell, lo[1] + float64(j)*b.cell, lo[2] + float64(k)*b.cell}
	}
	f := make([]float64, nx*ny*nz)
	for k := 0; k < nz; k++ {
		for j := 0; j < ny; j++ {
			for i := 0; i < nx; i++ {
				f[idx(i, j, k)] = b.sd(pos(i, j, k))
			}
		}
	}
	m := &mesh{}
	cache := map[[2]int]int{}
	corner := [8][3]int{{0, 0, 0}, {1, 0, 0}, {1, 1, 0}, {0, 1, 0}, {0, 0, 1}, {1, 0, 1}, {1, 1, 1}, {0, 1, 1}}
	tets := [6][4]int{{0, 5, 1, 6}, {0, 1, 2, 6}, {0, 2, 3, 6}, {0, 3, 7, 6}, {0, 7, 4, 6}, {0, 4, 5, 6}}
	vertex := func(ia, ib int, pa, pb v3) int {
		key := [2]int{ia, ib}
		if ia > ib {
			key = [2]int{ib, ia}
		}
		if v, ok := cache[key]; ok {
			return v
		}
		fa, fb := f[ia], f[ib]
		t := fa / (fa - fb)
		p := lerp3(pa, pb, t)
		for it := 0; it < 3; it++ { // pull onto the surface
			g := b.grad(p)
			gg := g.dot(g)
			if gg < 1e-12 {
				break
			}
			p = p.sub(g.mul(b.sd(p) / gg))
		}
		m.verts = append(m.verts, meshVert{p: p, n: b.grad(p).norm()})
		cache[key] = len(m.verts) - 1
		return len(m.verts) - 1
	}
	for k := 0; k < nz-1; k++ {
		for j := 0; j < ny-1; j++ {
			for i := 0; i < nx-1; i++ {
				var ci [8]int
				var cp [8]v3
				in := 0
				for c := 0; c < 8; c++ {
					ci[c] = idx(i+corner[c][0], j+corner[c][1], k+corner[c][2])
					cp[c] = pos(i+corner[c][0], j+corner[c][1], k+corner[c][2])
					if f[ci[c]] < 0 {
						in++
					}
				}
				if in == 0 || in == 8 {
					continue
				}
				for _, t := range tets {
					var inside, outside []int
					for _, c := range t {
						if f[ci[c]] < 0 {
							inside = append(inside, c)
						} else {
							outside = append(outside, c)
						}
					}
					var tri [][3]int
					switch len(inside) {
					case 1, 3:
						var a int
						var others []int
						if len(inside) == 1 {
							a, others = inside[0], outside
						} else {
							a, others = outside[0], inside
						}
						var v [3]int
						for q, o := range others {
							v[q] = vertex(ci[a], ci[o], cp[a], cp[o])
						}
						tri = append(tri, v)
					case 2:
						a0, a1 := inside[0], inside[1]
						b0, b1 := outside[0], outside[1]
						v00 := vertex(ci[a0], ci[b0], cp[a0], cp[b0])
						v01 := vertex(ci[a0], ci[b1], cp[a0], cp[b1])
						v10 := vertex(ci[a1], ci[b0], cp[a1], cp[b0])
						v11 := vertex(ci[a1], ci[b1], cp[a1], cp[b1])
						tri = append(tri, [3]int{v00, v01, v11}, [3]int{v00, v11, v10})
					}
					for _, tr := range tri {
						if tr[0] == tr[1] || tr[1] == tr[2] || tr[0] == tr[2] {
							continue
						}
						pa, pb, pc := m.verts[tr[0]].p, m.verts[tr[1]].p, m.verts[tr[2]].p
						cr := pb.sub(pa).cross(pc.sub(pa))
						if cr.len() < 1e-10 {
							continue
						}
						nn := m.verts[tr[0]].n.add(m.verts[tr[1]].n).add(m.verts[tr[2]].n)
						if cr.dot(nn) < 0 {
							tr[1], tr[2] = tr[2], tr[1]
						}
						m.faces = append(m.faces, tr)
					}
				}
			}
		}
	}
	return m
}

// skin assigns up to three bone influences per vertex from the shapes nearest to it.
func (b *bodyDef) skin(m *mesh) {
	for vi := range m.verts {
		p := m.verts[vi].p
		ds := make([]float64, len(b.shapes))
		dmin := math.Inf(1)
		for i := range b.shapes {
			ds[i] = b.shapes[i].sd(p)
			dmin = math.Min(dmin, ds[i])
		}
		acc := map[int]float64{}
		for i := range b.shapes {
			if ds[i]-dmin > 0.06 {
				continue
			}
			w := math.Exp(-(ds[i] - dmin) / 0.015)
			for _, x := range b.shapes[i].boneWeights(p) {
				acc[x.bone] += w * x.w
			}
		}
		var list []bw
		for bn, w := range acc {
			if w > 1e-4 {
				list = append(list, bw{bn, w})
			}
		}
		sort.Slice(list, func(i, j int) bool { return list[i].w > list[j].w })
		if len(list) > 3 {
			list = list[:3]
		}
		t := 0.0
		for _, x := range list {
			t += x.w
		}
		for i := range list {
			list[i].w /= t
		}
		m.verts[vi].bones = list
	}
}

// --- UV helpers ------------------------------------------------------------------------------

// boxUV projects each face on the plane facing its normal; the texture must tile.
func boxUV(lo v3, span float64) uvFunc {
	return func(p, n v3) (float64, float64, int) {
		ax := 0
		an := math.Abs(n[0])
		if math.Abs(n[1]) > an {
			ax, an = 1, math.Abs(n[1])
		}
		if math.Abs(n[2]) > an {
			ax = 2
		}
		q := p.sub(lo).mul(1 / span)
		switch ax {
		case 0:
			return clamp(q[1], 0, 0.999), clamp(q[2], 0, 0.999), 0
		case 1:
			return clamp(q[0], 0, 0.999), clamp(q[2], 0, 0.999), 1
		default:
			return clamp(q[0], 0, 0.999), clamp(q[1], 0, 0.999), 2
		}
	}
}

func frac(x float64) float64 { return x - math.Floor(x) }

// planarUV maps the part's extent in two axes onto the whole texture.
func planarUV(lo, hi v3, ax1, ax2 int, flip1 bool) uvFunc {
	return func(p, n v3) (float64, float64, int) {
		u := (p[ax1] - lo[ax1]) / (hi[ax1] - lo[ax1])
		v := (p[ax2] - lo[ax2]) / (hi[ax2] - lo[ax2])
		if flip1 {
			u = 1 - u
		}
		return clamp(u, 0, 0.999), clamp(1-v, 0, 0.999), 0
	}
}
