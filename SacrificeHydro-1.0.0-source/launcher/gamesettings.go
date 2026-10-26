package main

// The Game Settings page (as in the original game's Game Settings screen: game type, the
// tick-box rules and the scale bars) and the Spellbook page (a custom spellbook, which the original
// game unlocked after the campaign). Both are kept in launcher.ini and turned into Hydro.exe options.

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

var wizardNames = []string{"Random", "Abraxus", "Acheron", "Ambassador Buta", "Charlotte", "Eldred", "Grakkus", "Hachimen",
	"Jadugarr", "Marduk", "Mithras", "Seerix", "Shakti", "Sorcha", "The Ragman", "Yogo"}

var gameTypes = []string{"Skirmish", "Slaughter", "Soul Harvest", "Domination"}
var gameTypeGoal = []string{"", "Kills to win:", "Souls to win:", "Manaliths to win:"}

type tick struct {
	key, text, flag string // flag: the engine option (with --no- for off)
	def          bool
	tip          string
}

var ticks = []tick{
	{"randomAltar", "Random Altar", "shuffle-altars", false, "Players get random altars instead of altar 1 for slot 1 and so on."},
	{"observers", "Observers", "allow-observers", true, "Observers may join the game (multiplayer)."},
	{"observerChat", "Observer chat", "observer-chat", true, "Observers can chat with each other."},
	{"customSpellbooks", "Custom spellbooks", "custom-spellbooks", true, "Players may use their own spellbooks (see Spellbook...)."},
	{"alliedVision", "Allied vision", "allied-vision", true, "Allies share what they see."},
	{"collectAlliedSouls", "Collect Allied Souls", "collect-allied-souls", false, "Allies can collect the souls of each other's dead creatures."},
	{"friendlyFire", "Friendly Fire", "friendly-fire", true, "Your attacks can hurt your own and your allies' units."},
	{"alwaysGib", "Always Gib", "always-gib", false, "Every killed creature is gibbed; its soul is free for anyone."},
	{"noMapCreatures", "No map creatures", "map-creatures", false, "Removes the creatures the map maker placed."},
	{"noMapSouls", "No map souls", "map-souls", false, "Removes the blue souls the map maker placed."},
	{"revealBlueSouls", "Reveal blue souls", "reveal-blue-souls", true, "Blue souls show on the minimap from the start."},
}

type scale struct {
	key, text          string
	min, max, def, div int // value = position / div
	opt                string
}

var scales = []scale{
	{"souls", "Starting Souls", 0, 24, 12, 1, "souls"},
	{"minLevel", "Minimum Level", 1, 9, 1, 1, "min-level"},
	{"maxLevel", "Maximum Level", 1, 9, 9, 1, "max-level"},
	{"aiDamage", "AI damage multiplier", 50, 200, 100, 100, "bot-damage-factor"},
	{"gibChance", "Gib Chance", 0, 400, 100, 100, "gib-chance"},
	{"xpRate", "XP Rate", 0, 400, 100, 100, "xp-rate"},
	{"friendlyDamage", "Friendly Damage", 0, 200, 100, 100, "friendly-damage"},
}

func scaleText(s scale, v int) string {
	if s.div == 1 {
		return strconv.Itoa(v)
	}
	return fmt.Sprintf("%.2f", float64(v)/float64(s.div))
}

func cfgInt(cfg settings, key string, def int) int {
	if v, err := strconv.Atoi(cfg.get(key, "")); err == nil {
		return v
	}
	return def
}

// gameSettingArgs turns the Game Settings page into Hydro.exe options (for single player and host)
func gameSettingArgs(cfg settings) []string {
	var a []string
	gt := indexOf(gameTypes, cfg.get("gameType", "Skirmish"))
	goal := cfgInt(cfg, "gameGoal", 10)
	switch gt {
	case 1:
		a = append(a, fmt.Sprintf("--slaughter=%d", goal))
	case 2:
		a = append(a, fmt.Sprintf("--soul-harvest=%d", goal))
	case 3:
		a = append(a, fmt.Sprintf("--domination=%d", goal))
	default:
		a = append(a, "--skirmish")
	}
	for _, t := range ticks {
		on := cfg.get(t.key, strconv.FormatBool(t.def)) == "true"
		if strings.HasPrefix(t.key, "noMap") { // "No map creatures" ticked = map creatures off
			on = !on
		}
		if on {
			a = append(a, "--"+t.flag)
		} else {
			a = append(a, "--no-"+t.flag)
		}
	}
	for _, s := range scales {
		v := cfgInt(cfg, s.key, s.def)
		a = append(a, "--"+s.opt+"="+scaleText(s, v))
		if s.key == "minLevel" { // wizards start at the minimum level, as in the original game
			a = append(a, fmt.Sprintf("--level=%d", v))
		}
	}
	return a
}

func showGameSettings(owner walk.Form, cfg settings, save func()) {
	var dlg *walk.Dialog
	var typeCombo *walk.ComboBox
	var goalLabel *walk.Label
	var goalEdit *walk.NumberEdit
	checks := make([]*walk.CheckBox, len(ticks))
	sliders := make([]*walk.Slider, len(scales))
	values := make([]*walk.Label, len(scales))

	refresh := func() {
		gt := typeCombo.CurrentIndex()
		goalLabel.SetVisible(gt > 0)
		goalEdit.SetVisible(gt > 0)
		if gt > 0 {
			goalLabel.SetText(gameTypeGoal[gt])
		}
	}
	var tickWidgets []Widget
	for i, t := range ticks {
		i, t := i, t
		tickWidgets = append(tickWidgets, CheckBox{AssignTo: &checks[i], Text: t.text, ToolTipText: t.tip,
			Checked: cfg.get(t.key, strconv.FormatBool(t.def)) == "true"})
	}
	var scaleWidgets []Widget
	for i, s := range scales {
		i, s := i, s
		v := cfgInt(cfg, s.key, s.def)
		scaleWidgets = append(scaleWidgets,
			Label{Text: s.text + ":"},
			Slider{AssignTo: &sliders[i], MinValue: s.min, MaxValue: s.max, Value: v, ToolTipsHidden: true,
				OnValueChanged: func() {
					values[i].SetText(scaleText(s, sliders[i].Value()))
					// the minimum level cannot be above the maximum and the other way round
					if s.key == "minLevel" && sliders[2].Value() < sliders[1].Value() {
						sliders[2].SetValue(sliders[1].Value())
					}
					if s.key == "maxLevel" && sliders[1].Value() > sliders[2].Value() {
						sliders[1].SetValue(sliders[2].Value())
					}
				}},
			Label{AssignTo: &values[i], Text: scaleText(s, v), MinSize: Size{Width: 40}},
		)
	}
	ok := func() {
		cfg["gameType"] = typeCombo.Text()
		cfg["gameGoal"] = strconv.Itoa(int(goalEdit.Value()))
		for i, t := range ticks {
			cfg[t.key] = strconv.FormatBool(checks[i].Checked())
		}
		for i, s := range scales {
			cfg[s.key] = strconv.Itoa(sliders[i].Value())
		}
		save()
		dlg.Accept()
	}
	defaults := func() {
		typeCombo.SetCurrentIndex(0)
		goalEdit.SetValue(10)
		for i, t := range ticks {
			checks[i].SetChecked(t.def)
		}
		for i, s := range scales {
			sliders[i].SetValue(s.def)
		}
		refresh()
	}
	err := Dialog{
		AssignTo: &dlg,
		Title:    "Game Settings",
		MinSize:  Size{Width: 560, Height: 560},
		Layout:   VBox{Margins: Margins{Left: 14, Top: 12, Right: 14, Bottom: 12}},
		Children: []Widget{
			Label{Text: "Game Settings", Font: Font{PointSize: 14, Bold: true}},
			GroupBox{Title: "Game type", Layout: Grid{Columns: 4}, Children: []Widget{
				Label{Text: "Game type:"},
				ComboBox{AssignTo: &typeCombo, Model: gameTypes, CurrentIndex: indexOf(gameTypes, cfg.get("gameType", "Skirmish")), OnCurrentIndexChanged: refresh},
				Label{AssignTo: &goalLabel, Text: ""},
				NumberEdit{AssignTo: &goalEdit, Value: float64(cfgInt(cfg, "gameGoal", 10)), MinValue: 1, MaxValue: 999, Decimals: 0, SpinButtonsVisible: true},
				Label{ColumnSpan: 4, Text: "Skirmish: desecrate every enemy altar. Slaughter: reach the kill count first. Soul Harvest:\r\nhold that many souls first. Domination: build that many manaliths first."},
			}},
			GroupBox{Title: "Rules", Layout: Grid{Columns: 3}, Children: tickWidgets},
			GroupBox{Title: "Settings", Layout: Grid{Columns: 3}, Children: scaleWidgets},
			VSpacer{},
			Composite{Layout: HBox{MarginsZero: true}, Children: []Widget{
				PushButton{Text: "Restore defaults", OnClicked: defaults},
				HSpacer{},
				PushButton{Text: "OK", MinSize: Size{Width: 90}, OnClicked: ok},
				PushButton{Text: "Cancel", OnClicked: func() { dlg.Cancel() }},
			}},
		},
	}.Create(owner)
	if err != nil {
		walk.MsgBox(owner, "Game Settings", err.Error(), walk.MsgBoxIconError)
		return
	}
	refresh()
	dlg.Run()
}

// --- the spellbook ---

// a spellbook slot: what kind of entry it holds and which entry of each god's list goes there
type bookSlot struct {
	level    int
	creature bool
	index    int
}

// the slots of a spellbook, in the game's order (the same for every god)
var bookSlots = []bookSlot{
	{1, true, 1}, {1, true, 2}, {1, true, 3}, {1, false, 2},
	{2, true, 4}, {2, false, 3},
	{3, true, 5}, {3, false, 4},
	{4, true, 6}, {4, false, 5},
	{5, true, 7}, {5, false, 6},
	{6, true, 8}, {6, false, 7},
	{7, true, 9}, {7, false, 8},
	{8, false, 9}, {8, false, 10},
	{9, true, 10},
}

var bookGods = []string{"Persephone", "Pyro", "James", "Stratos", "Charnel", "Hydro"}

// slotChoices lists, for a slot, every god's entry: "Name (God)" and its tag
func slotChoices(s bookSlot) (names, tags []string) {
	for _, g := range bookGods {
		e := godSpells[g][s.index]
		if s.creature {
			e = godCreatures[g][s.index]
		}
		names = append(names, fmt.Sprintf("%s  (%s)", e[1], g))
		tags = append(tags, e[0])
	}
	return
}

// spellbookArg is the --spellbook option of the saved custom spellbook ("" if none is used)
func spellbookArg(cfg settings) string {
	if cfg.get("useSpellbook", "false") != "true" {
		return ""
	}
	var parts []string
	for i, s := range bookSlots {
		_, tags := slotChoices(s)
		g := indexOf(bookGods, cfg.get(fmt.Sprintf("book%d", i), "Hydro"))
		parts = append(parts, fmt.Sprintf("%d:@%s", s.level, tags[g]))
	}
	return "--spellbook=" + strings.Join(parts, ",")
}

func showSpellbook(owner walk.Form, cfg settings, save func()) {
	var dlg *walk.Dialog
	var use *walk.CheckBox
	var fill *walk.ComboBox
	combos := make([]*walk.ComboBox, len(bookSlots))
	var rows []Widget
	for i, s := range bookSlots {
		names, _ := slotChoices(s)
		kind := "Spell"
		if s.creature {
			kind = "Creature"
		}
		rows = append(rows,
			Label{Text: fmt.Sprintf("Level %d %s:", s.level, kind)},
			ComboBox{AssignTo: &combos[i], Model: names, CurrentIndex: indexOf(bookGods, cfg.get(fmt.Sprintf("book%d", i), "Hydro"))},
		)
	}
	enable := func() {
		for _, c := range combos {
			c.SetEnabled(use.Checked())
		}
		fill.SetEnabled(use.Checked())
	}
	err := Dialog{
		AssignTo: &dlg,
		Title:    "Spellbook",
		MinSize:  Size{Width: 500, Height: 640},
		Layout:   VBox{Margins: Margins{Left: 14, Top: 12, Right: 14, Bottom: 12}},
		Children: []Widget{
			Label{Text: "Spellbook", Font: Font{PointSize: 14, Bold: true}},
			Label{Text: "Build your own spellbook from the creatures and spells of every god, as after the\r\ncampaign in the original game. Each slot takes one god's entry of that level.\r\nThe host's \"Custom spellbooks\" rule must be on, otherwise your god's usual book is used."},
			CheckBox{AssignTo: &use, Text: "Use my custom spellbook", Checked: cfg.get("useSpellbook", "false") == "true", OnCheckedChanged: enable},
			Composite{Layout: HBox{MarginsZero: true}, Children: []Widget{
				Label{Text: "Fill every slot from:"},
				ComboBox{AssignTo: &fill, Model: bookGods, CurrentIndex: -1, OnCurrentIndexChanged: func() {
					if fill.CurrentIndex() >= 0 {
						for _, c := range combos {
							c.SetCurrentIndex(fill.CurrentIndex())
						}
					}
				}},
				HSpacer{},
			}},
			ScrollView{HorizontalFixed: true, Layout: Grid{Columns: 2}, Children: rows},
			Composite{Layout: HBox{MarginsZero: true}, Children: []Widget{
				HSpacer{},
				PushButton{Text: "OK", MinSize: Size{Width: 90}, OnClicked: func() {
					cfg["useSpellbook"] = strconv.FormatBool(use.Checked())
					for i, c := range combos {
						cfg[fmt.Sprintf("book%d", i)] = bookGods[max(0, c.CurrentIndex())]
					}
					save()
					dlg.Accept()
				}},
				PushButton{Text: "Cancel", OnClicked: func() { dlg.Cancel() }},
			}},
		},
	}.Create(owner)
	if err != nil {
		walk.MsgBox(owner, "Spellbook", err.Error(), walk.MsgBoxIconError)
		return
	}
	enable()
	dlg.Run()
}
