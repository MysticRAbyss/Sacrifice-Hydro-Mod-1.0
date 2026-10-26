package main

import (
	"fmt"
	"math"
	"path/filepath"
	"strings"
)

type gameData struct {
	cres   map[string]*cre8
	raw    map[string][]byte // CRE8 raw by tag
	models map[string]*modelFiles
	anims  map[string]map[string][2][]byte // model folder -> anim code -> sxsk,skel
	animPath map[string]map[string]string
}

type modelFiles struct {
	saxs, sxmd []byte
	pages      map[int][]byte
	path       string
}

func loadGame(data string) *gameData {
	g := &gameData{cres: map[string]*cre8{}, raw: map[string][]byte{}, models: map[string]*modelFiles{}, anims: map[string]map[string][2][]byte{}, animPath: map[string]map[string]string{}}
	sw, err := ReadWad(filepath.Join(data, "spells.wad"))
	if err != nil {
		panic(err)
	}
	sw.Root.Walk(func(e *Entry) {
		if e.IsDir() || e.Type != "CRE8" {
			return
		}
		b, _ := e.Data()
		if c := parseCre8(e.Path, b); c != nil && g.cres[c.tag] == nil {
			g.cres[c.tag] = c
			g.raw[c.tag] = b
		}
	})
	files, _ := filepath.Glob(filepath.Join(data, "saxs*.wad"))
	for _, f := range files {
		w, err := ReadWad(f)
		if err != nil {
			continue
		}
		w.Root.Walk(func(e *Entry) {
			if e.IsDir() {
				return
			}
			dir := filepath.Base(filepath.Dir(filepath.FromSlash(e.Path)))
			b, err := e.Data()
			if err != nil {
				return
			}
			get := func(k string) *modelFiles {
				mf := g.models[k]
				if mf == nil {
					mf = &modelFiles{pages: map[int][]byte{}}
					g.models[k] = mf
				}
				return mf
			}
			switch e.Type {
			case "SXMD":
				mf := get(e.Name)
				mf.sxmd = b
				mf.path = e.Path
			case "SAXS":
				get(e.Name).saxs = b
			case "SXTX":
				var n int
				if _, err := fmt.Sscanf(e.Name, ".%03d", &n); err == nil {
					get(dir).pages[n] = b
				}
			case "SXSK", "SKEL":
				if g.anims[dir] == nil {
					g.anims[dir] = map[string][2][]byte{}
				}
				p := g.anims[dir][e.Name]
				if e.Type == "SXSK" {
					p[0] = b
				} else {
					p[1] = b
				}
				g.anims[dir][e.Name] = p
			}
		})
	}
	return g
}

func dumpBones(g *gameData, tag string) {
	c := g.cres[tag]
	mf := g.models[c.model]
	m, _ := parseSXMD(mf.sxmd)
	sc := f32(mf.saxs, 0)
	ap := make([]v3, len(m.bones)+1)
	cnt := make([]int, len(m.bones)+1)
	for _, p := range m.positions {
		if int(p.bone) < len(cnt) {
			cnt[p.bone]++
		}
	}
	fmt.Printf("== %s model %s path %s scaling %.4f zf %.3f parts %d positions %d\n", tag, c.model, mf.path, sc, m.zfactor, len(m.parts), len(m.positions))
	for i, b := range m.bones {
		ap[i+1] = ap[b.parent].add(fromSXMD(b.pos[0], b.pos[1], b.pos[2]).mul(sc))
		w := ap[i+1]
		fmt.Printf("  bone %2d parent %2d  world (%6.2f %6.2f %6.2f)  verts %d\n", i+1, b.parent, w[0], w[1], w[2], cnt[i+1])
	}
	_ = math.Pi
	_ = strings.ToLower
}

func dumpBoneBoxes(g *gameData, tag string) {
	c := g.cres[tag]
	mf := g.models[c.model]
	m, _ := parseSXMD(mf.sxmd)
	sc := f32(mf.saxs, 0)
	ap := make([]v3, len(m.bones)+1)
	for i, b := range m.bones {
		ap[i+1] = ap[b.parent].add(fromSXMD(b.pos[0], b.pos[1], b.pos[2]).mul(sc))
	}
	lo := make([]v3, len(ap))
	hi := make([]v3, len(ap))
	for i := range lo {
		lo[i] = v3{9, 9, 9}
		hi[i] = v3{-9, -9, -9}
	}
	for _, p := range m.positions {
		if p.weight < 12000 || int(p.bone) >= len(ap) {
			continue
		}
		w := ap[p.bone].add(fromSXMD(float64(p.pos[0]), float64(p.pos[1]), float64(p.pos[2])).mul(sc))
		for k := 0; k < 3; k++ {
			lo[p.bone][k] = math.Min(lo[p.bone][k], w[k])
			hi[p.bone][k] = math.Max(hi[p.bone][k], w[k])
		}
	}
	for i := range ap {
		fmt.Printf("  bone %2d box x[%5.2f %5.2f] y[%5.2f %5.2f] z[%5.2f %5.2f]\n", i, lo[i][0], hi[i][0], lo[i][1], hi[i][1], lo[i][2], hi[i][2])
	}
	// animation codes
	for slot, code := range c.anims {
		if code != "" {
			fmt.Printf("%s=%s ", animSlots[slot], code)
		}
	}
	fmt.Println()
	fmt.Println("anim files in model folder:", len(g.anims[c.model]))
}
