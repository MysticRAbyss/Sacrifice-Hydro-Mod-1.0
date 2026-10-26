package main

// Load game: the games saved from the in-game menu (Esc > Save Game, or F1 for a quick save) are in
// the "saves" folder next to Hydro.exe: <name>.rcp with a <name>.txt description.

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

type savedGame struct {
	name, file, details string
	time                int64
}

func listSavedGames(dir string) []savedGame {
	var r []savedGame
	files, _ := filepath.Glob(filepath.Join(dir, "saves", "*.rcp"))
	for _, f := range files {
		st, err := os.Stat(f)
		if err != nil {
			continue
		}
		g := savedGame{name: strings.TrimSuffix(filepath.Base(f), ".rcp"), file: f, time: st.ModTime().Unix()}
		if b, err := os.ReadFile(strings.TrimSuffix(f, ".rcp") + ".txt"); err == nil {
			g.details = strings.ReplaceAll(strings.TrimSpace(strings.ReplaceAll(string(b), "\r\n", "\n")), "\n", "\r\n")
		} else {
			g.details = "Saved: " + st.ModTime().Format("2006-01-02 15:04")
		}
		r = append(r, g)
	}
	sort.Slice(r, func(i, j int) bool { return r[i].time > r[j].time })
	return r
}

// showLoadGame returns the saved game to continue, or ""
func showLoadGame(owner walk.Form, dir string) string {
	var dlg *walk.Dialog
	var list *walk.ListBox
	var details *walk.TextEdit
	var loadBtn, deleteBtn *walk.PushButton
	games := listSavedGames(dir)
	names := func() []string {
		n := []string{}
		for _, g := range games {
			n = append(n, g.name)
		}
		return n
	}
	result := ""
	selected := func() {
		i := list.CurrentIndex()
		ok := i >= 0 && i < len(games)
		loadBtn.SetEnabled(ok)
		deleteBtn.SetEnabled(ok)
		if ok {
			details.SetText(games[i].details)
		} else {
			details.SetText("")
		}
	}
	load := func() {
		if i := list.CurrentIndex(); i >= 0 && i < len(games) {
			result = games[i].file
			dlg.Accept()
		}
	}
	empty := "No saved games yet. In a game, press Esc and choose Save Game (or press F1 for a quick save)."
	_, err := Dialog{
		AssignTo: &dlg,
		Title:    "Load game",
		MinSize:  Size{Width: 520, Height: 360},
		Layout:   VBox{Margins: Margins{Left: 12, Top: 10, Right: 12, Bottom: 10}},
		Children: []Widget{
			Label{Text: "Saved games (single player)", Font: Font{Bold: true}},
			Label{Text: empty, Visible: len(games) == 0},
			Composite{Layout: HBox{MarginsZero: true}, Children: []Widget{
				ListBox{AssignTo: &list, Model: names(), MinSize: Size{Width: 220, Height: 240},
					OnCurrentIndexChanged: selected, OnItemActivated: load},
				TextEdit{AssignTo: &details, ReadOnly: true, MinSize: Size{Width: 240}},
			}},
			Composite{Layout: HBox{MarginsZero: true}, Children: []Widget{
				PushButton{AssignTo: &deleteBtn, Text: "Delete", Enabled: false, OnClicked: func() {
					i := list.CurrentIndex()
					if i < 0 || i >= len(games) {
						return
					}
					if walk.MsgBox(dlg, "Load game", "Delete the saved game \""+games[i].name+"\"?", walk.MsgBoxYesNo|walk.MsgBoxIconQuestion) != walk.DlgCmdYes {
						return
					}
					os.Remove(games[i].file)
					os.Remove(strings.TrimSuffix(games[i].file, ".rcp") + ".txt")
					games = listSavedGames(dir)
					list.SetModel(names())
					selected()
				}},
				HSpacer{},
				PushButton{AssignTo: &loadBtn, Text: "Load", Enabled: false, MinSize: Size{Width: 90}, OnClicked: load},
				PushButton{Text: "Cancel", OnClicked: func() { dlg.Cancel() }},
			}},
		},
	}.Run(owner)
	if err != nil {
		walk.MsgBox(owner, "Load game", err.Error(), walk.MsgBoxIconError)
		return ""
	}
	return result
}
