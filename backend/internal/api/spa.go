package api

import (
	"io/fs"
	"net/http"
	"path"
	"strings"
)

// SPAHandler serves an embedded single-page application. Requests that
// match a real file are served as-is; anything else falls back to
// index.html so that client-side routing (vue-router history mode)
// works on hard refreshes and deep links.
type SPAHandler struct {
	fsys       fs.FS
	fileServer http.Handler
}

// NewSPAHandler wraps the given filesystem (the root must contain
// index.html) in an SPA-aware static file handler.
func NewSPAHandler(fsys fs.FS) *SPAHandler {
	return &SPAHandler{
		fsys:       fsys,
		fileServer: http.FileServer(http.FS(fsys)),
	}
}

func (h *SPAHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	name := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if name == "" {
		name = "index.html"
	}

	if _, err := fs.Stat(h.fsys, name); err == nil {
		// Fingerprinted build assets (dist/assets/*) are safe to cache
		// aggressively; their names change on every content change.
		if strings.HasPrefix(name, "assets/") {
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		}
		h.fileServer.ServeHTTP(w, r)
		return
	}

	index, err := fs.ReadFile(h.fsys, "index.html")
	if err != nil {
		// The frontend has not been built into the binary yet.
		http.Error(w,
			"frontend not built: run `make build` to embed the Vue app",
			http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(index)
}
