package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"os"
)

// frameStrip renders n evenly spaced frames of each listed animation (one row per animation)
// and reports the lowest point of every part over all frames.
func frameStrip(cr *creature, anims []string, n int, yaw float64, path string) {
	const W, H = 300, 300
	var fr0 *animFrame
	if a := findAnim(cr, "stance"); a != nil && len(a.frames) > 0 {
		fr0 = &a.frames[0]
	}
	lo0, hi0 := bounds(cr, fr0)
	size := math.Max(hi0[0]-lo0[0], math.Max(hi0[1]-lo0[1], hi0[2]-lo0[2]))
	sheet := image.NewNRGBA(image.Rect(0, 0, W*n, H*len(anims)))
	for r, name := range anims {
		a := findAnim(cr, name)
		if a == nil || len(a.frames) == 0 {
			fmt.Println("no anim", name)
			continue
		}
		minz := make([]float64, len(cr.parts))
		for i := range minz {
			minz[i] = 99
		}
		for fi := range a.frames {
			for pi, pp := range pose(cr, &a.frames[fi]) {
				for _, p := range pp {
					minz[pi] = math.Min(minz[pi], p[2])
				}
			}
		}
		fmt.Printf("  %-10s %3d frames, lowest z per part %.2f\n", name, len(a.frames), minz)
		for i := 0; i < n; i++ {
			fr := &a.frames[i*(len(a.frames)-1)/max(1, n-1)]
			lo, hi := bounds(cr, fr)
			c := lo.add(hi).mul(0.5)
			c[2] = (lo0[2] + hi0[2]) / 2
			pitch := 0.25
			if ps := os.Getenv("STRIP_PITCH"); ps != "" {
				fmt.Sscan(ps, &pitch)
			}
			sc := 0.75 * W / size
			if fs := os.Getenv("STRIP_FOCUS"); fs != "" { // x,y,z,zoom
				var z float64
				fmt.Sscanf(fs, "%f,%f,%f,%f", &c[0], &c[1], &c[2], &z)
				sc *= z
			}
			img := renderCreature(cr, fr, camera{yaw: yaw, pitch: pitch, center: c, scale: sc}, W, H, color.NRGBA{58, 70, 86, 255})
			// ground line
			draw.Draw(sheet, image.Rect(W*i, H*r, W*(i+1), H*(r+1)), img, image.Point{}, draw.Src)
		}
	}
	writePNG(path, sheet)
}

func dumpRoot(cr *creature, name string) {
	a := findAnim(cr, name)
	for i := 0; i < len(a.frames); i += len(a.frames) / 8 {
		pos, rot := boneTransforms(cr, &a.frames[i])
		fmt.Printf("  %s %3d root %.2f %.2f  b1 %.2f\n", name, i, pos[0], rot[0], pos[1])
	}
}

// poseBlend holds a pose (some bones' rotations from one frame of another animation) through the
// middle of an animation, easing in and out: the same as the engine's rig.json "poses"
func poseBlend(cr *creature, target, from string, frac float64, bones []int, in, out float64) {
	src, dst := findAnim(cr, from), findAnim(cr, target)
	if src == nil || dst == nil {
		return
	}
	pf := src.frames[int(frac*float64(len(src.frames)-1))]
	n := len(dst.frames)
	for f := range dst.frames {
		t := float64(f) / float64(n-1)
		w := 1.0
		if t < in {
			w = smoothstep(t / in)
		} else if t > out {
			w = smoothstep((1 - t) / (1 - out))
		}
		for _, b := range bones {
			a, q := dst.frames[f].rot[b], pf.rot[b]
			if a[0]*q[0]+a[1]*q[1]+a[2]*q[2]+a[3]*q[3] < 0 {
				q = [4]float64{-q[0], -q[1], -q[2], -q[3]}
			}
			var r [4]float64
			l := 0.0
			for c := 0; c < 4; c++ {
				r[c] = a[c]*(1-w) + q[c]*w
				l += r[c] * r[c]
			}
			l = math.Sqrt(l)
			for c := 0; c < 4; c++ {
				r[c] /= l
			}
			dst.frames[f].rot[b] = r
		}
	}
}

func smoothstep(x float64) float64 {
	x = math.Max(0, math.Min(1, x))
	return x * x * (3 - 2*x)
}

type posePose struct {
	anims   []string
	from    string
	frame   float64
	bones   []int
	in, out float64
}
