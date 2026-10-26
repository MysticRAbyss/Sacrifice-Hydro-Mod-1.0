package main

// Quadric-error edge-collapse simplification (Garland & Heckbert), used to bring the
// marching-tetrahedra meshes down to game-sized vertex counts.

import (
	"container/heap"
	"math"
)

type quadric [10]float64 // symmetric 4x4: a2 ab ac ad b2 bc bd c2 cd d2

func planeQ(n v3, d float64) quadric {
	a, b, c := n[0], n[1], n[2]
	return quadric{a * a, a * b, a * c, a * d, b * b, b * c, b * d, c * c, c * d, d * d}
}
func (q *quadric) add(o quadric) {
	for i := range q {
		q[i] += o[i]
	}
}
func (q quadric) eval(p v3) float64 {
	x, y, z := p[0], p[1], p[2]
	return q[0]*x*x + 2*q[1]*x*y + 2*q[2]*x*z + 2*q[3]*x + q[4]*y*y + 2*q[5]*y*z + 2*q[6]*y + q[7]*z*z + 2*q[8]*z + q[9]
}
func (q quadric) optimum() (v3, bool) {
	a := [3][3]float64{{q[0], q[1], q[2]}, {q[1], q[4], q[5]}, {q[2], q[5], q[7]}}
	bv := v3{-q[3], -q[6], -q[8]}
	det := a[0][0]*(a[1][1]*a[2][2]-a[1][2]*a[2][1]) - a[0][1]*(a[1][0]*a[2][2]-a[1][2]*a[2][0]) + a[0][2]*(a[1][0]*a[2][1]-a[1][1]*a[2][0])
	if math.Abs(det) < 1e-12 {
		return v3{}, false
	}
	solve := func(col int) float64 {
		m := a
		for r := 0; r < 3; r++ {
			m[r][col] = bv[r]
		}
		return (m[0][0]*(m[1][1]*m[2][2]-m[1][2]*m[2][1]) - m[0][1]*(m[1][0]*m[2][2]-m[1][2]*m[2][0]) + m[0][2]*(m[1][0]*m[2][1]-m[1][1]*m[2][0])) / det
	}
	return v3{solve(0), solve(1), solve(2)}, true
}

type collapse struct {
	a, b    int
	cost    float64
	p       v3
	version [2]int
	index   int
}
type cheap []*collapse

func (h cheap) Len() int            { return len(h) }
func (h cheap) Less(i, j int) bool  { return h[i].cost < h[j].cost }
func (h cheap) Swap(i, j int)       { h[i], h[j] = h[j], h[i] }
func (h *cheap) Push(x interface{}) { *h = append(*h, x.(*collapse)) }
func (h *cheap) Pop() interface{} {
	old := *h
	x := old[len(old)-1]
	*h = old[:len(old)-1]
	return x
}

// decimate reduces m to about target vertices, then snaps the vertices back onto the surface.
func decimate(m *mesh, target int, surf *bodyDef) {
	nv := len(m.verts)
	if nv <= target {
		return
	}
	pos := make([]v3, nv)
	for i := range pos {
		pos[i] = m.verts[i].p
	}
	Q := make([]quadric, nv)
	vf := make([]map[int]bool, nv)
	for i := range vf {
		vf[i] = map[int]bool{}
	}
	faces := make([][3]int, len(m.faces))
	alive := make([]bool, len(m.faces))
	copy(faces, m.faces)
	for fi, f := range faces {
		alive[fi] = true
		n := pos[f[1]].sub(pos[f[0]]).cross(pos[f[2]].sub(pos[f[0]]))
		area := n.len()
		if area == 0 {
			continue
		}
		n = n.mul(1 / area)
		q := planeQ(n, -n.dot(pos[f[0]]))
		for k := range q {
			q[k] *= area
		}
		for _, v := range f {
			Q[v].add(q)
			vf[v][fi] = true
		}
	}
	version := make([]int, nv)
	dead := make([]bool, nv)
	h := &cheap{}
	push := func(a, b int) {
		q := Q[a]
		q.add(Q[b])
		p, ok := q.optimum()
		mid := pos[a].add(pos[b]).mul(0.5)
		if !ok || p.sub(mid).len() > pos[a].sub(pos[b]).len()*2 {
			p = mid
		}
		heap.Push(h, &collapse{a: a, b: b, cost: q.eval(p), p: p, version: [2]int{version[a], version[b]}})
	}
	seen := map[[2]int]bool{}
	for _, f := range faces {
		for k := 0; k < 3; k++ {
			a, b := f[k], f[(k+1)%3]
			if a > b {
				a, b = b, a
			}
			if !seen[[2]int{a, b}] {
				seen[[2]int{a, b}] = true
				push(a, b)
			}
		}
	}
	live := nv
	for live > target && h.Len() > 0 {
		c := heap.Pop(h).(*collapse)
		a, b := c.a, c.b
		if dead[a] || dead[b] || version[a] != c.version[0] || version[b] != c.version[1] {
			continue
		}
		// reject collapses that flip a face
		bad := false
		for _, v := range []int{a, b} {
			for fi := range vf[v] {
				if !alive[fi] {
					continue
				}
				f := faces[fi]
				if (f[0] == a || f[1] == a || f[2] == a) && (f[0] == b || f[1] == b || f[2] == b) {
					continue
				}
				var np [3]v3
				for k := 0; k < 3; k++ {
					np[k] = pos[f[k]]
					if f[k] == a || f[k] == b {
						np[k] = c.p
					}
				}
				n0 := pos[f[1]].sub(pos[f[0]]).cross(pos[f[2]].sub(pos[f[0]]))
				n1 := np[1].sub(np[0]).cross(np[2].sub(np[0]))
				if n0.dot(n1) <= 0.2*n0.len()*n1.len() {
					bad = true
					break
				}
			}
			if bad {
				break
			}
		}
		if bad {
			continue
		}
		// collapse b into a
		pos[a] = c.p
		Q[a].add(Q[b])
		dead[b] = true
		live--
		for fi := range vf[b] {
			if !alive[fi] {
				continue
			}
			f := &faces[fi]
			for k := 0; k < 3; k++ {
				if f[k] == b {
					f[k] = a
				}
			}
			if f[0] == f[1] || f[1] == f[2] || f[0] == f[2] {
				alive[fi] = false
				continue
			}
			vf[a][fi] = true
		}
		version[a]++
		nb := map[int]bool{}
		for fi := range vf[a] {
			if !alive[fi] {
				delete(vf[a], fi)
				continue
			}
			for _, v := range faces[fi] {
				if v != a {
					nb[v] = true
				}
			}
		}
		for v := range nb {
			x, y := a, v
			if x > y {
				x, y = y, x
			}
			push(x, y)
		}
	}
	// rebuild
	remap := make([]int, nv)
	var verts []meshVert
	for i := 0; i < nv; i++ {
		remap[i] = -1
		if dead[i] {
			continue
		}
		p := pos[i]
		for it := 0; it < 3; it++ {
			g := surf.grad(p)
			gg := g.dot(g)
			if gg < 1e-12 {
				break
			}
			p = p.sub(g.mul(surf.sd(p) / gg))
		}
		remap[i] = len(verts)
		verts = append(verts, meshVert{p: p, n: surf.grad(p).norm()})
	}
	var out [][3]int
	for fi, f := range faces {
		if !alive[fi] {
			continue
		}
		out = append(out, [3]int{remap[f[0]], remap[f[1]], remap[f[2]]})
	}
	m.verts, m.faces = verts, out
}
