package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func main() {
	data := flag.String("data", "/home/claude/hdmod/fakegame/data", "Sacrifice data folder")
	dump := flag.String("dump", "", "comma-separated creature tags to dump")
	stats := flag.String("stats", "", "comma-separated creature tags: print stats")
	findx := flag.Bool("findexplicit", false, "")
	gen := flag.String("gen", "", "output folder: generate the creatures named by -only (or all) into <folder>/<name>")
	only := flag.String("only", "", "comma-separated creature names (default: all)")
	sheetTags := flag.String("sheet", "", "contact sheet of these creatures")
	abil := flag.String("abil", "", "abilities of these creatures")
	rig := flag.String("rig", "", "rig view of this creature (rest and stance)")
	pv := flag.String("parts", "", "render each body part of this original creature")
	cmp := flag.String("cmp", "", "dir:design:obj  compare our skinning with an engine OBJ export")
	sw := flag.Bool("swatch", false, "render material swatches")
	strip := flag.String("strip", "", "dir:anim1,anim2[:yaw] frame strips of the -only creature")
	chk := flag.String("check", "", "mesh check: dir of generated creatures (with -only), or orig:tag1,tag2")
	iconRaw := flag.String("iconraw", "", "src:out[:yaw:pitch]  render the -only (or all) installed creatures in <src>/<dir> for round icons into <out> (black and white backgrounds)")
	flag.Parse()
	if *sw {
		swatches("/tmp/swatch.png")
		return
	}
	g := loadGame(*data)
	for _, t := range strings.Split(*abil, ",") {
		if t != "" {
			dumpAbil(g, t)
			fmt.Println("   anims:", g.cres[t].anims[:12])
		}
	}
	if *strip != "" {
		f := strings.Split(*strip, ":")
		yaw := 0.6
		if len(f) > 2 {
			fmt.Sscan(f[2], &yaw)
		}
		if f[0] == "orig" {
			origStrip(g, *only, strings.Split(f[1], ","), yaw)
			for _, an := range strings.Split(f[1], ",") {
				dumpPitch(g, *only, an)
			}
			return
		}
		for _, d := range designs() {
			if strings.EqualFold(d.dir, *only) {
				cr, err := readBack(g, d, filepath.Join(f[0], d.dir))
				if err != nil {
					panic(err)
				}
				if f[1] == "list" {
					for _, a := range cr.anims {
						fmt.Println(a.name, a.code, len(a.frames), a.fps)
					}
					return
				}
				if pb := os.Getenv("POSEBLEND"); pb != "" { // target:from:frac:in:out:b1,b2,...
					p := strings.Split(pb, ":")
					var frac, in, out float64
					fmt.Sscan(p[2], &frac)
					fmt.Sscan(p[3], &in)
					fmt.Sscan(p[4], &out)
					var bones []int
					for _, b := range strings.Split(p[5], ",") {
						var bi int
						fmt.Sscan(b, &bi)
						bones = append(bones, bi)
					}
					poseBlend(cr, p[0], p[1], frac, bones, in, out)
				}
				for _, an := range strings.Split(f[1], ",") {
					if findAnim(cr, an) != nil {
						dumpRoot(cr, an)
					}
				}
				if os.Getenv("HEADBOB") != "" {
					headBob(cr)
				}
				frameStrip(cr, strings.Split(f[1], ","), 7, yaw, "/tmp/strip_"+d.dir+".png")
			}
		}
		return
	}
	if strings.HasPrefix(*chk, "orig:") {
		for _, t := range strings.Split((*chk)[5:], ",") {
			cr, err := loadOriginal(g, t)
			if err == nil {
				meshCheck(cr, t)
			}
		}
		return
	}
	if *chk != "" {
		for _, d := range designs() {
			if *only != "" && !strings.Contains(","+strings.ToLower(*only)+",", ","+strings.ToLower(d.dir)+",") {
				continue
			}
			cr, err := readBack(g, d, filepath.Join(*chk, d.dir))
			if err != nil {
				fmt.Println(d.name, err)
				continue
			}
			meshCheck(cr, d.name)
		}
		return
	}
	if *pv != "" {
		partsView(g, *pv, "/tmp/parts_"+*pv+".png")
		return
	}
	if *cmp != "" {
		f := strings.SplitN(*cmp, ":", 3)
		for _, d := range designs() {
			if d.dir == f[1] {
				compareOBJ(g, d, f[0], f[2])
			}
		}
		return
	}
	if *rig != "" && *only != "" {
		for _, p := range strings.Split(*only, ",") {
			var pi int
			fmt.Sscan(p, &pi)
			dbgPart(g, *rig, pi)
		}
		return
	}
	if *rig != "" {
		cr, err := loadOriginal(g, *rig)
		if err != nil {
			panic(err)
		}
		rigView(cr, "", "/tmp/rig_"+*rig+"_rest.png")
		rigView(cr, "stance", "/tmp/rig_"+*rig+"_stance.png")
		if a := findAnim(cr, "stance"); a != nil {
			pos, _ := boneTransforms(cr, &a.frames[0])
			for i := 1; i < len(cr.bones); i++ {
				fmt.Printf("  stance bone %2d parent %2d (%6.2f %6.2f %6.2f)\n", i, cr.bones[i].parent, pos[i][0], pos[i][1], pos[i][2])
			}
			lo, hi := bounds(cr, &a.frames[0])
			fmt.Printf("  stance bounds %v %v\n", lo, hi)
		}
		return
	}
	if *sheetTags != "" {
		contactSheet(g, strings.Split(*sheetTags, ","), "/tmp/sheet.png", "stance")
		return
	}
	if *iconRaw != "" {
		parts := strings.Split(*iconRaw, ":")
		for _, d := range designs() {
			if *only != "" && !strings.Contains(","+strings.ToLower(*only)+",", ","+strings.ToLower(d.dir)+",") {
				continue
			}
			cr, err := readBack(g, d, filepath.Join(parts[0], d.dir))
			if err != nil {
				panic(err)
			}
			yaw, pitch := d.iconYaw, 0.3
			if len(parts) > 2 {
				fmt.Sscan(parts[2], &yaw)
			}
			if len(parts) > 3 {
				fmt.Sscan(parts[3], &pitch)
			}
			if err := iconRawRender(cr, d, filepath.Join(parts[1], d.dir), yaw, pitch); err != nil {
				panic(err)
			}
		}
		return
	}
	if *gen != "" {
		for _, d := range designs() {
			if *only != "" && !strings.Contains(","+strings.ToLower(*only)+",", ","+strings.ToLower(d.dir)+",") {
				continue
			}
			fmt.Println("==", d.name)
			out := filepath.Join(*gen, d.dir)
			if err := generate(g, d, out); err != nil {
				panic(err)
			}
			cr, err := readBack(g, d, out)
			if err != nil {
				panic(err)
			}
			fmt.Printf("read back: %d parts, %d bones, %d animations\n", len(cr.parts), len(cr.bones), len(cr.anims))
			if err := previews(cr, d, filepath.Join(*gen, d.dir+"_preview.png")); err != nil {
				panic(err)
			}
			closeup(cr, filepath.Join(*gen, d.dir+"_close.png"), 0.6)
			if err := icon(cr, d, filepath.Join(out, "icon.png"), 64); err != nil {
				panic(err)
			}
		}
		return
	}
	if *findx {
		findExplicit(g)
	}
	for _, t := range strings.Split(*stats, ",") {
		if t != "" {
			dumpStats(g, t)
		}
	}
	if *dump != "" {
		for _, t := range strings.Split(*dump, ",") {
			analyze(g, t)
		}
		for _, t := range strings.Split(*dump, ",") {
			if g.cres[t] != nil {
				dumpBones(g, t)
				dumpBoneBoxes(g, t)
			}
		}
	}
}
