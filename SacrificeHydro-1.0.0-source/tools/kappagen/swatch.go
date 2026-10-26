package main

import (
	"image"
	"image/draw"
)

func swatches2(path string) {
	mats := []func(w, h int) [][3]uint8{
		matScales(palette{rgb{44, 98, 88}, rgb{14, 40, 38}, rgb{120, 170, 140}, rgb{80, 70, 50}}, 64, 1001),
		matReptile(palette{rgb{44, 98, 88}, rgb{14, 40, 38}, rgb{120, 170, 140}, rgb{80, 70, 50}}, 40, 60, 1001),
	}
	const S = 512
	img := image.NewNRGBA(image.Rect(0, 0, S*2, S))
	for i, m := range mats {
		draw.Draw(img, image.Rect(S*i, 0, S*(i+1), S), toImage(S, S, m(S, S)), image.Point{}, draw.Src)
	}
	writePNG(path, img)
}

func swatches(path string) {
	if true {
		swatches2(path)
		return
	}
	mats := []func(w, h int) [][3]uint8{
		matScales(palette{rgb{44, 98, 88}, rgb{14, 40, 38}, rgb{120, 170, 140}, rgb{80, 70, 50}}, 30, 1001),
		matScales(palette{rgb{50, 96, 112}, rgb{18, 40, 56}, rgb{120, 170, 170}, rgb{70, 60, 90}}, 18, 12),
		matChitin(palette{rgb{170, 60, 40}, rgb{70, 20, 16}, rgb{230, 150, 90}, rgb{60, 80, 60}}, 7, 13),
		matShell(palette{rgb{96, 74, 44}, rgb{40, 28, 18}, rgb{160, 130, 80}, rgb{70, 80, 40}}, 4, 14),
		matCoral(palette{rgb{210, 120, 110}, rgb{110, 50, 60}, rgb{250, 200, 170}, rgb{200, 160, 120}}, 15),
		matFin(palette{rgb{90, 150, 170}, rgb{30, 60, 90}, rgb{180, 220, 230}, rgb{120, 80, 140}}, 9, 16),
		matStreak(palette{rgb{220, 210, 180}, rgb{120, 100, 80}, rgb{250, 245, 230}, rgb{150, 130, 90}}, 17),
		matReptile(palette{rgb{44, 98, 88}, rgb{14, 40, 38}, rgb{120, 170, 140}, rgb{80, 70, 50}}, 40, 64, 1001),
	}
	const S = 256
	img := image.NewNRGBA(image.Rect(0, 0, S*4, S*2))
	for i, m := range mats {
		px := m(S, S)
		draw.Draw(img, image.Rect(S*(i%4), S*(i/4), S*(i%4+1), S*(i/4+1)), toImage(S, S, px), image.Point{}, draw.Src)
	}
	writePNG(path, img)
}
