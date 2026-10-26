package main

// Gritty, painted-relief textures in the spirit of Sacrifice's creature skins: an albedo with
// colour variation and grime, lit by a baked relief (scales, plates, wrinkles, pores, rays),
// with dark cavities, a little wet sheen and fine grain. All tile seamlessly.

import "math"

type palette struct {
	base, dark, light, accent rgb
}

type relief func(u, v float64) (h float64, cell int)

// ridged periodic noise, 0..1, sharp crests
func ridged(u, v float64, period, oct, seed int) float64 {
	s, amp, tot := 0.0, 1.0, 0.0
	for o := 0; o < oct; o++ {
		n := 1 - math.Abs(pnoise(u, v, period<<o, seed+o*31))
		s += amp * n * n
		tot += amp
		amp *= 0.5
	}
	return s / tot
}

func scaleRelief(period int, elong float64, seed int) relief {
	return func(u, v float64) (float64, int) {
		// offset rows so the scales overlap like shingles
		d1, d2, c := voronoi(u, v, period, seed)
		e := clamp((d2-d1)/0.35, 0, 1)
		h := math.Sqrt(e)*0.8 + 0.2*clamp(1-d1/0.7, 0, 1)
		h += 0.08 * pnoise(u, v, period*4, seed+3)
		_ = elong
		return h, c
	}
}

func wrinkleRelief(period int, seed int) relief {
	return func(u, v float64) (float64, int) {
		w := ridged(u, v, period, 3, seed)
		d1, _, c := voronoi(u, v, period*3, seed+7) // small bumps / warts
		b := clamp(1-d1/0.45, 0, 1)
		pore := 0.0
		if pnoise(u, v, period*16, seed+9) > 0.82 {
			pore = -0.1
		}
		return 0.55*w + 0.35*b*b + pore, c
	}
}

func plateRelief(rows float64, pits int, seed int) relief {
	return func(u, v float64) (float64, int) {
		f := v*rows + 0.35*pnoise(u, v, 3, seed)
		seg := math.Abs(math.Mod(f, 1)-0.5) * 2 // 1 at plate centre, 0 at seams
		h := math.Sqrt(clamp(seg/0.9, 0, 1))
		d1, _, c := voronoi(u, v, pits, seed+5)
		if d1 < 0.11 {
			h -= 0.15 * (1 - d1/0.11)
		}
		// scratches
		s := math.Abs(pnoise(u*3, v*0.4, 24, seed+8))
		if s < 0.02 {
			h -= 0.2
		}
		return h, int(f) + c*7
	}
}

func shellRelief(cells int, seed int) relief {
	return func(u, v float64) (float64, int) {
		d1, d2, c := voronoi(u, v, cells, seed)
		e := clamp((d2-d1)/0.25, 0, 1)
		rings := 0.5 + 0.5*math.Sin(d1*55+2*pnoise(u, v, 8, seed+2))
		return 0.7*math.Sqrt(e) + 0.15*rings*e + 0.1*pnoise(u, v, 32, seed+4), c
	}
}

func coralRelief(seed int) relief {
	return func(u, v float64) (float64, int) {
		d1, _, c := voronoi(u, v, 14, seed)
		h := 0.6 + 0.3*fbm(u, v, 6, 3, seed+1)
		if d1 < 0.22 {
			h -= 0.6 * (1 - d1/0.22)
		}
		return h, c
	}
}

func finRelief(rays float64, seed int) relief {
	return func(u, v float64) (float64, int) {
		r := 0.5 + 0.5*math.Cos(2*math.Pi*(u*rays+0.15*pnoise(u, v, 4, seed)))
		vein := ridged(u, v, 6, 2, seed+3)
		return 0.6*r*r + 0.3*vein, 0
	}
}

func streakRelief(seed int) relief { // horn, bone, wood, hair
	return func(u, v float64) (float64, int) {
		s := pnoise(u, v*0.08, 48, seed)
		c := 0.0
		if math.Abs(pnoise(u, v, 12, seed+4)) < 0.03 {
			c = -0.4
		}
		return 0.5 + 0.4*s + c, 0
	}
}

// gritty renders a material from a relief and a palette
func gritty(rf relief, pal palette, bumpStrength, sheen float64, seed int) func(w, h int) [][3]uint8 {
	return func(w, h int) [][3]uint8 {
		hf := make([]float64, w*h)
		cells := make([]int, w*h)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				u, v := (float64(x)+0.5)/float64(w), (float64(y)+0.5)/float64(h)
				hf[y*w+x], cells[y*w+x] = rf(u, v)
			}
		}
		L := v3{-0.45, -0.55, 0.7}.norm()
		H := L.add(v3{0, 0, 1}).norm()
		out := make([][3]uint8, w*h)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				at := func(dx, dy int) float64 { return hf[((y+dy+h)%h)*w+(x+dx+w)%w] }
				gx := (at(1, 0) - at(-1, 0)) * bumpStrength * float64(w) / 256
				gy := (at(0, 1) - at(0, -1)) * bumpStrength * float64(h) / 256
				n := v3{-gx, -gy, 1}.norm()
				u, v := (float64(x)+0.5)/float64(w), (float64(y)+0.5)/float64(h)
				hh := hf[y*w+x]
				// albedo: colour drift, per-cell variation, accent mottling, grime
				c := mix(pal.base, pal.dark, 0.45+0.9*fbm(u, v, 3, 4, seed))
				c = mix(c, pal.light, 0.25*(0.5+0.5*hash3(cells[y*w+x], 11, seed))*clamp(hh, 0, 1))
				if m := fbm(u, v, 5, 3, seed+20); m > 0.25 {
					c = mix(c, pal.accent, clamp((m-0.25)*2.2, 0, 0.6))
				}
				c = mix(c, scale(pal.dark, 0.6), clamp((fbm(u, v, 2, 3, seed+40)-0.1)*1.2, 0, 0.45))
				// lighting baked from the relief
				diff := 0.42 + 0.75*math.Max(0, n.dot(L))
				cav := 0.55 + 0.45*clamp(hh*1.25, 0, 1)
				spec := sheen * math.Pow(math.Max(0, n.dot(H)), 28)
				grain := 1 + 0.07*hash3(x*7+y*13, y, seed)
				col := scale(c, diff*cav*grain)
				col = rgb{col[0] + 255*spec, col[1] + 255*spec, col[2] + 255*spec}
				for k := 0; k < 3; k++ {
					out[y*w+x][k] = uint8(clamp(col[k], 0, 255))
				}
			}
		}
		return out
	}
}

// monster eye: no white, veined iris, dark rim, slit or round pupil, small wet highlight
func monsterEye(iris, deep rgb, slit bool, glow float64) func(w, h int) [][3]uint8 {
	return func(w, h int) [][3]uint8 {
		return render(w, h, func(u, v float64) rgb {
			x, y := (u-0.5)*2, (v-0.5)*2
			r := math.Hypot(x, y)
			a := math.Atan2(y, x)
			streak := 0.5 + 0.5*math.Sin(a*23+3*pnoise(u, v, 6, 9))
			c := mix(iris, deep, clamp(r*0.9+0.25*streak-0.1, 0, 1))
			c = scale(c, 1+glow*(1-r))
			c = mix(c, rgb{10, 8, 6}, clamp((r-0.82)*6, 0, 1)) // dark rim / lid shadow
			if slit {
				if math.Abs(x) < 0.07*(1-y*y*0.9)+0.012 && math.Abs(y) < 0.8 {
					c = rgb{6, 5, 5}
				}
			} else if r < 0.3 {
				c = rgb{6, 5, 5}
			}
			if math.Hypot(x+0.28, y+0.3) < 0.09 {
				c = mix(c, rgb{255, 255, 245}, 0.85)
			}
			return c
		})
	}
}

// --- ready-made materials ----------------------------------------------------------------

func matScales(p palette, period int, seed int) func(w, h int) [][3]uint8 {
	return gritty(scaleRelief(period, 1, seed), p, 2.2, 0.35, seed)
}
func matSkin(p palette, period int, seed int) func(w, h int) [][3]uint8 {
	return gritty(wrinkleRelief(period, seed), p, 2.6, 0.45, seed)
}
func matChitin(p palette, rows float64, seed int) func(w, h int) [][3]uint8 {
	return gritty(plateRelief(rows, 20, seed), p, 2.0, 0.5, seed)
}
func matShell(p palette, cells int, seed int) func(w, h int) [][3]uint8 {
	return gritty(shellRelief(cells, seed), p, 2.4, 0.25, seed)
}
func matCoral(p palette, seed int) func(w, h int) [][3]uint8 {
	return gritty(coralRelief(seed), p, 2.8, 0.15, seed)
}
func matFin(p palette, rays float64, seed int) func(w, h int) [][3]uint8 {
	return gritty(finRelief(rays, seed), p, 1.4, 0.4, seed)
}
func matStreak(p palette, seed int) func(w, h int) [][3]uint8 {
	return gritty(streakRelief(seed), p, 1.6, 0.3, seed)
}

// overlapping fish/reptile scales (scalloped rows), periodic
func scallopRelief(cols, rows int, seed int) relief {
	return func(u, v float64) (float64, int) {
		x, y := u*float64(cols), v*float64(rows)
		best, bh, bc := -1e9, 0.0, 0
		jy := int(math.Floor(y))
		for dj := -1; dj <= 2; dj++ {
			j := jy + dj
			off := 0.0
			if ((j%rows)+rows)%2 == 1 {
				off = 0.5
			}
			ix := int(math.Floor(x - off))
			for di := -1; di <= 1; di++ {
				i := ix + di
				cx, cy := float64(i)+0.5+off, float64(j)
				d := math.Hypot((x-cx)*1.0, (y-cy)*1.25)
				if d < 0.78 {
					// the scale of the row further down (larger j) lies on top
					prio := float64(j)*10 - d
					if prio > best {
						best = prio
						bh = 1 - d/0.78
						wi, wj := ((i%cols)+cols)%cols, ((j%rows)+rows)%rows
						bc = wj*cols + wi
					}
				}
			}
		}
		h := math.Sqrt(clamp(bh, 0, 1))*0.9 + 0.08*pnoise(u, v, cols*3, seed)
		return h, bc
	}
}

func matFishScales(p palette, cols, rows int, seed int) func(w, h int) [][3]uint8 {
	return gritty(scallopRelief(cols, rows, seed), p, 1.3, 0.45, seed)
}

// bright bioluminescent spots on top of another material
func withGlowSpots(base func(w, h int) [][3]uint8, spot rgb, period int, size float64, seed int) func(w, h int) [][3]uint8 {
	return func(w, h int) [][3]uint8 {
		px := base(w, h)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				u, v := (float64(x)+0.5)/float64(w), (float64(y)+0.5)/float64(h)
				d1, _, cell := voronoi(u, v, period, seed)
				if hash3(cell, 5, seed) < 0.1 {
					continue
				}
				r := size * (0.5 + 0.5*hash3(cell, 1, seed))
				t := clamp(1-d1/r, 0, 1)
				if t <= 0 {
					continue
				}
				c := rgb{float64(px[y*w+x][0]), float64(px[y*w+x][1]), float64(px[y*w+x][2])}
				c = mix(c, spot, t*1.6)
				c = mix(c, rgb{240, 255, 255}, clamp(t*3-2, 0, 1))
				for k := 0; k < 3; k++ {
					px[y*w+x][k] = uint8(clamp(c[k], 0, 255))
				}
			}
		}
		return px
	}
}

// matReptile: small overlapping scales with soft shading, per-scale tint and large mottling,
// closer to real reptile skin than the cell-like matScales.
func matReptile(p palette, cols, rows int, seed int) func(w, h int) [][3]uint8 {
	rf := scallopRelief(cols, rows, seed)
	return func(w, h int) [][3]uint8 {
		hf := make([]float64, w*h)
		cells := make([]int, w*h)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				u, v := (float64(x)+0.5)/float64(w), (float64(y)+0.5)/float64(h)
				hh, c := rf(u, v)
				hf[y*w+x], cells[y*w+x] = math.Pow(clamp(hh, 0, 1), 0.7), c
			}
		}
		L := v3{-0.45, -0.55, 0.7}.norm()
		H := L.add(v3{0, 0, 1}).norm()
		out := make([][3]uint8, w*h)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				at := func(dx, dy int) float64 { return hf[((y+dy+h)%h)*w+(x+dx+w)%w] }
				gx := (at(1, 0) - at(-1, 0)) * 2.2 * float64(w) / 256
				gy := (at(0, 1) - at(0, -1)) * 2.2 * float64(h) / 256
				n := v3{-gx, -gy, 1}.norm()
				u, v := (float64(x)+0.5)/float64(w), (float64(y)+0.5)/float64(h)
				hh := hf[y*w+x]
				cell := cells[y*w+x]
				// mottling: large soft blotches and finer speckle, then per-scale tint
				mot := fbm(u, v, 3, 4, seed+5)
				c := mix(p.base, p.dark, clamp(0.35+0.9*mot, 0, 1))
				if sp := fbm(u, v, 9, 2, seed+9); sp > 0.3 {
					c = mix(c, p.accent, clamp((sp-0.3)*2.5, 0, 0.7))
				}
				c = mix(c, p.dark, 0.25*(0.5+0.5*hash3(cell, 3, seed)))
				c = mix(c, p.light, 0.22*hh*(0.6+0.4*hash3(cell, 7, seed)))
				diff := 0.55 + 0.55*math.Max(0, n.dot(L))
				cav := 0.72 + 0.28*hh
				spec := 0.22 * math.Pow(math.Max(0, n.dot(H)), 20) * hh
				grain := 1 + 0.05*hash3(x*7+y*13, y, seed)
				col := scale(c, diff*cav*grain)
				for k := 0; k < 3; k++ {
					out[y*w+x][k] = uint8(clamp(col[k]+255*spec, 0, 255))
				}
			}
		}
		return out
	}
}

// withStripes lays pale, wavy bands across a material (along the texture's u axis)
func withStripes(base func(w, h int) [][3]uint8, stripe rgb, freq float64, seed int) func(w, h int) [][3]uint8 {
	return func(w, h int) [][3]uint8 {
		px := base(w, h)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				u, v := (float64(x)+0.5)/float64(w), (float64(y)+0.5)/float64(h)
				ph := u*freq + 0.6*fbm(u, v, 3, 3, seed) + 0.25*math.Sin(v*2*math.Pi*2)
				s := math.Sin(2 * math.Pi * ph)
				t := clamp((s-0.35)*2.2, 0, 1) * (0.55 + 0.45*fbm(u, v, 6, 2, seed+3))
				if t <= 0 {
					continue
				}
				c := rgb{float64(px[y*w+x][0]), float64(px[y*w+x][1]), float64(px[y*w+x][2])}
				l := (c[0] + c[1] + c[2]) / (3 * 128) // keep the scale shading
				c = mix(c, scale(stripe, 0.6+0.5*l), 0.7*t)
				for k := 0; k < 3; k++ {
					px[y*w+x][k] = uint8(clamp(c[k], 0, 255))
				}
			}
		}
		return px
	}
}

// humanEye: white with a coloured, streaked iris, dark pupil and a highlight
func humanEye(iris rgb) func(w, h int) [][3]uint8 {
	return func(w, h int) [][3]uint8 {
		return render(w, h, func(u, v float64) rgb {
			x, y := (u-0.5)*2, (v-0.5)*2
			r := math.Hypot(x, y)
			c := rgb{236, 232, 226}
			c = mix(c, rgb{190, 170, 160}, clamp((r-0.75)*3, 0, 1)) // shaded towards the lids
			if r < 0.48 {
				a := math.Atan2(y, x)
				c = mix(iris, scale(iris, 0.45), clamp(r/0.48*0.8+0.2*math.Sin(a*19), 0, 1))
				c = mix(c, rgb{20, 30, 30}, clamp((r-0.42)*16, 0, 1)) // dark ring
			}
			if r < 0.2 {
				c = rgb{8, 8, 10}
			}
			if math.Hypot(x+0.16, y+0.18) < 0.09 {
				c = rgb{255, 255, 250}
			}
			return c
		})
	}
}

// armourRelief: bands of riveted plate with raised edges alternating with bands of chain mail
// (rows of interlocking rings, every other row offset by half a ring), as on a worn armour
func armourRelief(rows, rings, seed int) relief {
	return func(u, v float64) (float64, int) {
		fv := v * float64(rows)
		band := int(math.Floor(fv))
		t := fv - math.Floor(fv)
		if band%2 == 0 { // plate
			edge := math.Min(t, 1-t)
			h := 0.45 + 0.35*clamp(edge/0.1, 0, 1)
			if edge < 0.035 { // a raised rim along each edge
				h = 0.85
			}
			h += 0.06 * pnoise(u, v, 24, seed)
			nr := rings / 4
			for _, tr := range []float64{0.22, 0.78} {
				cu := (math.Floor(u*float64(nr)) + 0.5) / float64(nr)
				d := math.Hypot(u-cu, (t-tr)/float64(rows))
				const r = 0.007
				if d < r {
					h = math.Max(h, 0.75+0.3*math.Sqrt(1-(d/r)*(d/r)))
				}
			}
			return h, band
		}
		// chain mail
		s := 1 / float64(rings)
		bandH := 1 / float64(rows)
		nrow := int(math.Max(2, math.Round(bandH/(s*0.62))))
		rowH := bandH / float64(nrow)
		y := t * bandH
		h := 0.12
		jr := int(math.Floor(y / rowH))
		for j := jr - 1; j <= jr+1; j++ {
			cv := (float64(j) + 0.5) * rowH
			off := 0.0
			if ((j%2)+2)%2 == 1 {
				off = 0.5
			}
			ir := int(math.Floor(u/s - off))
			for i := ir - 1; i <= ir+1; i++ {
				cu := (float64(i) + 0.5 + off) * s
				d := math.Hypot(u-cu, y-cv)
				p := 1 - math.Abs(d-0.4*s)/(0.13*s)
				if p > 0 {
					h = math.Max(h, 0.3+0.6*math.Sqrt(p))
				}
			}
		}
		return h, band + 1000*jr
	}
}
