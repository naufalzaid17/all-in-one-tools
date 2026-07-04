// Package public embeds the built Vue frontend.
//
// The Makefile copies the output of `npm run build` (frontend/dist) into
// backend/public/dist before compiling the Go binary, so the whole
// application ships as a single executable.
package public

import (
	"embed"
	"io/fs"
)

// assets embeds the frontend build output. The `all:` prefix includes
// files that start with `.` or `_`, which plain patterns would skip.
//
//go:embed all:dist
var assets embed.FS

// FS returns the frontend filesystem rooted at the build output, so
// callers see index.html at the root.
func FS() fs.FS {
	sub, err := fs.Sub(assets, "dist")
	if err != nil {
		// Unreachable: "dist" is embedded at compile time.
		panic(err)
	}
	return sub
}
