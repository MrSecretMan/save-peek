package web

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MrSecretMan/save-peek/internal/stardew"
)

func TestMultiServerListsSavesInDiscoveryOrder(t *testing.T) {
	dir := t.TempDir()
	newer := multiSave(t, dir, "NewFarm_2", "New", 200, time.Now())
	older := multiSave(t, dir, "OldFarm_1", "Old", 100, time.Now().Add(-time.Hour))

	handler := NewMultiServer([]stardew.Save{newer, older}).Handler()
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/saves", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}

	var saves []struct {
		ID         string `json:"id"`
		PlayerName string `json:"player_name"`
		FarmName   string `json:"farm_name"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &saves); err != nil {
		t.Fatal(err)
	}
	if len(saves) != 2 {
		t.Fatalf("got %d saves, want 2", len(saves))
	}
	if saves[0].ID != "NewFarm_2" || saves[0].PlayerName != "New" {
		t.Fatalf("first save = %#v", saves[0])
	}
	if saves[1].ID != "OldFarm_1" || saves[1].PlayerName != "Old" {
		t.Fatalf("second save = %#v", saves[1])
	}
}

func TestMultiServerSelectsRequestedSave(t *testing.T) {
	dir := t.TempDir()
	first := multiSave(t, dir, "One_1", "One", 100, time.Now())
	second := multiSave(t, dir, "Two_2", "Two", 200, time.Now().Add(-time.Minute))

	handler := NewMultiServer([]stardew.Save{first, second}).Handler()
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/save?id=Two_2", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d", rec.Code)
	}

	var progress stardew.Progress
	if err := json.Unmarshal(rec.Body.Bytes(), &progress); err != nil {
		t.Fatal(err)
	}
	if progress.PlayerName != "Two" || progress.Money != 200 {
		t.Fatalf("selected wrong save: %#v", progress)
	}
}

func TestMultiServerDefaultsToNewestSave(t *testing.T) {
	dir := t.TempDir()
	newer := multiSave(t, dir, "New_2", "New", 200, time.Now())
	older := multiSave(t, dir, "Old_1", "Old", 100, time.Now().Add(-time.Hour))

	handler := NewMultiServer([]stardew.Save{newer, older}).Handler()
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/save", nil))

	var progress stardew.Progress
	if err := json.Unmarshal(rec.Body.Bytes(), &progress); err != nil {
		t.Fatal(err)
	}
	if progress.PlayerName != "New" {
		t.Fatalf("defaulted to %q, want newest save", progress.PlayerName)
	}
}

func multiSave(t *testing.T, root, folder, player string, money int, modified time.Time) stardew.Save {
	t.Helper()
	dir := filepath.Join(root, folder)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(dir, folder)
	data := []byte("<SaveGame><player><name>" + player + "</name><money>" + itoa(money) + "</money></player><farmName>" + folder + "</farmName><currentSeason>spring</currentSeason><dayOfMonth>1</dayOfMonth><year>1</year></SaveGame>")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, modified, modified); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	return stardew.Save{Path: path, Folder: folder, ModifiedAt: info.ModTime(), Size: info.Size()}
}
