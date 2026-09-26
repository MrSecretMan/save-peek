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

func TestSnapshotKeepsUnchangedSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Farm_1")
	writeSave(t, path, "One", 100)

	save, initial := openSave(t, path)
	s := NewServer(save, initial)

	original, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	broken := make([]byte, len(original))
	for i := range broken {
		broken[i] = 'x'
	}
	if err := os.WriteFile(path, broken, 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, save.ModifiedAt, save.ModifiedAt); err != nil {
		t.Fatal(err)
	}

	s.nextCheck.Store(0)
	entry, err := s.snapshot()
	if err != nil {
		t.Fatal(err)
	}
	progress := decodeProgress(t, entry.body)
	if progress.PlayerName != "One" {
		t.Fatalf("cache miss: %#v", progress)
	}
}

func TestSnapshotReloadsChangedSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Farm_1")
	writeSave(t, path, "One", 100)

	save, initial := openSave(t, path)
	s := NewServer(save, initial)

	writeSave(t, path, "Two", 200)
	newTime := save.ModifiedAt.Add(2 * time.Second)
	if err := os.Chtimes(path, newTime, newTime); err != nil {
		t.Fatal(err)
	}

	s.nextCheck.Store(0)
	entry, err := s.snapshot()
	if err != nil {
		t.Fatal(err)
	}
	progress := decodeProgress(t, entry.body)
	if progress.PlayerName != "Two" || progress.Money != 200 {
		t.Fatalf("stale progress: %#v", progress)
	}
	if !progress.Source.ModifiedAt.Equal(newTime) {
		t.Fatalf("source mtime = %v, want %v", progress.Source.ModifiedAt, newTime)
	}
}

func openSave(t *testing.T, path string) (stardew.Save, stardew.Progress) {
	t.Helper()
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	save := stardew.Save{Path: path, Folder: "Farm_1", ModifiedAt: info.ModTime(), Size: info.Size()}
	progress, err := stardew.Parse(save)
	if err != nil {
		t.Fatal(err)
	}
	return save, progress
}

func decodeProgress(t *testing.T, body []byte) stardew.Progress {
	t.Helper()
	var progress stardew.Progress
	if err := json.Unmarshal(body, &progress); err != nil {
		t.Fatal(err)
	}
	return progress
}

func writeSave(t *testing.T, path, name string, money int) {
	t.Helper()
	data := []byte("<SaveGame><player><name>" + name + "</name><money>" + itoa(money) + "</money></player><farmName>Farm</farmName><currentSeason>spring</currentSeason><dayOfMonth>1</dayOfMonth><year>1</year></SaveGame>")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatal(err)
	}
}

func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	buf := [20]byte{}
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

func BenchmarkSnapshotCached(b *testing.B) {
	dir := b.TempDir()
	path := filepath.Join(dir, "Farm_1")
	data := []byte("<SaveGame><player><name>One</name><money>100</money></player><farmName>Farm</farmName><currentSeason>spring</currentSeason><dayOfMonth>1</dayOfMonth><year>1</year></SaveGame>")
	if err := os.WriteFile(path, data, 0o644); err != nil {
		b.Fatal(err)
	}

	info, err := os.Stat(path)
	if err != nil {
		b.Fatal(err)
	}
	save := stardew.Save{Path: path, Folder: "Farm_1", ModifiedAt: info.ModTime(), Size: info.Size()}
	progress, err := stardew.Parse(save)
	if err != nil {
		b.Fatal(err)
	}
	s := NewServer(save, progress)

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.snapshot(); err != nil {
			b.Fatal(err)
		}
	}
}

func TestSaveHandlerUsesETag(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Farm_1")
	writeSave(t, path, "One", 100)
	save, initial := openSave(t, path)
	handler := NewServer(save, initial).Handler()

	first := httptest.NewRecorder()
	handler.ServeHTTP(first, httptest.NewRequest(http.MethodGet, "/api/save", nil))
	if first.Code != http.StatusOK {
		t.Fatalf("first status = %d", first.Code)
	}
	etag := first.Header().Get("etag")
	if etag == "" {
		t.Fatal("missing etag")
	}

	request := httptest.NewRequest(http.MethodGet, "/api/save", nil)
	request.Header.Set("if-none-match", etag)
	second := httptest.NewRecorder()
	handler.ServeHTTP(second, request)
	if second.Code != http.StatusNotModified {
		t.Fatalf("second status = %d", second.Code)
	}
	if second.Body.Len() != 0 {
		t.Fatalf("304 body = %q", second.Body.String())
	}
}
