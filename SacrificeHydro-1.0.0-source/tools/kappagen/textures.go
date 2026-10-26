package main

// Procedural textures for the Tidekin (all tile seamlessly where box-projected).

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
)

type rgb = [3]float64

func mix(a, b rgb, t float64) rgb {
	t = clamp(t, 0, 1)
	return rgb{a[0] + (b[0]-a[0])*t, a[1] + (b[1]-a[1])*t, a[2] + (b[2]-a[2])*t}
}
func scale(a rgb, s float64) rgb { return rgb{a[0] * s, a[1] * s, a[2] * s} }

// periodic value noise on [0,1)^2 with integer period
func pnoise(x, y float64, period int, seed int) float64 {
	x *= float64(period)
	y *= float64(period)
	x0, y0 := math.Floor(x), math.Floor(y)
	fx, fy := x-x0, y-y0
	sx, sy := fx*fx*(3-2*fx), fy*fy*(3-2*fy)
	w := func(i, j int) float64 {
		i = ((i % period) + period) % period
		j = ((j % period) + period) % period
		return hash3(i, j, seed)
	}
	ix, iy := int(x0), int(y0)
	a := w(ix, iy) + (w(ix+1, iy)-w(ix, iy))*sx
	b := w(ix, iy+1) + (w(ix+1, iy+1)-w(ix, iy+1))*sx
	return a + (b-a)*sy
}

func fbm(x, y float64, base, oct, seed int) float64 {
	s, amp, tot := 0.0, 1.0, 0.0
	for o := 0; o < oct; o++ {
		s += amp * pnoise(x, y, base<<o, seed+o*17)
		tot += amp
		amp *= 0.5
	}
	return s / tot
}

// periodic Voronoi: distance to nearest and second-nearest feature point, and the cell id
func voronoi(x, y float64, period int, seed int) (d1, d2 float64, cell int) {
	x *= float64(period)
	y *= float64(period)
	ix, iy := int(math.Floor(x)), int(math.Floor(y))
	d1, d2 = 9, 9
	for dj := -1; dj <= 1; dj++ {
		for di := -1; di <= 1; di++ {
			ci, cj := ix+di, iy+dj
			wi, wj := ((ci%period)+period)%period, ((cj%period)+period)%period
			px := float64(ci) + 0.5 + 0.45*hash3(wi, wj, seed)
			py := float64(cj) + 0.5 + 0.45*hash3(wi, wj, seed+1)
			d := math.Hypot(x-px, y-py)
			if d < d1 {
				d2 = d1
				d1 = d
				cell = wj*period + wi
			} else if d < d2 {
				d2 = d
			}
		}
	}
	return
}

func render(w, h int, f func(u, v float64) rgb) [][3]uint8 {
	out := make([][3]uint8, w*h)
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := f((float64(x)+0.5)/float64(w), (float64(y)+0.5)/float64(h))
			for k := 0; k < 3; k++ {
				out[y*w+x][k] = uint8(clamp(c[k], 0, 255))
			}
		}
	}
	return out
}

func skinTex(light float64) func(w, h int) [][3]uint8 {
	return func(w, h int) [][3]uint8 {
		base := rgb{74, 124, 70}
		dark := rgb{38, 74, 46}
		pale := rgb{128, 168, 96}
		return render(w, h, func(u, v float64) rgb {
			m := fbm(u, v, 4, 4, 11)
			c := mix(base, dark, 0.5+0.9*m)
			d1, d2, cell := voronoi(u, v, 22, 5)
			edge := clamp((d2-d1)/0.18, 0, 1) // 0 at scale borders
			c = scale(c, 0.78+0.22*edge)
			c = mix(c, pale, 0.18*(0.5+0.5*hash3(cell, 3, 9))*edge)
			sp := pnoise(u, v, 48, 21)
			if sp > 0.82 {
				c = mix(c, dark, 0.6)
			}
			return scale(c, light)
		})
	}
}

func shellTex(w, h int) [][3]uint8 {
	base := rgb{104, 82, 44}
	dark := rgb{48, 36, 20}
	light := rgb{150, 124, 70}
	olive := rgb{86, 92, 46}
	return render(w, h, func(u, v float64) rgb {
		// hexagonal scutes: three columns of plates and a marginal ring
		x, y := (u-0.5)*2, (v-0.5)*2
		r := math.Hypot(x*0.95, y)
		var d1, d2 float64
		var cell int
		if r > 0.82 { // marginal scutes around the rim
			a := math.Atan2(y, x)/(2*math.Pi) + 0.5
			seg := a * 22
			fseg := seg - math.Floor(seg)
			d1 = math.Min(fseg, 1-fseg) * 0.5
			d1 = math.Min(d1, (r-0.82)*1.2)
			d2 = d1 + 0.2
			cell = int(seg)
			d1, d2 = 0.3-d1, 0.3
		} else {
			d1, d2, cell = voronoi(u*0.999, v*0.999, 4, 31)
		}
		edge := clamp((d2-d1)/0.12, 0, 1)
		ring := 0.5 + 0.5*math.Sin(d1*60)
		c := mix(base, olive, 0.5+0.5*hash3(cell, 1, 2))
		c = mix(c, light, 0.25*ring*edge)
		c = mix(dark, c, edge)
		c = scale(c, 0.85+0.25*fbm(u, v, 8, 3, 41))
		if r > 0.97 {
			c = mix(c, dark, (r-0.97)*20)
		}
		return c
	})
}

func plastronTex(w, h int) [][3]uint8 {
	base := rgb{206, 182, 112}
	seam := rgb{120, 96, 52}
	return render(w, h, func(u, v float64) rgb {
		c := scale(base, 0.88+0.2*fbm(u, v, 8, 3, 51))
		dv := math.Abs(math.Mod(v*5+0.5, 1)-0.5) * 2 // horizontal plate seams
		du := math.Abs(u-0.5) * 2
		s := math.Min(1, math.Min(clamp(dv/0.06, 0, 1), clamp(du/0.015+0.0, 0, 1)))
		c = mix(seam, c, s)
		c = mix(c, scale(base, 0.7), clamp((du-0.8)*4, 0, 1))
		return c
	})
}

func beakTex(w, h int) [][3]uint8 {
	a := rgb{232, 176, 58}
	b := rgb{190, 118, 36}
	return render(w, h, func(u, v float64) rgb {
		c := mix(a, b, 0.5+0.8*fbm(u, v, 4, 3, 61))
		c = scale(c, 0.9+0.1*pnoise(u, v, 64, 62))
		return c
	})
}

func dishTex(w, h int) [][3]uint8 {
	rim := rgb{214, 208, 186}
	water := rgb{96, 158, 196}
	deep := rgb{48, 96, 150}
	return render(w, h, func(u, v float64) rgb {
		x, y := (u-0.5)*2, (v-0.5)*2
		r := math.Hypot(x, y)
		c := mix(water, deep, r*1.1+0.25*fbm(u, v, 4, 3, 71))
		c = mix(c, rgb{235, 245, 255}, clamp(1-math.Hypot(x+0.25, y+0.3)/0.22, 0, 1)) // highlight
		c = mix(c, rim, clamp((r-0.66)*8, 0, 1))
		c = scale(c, 0.92+0.1*fbm(u, v, 8, 2, 72))
		return c
	})
}

func hairTex(w, h int) [][3]uint8 {
	a := rgb{30, 46, 32}
	b := rgb{12, 20, 14}
	return render(w, h, func(u, v float64) rgb {
		s := pnoise(u, v*0.02, 64, 81)
		return mix(a, b, 0.5+0.9*s+0.3*pnoise(u, v, 32, 82))
	})
}

func eyeTex(w, h int) [][3]uint8 {
	iris := rgb{236, 196, 52}
	ring := rgb{140, 90, 20}
	return render(w, h, func(u, v float64) rgb {
		x, y := (u-0.5)*2, (v-0.5)*2
		r := math.Hypot(x, y)
		c := mix(iris, ring, clamp((r-0.55)*3, 0, 1))
		if math.Abs(x) < 0.14*(1-y*y)+0.02 && math.Abs(y) < 0.75 { // slit pupil
			c = rgb{8, 8, 10}
		}
		if math.Hypot(x+0.3, y+0.35) < 0.13 {
			c = rgb{250, 250, 250}
		}
		return c
	})
}

// TGA (24-bit, bottom-left origin, as the game's own SXTX files)
func writeTGA(path string, w, h int, px [][3]uint8) error {
	var b bytes.Buffer
	hdr := make([]byte, 18)
	hdr[2] = 2
	binary.LittleEndian.PutUint16(hdr[12:], uint16(w))
	binary.LittleEndian.PutUint16(hdr[14:], uint16(h))
	hdr[16] = 24
	b.Write(hdr)
	for y := h - 1; y >= 0; y-- {
		for x := 0; x < w; x++ {
			c := px[y*w+x]
			b.Write([]byte{c[2], c[1], c[0]})
		}
	}
	return os.WriteFile(path, b.Bytes(), 0o644)
}

func writePNG(path string, img image.Image) error {
	var b bytes.Buffer
	if err := png.Encode(&b, img); err != nil {
		return err
	}
	return os.WriteFile(path, b.Bytes(), 0o644)
}

func toImage(w, h int, px [][3]uint8) *image.NRGBA {
	img := image.NewNRGBA(image.Rect(0, 0, w, h))
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			c := px[y*w+x]
			img.Set(x, y, color.NRGBA{c[0], c[1], c[2], 255})
		}
	}
	return img
}
