package main

import "math"

// more procedural textures for the Hydro creatures (all tileable unless noted)

func spotsTex(base, dark, spot rgb, period, spotSize float64, seed int) func(w, h int) [][3]uint8 {
	return func(w, h int) [][3]uint8 {
		return render(w, h, func(u, v float64) rgb {
			c := mix(base, dark, 0.5+0.9*fbm(u, v, 4, 4, seed))
			d1, _, cell := voronoi(u, v, int(period), seed+3)
			r := spotSize * (0.6 + 0.4*hash3(cell, 1, seed))
			if hash3(cell, 7, seed) > -0.2 {
				t := clamp(1-d1/r, 0, 1)
				c = mix(c, spot, t*t*2)
				c = mix(c, rgb{255, 255, 255}, clamp(t*3-2.2, 0, 1))
			}
			return c
		})
	}
}

// soft vertical bands (fins, tendrils, gills)
func bandsTex(a, b rgb, n float64, seed int) func(w, h int) [][3]uint8 {
	return func(w, h int) [][3]uint8 {
		return render(w, h, func(u, v float64) rgb {
			t := 0.5 + 0.5*math.Sin(2*math.Pi*n*u+3*pnoise(u, v, 4, seed))
			c := mix(a, b, t)
			return scale(c, 0.9+0.15*fbm(u, v, 8, 3, seed+5))
		})
	}
}

// scaly skin with configurable colours
func scalesTex(base, dark, pale rgb, period int, seed int) func(w, h int) [][3]uint8 {
	return func(w, h int) [][3]uint8 {
		return render(w, h, func(u, v float64) rgb {
			m := fbm(u, v, 4, 4, seed)
			c := mix(base, dark, 0.5+0.9*m)
			d1, d2, cell := voronoi(u, v, period, seed+5)
			edge := clamp((d2-d1)/0.18, 0, 1)
			c = scale(c, 0.75+0.25*edge)
			c = mix(c, pale, 0.25*(0.5+0.5*hash3(cell, 3, seed))*edge)
			return c
		})
	}
}

// crab/shrimp shell: segmented plates with speckles
func carapaceTex(base, dark, light rgb, plates float64, seed int) func(w, h int) [][3]uint8 {
	return func(w, h int) [][3]uint8 {
		return render(w, h, func(u, v float64) rgb {
			c := mix(base, dark, 0.5+0.8*fbm(u, v, 4, 4, seed))
			seg := math.Abs(math.Mod(v*plates+0.5*pnoise(u, v, 2, seed), 1)-0.5) * 2
			c = mix(scale(c, 0.55), c, clamp(seg/0.12, 0, 1))
			c = mix(c, light, clamp((seg-0.7)*1.5, 0, 0.5))
			if pnoise(u, v, 64, seed+9) > 0.75 {
				c = mix(c, dark, 0.7)
			}
			if pnoise(u, v, 48, seed+11) > 0.8 {
				c = mix(c, light, 0.6)
			}
			return c
		})
	}
}

// coral: bumpy pink/orange with pores
func coralTex(base, dark, tip rgb, seed int) func(w, h int) [][3]uint8 {
	return func(w, h int) [][3]uint8 {
		return render(w, h, func(u, v float64) rgb {
			c := mix(base, tip, 0.5+0.8*fbm(u, v, 6, 3, seed))
			d1, _, _ := voronoi(u, v, 18, seed+2)
			if d1 < 0.18 {
				c = mix(c, dark, 1-d1/0.18)
			}
			return c
		})
	}
}

func plainTex(a, b rgb, seed int) func(w, h int) [][3]uint8 {
	return func(w, h int) [][3]uint8 {
		return render(w, h, func(u, v float64) rgb {
			return mix(a, b, 0.5+0.8*fbm(u, v, 4, 3, seed))
		})
	}
}

// eyes: iris colour, round or slit pupil
func eyeTexC(iris, ring rgb, slit bool) func(w, h int) [][3]uint8 {
	return func(w, h int) [][3]uint8 {
		return render(w, h, func(u, v float64) rgb {
			x, y := (u-0.5)*2, (v-0.5)*2
			r := math.Hypot(x, y)
			c := mix(iris, ring, clamp((r-0.5)*3, 0, 1))
			if slit {
				if math.Abs(x) < 0.14*(1-y*y)+0.02 && math.Abs(y) < 0.75 {
					c = rgb{8, 8, 10}
				}
			} else if r < 0.36 {
				c = rgb{8, 8, 10}
			}
			if math.Hypot(x+0.3, y+0.35) < 0.14 {
				c = rgb{250, 250, 250}
			}
			return c
		})
	}
}

// googly eye: big white with small pupil (goofy)
func googlyTex(pupil rgb) func(w, h int) [][3]uint8 {
	return func(w, h int) [][3]uint8 {
		return render(w, h, func(u, v float64) rgb {
			x, y := (u-0.5)*2, (v-0.5)*2
			c := rgb{238, 236, 222}
			if math.Hypot(x-0.05, y-0.1) < 0.42 {
				c = pupil
			}
			if math.Hypot(x-0.05, y-0.1) < 0.24 {
				c = rgb{6, 6, 8}
			}
			if math.Hypot(x+0.12, y-0.0) < 0.09 {
				c = rgb{255, 255, 255}
			}
			return c
		})
	}
}
