package stardew

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

func DefaultSaveDirs() []string {
	home, _ := os.UserHomeDir()
	var dirs []string
	switch runtime.GOOS {
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" { dirs = append(dirs, filepath.Join(appData, "StardewValley", "Saves")) }
	case "darwin", "linux":
		if home != "" { dirs = append(dirs, filepath.Join(home, ".config", "StardewValley", "Saves")) }
	}
	if home != "" { dirs = append(dirs, filepath.Join(home, "StardewValley", "Saves")) }
	return unique(dirs)
}

func FindAll(extraDir string) ([]Save, error) {
	dirs := DefaultSaveDirs()
	if extraDir != "" { dirs = append([]string{extraDir}, dirs...) }
	seen := make(map[string]struct{})
	var saves []Save
	for _, dir := range unique(dirs) {
		found, err := savesIn(dir)
		if err != nil { continue }
		for _, save := range found {
			clean := filepath.Clean(save.Path)
			if _, ok := seen[clean]; ok { continue }
			seen[clean] = struct{}{}
			saves = append(saves, save)
		}
	}
	if len(saves) == 0 { return nil, errors.New("no Stardew Valley saves found") }
	sort.SliceStable(saves, func(i, j int) bool { return saves[i].ModifiedAt.After(saves[j].ModifiedAt) })
	return saves, nil
}

func FindLatest(extraDir string) (Save, error) {
	saves, err := FindAll(extraDir)
	if err != nil { return Save{}, err }
	return saves[0], nil
}

func savesIn(root string) ([]Save, error) {
	entries, err := os.ReadDir(root)
	if err != nil { return nil, err }
	saves := make([]Save, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() { continue }
		folder := filepath.Join(root, entry.Name())
		main, err := mainSaveFile(folder, entry.Name())
		if err != nil { continue }
		info, err := os.Stat(main)
		if err != nil { continue }
		saves = append(saves, Save{Path: main, Folder: entry.Name(), ModifiedAt: info.ModTime(), Size: info.Size()})
	}
	return saves, nil
}

func latestIn(root string) (Save, bool, error) {
	saves, err := savesIn(root)
	if err != nil || len(saves) == 0 { return Save{}, false, err }
	sort.SliceStable(saves, func(i, j int) bool { return saves[i].ModifiedAt.After(saves[j].ModifiedAt) })
	return saves[0], true, nil
}

func mainSaveFile(folder, folderName string) (string, error) {
	candidate := filepath.Join(folder, folderName)
	if info, err := os.Stat(candidate); err == nil && !info.IsDir() { return candidate, nil }
	entries, err := os.ReadDir(folder)
	if err != nil { return "", err }
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || name == "SaveGameInfo" || strings.HasSuffix(name, "_old") { continue }
		if strings.Contains(name, "_") { return filepath.Join(folder, name), nil }
	}
	return "", os.ErrNotExist
}

func unique(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" { continue }
		clean := filepath.Clean(value)
		if _, ok := seen[clean]; ok { continue }
		seen[clean] = struct{}{}
		out = append(out, clean)
	}
	return out
}
