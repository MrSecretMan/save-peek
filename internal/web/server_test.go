package web

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/MrSecretMan/save-peek/internal/stardew"
)

func TestProgressCachesUnchangedSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Farm_1")
	writeSave(t, path, "One", 100)

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	s := &Server{Save: stardew.Save{Path: path, Folder: "Farm_1", ModifiedAt: info.ModTime()}}

	first, err := s.progress()
	if err != nil {
		t.Fatal(err)
	}
	if first.PlayerName != "One" {
		t.Fatalf("player = %q", first.PlayerName)
	}

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
	if err := os.Chtimes(path, info.ModTime(), info.ModTime()); err != nil {
		t.Fatal(err)
	}

	second, err := s.progress()
	if err != nil {
		t.Fatal(err)
	}
	if second.PlayerName != "One" {
		t.Fatalf("cache miss: %#v", second)
	}
}

func TestProgressReloadsChangedSave(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "Farm_1")
	writeSave(t, path, "One", 100)

	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	s := &Server{Save: stardew.Save{Path: path, Folder: "Farm_1", ModifiedAt: info.ModTime()}}

	if _, err := s.progress(); err != nil {
		t.Fatal(err)
	}
	writeSave(t, path, "Two", 200)
	newTime := info.ModTime().Add(2 * time.Second)
	if err := os.Chtimes(path, newTime, newTime); err != nil {
		t.Fatal(err)
	}

	progress, err := s.progress()
	if err != nil {
		t.Fatal(err)
	}
	if progress.PlayerName != "Two" || progress.Money != 200 {
		t.Fatalf("stale progress: %#v", progress)
	}
	if !progress.Source.ModifiedAt.Equal(newTime) {
		t.Fatalf("source mtime = %v, want %v", progress.Source.ModifiedAt, newTime)
	}
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

func BenchmarkProgressCached(b *testing.B) {
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
	s := &Server{Save: stardew.Save{Path: path, Folder: "Farm_1", ModifiedAt: info.ModTime()}}
	if _, err := s.progress(); err != nil {
		b.Fatal(err)
	}

	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := s.progress(); err != nil {
			b.Fatal(err)
		}
	}
}
