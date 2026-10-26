package main

// The Key bindings page of the Options window: shows hotkeys.txt (next to Hydro.exe) as a table and
// lets the player change, add and remove keys. The file keeps its layout; only the key part of the
// changed lines is rewritten. The game reads hotkeys.txt every time it starts.

import (
	_ "embed"
	"os"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

//go:embed hotkeys_default.txt
var defaultHotkeys string

type keyRow struct {
	line   int    // line in the file
	group  string // "Orders", "Formations", ...
	action string // as written in the file, e.g. `Attack` or `Send Chat Message "gg"`
	keys   []string
}

type keyBindings struct {
	lines []string
	rows  []keyRow
}

func parseKeyBindings(text string) *keyBindings {
	kb := &keyBindings{}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	kb.lines = strings.Split(strings.TrimRight(text, "\n"), "\n")
	group := ""
	for i, l := range kb.lines {
		t := strings.TrimSpace(l)
		if strings.HasPrefix(t, "#") {
			group = strings.TrimSpace(strings.TrimLeft(t, "#"))
			continue
		}
		c := strings.LastIndex(t, ":")
		if t == "" || c < 0 {
			continue
		}
		r := keyRow{line: i, group: group, action: strings.TrimSpace(t[:c])}
		for _, k := range strings.Split(t[c+1:], " or ") {
			if k = strings.TrimSpace(k); k != "" {
				r.keys = append(r.keys, k)
			}
		}
		kb.rows = append(kb.rows, r)
	}
	return kb
}

func loadKeyBindings(p string) *keyBindings {
	b, err := os.ReadFile(p)
	if err != nil || len(strings.TrimSpace(string(b))) == 0 {
		return parseKeyBindings(defaultHotkeys)
	}
	return parseKeyBindings(string(b))
}

func (kb *keyBindings) save(p string) error {
	for _, r := range kb.rows {
		kb.lines[r.line] = r.action + ": " + strings.Join(r.keys, " or ")
	}
	return os.WriteFile(p, []byte(strings.Join(kb.lines, "\r\n")+"\r\n"), 0644)
}

// keyNames are the key names the game understands (keycodes.d), in the order the list shows them
var keyNames = func() []string {
	l := []string{}
	for c := 'a'; c <= 'z'; c++ {
		l = append(l, string(c))
	}
	for c := '0'; c <= '9'; c++ {
		l = append(l, string(c))
	}
	for i := 1; i <= 12; i++ {
		l = append(l, "F"+itoa(i))
	}
	l = append(l, "Space", "Tab", "Return", "Backspace", "Escape", "Alt", "Ctrl", "Shift", "CAPS",
		"Up Arrow", "Down Arrow", "Left Arrow", "Right Arrow", "Insert", "Delete", "Home", "End", "Page Up", "Page Down",
		"Pause", "Scroll Lock", "Num-lock", "Print Screen")
	for c := '0'; c <= '9'; c++ {
		l = append(l, "Keypad-"+string(c))
	}
	l = append(l, "Keypad-+", "Keypad--", "Keypad-*", "Keypad-/", "Keypad-.", "Keypad-enter",
		"-", "=", "[", "]", "\\", ";", "'", "`", ",", ".", "/")
	return l
}()

func itoa(i int) string {
	if i < 10 {
		return string(rune('0' + i))
	}
	return string(rune('0'+i/10)) + string(rune('0'+i%10))
}

// keyFromWalk turns a key press into the game's name for it ("" if the game has no name for it)
func keyFromWalk(k walk.Key) string {
	switch {
	case k >= walk.KeyA && k <= walk.KeyZ:
		return string(rune('a' + (k - walk.KeyA)))
	case k >= walk.Key0 && k <= walk.Key9:
		return string(rune('0' + (k - walk.Key0)))
	case k >= walk.KeyF1 && k <= walk.KeyF12:
		return "F" + itoa(int(k-walk.KeyF1)+1)
	case k >= walk.KeyNumpad0 && k <= walk.KeyNumpad9:
		return "Keypad-" + string(rune('0'+(k-walk.KeyNumpad0)))
	}
	names := map[walk.Key]string{
		walk.KeySpace: "Space", walk.KeyTab: "Tab", walk.KeyReturn: "Return", walk.KeyBack: "Backspace",
		walk.KeyMenu: "Alt", walk.KeyCapital: "CAPS",
		walk.KeyUp: "Up Arrow", walk.KeyDown: "Down Arrow", walk.KeyLeft: "Left Arrow", walk.KeyRight: "Right Arrow",
		walk.KeyInsert: "Insert", walk.KeyDelete: "Delete", walk.KeyHome: "Home", walk.KeyEnd: "End",
		walk.KeyPrior: "Page Up", walk.KeyNext: "Page Down", walk.KeyPause: "Pause", walk.KeyScroll: "Scroll Lock",
		walk.KeyNumlock: "Num-lock", walk.KeySnapshot: "Print Screen",
		walk.KeyAdd: "Keypad-+", walk.KeySubtract: "Keypad--", walk.KeyMultiply: "Keypad-*", walk.KeyDivide: "Keypad-/",
		walk.KeyDecimal:  "Keypad-.",
		walk.KeyOEMMinus: "-", walk.KeyOEMPlus: "=", walk.KeyOEM4: "[", walk.KeyOEM6: "]", walk.KeyOEM5: "\\",
		walk.KeyOEM1: ";", walk.KeyOEM7: "'", walk.KeyOEM3: "`", walk.KeyOEMComma: ",", walk.KeyOEMPeriod: ".", walk.KeyOEM2: "/",
	}
	return names[k]
}

// splitKey splits "Shift-Ctrl-r" into its modifiers and key
func splitKey(k string) (ctrl, shift bool, key string) {
	for {
		lk := strings.ToLower(k)
		switch {
		case strings.HasPrefix(lk, "ctrl-") && len(k) > 5:
			ctrl, k = true, k[5:]
		case strings.HasPrefix(lk, "shift-") && len(k) > 6:
			shift, k = true, k[6:]
		default:
			return ctrl, shift, k
		}
	}
}

func joinKey(ctrl, shift bool, key string) string {
	s := ""
	if shift {
		s += "Shift-"
	}
	if ctrl {
		s += "Ctrl-"
	}
	return s + key
}

func sameKey(a, b string) bool {
	ac, as, ak := splitKey(a)
	bc, bs, bk := splitKey(b)
	return ac == bc && as == bs && strings.EqualFold(ak, bk)
}

type keyModel struct {
	walk.TableModelBase
	kb *keyBindings
}

func (m *keyModel) RowCount() int { return len(m.kb.rows) }
func (m *keyModel) Value(row, col int) interface{} {
	r := m.kb.rows[row]
	switch col {
	case 0:
		return r.group
	case 1:
		return r.action
	}
	if len(r.keys) == 0 {
		return "(none)"
	}
	return strings.Join(r.keys, "  or  ")
}

// askKey shows the small window that picks one key; "" if cancelled
func askKey(owner walk.Form, action, current string) string {
	var dlg *walk.Dialog
	var ctrlBox, shiftBox *walk.CheckBox
	var keyCombo *walk.ComboBox
	c, s, k := splitKey(current)
	idx := 0
	for i, n := range keyNames {
		if strings.EqualFold(n, k) {
			idx = i
		}
	}
	result := ""
	_, err := Dialog{
		AssignTo: &dlg,
		Title:    "Key for " + action,
		MinSize:  Size{Width: 360, Height: 200},
		Layout:   VBox{Margins: Margins{Left: 12, Top: 10, Right: 12, Bottom: 10}},
		Children: []Widget{
			Label{Text: "Press the key here (hold Ctrl or Shift for a combination),\nor choose it from the list. Alt can only be chosen from the list."},
			LineEdit{ReadOnly: true, Text: "Press a key...", OnKeyDown: func(key walk.Key) {
				name := keyFromWalk(key)
				if name == "" {
					return
				}
				mods := walk.ModifiersDown()
				ctrlBox.SetChecked(mods&walk.ModControl != 0)
				shiftBox.SetChecked(mods&walk.ModShift != 0)
				for i, n := range keyNames {
					if n == name {
						keyCombo.SetCurrentIndex(i)
					}
				}
			}},
			Composite{Layout: HBox{MarginsZero: true}, Children: []Widget{
				CheckBox{AssignTo: &ctrlBox, Text: "Ctrl", Checked: c},
				CheckBox{AssignTo: &shiftBox, Text: "Shift", Checked: s},
				ComboBox{AssignTo: &keyCombo, Model: keyNames, CurrentIndex: idx},
				HSpacer{},
			}},
			VSpacer{},
			Composite{Layout: HBox{MarginsZero: true}, Children: []Widget{
				HSpacer{},
				PushButton{Text: "OK", MinSize: Size{Width: 80}, OnClicked: func() {
					result = joinKey(ctrlBox.Checked(), shiftBox.Checked(), keyNames[max(0, keyCombo.CurrentIndex())])
					dlg.Accept()
				}},
				PushButton{Text: "Cancel", OnClicked: func() { dlg.Cancel() }},
			}},
		},
	}.Run(owner)
	if err != nil {
		return ""
	}
	return result
}

// keyBindingsPage builds the page; kb is changed in place and saved by the Options window's OK
func keyBindingsPage(owner func() walk.Form, kb *keyBindings) TabPage {
	var table *walk.TableView
	model := &keyModel{kb: kb}
	current := func() int {
		i := table.CurrentIndex()
		if i < 0 || i >= len(kb.rows) {
			return -1
		}
		return i
	}
	// gives the key to row i, taking it away from any other action (after asking)
	assign := func(i, keep int, key string) {
		for j := range kb.rows {
			if j == i {
				continue
			}
			for n, k := range kb.rows[j].keys {
				if sameKey(k, key) {
					if walk.MsgBox(owner(), "Key bindings", key+" is used for "+kb.rows[j].action+".\r\nUse it for "+kb.rows[i].action+" instead?", walk.MsgBoxYesNo|walk.MsgBoxIconQuestion) != walk.DlgCmdYes {
						return
					}
					kb.rows[j].keys = append(kb.rows[j].keys[:n:n], kb.rows[j].keys[n+1:]...)
					break
				}
			}
		}
		r := &kb.rows[i]
		if keep >= 0 && keep < len(r.keys) {
			r.keys[keep] = key
		} else {
			r.keys = append(r.keys, key)
		}
		model.PublishRowsReset()
		table.SetCurrentIndex(i)
	}
	change := func() {
		i := current()
		if i < 0 {
			return
		}
		cur := ""
		if len(kb.rows[i].keys) > 0 {
			cur = kb.rows[i].keys[0]
		}
		if k := askKey(owner(), kb.rows[i].action, cur); k != "" {
			if len(kb.rows[i].keys) > 1 {
				kb.rows[i].keys = kb.rows[i].keys[:1]
			}
			assign(i, 0, k)
		}
	}
	return TabPage{
		Title:  "Key bindings",
		Layout: VBox{},
		Children: []Widget{
			Label{Text: "Double-click an action (or select it and click Change) to give it a new key."},
			TableView{
				AssignTo:         &table,
				AlternatingRowBG: true,
				MinSize:          Size{Width: 400, Height: 380},
				Columns: []TableViewColumn{
					{Title: "Group", Width: 110},
					{Title: "Action", Width: 160},
					{Title: "Keys", Width: 150},
				},
				Model:           model,
				OnItemActivated: change,
			},
			Composite{Layout: HBox{MarginsZero: true}, Children: []Widget{
				PushButton{Text: "Change...", OnClicked: change},
				PushButton{Text: "Add another key...", OnClicked: func() {
					if i := current(); i >= 0 {
						if k := askKey(owner(), kb.rows[i].action, ""); k != "" {
							for _, have := range kb.rows[i].keys {
								if sameKey(have, k) {
									return
								}
							}
							assign(i, -1, k)
						}
					}
				}},
				PushButton{Text: "Remove keys", OnClicked: func() {
					if i := current(); i >= 0 {
						kb.rows[i].keys = nil
						model.PublishRowsReset()
						table.SetCurrentIndex(i)
					}
				}},
				HSpacer{},
				PushButton{Text: "Default keys", OnClicked: func() {
					d := parseKeyBindings(defaultHotkeys)
					*kb = *d
					model.PublishRowsReset()
				}},
			}},
		},
	}
}
