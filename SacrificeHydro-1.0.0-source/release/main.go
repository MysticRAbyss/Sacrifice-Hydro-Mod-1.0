// Builds the Sacrifice: Hydro 1.0.0 and Sacrifice Hydro HD 1.0.0 downloads (Setup and Manual zips)
// from this source folder - see BUILD.txt:
//
//	go run . -engine <Hydro.exe or 3d.exe you built> -assets <SacrificeHydro-1.0.0-Manual.zip>
//	         -sachd <SacrificeHD-1.1.1-Manual.zip> [-source <SacrificeHydro-1.0.0-source.zip>] -out <folder>
//
// It builds the launcher and the installers, takes the mod's creature/spell files from the published
// Manual zip, the HD texture builder from the Sacrifice HD Manual zip, puts everything together and
// prints the checksum of every program it made.
package main

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

const version = "1.0.0"

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
}

func fileSum(p string) string {
	b, err := os.ReadFile(p)
	must(err)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

func write(p string, b []byte) {
	must(os.MkdirAll(filepath.Dir(p), 0755))
	must(os.WriteFile(p, b, 0644))
}

func copyFile(src, dst string) {
	b, err := os.ReadFile(src)
	must(err)
	write(dst, b)
}

func copyDir(src, dst string) {
	must(filepath.Walk(src, func(p string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return err
		}
		rel, _ := filepath.Rel(src, p)
		copyFile(p, filepath.Join(dst, rel))
		return nil
	}))
}

// extract copies the zip entries under prefix (e.g. "SacrificeHydro/mods/") to dst, without the prefix;
// if prefix is one file's full name, dst is that file's new path
func extract(zipPath, prefix, dst string) int {
	r, err := zip.OpenReader(zipPath)
	must(err)
	defer r.Close()
	n := 0
	for _, f := range r.File {
		if f.FileInfo().IsDir() || !strings.HasPrefix(f.Name, prefix) {
			continue
		}
		rc, err := f.Open()
		must(err)
		b, err := io.ReadAll(rc)
		rc.Close()
		must(err)
		write(filepath.Join(dst, filepath.FromSlash(strings.TrimPrefix(f.Name, prefix))), b)
		n++
	}
	return n
}

// dirBytes is the total size of the files in dir (0 if it does not exist)
func dirBytes(dir string) int64 {
	var n int64
	filepath.Walk(dir, func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			n += info.Size()
		}
		return nil
	})
	return n
}

func run(dir string, env []string, name string, args ...string) {
	cmd := exec.Command(name, args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), env...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	fmt.Println("   >", name, strings.Join(args, " "))
	must(cmd.Run())
}

func python() string {
	if _, err := exec.LookPath("python3"); err == nil {
		return "python3"
	}
	return "python"
}

func zipFolder(srcParent, folder, out string) {
	f, err := os.Create(out)
	must(err)
	w := zip.NewWriter(f)
	var files []string
	filepath.Walk(filepath.Join(srcParent, folder), func(p string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() {
			files = append(files, p)
		}
		return err
	})
	sort.Strings(files)
	for _, p := range files {
		rel, _ := filepath.Rel(srcParent, p)
		zw, err := w.Create(filepath.ToSlash(rel))
		must(err)
		b, err := os.ReadFile(p)
		must(err)
		_, err = zw.Write(b)
		must(err)
	}
	must(w.Close())
	must(f.Close())
}

var win = []string{"GOOS=windows", "GOARCH=amd64", "CGO_ENABLED=0"}

// resource writes rsrc_windows_amd64.syso (icon, manifest, version information) into dir
func resource(root, work, dir, title, desc, exe, icon, manifest string) {
	tool := filepath.Join(work, "verinfo")
	if runtime.GOOS == "windows" {
		tool += ".exe"
	}
	if _, err := os.Stat(tool); err != nil {
		run(filepath.Join(root, "installer", "verinfo"), nil, "go", "build", "-o", tool, ".")
	}
	vi := filepath.Join(work, "vi.json")
	run(filepath.Join(root, "installer"), nil, python(), "mkver.py", title, version, desc, exe, vi, icon, manifest)
	run(dir, nil, tool, vi, "rsrc_windows_amd64.syso")
}

func main() {
	engine := flag.String("engine", "", "the engine you built (engine\\3d.exe); it becomes Hydro.exe")
	assets := flag.String("assets", "", "SacrificeHydro-1.0.0-Manual.zip (its mods\\hydro folder is used)")
	sachd := flag.String("sachd", "", "SacrificeHD-1.1.1-Manual.zip (its texture builder goes into Hydro HD)")
	source := flag.String("source", "", "optional: SacrificeHydro-1.0.0-source.zip, copied into the mod's source folder")
	out := flag.String("out", "out", "output folder")
	flag.Parse()
	if *engine == "" || *assets == "" || *sachd == "" {
		fmt.Println("usage: go run . -engine <3d.exe> -assets <SacrificeHydro-1.0.0-Manual.zip> -sachd <SacrificeHD-1.1.1-Manual.zip> [-source <source zip>] -out <folder>")
		os.Exit(2)
	}
	here, _ := os.Getwd()
	root := filepath.Dir(here)
	outDir, _ := filepath.Abs(*out)
	work := filepath.Join(outDir, "work")
	os.RemoveAll(work)
	must(os.MkdirAll(work, 0755))

	fmt.Println("1. Launcher")
	lch := filepath.Join(root, "launcher")
	resource(root, work, lch, "Sacrifice Hydro", "Sacrifice Hydro launcher", "Sacrifice Hydro Launcher.exe",
		filepath.Join(lch, "hydro.ico"), filepath.Join(lch, "launcher.manifest"))
	launcher := filepath.Join(work, "Sacrifice Hydro Launcher.exe")
	run(lch, win, "go", "build", "-trimpath", "-ldflags", "-H windowsgui -s -w", "-o", launcher, ".")

	type ed struct{ folder, title, edition, json string }
	sums := map[string]string{"Hydro.exe (yours)": fileSum(*engine), "Sacrifice Hydro Launcher.exe": fileSum(launcher)}
	for _, e := range []ed{{"SacrificeHydro", "Sacrifice Hydro", "plain", "hydro.json"}, {"SacrificeHydroHD", "Sacrifice Hydro HD", "hd", "hydrohd.json"}} {
		fmt.Println("2. " + e.title)
		setupName := e.title + " " + version + " Setup"
		f := filepath.Join(work, e.edition, setupName, "files", e.folder) // the mod's folder
		pkg := filepath.Join(root, "package")
		copyFile(*engine, filepath.Join(f, "Hydro.exe"))
		copyFile(launcher, filepath.Join(f, "Sacrifice Hydro Launcher.exe"))
		for _, d := range []string{"SDL2.dll", "OpenAL32.dll", "freetype.dll"} {
			copyFile(filepath.Join(root, "runtime", d), filepath.Join(f, d))
		}
		copyFile(filepath.Join(pkg, "hotkeys.txt"), filepath.Join(f, "hotkeys.txt"))
		if e.edition == "hd" {
			copyFile(filepath.Join(pkg, "hd", "settings.txt"), filepath.Join(f, "settings.txt"))
		} else {
			copyFile(filepath.Join(pkg, "settings.txt"), filepath.Join(f, "settings.txt"))
		}
		copyDir(filepath.Join(pkg, "licenses"), filepath.Join(f, "licenses"))
		if n := extract(*assets, "SacrificeHydro/mods/", filepath.Join(f, "mods")); n == 0 {
			must(fmt.Errorf("no SacrificeHydro/mods/ files in %s", *assets))
		}
		if *source != "" {
			copyFile(*source, filepath.Join(f, "source", "SacrificeHydro-source.zip"))
		}
		if e.edition == "hd" {
			hd := filepath.Join(f, "HDTextures")
			pre := "SacrificeHD-1.1.1-Manual/SacrificeHD/"
			if extract(*sachd, pre+"SacrificeHD.exe", filepath.Join(hd, "SacrificeHD.exe")) == 0 {
				must(fmt.Errorf("no %sSacrificeHD.exe in %s", pre, *sachd))
			}
			extract(*sachd, pre+"esrgan/", filepath.Join(hd, "esrgan"))
			for _, l := range []string{"Real-ESRGAN-BSD-3-Clause.txt", "Real-ESRGAN-ncnn-vulkan-MIT.txt"} {
				extract(*sachd, pre+"licenses/"+l, filepath.Join(hd, "licenses", l))
			}
			copyFile(filepath.Join(pkg, "hd", "THIRD-PARTY.txt"), filepath.Join(hd, "licenses", "THIRD-PARTY.txt"))
			copyFile(filepath.Join(pkg, "hd", "prepare.bat"), filepath.Join(hd, "prepare.bat"))
		}
		run(root, nil, python(), filepath.Join("tools", "release", "make_readme.py"), e.edition, version, filepath.Join(f, "README.txt"))

		// the installer's settings, plus the size of the mod (shown in Windows' list of installed programs);
		// the source zip is left out of the size so that the installer does not depend on it
		inst := filepath.Join(root, "installer")
		edJSON, err := os.ReadFile(filepath.Join(inst, "editions", e.json))
		must(err)
		edText := strings.TrimRight(string(edJSON), " \t\r\n")
		edText = strings.TrimSuffix(edText, "}") + fmt.Sprintf(",\"sizeKB\":%d}\n", (dirBytes(f)-dirBytes(filepath.Join(f, "source"))+1023)/1024)
		write(filepath.Join(inst, "edition.json"), []byte(edText))
		resource(root, work, inst, e.title, e.title+" setup", "Install "+e.title+".exe",
			filepath.Join(inst, "hydro.ico"), filepath.Join(inst, "setup.manifest"))
		setupDir := filepath.Join(work, e.edition, setupName)
		installer := filepath.Join(setupDir, "Install "+e.title+".exe")
		run(inst, win, "go", "build", "-trimpath", "-ldflags", "-H windowsgui", "-o", installer, ".")
		sums["Install "+e.title+".exe"] = fileSum(installer)

		// Manual zip: the mod's folder without the uninstaller; Setup zip: installer + files + README
		manual := filepath.Join(work, e.edition, "manual")
		copyDir(f, filepath.Join(manual, e.folder))
		copyFile(installer, filepath.Join(f, "Uninstall "+e.title+".exe"))
		copyFile(filepath.Join(f, "README.txt"), filepath.Join(setupDir, "README.txt"))
		zipFolder(filepath.Join(work, e.edition), setupName, filepath.Join(outDir, e.folder+"-"+version+"-Setup.zip"))
		zipFolder(manual, e.folder, filepath.Join(outDir, e.folder+"-"+version+"-Manual.zip"))
	}
	copyFile(filepath.Join(root, "installer", "editions", "hydro.json"), filepath.Join(root, "installer", "edition.json"))

	fmt.Println("3. Checksums (compare with CHECKSUMS.txt)")
	var names []string
	for n := range sums {
		names = append(names, n)
	}
	sort.Strings(names)
	for _, n := range names {
		fmt.Printf("   %s  %s\n", sums[n], n)
	}
	fmt.Println("done:", outDir)
}
