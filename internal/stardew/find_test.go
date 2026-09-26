package stardew

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLatestInPicksNewestSave(t *testing.T) {
	root := t.TempDir()
	older := filepath.Join(root, "OldFarm_1")
	newer := filepath.Join(root, "NewFarm_2")
	for _, dir := range []string{older, newer} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatal(err)
		}
		name := filepath.Base(dir)
		if err := os.WriteFile(filepath.Join(dir, name), []byte("<SaveGame/>"), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	oldTime := time.Now().Add(-time.Hour)
	newTime := time.Now()
	if err := os.Chtimes(filepath.Join(older, filepath.Base(older)), oldTime, oldTime); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(filepath.Join(newer, filepath.Base(newer)), newTime, newTime); err != nil {
		t.Fatal(err)
	}

	save, ok, err := latestIn(root)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected a save")
	}
	if save.Folder != "NewFarm_2" {
		t.Fatalf("got %q", save.Folder)
	}
}
