package web

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"log/slog"
	"net/http"
	"time"

	"github.com/MrSecretMan/save-peek/internal/stardew"
)

//go:embed static/*
var files embed.FS

type Server struct {
	Save stardew.Save
}

func (s Server) Handler() http.Handler {
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

func (s Server) save(w http.ResponseWriter, _ *http.Request) {
	progress, err := stardew.Parse(s.Save)
	if err != nil {
		http.Error(w, "could not read save", http.StatusInternalServerError)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.Header().Set("cache-control", "no-store")
	if err := json.NewEncoder(w).Encode(progress); err != nil {
		slog.Error("write response", "error", err)
	}
}

func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		if r.URL.Path != "/healthz" {
			slog.Debug("request", "method", r.Method, "path", r.URL.Path, "took", fmt.Sprint(time.Since(start).Round(time.Millisecond)))
		}
	})
}
