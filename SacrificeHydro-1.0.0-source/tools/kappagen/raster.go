package main

// A small software renderer used to check generated creatures (read back from the written
// SXMD/SAXS/SXTX files) and to draw the spell-bar icon.

import (
	"image"
	"image/color"
	"math"
)

func qrot(q [4]float64, v v3) v3 {
	u := v3{q[0], q[1], q[2]}
	s := q[3]
	return u.mul(2 * u.dot(v)).add(v.mul(s*s - u.dot(u))).add(u.cross(v).mul(2 * s))
}

func qmul(a, b [4]float64) [4]float64 {
	return [4]float64{
		a[3]*b[0] + a[0]*b[3] + a[1]*b[2] - a[2]*b[1],
		a[3]*b[1] - a[0]*b[2] + a[1]*b[3] + a[2]*b[0],
		a[3]*b[2] + a[0]*b[1] - a[1]*b[0] + a[2]*b[3],
		a[3]*b[3] - a[0]*b[0] - a[1]*b[1] - a[2]*b[2],
	}
}

// pose returns posed vertex positions per part (frame nil = rest pose)
// boneTransforms gives every bone's world position and rotation in a frame (nil = rest pose)
func boneTransforms(cr *creature, fr *animFrame) ([]v3, [][4]float64) {
	nb := len(cr.bones)
	rot := make([][4]float64, nb)
	pos := make([]v3, nb)
	rot[0] = [4]float64{0, 0, 0, 1}
	if fr != nil && len(fr.rot) == nb {
		rot[0] = fr.rot[0]
		pos[0] = fr.disp
	}
	for i := 1; i < nb; i++ {
		p := cr.bones[i].parent
		r := [4]float64{0, 0, 0, 1}
		if fr != nil && len(fr.rot) == nb {
			r = fr.rot[i]
		}
		rot[i] = qmul(rot[p], r)
		pos[i] = pos[p].add(qrot(rot[p], cr.bones[i].local))
	}
	if fr != nil && len(cr.noDisp) > 0 {
		for i := range pos {
			if cr.noDisp[i] {
				pos[i] = pos[i].sub(fr.disp)
			}
		}
	}
	// the engine's extra static "ground" bone (index = number of bones): never moves
	return append(pos, v3{}), append(rot, [4]float64{0, 0, 0, 1})
}

// pose returns posed vertex positions per part (frame nil = rest pose), skinned like the engine
func pose(cr *creature, fr *animFrame) [][]v3 {
	pos, rot := boneTransforms(cr, fr)
	out := make([][]v3, len(cr.parts))
	for pi, pt := range cr.parts {
		for _, v := range pt.verts {
			var acc v3
			for k := 0; k < 4; k++ {
				if v.weights[k] == 0 {
					continue
				}
				b := v.joints[k]
				acc = acc.add(pos[b].add(qrot(rot[b], v.offs[k])).mul(v.weights[k]))
			}
			out[pi] = append(out[pi], acc)
		}
	}
	return out
}

type camera struct {
	yaw, pitch float64 // radians; yaw 0 looks at the creature's front
	center     v3
	scale      float64 // pixels per unit
}

func renderCreature(cr *creature, fr *animFrame, cam camera, w, h int, bg color.NRGBA) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for i := range img.Pix {
		img.Pix[i] = []uint8{bg.R, bg.G, bg.B, bg.A}[i%4]
	}
	zb := make([]float64, w*h)
	for i := range zb {
		zb[i] = math.Inf(1)
	}
	// view basis: camera in front (+y) of the creature looking back towards -y
	cy, sy := math.Cos(cam.yaw), math.Sin(cam.yaw)
	cp, sp := math.Cos(cam.pitch), math.Sin(cam.pitch)
	fwd := v3{-sy * cp, -cy * cp, -sp} // from camera into the scene
	right := v3{cy, -sy, 0}.mul(-1)
	up := right.cross(fwd)
	light := v3{0.4, 0.8, 0.6}.norm()
	project := func(p v3) (float64, float64, float64) {
		q := p.sub(cam.center)
		return float64(w)/2 + q.dot(right)*cam.scale, float64(h)/2 - q.dot(up)*cam.scale, q.dot(fwd)
	}
	posed := pose(cr, fr)
	for pi, pt := range cr.parts {
		tex := pt.tex
		tw, th := tex.Bounds().Dx(), tex.Bounds().Dy()
		for _, f := range pt.faces {
			var sx, sy2, sz [3]float64
			var uv [3][2]float64
			for k := 0; k < 3; k++ {
				sx[k], sy2[k], sz[k] = project(posed[pi][f[k]])
				uv[k] = pt.verts[f[k]].uv
			}
			a, b, c := posed[pi][f[0]], posed[pi][f[1]], posed[pi][f[2]]
			n := b.sub(a).cross(c.sub(a)).norm()
			if n.dot(fwd) > 0 { // back face: culled, as in the game
				continue
			}
			shade := 0.45 + 0.65*math.Max(0, n.dot(light))
			minx := int(math.Max(0, math.Floor(math.Min(sx[0], math.Min(sx[1], sx[2])))))
			maxx := int(math.Min(float64(w-1), math.Ceil(math.Max(sx[0], math.Max(sx[1], sx[2])))))
			miny := int(math.Max(0, math.Floor(math.Min(sy2[0], math.Min(sy2[1], sy2[2])))))
			maxy := int(math.Min(float64(h-1), math.Ceil(math.Max(sy2[0], math.Max(sy2[1], sy2[2])))))
			area := (sx[1]-sx[0])*(sy2[2]-sy2[0]) - (sx[2]-sx[0])*(sy2[1]-sy2[0])
			if math.Abs(area) < 1e-9 {
				continue
			}
			for y := miny; y <= maxy; y++ {
				for x := minx; x <= maxx; x++ {
					px, py := float64(x)+0.5, float64(y)+0.5
					w0 := ((sx[1]-px)*(sy2[2]-py) - (sx[2]-px)*(sy2[1]-py)) / area
					w1 := ((sx[2]-px)*(sy2[0]-py) - (sx[0]-px)*(sy2[2]-py)) / area
					w2 := 1 - w0 - w1
					if w0 < 0 || w1 < 0 || w2 < 0 {
						continue
					}
					z := w0*sz[0] + w1*sz[1] + w2*sz[2]
					if z >= zb[y*w+x] {
						continue
					}
					zb[y*w+x] = z
					u := w0*uv[0][0] + w1*uv[1][0] + w2*uv[2][0]
					v := w0*uv[0][1] + w1*uv[1][1] + w2*uv[2][1]
					tx := int(clamp(u*float64(tw), 0, float64(tw-1)))
					ty := int(clamp(v*float64(th), 0, float64(th-1)))
					col := tex.NRGBAAt(tx, ty)
					img.SetNRGBA(x, y, color.NRGBA{
						uint8(clamp(float64(col.R)*shade, 0, 255)),
						uint8(clamp(float64(col.G)*shade, 0, 255)),
						uint8(clamp(float64(col.B)*shade, 0, 255)), 255})
				}
			}
		}
	}
	return img
}
