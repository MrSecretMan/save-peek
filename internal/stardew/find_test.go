package stardew

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFindLatestInCustomDir(t *testing.T) {
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
	_ = os.Chtimes(filepath.Join(older, filepath.Base(older)), oldTime, oldTime)
	_ = os.Chtimes(filepath.Join(newer, filepath.Base(newer)), newTime, newTime)

	saves, err := findIn(root)
	if err != nil {
		t.Fatal(err)
	}
	if len(saves) != 2 {
		t.Fatalf("got %d saves", len(saves))
	}
}
