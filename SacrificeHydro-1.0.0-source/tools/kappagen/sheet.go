package main

import (
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"math"
	"strings"
)

func loadOriginal(g *gameData, tag string) (*creature, error) {
	c := g.cres[tag]
	if c == nil {
		return nil, fmt.Errorf("no creature %s", tag)
	}
	mf := g.models[c.model]
	animDir := strings.Split(strings.TrimPrefix(mf.path, "/"), "/")[1]
	return buildCreature(c, mf.saxs, mf.sxmd, mf.pages, g.anims[animDir])
}

// contact sheet of original creatures in their stance pose, auto-framed
func contactSheet(g *gameData, tags []string, path string, anim string) {
	const W, H = 300, 300
	cols := 6
	rows := (len(tags) + cols - 1) / cols
	sheet := image.NewNRGBA(image.Rect(0, 0, W*cols, H*rows))
	for i, t := range tags {
		cr, err := loadOriginal(g, t)
		if err != nil {
			fmt.Println(t, err)
			continue
		}
		var fr *animFrame
		if a := findAnim(cr, anim); a != nil && len(a.frames) > 0 {
			fr = &a.frames[0]
		}
		posed := pose(cr, fr)
		lo, hi := v3{9, 9, 9}, v3{-9, -9, -9}
		for _, pp := range posed {
			for _, p := range pp {
				for k := 0; k < 3; k++ {
					lo[k] = math.Min(lo[k], p[k])
					hi[k] = math.Max(hi[k], p[k])
				}
			}
		}
		size := math.Max(hi[0]-lo[0], math.Max(hi[1]-lo[1], hi[2]-lo[2]))
		img := renderCreature(cr, fr, camera{yaw: 0.6, pitch: 0.25, center: lo.add(hi).mul(0.5), scale: 0.8 * W / size}, W, H, color.NRGBA{58, 70, 86, 255})
		draw.Draw(sheet, image.Rect(W*(i%cols), H*(i/cols), W*(i%cols+1), H*(i/cols+1)), img, image.Point{}, draw.Src)
		var an []string
		for _, a := range cr.anims {
			an = append(an, a.name)
		}
		fmt.Printf("%s %-14s bones %2d size %.2f anims %s\n", t, g.cres[t].model, len(cr.bones)-1, size, strings.Join(an, ","))
	}
	writePNG(path, sheet)
}
