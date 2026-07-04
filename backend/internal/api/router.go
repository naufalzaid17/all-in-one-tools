// Package api wires the HTTP surface of the application: the versioned
// REST API for tools and the embedded single-page frontend.
package api

import (
	"io/fs"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/naufalzaid17/all-in-one-tools/backend/internal/core"
)

// toolInfo is the discovery metadata exposed for each registered tool.
type toolInfo struct {
	ID          string        `json:"id"`
	Name        string        `json:"name"`
	Category    core.Category `json:"category"`
	Description string        `json:"description"`
}

// NewRouter builds the application router: middleware, health check,
// the /api/v1 tool endpoints and the SPA fallback serving staticFS.
func NewRouter(registry *core.Registry, staticFS fs.FS) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Route("/api/v1", func(v1 chi.Router) {
		v1.Get("/health", handleHealth)

		v1.Route("/tools", func(tr chi.Router) {
			tr.Get("/", handleListTools(registry))
			for _, tool := range registry.Tools() {
				tr.Route("/"+tool.ID(), tool.RegisterRoutes)
			}
		})

		// Unknown API routes must return JSON, never the SPA fallback.
		v1.NotFound(func(w http.ResponseWriter, _ *http.Request) {
			core.RespondError(w, http.StatusNotFound, core.CodeNotFound, "endpoint not found")
		})
		v1.MethodNotAllowed(func(w http.ResponseWriter, _ *http.Request) {
			core.RespondError(w, http.StatusMethodNotAllowed, core.CodeInvalidRequest, "method not allowed")
		})
	})

	// Everything else is the embedded Vue SPA.
	r.NotFound(NewSPAHandler(staticFS).ServeHTTP)

	return r
}

func handleHealth(w http.ResponseWriter, _ *http.Request) {
	core.RespondOK(w, map[string]string{"status": "ok"})
}

// handleListTools exposes tool metadata so the frontend (or any client)
// can discover what is available without hardcoding the catalogue.
func handleListTools(registry *core.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, _ *http.Request) {
		tools := registry.Tools()
		infos := make([]toolInfo, 0, len(tools))
		for _, t := range tools {
			infos = append(infos, toolInfo{
				ID:          t.ID(),
				Name:        t.Name(),
				Category:    t.Category(),
				Description: t.Description(),
			})
		}
		core.RespondOK(w, map[string]any{"tools": infos})
	}
}
