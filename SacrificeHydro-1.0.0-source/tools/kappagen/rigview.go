package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
)

// 3x5 digit font
var digits = []string{"111101101101111", "010110010010111", "111001111100111", "111001111001111", "101101111001001", "111100111001111", "111100111101111", "111001001001001", "111101111101111", "111101111001111"}

func drawNum(img *image.NRGBA, x, y, n int, c color.NRGBA) {
	s := fmt.Sprint(n)
	for i, ch := range s {
		d := digits[ch-'0']
		for k := 0; k < 15; k++ {
			if d[k] == '1' {
				for sy := 0; sy < 2; sy++ {
					for sx := 0; sx < 2; sx++ {
						img.SetNRGBA(x+i*8+(k%3)*2+sx, y+(k/3)*2+sy, c)
					}
				}
			}
		}
	}
}

func drawLine(img *image.NRGBA, x0, y0, x1, y1 float64, c color.NRGBA) {
	n := int(math.Max(math.Abs(x1-x0), math.Abs(y1-y0))) + 1
	for i := 0; i <= n; i++ {
		t := float64(i) / float64(n)
		img.SetNRGBA(int(x0+(x1-x0)*t), int(y0+(y1-y0)*t), c)
	}
}

// rigView renders a creature (rest pose or animation frame) from several angles with its bones numbered
func rigView(cr *creature, animName string, path string) {
	var fr *animFrame
	if a := findAnim(cr, animName); a != nil && len(a.frames) > 0 {
		fr = &a.frames[0]
	}
	posed := pose(cr, fr)
	lo, hi := v3{99, 99, 99}, v3{-99, -99, -99}
	for _, pp := range posed {
		for _, p := range pp {
			for k := 0; k < 3; k++ {
				lo[k] = math.Min(lo[k], p[k])
				hi[k] = math.Max(hi[k], p[k])
			}
		}
	}
	size := math.Max(hi[0]-lo[0], math.Max(hi[1]-lo[1], hi[2]-lo[2]))
	pos, _ := boneTransforms(cr, fr)
	nb := len(cr.bones)
	const W, H = 420, 420
	yaws := []float64{0, math.Pi / 2, math.Pi, 0.7}
	sheet := image.NewNRGBA(image.Rect(0, 0, W*len(yaws), H))
	for vi, yaw := range yaws {
		cam := camera{yaw: yaw, pitch: 0.1, center: lo.add(hi).mul(0.5), scale: 0.85 * W / size}
		img := renderCreature(cr, fr, cam, W, H, color.NRGBA{58, 70, 86, 255})
		cy, sy := math.Cos(cam.yaw), math.Sin(cam.yaw)
		cp, sp := math.Cos(cam.pitch), math.Sin(cam.pitch)
		fwd := v3{-sy * cp, -cy * cp, -sp}
		right := v3{-cy, sy, 0}
		up := right.cross(fwd)
		proj := func(p v3) (float64, float64) {
			q := p.sub(cam.center)
			return W/2 + q.dot(right)*cam.scale, H/2 - q.dot(up)*cam.scale
		}
		for i := 1; i < nb; i++ {
			x0, y0 := proj(pos[cr.bones[i].parent])
			x1, y1 := proj(pos[i])
			drawLine(img, x0, y0, x1, y1, color.NRGBA{255, 255, 0, 255})
		}
		for i := 1; i < nb; i++ {
			x, y := proj(pos[i])
			drawNum(img, int(x)+2, int(y)+2, i, color.NRGBA{255, 80, 80, 255})
		}
		draw.Draw(sheet, image.Rect(W*vi, 0, W*(vi+1), H), img, image.Point{}, draw.Src)
	}
	writePNG(path, sheet)
}
