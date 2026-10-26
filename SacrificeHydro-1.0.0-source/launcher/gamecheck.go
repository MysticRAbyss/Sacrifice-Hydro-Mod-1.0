package main

import (
	"os"
	"path/filepath"
)

// originalGameFiles are files of the original Sacrifice that the mod cannot run without
var originalGameFiles = []string{
	"data/menus.wad", "data/main.wad", "data/models.wad", "data/spells.wad", "data/sounds.wad",
	"data/saxs.wad", "data/interfac.wad",
}

// missingGameFiles lists the original game's files that are not in folder (none: it is a Sacrifice folder)
func missingGameFiles(folder string) []string {
	var missing []string
	for _, f := range originalGameFiles {
		if st, err := os.Stat(filepath.Join(folder, filepath.FromSlash(f))); err != nil || st.Size() == 0 {
			missing = append(missing, f)
		}
	}
	if st, err := os.Stat(filepath.Join(folder, "maps")); err != nil || !st.IsDir() {
		missing = append(missing, "maps")
	}
	return missing
}
