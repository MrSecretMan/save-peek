package stardew

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestLatestInPicksNewestSave(t *testing.T) {
	root := t.TempDir()
	writeFoundSave(t, root, "OldFarm_1", time.Now().Add(-time.Hour))
	writeFoundSave(t, root, "NewFarm_2", time.Now())
	save, ok, err := latestIn(root)
	if err != nil { t.Fatal(err) }
	if !ok { t.Fatal("expected a save") }
	if save.Folder != "NewFarm_2" { t.Fatalf("got %q", save.Folder) }
}

func TestSavesInReturnsEverySave(t *testing.T) {
	root := t.TempDir()
	writeFoundSave(t, root, "One_1", time.Now().Add(-time.Hour))
	writeFoundSave(t, root, "Two_2", time.Now())
	saves, err := savesIn(root)
	if err != nil { t.Fatal(err) }
	if len(saves) != 2 { t.Fatalf("got %d saves, want 2", len(saves)) }
	seen := map[string]bool{}
	for _, save := range saves { seen[save.Folder] = true }
	if !seen["One_1"] || !seen["Two_2"] { t.Fatalf("unexpected saves: %#v", saves) }
}

func writeFoundSave(t *testing.T, root, name string, modified time.Time) {
	t.Helper()
	dir := filepath.Join(root, name)
	if err := os.MkdirAll(dir, 0o755); err != nil { t.Fatal(err) }
	path := filepath.Join(dir, name)
	if err := os.WriteFile(path, []byte("<SaveGame/>"), 0o644); err != nil { t.Fatal(err) }
	if err := os.Chtimes(path, modified, modified); err != nil { t.Fatal(err) }
}
