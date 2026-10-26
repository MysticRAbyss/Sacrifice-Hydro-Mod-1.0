package main

// The Options page: the settings the original game had in its Options menu (screen, sound, mouse),
// written to settings.txt next to Hydro.exe, which the game reads every time it starts. Lines in
// settings.txt that this page does not manage are kept as they are.

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// gameSettings are the managed settings, with the engine's defaults
type gameSettings struct {
	resolution  string // "" = desktop resolution
	fullscreen  bool
	hd          bool
	shadows     int     // shadow map resolution
	glow        float64 // 0..1
	volume      float64 // 0..1
	music       float64 // 0..1
	sound       float64 // 0..1
	camera      float64 // camera mouse sensitivity, 1 = normal
	wheel       float64 // mouse wheel sensitivity, 1 = normal
	doubleClick int     // frames at 60 per second
	cursor      int     // -1 = automatic
	capture     bool    // keep the mouse inside the window
	// picture and frame rate
	vivid      bool    // colours as in the original game (no filmic tone curve)
	exposure   float64 // brightness
	saturation float64 // colour strength
	contrast   float64
	anisotropy int // texture sharpness: 1 (off), 2, 4, 8, 16
	vsync      bool
	fpsLimit   int // 0: no limit
	antialias  bool
}

func defaultSettings(isHD bool) gameSettings {
	return gameSettings{fullscreen: true, hd: isHD, shadows: 1024, glow: 0.5, volume: 1, music: 1, sound: 1, camera: 1, wheel: 1, doubleClick: 20, cursor: -1,
		vivid: true, exposure: 0.7, saturation: 1.5, contrast: 1.2, anisotropy: 16, vsync: true, antialias: true}
}

// options this page writes; any other line of settings.txt is left alone
var managed = []string{"--detect-resolution", "--resolution=", "--fullscreen", "--no-fullscreen", "--hd", "--no-hd", "--hd-dir=",
	"--shadow-map-resolution=", "--glow-brightness=", "--volume=", "--music-volume=", "--sound-volume=",
	"--camera-mouse-sensitivity=", "--mouse-wheel-sensitivity=", "--double-click-speed=", "--cursor-size=",
	"--capture-mouse", "--no-capture-mouse",
	"--original-colors", "--no-original-colors", "--exposure=", "--saturation=", "--contrast=", "--anisotropy=",
	"--vsync", "--no-vsync", "--fps-limit=", "--antialiasing", "--no-antialiasing"}

func isManaged(opt string) bool {
	for _, m := range managed {
		if strings.HasSuffix(m, "=") && strings.HasPrefix(opt, m) || opt == m {
			return true
		}
	}
	return false
}

func readSettingsTxt(p string, isHD bool) (gameSettings, []string) {
	s := defaultSettings(isHD)
	s.fullscreen = false // only if the file asks for it, as the engine does
	s.hd = false
	b, err := os.ReadFile(p)
	if err != nil {
		return defaultSettings(isHD), nil
	}
	var others []string
	f := func(v string, d float64) float64 {
		if x, err := strconv.ParseFloat(v, 64); err == nil {
			return x
		}
		return d
	}
	for _, line := range strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n") {
		opt := strings.TrimSpace(line)
		if i := strings.Index(opt, "#"); i >= 0 {
			opt = strings.TrimSpace(opt[:i])
		}
		if opt == "" || !isManaged(opt) {
			if strings.TrimSpace(line) != "" && !strings.HasPrefix(strings.TrimSpace(line), "# Options page") {
				others = append(others, strings.TrimRight(line, " "))
			}
			continue
		}
		k, v, _ := strings.Cut(opt, "=")
		switch k {
		case "--detect-resolution":
			s.resolution = ""
		case "--resolution":
			s.resolution = v
		case "--fullscreen":
			s.fullscreen = true
		case "--no-fullscreen":
			s.fullscreen = false
		case "--hd", "--hd-dir":
			s.hd = true
		case "--no-hd":
			s.hd = false
		case "--shadow-map-resolution":
			s.shadows = int(f(v, 1024))
		case "--glow-brightness":
			s.glow = f(v, 0.5)
		case "--volume":
			s.volume = f(v, 1)
		case "--music-volume":
			s.music = f(v, 1)
		case "--sound-volume":
			s.sound = f(v, 1)
		case "--camera-mouse-sensitivity":
			s.camera = f(v, 1)
		case "--mouse-wheel-sensitivity":
			s.wheel = f(v, 1)
		case "--double-click-speed":
			s.doubleClick = int(f(v, 20))
		case "--cursor-size":
			s.cursor = int(f(v, -1))
		case "--capture-mouse":
			s.capture = true
		case "--no-capture-mouse":
			s.capture = false
		case "--original-colors":
			s.vivid = true
		case "--no-original-colors":
			s.vivid = false
		case "--exposure":
			s.exposure = f(v, 0.7)
		case "--saturation":
			s.saturation = f(v, 1.5)
		case "--contrast":
			s.contrast = f(v, 1.2)
		case "--anisotropy":
			s.anisotropy = int(f(v, 16))
		case "--vsync":
			s.vsync = true
		case "--no-vsync":
			s.vsync = false
		case "--fps-limit":
			s.fpsLimit = int(f(v, 0))
		case "--antialiasing":
			s.antialias = true
		case "--no-antialiasing":
			s.antialias = false
		}
	}
	return s, others
}

func writeSettingsTxt(p string, s gameSettings, others []string, isHD bool) error {
	num := func(x float64) string { return strconv.FormatFloat(math.Round(x*100)/100, 'f', -1, 64) }
	lines := []string{"# Options page of the launcher (the lines from here on are rewritten when you press OK there)"}
	if s.resolution == "" {
		lines = append(lines, "--detect-resolution")
	} else {
		lines = append(lines, "--resolution="+s.resolution)
	}
	if s.fullscreen {
		lines = append(lines, "--fullscreen")
	} else {
		lines = append(lines, "--no-fullscreen")
	}
	if isHD {
		if s.hd {
			lines = append(lines, "--hd")
		} else {
			lines = append(lines, "--no-hd")
		}
	}
	lines = append(lines,
		fmt.Sprintf("--shadow-map-resolution=%d", s.shadows),
		"--glow-brightness="+num(s.glow),
		"--volume="+num(s.volume),
		"--music-volume="+num(s.music),
		"--sound-volume="+num(s.sound),
		"--camera-mouse-sensitivity="+num(s.camera),
		"--mouse-wheel-sensitivity="+num(s.wheel),
		fmt.Sprintf("--double-click-speed=%d", s.doubleClick),
		fmt.Sprintf("--cursor-size=%d", s.cursor))
	if s.capture {
		lines = append(lines, "--capture-mouse")
	} else {
		lines = append(lines, "--no-capture-mouse")
	}
	flag := func(on bool, name string) string {
		if on {
			return "--" + name
		}
		return "--no-" + name
	}
	lines = append(lines, flag(s.vivid, "original-colors"),
		"--exposure="+num(s.exposure), "--saturation="+num(s.saturation), "--contrast="+num(s.contrast),
		fmt.Sprintf("--anisotropy=%d", s.anisotropy), flag(s.vsync, "vsync"), fmt.Sprintf("--fps-limit=%d", s.fpsLimit),
		flag(s.antialias, "antialiasing"))
	all := append(append([]string{}, others...), lines...)
	return os.WriteFile(p, []byte(strings.Join(all, "\r\n")+"\r\n"), 0o644)
}

var resolutions = []string{"Desktop resolution", "1280x720", "1366x768", "1600x900", "1920x1080", "2560x1080", "2560x1440", "3440x1440", "3840x2160", "1024x768", "1280x1024", "1600x1200"}
var shadowNames = []string{"Low", "Medium", "High", "Very high"}
var shadowSizes = []int{512, 1024, 2048, 4096}
var colourNames = []string{"Vivid (like the original game)", "Filmic (darker, softer highlights)"}
var anisoNames = []string{"Off", "2x", "4x", "8x", "16x (sharpest)"}
var anisoValues = []int{1, 2, 4, 8, 16}
var fpsNames = []string{"No limit", "30", "60", "75", "100", "120", "144", "165", "240"}
var fpsValues = []int{0, 30, 60, 75, 100, 120, 144, 165, 240}
var cursorNames = []string{"Automatic", "Small (32)", "Medium (48)", "Large (64)"}
var cursorSizes = []int{-1, 32, 48, 64}

// slider positions are whole numbers: percentages, or tenths for the mouse speeds
type sliderRow struct {
	slider *walk.Slider
	label  *walk.Label
	format func(int) string
}

func showOptions(owner walk.Form, dir string, isHD bool) {
	p := filepath.Join(dir, "settings.txt")
	s, others := readSettingsTxt(p, isHD)
	hotkeysPath := filepath.Join(dir, "hotkeys.txt")
	kb := loadKeyBindings(hotkeysPath)

	var dlg *walk.Dialog
	var resCombo, shadowCombo, cursorCombo *walk.ComboBox
	var fullscreen, hd, capture *walk.CheckBox
	var colourCombo, anisoCombo, fpsCombo *walk.ComboBox
	var vsyncBox, aaBox *walk.CheckBox
	rows := map[string]*sliderRow{}
	pct := func(v int) string { return fmt.Sprintf("%d%%", v) }
	speed := func(v int) string { return fmt.Sprintf("%.1fx", float64(v)/10) }
	ms := func(v int) string { return fmt.Sprintf("%d ms", v*1000/60) }
	rows["glow"] = &sliderRow{format: pct}
	rows["volume"] = &sliderRow{format: pct}
	rows["music"] = &sliderRow{format: pct}
	rows["sound"] = &sliderRow{format: pct}
	rows["camera"] = &sliderRow{format: speed}
	rows["wheel"] = &sliderRow{format: speed}
	rows["double"] = &sliderRow{format: ms}
	rows["exposure"] = &sliderRow{format: pct}
	rows["saturation"] = &sliderRow{format: pct}
	rows["contrast"] = &sliderRow{format: pct}

	slider := func(key, text string, min, max, value int) []Widget {
		r := rows[key]
		return []Widget{
			Label{Text: text},
			Slider{AssignTo: &r.slider, MinValue: min, MaxValue: max, Value: value, ToolTipsHidden: true,
				OnValueChanged: func() { r.label.SetText(r.format(r.slider.Value())) }},
			Label{AssignTo: &r.label, Text: r.format(value), MinSize: Size{Width: 50}},
		}
	}

	resIdx := 0
	if s.resolution != "" {
		resIdx = indexOf(resolutions, s.resolution)
		if resolutions[resIdx] != s.resolution {
			resolutions = append(resolutions, s.resolution)
			resIdx = len(resolutions) - 1
		}
	}
	shadowIdx := 1
	for i, v := range shadowSizes {
		if v == s.shadows {
			shadowIdx = i
		}
	}
	cursorIdx := 0
	for i, v := range cursorSizes {
		if v == s.cursor {
			cursorIdx = i
		}
	}
	pc := func(x float64) int { return int(math.Round(x * 100)) }
	tenth := func(x float64) int { return int(math.Round(x * 10)) }

	var applyPicture func(d gameSettings)
	apply := func(d gameSettings) { // shows the given settings in the page
		ri := 0
		if d.resolution != "" {
			ri = indexOf(resolutions, d.resolution)
		}
		resCombo.SetCurrentIndex(ri)
		fullscreen.SetChecked(d.fullscreen)
		hd.SetChecked(d.hd)
		for i, v := range shadowSizes {
			if v == d.shadows {
				shadowCombo.SetCurrentIndex(i)
			}
		}
		for i, v := range cursorSizes {
			if v == d.cursor {
				cursorCombo.SetCurrentIndex(i)
			}
		}
		capture.SetChecked(d.capture)
		rows["glow"].slider.SetValue(pc(d.glow))
		rows["volume"].slider.SetValue(pc(d.volume))
		rows["music"].slider.SetValue(pc(d.music))
		rows["sound"].slider.SetValue(pc(d.sound))
		rows["camera"].slider.SetValue(tenth(d.camera))
		rows["wheel"].slider.SetValue(tenth(d.wheel))
		rows["double"].slider.SetValue(d.doubleClick)
		applyPicture(d)
	}

	ok := func() {
		n := s
		n.resolution = ""
		if i := resCombo.CurrentIndex(); i > 0 {
			n.resolution = resolutions[i]
		}
		n.fullscreen = fullscreen.Checked()
		n.hd = hd.Checked()
		n.shadows = shadowSizes[max(0, shadowCombo.CurrentIndex())]
		n.cursor = cursorSizes[max(0, cursorCombo.CurrentIndex())]
		n.capture = capture.Checked()
		n.glow = float64(rows["glow"].slider.Value()) / 100
		n.volume = float64(rows["volume"].slider.Value()) / 100
		n.music = float64(rows["music"].slider.Value()) / 100
		n.sound = float64(rows["sound"].slider.Value()) / 100
		n.camera = float64(rows["camera"].slider.Value()) / 10
		n.wheel = float64(rows["wheel"].slider.Value()) / 10
		n.doubleClick = rows["double"].slider.Value()
		n.vivid = colourCombo.CurrentIndex() != 1
		n.exposure = float64(rows["exposure"].slider.Value()) / 100
		n.saturation = float64(rows["saturation"].slider.Value()) / 100
		n.contrast = float64(rows["contrast"].slider.Value()) / 100
		n.anisotropy = anisoValues[max(0, anisoCombo.CurrentIndex())]
		n.vsync = vsyncBox.Checked()
		n.fpsLimit = fpsValues[max(0, fpsCombo.CurrentIndex())]
		n.antialias = aaBox.Checked()
		if err := writeSettingsTxt(p, n, others, isHD); err != nil {
			walk.MsgBox(dlg, "Options", "Could not save settings.txt: "+err.Error(), walk.MsgBoxIconError)
			return
		}
		if err := kb.save(hotkeysPath); err != nil {
			walk.MsgBox(dlg, "Options", "Could not save hotkeys.txt: "+err.Error(), walk.MsgBoxIconError)
			return
		}
		dlg.Accept()
	}

	screen := []Widget{
		Label{Text: "Resolution:"},
		ComboBox{AssignTo: &resCombo, Model: resolutions, CurrentIndex: resIdx, ColumnSpan: 2},
		Label{Text: "Shadows:"},
		ComboBox{AssignTo: &shadowCombo, Model: shadowNames, CurrentIndex: shadowIdx, ColumnSpan: 2},
	}
	screen = append(screen, slider("glow", "Glow of spells:", 0, 100, pc(s.glow))...)
	screen = append(screen,
		CheckBox{AssignTo: &fullscreen, Text: "Full screen", Checked: s.fullscreen, ColumnSpan: 3},
		CheckBox{AssignTo: &hd, Text: "HD textures (Sacrifice HD)", Checked: s.hd, Visible: isHD, ColumnSpan: 3},
	)
	sound := append(append(slider("volume", "Master volume:", 0, 100, pc(s.volume)),
		slider("music", "Music:", 0, 100, pc(s.music))...),
		slider("sound", "Sound effects:", 0, 100, pc(s.sound))...)
	mouse := append(append(slider("camera", "Camera speed:", 2, 30, tenth(s.camera)),
		slider("wheel", "Mouse wheel speed:", 2, 30, tenth(s.wheel))...),
		slider("double", "Double-click time:", 6, 40, s.doubleClick)...)
	mouse = append(mouse,
		Label{Text: "Cursor size:"},
		ComboBox{AssignTo: &cursorCombo, Model: cursorNames, CurrentIndex: cursorIdx, ColumnSpan: 2},
		CheckBox{AssignTo: &capture, Text: "Keep the mouse inside the game window", Checked: s.capture, ColumnSpan: 3},
	)

	idxOf := func(l []int, v int) int {
		for i, x := range l {
			if x == v {
				return i
			}
		}
		return 0
	}
	colourIdx := 0
	if !s.vivid {
		colourIdx = 1
	}
	picture := []Widget{
		Label{Text: "Colours:"},
		ComboBox{AssignTo: &colourCombo, Model: colourNames, CurrentIndex: colourIdx, ColumnSpan: 2},
	}
	picture = append(picture, slider("exposure", "Brightness:", 30, 150, pc(s.exposure))...)
	picture = append(picture, slider("saturation", "Colour strength:", 50, 200, pc(s.saturation))...)
	picture = append(picture, slider("contrast", "Contrast:", 70, 160, pc(s.contrast))...)
	picture = append(picture,
		Label{Text: "Texture sharpness:"},
		ComboBox{AssignTo: &anisoCombo, Model: anisoNames, CurrentIndex: idxOf(anisoValues, s.anisotropy), ColumnSpan: 2},
		CheckBox{AssignTo: &aaBox, Text: "Smooth edges (anti-aliasing)", Checked: s.antialias, ColumnSpan: 3},
	)
	framerate := []Widget{
		CheckBox{AssignTo: &vsyncBox, Text: "VSync (no tearing; frames that come late are not held back)", Checked: s.vsync, ColumnSpan: 3},
		Label{Text: "Frame rate limit:"},
		ComboBox{AssignTo: &fpsCombo, Model: fpsNames, CurrentIndex: idxOf(fpsValues, s.fpsLimit), ColumnSpan: 2},
		Label{Text: "For more frames per second: lower Shadows and Glow on the Game page.", ColumnSpan: 3},
	}
	applyPicture = func(d gameSettings) {
		if d.vivid {
			colourCombo.SetCurrentIndex(0)
		} else {
			colourCombo.SetCurrentIndex(1)
		}
		rows["exposure"].slider.SetValue(pc(d.exposure))
		rows["saturation"].slider.SetValue(pc(d.saturation))
		rows["contrast"].slider.SetValue(pc(d.contrast))
		anisoCombo.SetCurrentIndex(idxOf(anisoValues, d.anisotropy))
		aaBox.SetChecked(d.antialias)
		vsyncBox.SetChecked(d.vsync)
		fpsCombo.SetCurrentIndex(idxOf(fpsValues, d.fpsLimit))
	}

	_, err := Dialog{
		AssignTo: &dlg,
		Title:    "Options",
		MinSize:  Size{Width: 500, Height: 600},
		Layout:   VBox{Margins: Margins{Left: 14, Top: 12, Right: 14, Bottom: 12}},
		Children: []Widget{
			Label{Text: "Options", Font: Font{PointSize: 14, Bold: true}},
			Label{Text: "They apply the next time the game starts."},
			TabWidget{Pages: []TabPage{
				{Title: "Game", Layout: VBox{}, Children: []Widget{
					GroupBox{Title: "Screen", Layout: Grid{Columns: 3}, Children: screen},
					GroupBox{Title: "Sound", Layout: Grid{Columns: 3}, Children: sound},
					GroupBox{Title: "Mouse", Layout: Grid{Columns: 3}, Children: mouse},
					VSpacer{},
				}},
				{Title: "Picture", Layout: VBox{}, Children: []Widget{
					GroupBox{Title: "Picture", Layout: Grid{Columns: 3}, Children: picture},
					GroupBox{Title: "Frame rate", Layout: Grid{Columns: 3}, Children: framerate},
					VSpacer{},
				}},
				keyBindingsPage(func() walk.Form { return dlg }, kb),
			}},
			Composite{Layout: HBox{MarginsZero: true}, Children: []Widget{
				PushButton{Text: "Restore defaults", ToolTipText: "Screen, sound and mouse settings (the Key bindings page has its own button)", OnClicked: func() { apply(defaultSettings(isHD)) }},
				HSpacer{},
				PushButton{Text: "OK", MinSize: Size{Width: 90}, OnClicked: ok},
				PushButton{Text: "Cancel", OnClicked: func() { dlg.Cancel() }},
			}},
		},
	}.Run(owner)
	if err != nil {
		walk.MsgBox(owner, "Options", err.Error(), walk.MsgBoxIconError)
	}
}

// hdEnabled tells whether settings.txt turns the HD textures on
func hdEnabled(dir string, isHD bool) bool {
	s, _ := readSettingsTxt(filepath.Join(dir, "settings.txt"), isHD)
	return isHD && s.hd
}
