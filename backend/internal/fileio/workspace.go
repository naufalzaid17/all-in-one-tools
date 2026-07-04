// Package fileio provides OS-agnostic file handling for tools that
// process uploads: a per-request temporary workspace and streaming
// multipart parsing.
//
// All paths are derived from os.TempDir via os.MkdirTemp and joined with
// path/filepath, so behaviour is identical on Linux, Windows and macOS.
// Nothing in this package ever hardcodes a directory separator or a
// platform-specific location.
package fileio

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Workspace is an isolated temporary directory scoped to a single
// request. Large files are staged here instead of in memory; Close
// removes the directory and everything in it.
type Workspace struct {
	dir    string
	nextID int
}

// NewWorkspace creates a fresh temporary directory under the platform's
// temp location (os.TempDir on every OS).
func NewWorkspace() (*Workspace, error) {
	dir, err := os.MkdirTemp("", "aio-tools-*")
	if err != nil {
		return nil, fmt.Errorf("fileio: create workspace: %w", err)
	}
	return &Workspace{dir: dir}, nil
}

// Close removes the workspace directory and all files inside it. It is
// safe to call multiple times.
func (ws *Workspace) Close() error {
	if ws.dir == "" {
		return nil
	}
	dir := ws.dir
	ws.dir = ""
	return os.RemoveAll(dir)
}

// Path returns an absolute path inside the workspace for the given file
// name. The name is sanitised to its base component, so uploaded names
// like "../../etc/passwd" cannot escape the workspace.
func (ws *Workspace) Path(name string) string {
	return filepath.Join(ws.dir, sanitizeName(name, ws.bump()))
}

// CreateFile creates a new file inside the workspace and returns it
// along with its absolute path. The caller must close the file.
func (ws *Workspace) CreateFile(name string) (*os.File, string, error) {
	p := ws.Path(name)
	f, err := os.Create(p)
	if err != nil {
		return nil, "", fmt.Errorf("fileio: create %s: %w", filepath.Base(p), err)
	}
	return f, p, nil
}

// SaveReader streams r into a new workspace file, enforcing limit bytes
// (0 means unlimited). It returns the absolute path and the number of
// bytes written. ErrTooLarge is returned when the limit is exceeded.
func (ws *Workspace) SaveReader(name string, r io.Reader, limit int64) (string, int64, error) {
	f, p, err := ws.CreateFile(name)
	if err != nil {
		return "", 0, err
	}
	defer f.Close()

	src := r
	if limit > 0 {
		// Read one extra byte so we can distinguish "exactly at the
		// limit" from "over it".
		src = io.LimitReader(r, limit+1)
	}
	n, err := io.Copy(f, src)
	if err != nil {
		return "", 0, fmt.Errorf("fileio: save %s: %w", filepath.Base(p), err)
	}
	if limit > 0 && n > limit {
		return "", 0, ErrTooLarge
	}
	return p, n, nil
}

// bump returns a unique per-workspace counter used to disambiguate
// duplicate or empty upload names.
func (ws *Workspace) bump() int {
	ws.nextID++
	return ws.nextID
}

// sanitizeName reduces an untrusted file name to a safe on-disk name.
// The per-workspace counter prefix guarantees two uploads with the same
// name never overwrite each other.
func sanitizeName(name string, id int) string {
	base := SafeBaseName(name)
	if base == "" {
		return fmt.Sprintf("upload-%d", id)
	}
	return fmt.Sprintf("%d-%s", id, base)
}

// SafeBaseName reduces an untrusted file name to its base component so
// names like "../../etc/passwd" or "C:\evil.exe" cannot traverse
// directories on any platform. It returns "" when nothing usable
// remains.
func SafeBaseName(name string) string {
	// filepath.Base handles the native separator; normalise backslashes
	// explicitly so unix hosts also strip Windows-style paths.
	base := filepath.Base(strings.ReplaceAll(name, "\\", "/"))
	base = strings.TrimSpace(base)
	if base == "" || base == "." || base == ".." || base == "/" {
		return ""
	}
	return base
}
