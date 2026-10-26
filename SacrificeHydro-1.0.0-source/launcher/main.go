// Sacrifice Hydro launcher: the start window of the Hydro mod (and of the Hydro HD package).
// sacengine has no main menu of its own, so this window picks the game - single player against the
// computer, hosting or joining a multiplayer game - your god, the map and the opponents, then starts
// Hydro.exe with those settings. In the HD package (an HDTextures folder next to it) it first makes
// sure the high-resolution textures are built (Sacrifice HD's texture builder) and turns them on.
package main

import (
	"bufio"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

var gods = []string{"Random", "Persephone", "Pyro", "James", "Stratos", "Charnel", "Hydro"}

const maxBots = 5

type mapInfo struct {
	name, file string
	players    int
}

type settings map[string]string

func (s settings) get(k, def string) string {
	if v, ok := s[k]; ok && v != "" {
		return v
	}
	return def
}

func loadSettings(p string) settings {
	s := settings{}
	f, err := os.Open(p)
	if err != nil {
		return s
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	for sc.Scan() {
		if k, v, ok := strings.Cut(sc.Text(), "="); ok {
			s[strings.TrimSpace(k)] = strings.TrimSpace(v)
		}
	}
	return s
}

func saveSettings(p string, s settings) {
	var keys []string
	for k := range s {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for _, k := range keys {
		fmt.Fprintf(&b, "%s=%s\r\n", k, s[k])
	}
	os.WriteFile(p, []byte(b.String()), 0o644)
}

// the maps of the game: <Sacrifice>\maps\(N) Name.scp, N = number of players
func findMaps(game string) []mapInfo {
	re := regexp.MustCompile(`^\((\d+)\)\s*(.+)\.scp$`)
	fs, _ := filepath.Glob(filepath.Join(game, "maps", "*.scp"))
	var ms []mapInfo
	for _, f := range fs {
		b := filepath.Base(f)
		m := re.FindStringSubmatch(b)
		if m == nil {
			continue
		}
		n, _ := strconv.Atoi(m[1])
		if n < 2 {
			continue
		}
		ms = append(ms, mapInfo{name: fmt.Sprintf("%s  (%d players)", m[2], n), file: b, players: n})
	}
	sort.Slice(ms, func(i, j int) bool {
		if ms[i].players != ms[j].players {
			return ms[i].players < ms[j].players
		}
		return ms[i].name < ms[j].name
	})
	return ms
}

func indexOf(l []string, s string) int {
	for i, x := range l {
		if strings.EqualFold(x, s) {
			return i
		}
	}
	return 0
}

func main() {
	exe, _ := os.Executable()
	dir := filepath.Dir(exe)
	game := filepath.Dir(dir)
	hdTools := filepath.Join(dir, "HDTextures", "SacrificeHD.exe")
	_, err := os.Stat(hdTools)
	isHD := err == nil
	title := "Sacrifice: Hydro"
	if isHD {
		title = "Sacrifice Hydro HD"
	}
	if missing := missingGameFiles(game); len(missing) > 0 {
		walk.MsgBox(nil, title, "Sacrifice: Hydro needs the original Sacrifice (GOG or Steam), and this folder\r\n"+
			"has to be inside your Sacrifice folder, next to its \"data\" and \"maps\" folders.\r\n\r\n"+
			"Not found in "+game+":\r\n"+strings.Join(missing, "\r\n")+"\r\n\r\n"+
			"Run the Setup again and pick your Sacrifice folder, or move this folder there.", walk.MsgBoxIconError)
		return
	}
	cfgPath := filepath.Join(dir, "launcher.ini")
	cfg := loadSettings(cfgPath)

	maps := findMaps(game)
	var mapNames []string
	for _, m := range maps {
		mapNames = append(mapNames, m.name)
	}

	var mw *walk.MainWindow
	var rbSingle, rbHost, rbJoin *walk.RadioButton
	var nameEdit, ipEdit *walk.LineEdit
	var godCombo, mapCombo, wizardCombo *walk.ComboBox
	var gameSettingsBtn *walk.PushButton
	var botCount, playerCount *walk.NumberEdit
	var botCombos [maxBots]*walk.ComboBox
	var botLabels [maxBots]*walk.Label
	var singleBox, hostBox, joinBox, mapBox *walk.GroupBox
	var status *walk.Label

	botChildren := []Widget{}
	for i := 0; i < maxBots; i++ {
		i := i
		botChildren = append(botChildren,
			Label{AssignTo: &botLabels[i], Text: fmt.Sprintf("Computer %d:", i+1)},
			ComboBox{AssignTo: &botCombos[i], Model: gods, CurrentIndex: indexOf(gods, cfg.get(fmt.Sprintf("bot%d", i+1), "Random"))},
		)
	}

	mapPlayers := func() int {
		if mapCombo == nil || mapCombo.CurrentIndex() < 0 || mapCombo.CurrentIndex() >= len(maps) {
			return 2
		}
		return maps[mapCombo.CurrentIndex()].players
	}
	refresh := func() {
		if mw == nil {
			return
		}
		single, host, join := rbSingle.Checked(), rbHost.Checked(), rbJoin.Checked()
		singleBox.SetVisible(single)
		hostBox.SetVisible(host)
		joinBox.SetVisible(join)
		mapBox.SetVisible(!join)
		gameSettingsBtn.SetEnabled(!join) // the host decides the rules
		n := mapPlayers()
		botCount.SetRange(1, float64(n-1))
		if botCount.Value() > float64(n-1) {
			botCount.SetValue(float64(n - 1))
		}
		playerCount.SetRange(2, float64(n))
		if playerCount.Value() > float64(n) {
			playerCount.SetValue(float64(n))
		}
		for i := 0; i < maxBots; i++ {
			show := i < int(botCount.Value())
			botLabels[i].SetVisible(show)
			botCombos[i].SetVisible(show)
		}
	}

	godArg := func(name string) string { return strings.ToLower(name) }

	var launch func(args []string)
	play := func() {
		args := []string{}
		if n := strings.TrimSpace(nameEdit.Text()); n != "" {
			args = append(args, "--name="+n)
		}
		if g := godCombo.Text(); g != "Random" {
			args = append(args, "--god="+godArg(g))
		}
		if w := wizardCombo.Text(); w != "Random" {
			args = append(args, "--wizard="+w)
		}
		if b := spellbookArg(cfg); b != "" {
			args = append(args, b)
		}
		if !rbJoin.Checked() {
			args = append(args, gameSettingArgs(cfg)...)
		}
		mapFile := ""
		if i := mapCombo.CurrentIndex(); i >= 0 && i < len(maps) {
			mapFile = filepath.Join("maps", maps[i].file)
		}
		switch {
		case rbSingle.Checked():
			nb := int(botCount.Value())
			var slots, bg []string
			for i := 0; i < nb; i++ {
				slots = append(slots, strconv.Itoa(i+1))
				bg = append(bg, godArg(botCombos[i].Text()))
			}
			args = append(args, fmt.Sprintf("--num-slots=%d", nb+1), "--bot-slots="+strings.Join(slots, ","), "--bot-gods="+strings.Join(bg, ","))
			if mapFile != "" {
				args = append(args, mapFile)
			}
		case rbHost.Checked():
			args = append(args, fmt.Sprintf("--host=%d", int(playerCount.Value())))
			if mapFile != "" {
				args = append(args, mapFile)
			}
		case rbJoin.Checked():
			ip := strings.TrimSpace(ipEdit.Text())
			if ip == "" {
				walk.MsgBox(mw, title, "Enter the address of the player who hosts the game.", walk.MsgBoxIconInformation)
				return
			}
			args = append(args, "--join="+ip)
		}
		// remember the choices for next time
		mode := "single"
		if rbHost.Checked() {
			mode = "host"
		} else if rbJoin.Checked() {
			mode = "join"
		}
		cfg["mode"], cfg["name"], cfg["god"], cfg["join"] = mode, nameEdit.Text(), godCombo.Text(), ipEdit.Text()
		cfg["wizard"] = wizardCombo.Text()
		if i := mapCombo.CurrentIndex(); i >= 0 && i < len(maps) {
			cfg["map"] = maps[i].file
		}
		cfg["bots"] = strconv.Itoa(int(botCount.Value()))
		cfg["players"] = strconv.Itoa(int(playerCount.Value()))
		for i := 0; i < maxBots; i++ {
			cfg[fmt.Sprintf("bot%d", i+1)] = botCombos[i].Text()
		}
		saveSettings(cfgPath, cfg)
		launch(args)
	}

	launch = func(args []string) {
		mw.SetEnabled(false)
		withHD := hdEnabled(dir, isHD)
		if withHD {
			status.SetText("Checking the HD textures (the first time this builds them: a few minutes)...")
		} else {
			status.SetText("Starting...")
		}
		go func() {
			if withHD {
				prepare(dir, game, func(line string) {
					mw.Synchronize(func() { status.SetText(line) })
				})
				mw.Synchronize(func() { status.SetText("Starting the game...") })
			}
			msg := runGame(dir, game, args, func() { mw.Synchronize(func() { mw.SetVisible(false) }) })
			mw.Synchronize(func() {
				if msg == "" {
					mw.Close()
					return
				}
				mw.SetVisible(true)
				walk.MsgBox(mw, title, msg, walk.MsgBoxIconError)
				mw.SetEnabled(true)
				status.SetText("")
			})
		}()
	}

	modeNow := cfg.get("mode", "single")
	mapIdx := 0
	for i, m := range maps {
		if m.file == cfg.get("map", "(2) Alpine Meadow.scp") {
			mapIdx = i
		}
	}
	atoi := func(s string, d int) float64 {
		if v, err := strconv.Atoi(s); err == nil {
			return float64(v)
		}
		return float64(d)
	}

	header := "Hydro, god of water, joins the five gods of Sacrifice."
	if isHD {
		header += " With Sacrifice HD's high-resolution textures."
	}
	err = MainWindow{
		AssignTo: &mw,
		Title:    title,
		MinSize:  Size{Width: 460, Height: 520},
		Size:     Size{Width: 480, Height: 560},
		Layout:   VBox{Margins: Margins{Left: 14, Top: 12, Right: 14, Bottom: 12}},
		Children: []Widget{
			Label{Text: title, Font: Font{PointSize: 16, Bold: true}},
			Label{Text: header},
			GroupBox{
				Title:  "Game",
				Layout: VBox{},
				Children: []Widget{
					RadioButton{AssignTo: &rbSingle, Text: "Single player - play against the computer", OnClicked: refresh},
					RadioButton{AssignTo: &rbHost, Text: "Multiplayer - host a game", OnClicked: refresh},
					RadioButton{AssignTo: &rbJoin, Text: "Multiplayer - join a game", OnClicked: refresh},
				},
			},
			GroupBox{
				Title:  "You",
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{Text: "Name:"},
					LineEdit{AssignTo: &nameEdit, Text: cfg.get("name", "")},
					Label{Text: "God:"},
					ComboBox{AssignTo: &godCombo, Model: gods, CurrentIndex: indexOf(gods, cfg.get("god", "Hydro"))},
					Label{Text: "Wizard:"},
					ComboBox{AssignTo: &wizardCombo, Model: wizardNames, CurrentIndex: indexOf(wizardNames, cfg.get("wizard", "Random"))},
					Label{Text: ""},
					PushButton{Text: "Spellbook...", OnClicked: func() { showSpellbook(mw, cfg, func() { saveSettings(cfgPath, cfg) }) }},
				},
			},
			GroupBox{
				AssignTo: &mapBox,
				Title:    "Map",
				Layout:   VBox{},
				Children: []Widget{
					ComboBox{AssignTo: &mapCombo, Model: mapNames, CurrentIndex: mapIdx, OnCurrentIndexChanged: refresh},
				},
			},
			GroupBox{
				AssignTo: &singleBox,
				Title:    "Computer opponents",
				Layout:   Grid{Columns: 2},
				Children: append([]Widget{
					Label{Text: "How many:"},
					NumberEdit{AssignTo: &botCount, Value: atoi(cfg.get("bots", "1"), 1), MinValue: 1, MaxValue: maxBots, Decimals: 0, SpinButtonsVisible: true, OnValueChanged: refresh},
				}, botChildren...),
			},
			GroupBox{
				AssignTo: &hostBox,
				Title:    "Host",
				Layout:   Grid{Columns: 2},
				Children: []Widget{
					Label{Text: "Players:"},
					NumberEdit{AssignTo: &playerCount, Value: atoi(cfg.get("players", "2"), 2), MinValue: 2, MaxValue: 6, Decimals: 0, SpinButtonsVisible: true},
					Label{Text: "The others join with your IP address; the game starts when everyone is in.", ColumnSpan: 2},
				},
			},
			GroupBox{
				AssignTo: &joinBox,
				Title:    "Join",
				Layout:   Grid{Columns: 2},
				Children: []Widget{
					Label{Text: "Host address:"},
					LineEdit{AssignTo: &ipEdit, Text: cfg.get("join", "")},
					Label{Text: "The host picks the map. Everyone needs the same version of the mod.", ColumnSpan: 2},
				},
			},
			VSpacer{},
			Label{AssignTo: &status, Text: ""},
			Composite{
				Layout: HBox{MarginsZero: true},
				Children: []Widget{
					PushButton{Text: "Options...", OnClicked: func() { showOptions(mw, dir, isHD) }},
					PushButton{AssignTo: &gameSettingsBtn, Text: "Game settings...", OnClicked: func() { showGameSettings(mw, cfg, func() { saveSettings(cfgPath, cfg) }) }},
					HSpacer{},
					PushButton{Text: "Load game...", OnClicked: func() {
						if f := showLoadGame(mw, dir); f != "" {
							launch([]string{"--continue=" + f})
						}
					}},
					PushButton{Text: "Play", MinSize: Size{Width: 110}, OnClicked: play},
					PushButton{Text: "Quit", OnClicked: func() { mw.Close() }},
				},
			},
		},
	}.Create()
	if err != nil {
		walk.MsgBox(nil, title, err.Error(), walk.MsgBoxIconError)
		return
	}
	if len(maps) == 0 {
		walk.MsgBox(mw, title, "No maps found in "+filepath.Join(game, "maps")+".\r\nPut this folder inside your Sacrifice folder (next to Sacrifice.exe).", walk.MsgBoxIconWarning)
	}
	switch modeNow {
	case "host":
		rbHost.SetChecked(true)
	case "join":
		rbJoin.SetChecked(true)
	default:
		rbSingle.SetChecked(true)
	}
	refresh()
	mw.Run()
}

const createNoWindow = 0x08000000

// prepare builds or checks the HD textures with Sacrifice HD's texture builder; its progress is
// shown in the launcher window (the first build takes a few minutes)
func prepare(dir, game string, progress func(string)) {
	// if Sacrifice HD is installed and one of its games did not close properly, let it put the
	// original files back before anything reads them
	hdMod := filepath.Join(game, "SacrificeHD")
	if _, err := os.Stat(filepath.Join(hdMod, "swapped.json")); err == nil {
		c := exec.Command(filepath.Join(hdMod, "SacrificeHD.exe"), "-restore")
		c.Dir = hdMod
		c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
		c.Run()
	}
	progress("Checking the HD textures...")
	c := exec.Command(filepath.Join(dir, "HDTextures", "SacrificeHD.exe"), "-generate-only", "-game", game)
	c.Dir = filepath.Join(dir, "HDTextures")
	c.SysProcAttr = &syscall.SysProcAttr{HideWindow: true, CreationFlags: createNoWindow}
	out, err := c.StdoutPipe()
	if err != nil {
		return
	}
	c.Stderr = c.Stdout
	if err := c.Start(); err != nil {
		return
	}
	step := ""
	sc := bufio.NewScanner(out)
	for sc.Scan() {
		line := strings.TrimSpace(sc.Text())
		if len(line) > 9 && line[2] == ':' && line[5] == ':' { // drop the time stamp
			line = strings.TrimSpace(line[9:])
		}
		switch {
		case line == "":
		case strings.HasPrefix(line, "["):
			step = line
			progress("Building the HD textures (once, a few minutes): " + step)
		case strings.HasPrefix(line, "Upscaling") || strings.HasPrefix(line, "Preparing"):
			progress("Building the HD textures (once, a few minutes): " + step + " - " + line)
		case strings.HasPrefix(line, "HD textures"):
			progress(line)
		}
	}
	c.Wait()
}

// runGame starts Hydro.exe without a console window (its messages go to SacEngine.out.txt and
// SacEngine.err.txt in the Sacrifice folder). It waits until the game ends and returns what went
// wrong if it ended with an error ("" otherwise); running is called once the game is up.
func runGame(dir, game string, args []string, running func()) string {
	errFile := filepath.Join(game, "SacEngine.err.txt")
	os.Remove(errFile)
	os.Remove(filepath.Join(game, "SacEngine.out.txt"))
	cmd := exec.Command(filepath.Join(dir, "Hydro.exe"), append([]string{"--redirect-output"}, args...)...)
	cmd.Dir = dir
	// no console window, but the game window must be shown normally: HideWindow would also hide it
	// (Windows applies the start-up show state to a program's first window)
	cmd.SysProcAttr = &syscall.SysProcAttr{CreationFlags: createNoWindow}
	if err := cmd.Start(); err != nil {
		return "Could not start Hydro.exe: " + err.Error()
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	select {
	case err := <-done:
		return crashReport(game, errFile, err)
	case <-time.After(20 * time.Second):
		running() // the game is up: the launcher hides, and stays to report a crash later
		return crashReport(game, errFile, <-done)
	}
}

// crashReport tells what went wrong when the game ended with an error ("" if it ended normally)
func crashReport(game, errFile string, err error) string {
	{
		if err == nil {
			return "" // closed normally (for example with Alt+F4 or Quit) - nothing to report
		}
		b, _ := os.ReadFile(errFile)
		var lines []string
		for _, l := range strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n") {
			if l = strings.TrimSpace(l); l != "" && !strings.HasPrefix(l, "warning:") && !strings.HasPrefix(l, "hydro") {
				lines = append(lines, l)
			}
		}
		// a crash report goes to SacEngine.out.txt: show its first line (the reason)
		outFile := filepath.Join(game, "SacEngine.out.txt")
		if b, err := os.ReadFile(outFile); err == nil {
			for _, l := range strings.Split(strings.ReplaceAll(string(b), "\r\n", "\n"), "\n") {
				if strings.Contains(l, "Exception@") || strings.Contains(l, "Error@") {
					lines = append(lines, strings.TrimSpace(l))
					errFile = outFile
					break
				}
			}
		}
		if len(lines) > 25 {
			lines = lines[:25]
		}
		msg := "The game stopped (" + err.Error() + ")."
		if len(lines) > 0 {
			msg += "\r\n\r\n" + strings.Join(lines, "\r\n")
		}
		return msg + "\r\n\r\nThe full message is in " + errFile + "."
	}
}
