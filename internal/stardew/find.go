package stardew

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

func DefaultSaveDirs() []string {
	home, _ := os.UserHomeDir()
	var dirs []string

	switch runtime.GOOS {
	case "windows":
		if appData := os.Getenv("APPDATA"); appData != "" {
			dirs = append(dirs, filepath.Join(appData, "StardewValley", "Saves"))
		}
	case "darwin", "linux":
		if home != "" {
			dirs = append(dirs, filepath.Join(home, ".config", "StardewValley", "Saves"))
		}
	}

	if home != "" {
		dirs = append(dirs, filepath.Join(home, "StardewValley", "Saves"))
	}
	return unique(dirs)
}

func FindLatest(extraDir string) (Save, error) {
	dirs := DefaultSaveDirs()
	if extraDir != "" {
		dirs = append([]string{extraDir}, dirs...)
	}

	var latest Save
	found := false
	for _, dir := range unique(dirs) {
		save, ok, err := latestIn(dir)
		if err != nil || !ok {
			continue
		}
		if !found || save.ModifiedAt.After(latest.ModifiedAt) {
			latest = save
			found = true
		}
	}

	if !found {
		return Save{}, errors.New("no Stardew Valley saves found")
	}
	return latest, nil
}

func latestIn(root string) (Save, bool, error) {
	entries, err := os.ReadDir(root)
	if err != nil {
		return Save{}, false, err
	}

	var latest Save
	found := false
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		folder := filepath.Join(root, entry.Name())
		main, err := mainSaveFile(folder, entry.Name())
		if err != nil {
			continue
		}
		info, err := os.Stat(main)
		if err != nil {
			continue
		}

		save := Save{Path: main, Folder: entry.Name(), ModifiedAt: info.ModTime()}
		if !found || save.ModifiedAt.After(latest.ModifiedAt) {
			latest = save
			found = true
		}
	}
	return latest, found, nil
}

func mainSaveFile(folder, folderName string) (string, error) {
	candidate := filepath.Join(folder, folderName)
	if info, err := os.Stat(candidate); err == nil && !info.IsDir() {
		return candidate, nil
	}

	entries, err := os.ReadDir(folder)
	if err != nil {
		return "", err
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || name == "SaveGameInfo" || strings.HasSuffix(name, "_old") {
			continue
		}
		if strings.Contains(name, "_") {
			return filepath.Join(folder, name), nil
		}
	}
	return "", os.ErrNotExist
}

func unique(values []string) []string {
	seen := make(map[string]struct{}, len(values))
	out := make([]string, 0, len(values))
	for _, value := range values {
		if value == "" {
			continue
		}
		clean := filepath.Clean(value)
		if _, ok := seen[clean]; ok {
			continue
		}
		seen[clean] = struct{}{}
		out = append(out, clean)
	}
	return out
}
