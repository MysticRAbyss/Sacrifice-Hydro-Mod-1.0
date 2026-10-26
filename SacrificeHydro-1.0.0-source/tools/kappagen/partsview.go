package main

import (
	"image"
	"image/color"
	"image/draw"
)

// partsView renders each body part of an original creature on its own (stance pose)
func partsView(g *gameData, tag, path string) {
	cr, err := loadOriginal(g, tag)
	if err != nil {
		panic(err)
	}
	var fr *animFrame
	if a := findAnim(cr, "stance"); a != nil {
		fr = &a.frames[0]
	}
	lo, hi := bounds(cr, fr)
	size := hi.sub(lo)
	s := size[0]
	if size[1] > s {
		s = size[1]
	}
	if size[2] > s {
		s = size[2]
	}
	const W = 300
	n := len(cr.parts)
	sheet := image.NewNRGBA(image.Rect(0, 0, W*n, W))
	for i := range cr.parts {
		one := *cr
		one.parts = cr.parts[i : i+1]
		img := renderCreature(&one, fr, camera{yaw: 0.6, pitch: 0.2, center: lo.add(hi).mul(0.5), scale: 0.8 * W / s}, W, W, color.NRGBA{58, 70, 86, 255})
		drawNum(img, 6, 6, i, color.NRGBA{255, 255, 0, 255})
		draw.Draw(sheet, image.Rect(W*i, 0, W*(i+1), W), img, image.Point{}, draw.Src)
	}
	writePNG(path, sheet)
}
