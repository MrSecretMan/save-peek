package web

import (
	"embed"
	"encoding/json"
	"io/fs"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/MrSecretMan/save-peek/internal/stardew"
)

//go:embed static/*
var files embed.FS

const saveCheckInterval = 500 * time.Millisecond

type Server struct {
	Save stardew.Save

	mu        sync.Mutex
	cached    atomic.Pointer[cacheEntry]
	nextCheck atomic.Int64
}

type cacheEntry struct {
	modTime time.Time
	size    int64
	body    []byte
	etag    string
}

func NewServer(save stardew.Save, progress stardew.Progress) *Server {
	s := &Server{Save: save}
	if body, err := json.Marshal(progress); err == nil {
		body = append(body, '\n')
		s.cached.Store(&cacheEntry{
			modTime: save.ModifiedAt,
			size:    save.Size,
			body:    body,
			etag:    makeETag(save.ModifiedAt, save.Size),
		})
		s.nextCheck.Store(time.Now().Add(saveCheckInterval).UnixNano())
	}
	return s
}

func (s *Server) Handler() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/save", s.save)
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("content-type", "text/plain; charset=utf-8")
		_, _ = w.Write([]byte("ok\n"))
	})

	static, err := fs.Sub(files, "static")
	if err != nil {
		panic(err)
	}
	mux.Handle("/", http.FileServer(http.FS(static)))
	return logRequests(mux)
}

func (s *Server) save(w http.ResponseWriter, r *http.Request) {
	entry, err := s.snapshot()
	if err != nil {
		http.Error(w, "could not read save", http.StatusInternalServerError)
		return
	}

	w.Header().Set("content-type", "application/json")
	w.Header().Set("cache-control", "private, max-age=0, must-revalidate")
	w.Header().Set("etag", entry.etag)
	if r.Header.Get("if-none-match") == entry.etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	if _, err := w.Write(entry.body); err != nil {
		slog.Error("write response", "error", err)
	}
}

func (s *Server) snapshot() (*cacheEntry, error) {
	now := time.Now()
	if entry := s.cached.Load(); entry != nil && now.UnixNano() < s.nextCheck.Load() {
		return entry, nil
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	now = time.Now()
	if entry := s.cached.Load(); entry != nil && now.UnixNano() < s.nextCheck.Load() {
		return entry, nil
	}

	info, err := os.Stat(s.Save.Path)
	if err != nil {
		return nil, err
	}
	s.nextCheck.Store(now.Add(saveCheckInterval).UnixNano())

	if entry := s.cached.Load(); entry != nil && entry.size == info.Size() && entry.modTime.Equal(info.ModTime()) {
		return entry, nil
	}

	save := s.Save
	save.ModifiedAt = info.ModTime()
	save.Size = info.Size()
	progress, err := stardew.Parse(save)
	if err != nil {
		return nil, err
	}
	body, err := json.Marshal(progress)
	if err != nil {
		return nil, err
	}
	body = append(body, '\n')

	entry := &cacheEntry{
		modTime: info.ModTime(),
		size:    info.Size(),
		body:    body,
		etag:    makeETag(info.ModTime(), info.Size()),
	}
	s.cached.Store(entry)
	return entry, nil
}

func makeETag(modTime time.Time, size int64) string {
	return `"` + strconv.FormatInt(modTime.UnixNano(), 36) + `-` + strconv.FormatInt(size, 36) + `"`
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		if r.URL.Path != "/healthz" {
			slog.Debug("request", "method", r.Method, "path", r.URL.Path, "took", time.Since(start).Round(time.Millisecond))
		}
	})
}
