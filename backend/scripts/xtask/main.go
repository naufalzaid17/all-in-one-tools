// Command xtask implements the OS-agnostic build steps invoked by the
// root Makefile. Make recipes run through cmd.exe on Windows and sh on
// unix, so anything beyond "cd && go/npm ..." lives here instead of in
// shell syntax. Stdlib only; run from the backend directory.
//
// Subcommands:
//
//	embed-assets  copy ../frontend/dist into public/dist (go:embed root)
//	ensure-bin    create ../bin for the compiled binary
//	clean         remove build artifacts (bin, frontend/dist, embedded dist)
package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

const (
	frontendDist = "../frontend/dist"
	embedDist    = "public/dist"
	binDir       = "../bin"
	keepFile     = ".gitkeep"
)

func main() {
	if len(os.Args) != 2 {
		fatal(fmt.Errorf("usage: xtask <embed-assets|ensure-bin|clean>"))
	}
	switch os.Args[1] {
	case "embed-assets":
		if _, err := os.Stat(frontendDist); err != nil {
			fatal(fmt.Errorf("frontend build output not found at %s — run the frontend build first", frontendDist))
		}
		fatalIf(clearDir(embedDist))
		fatalIf(copyTree(frontendDist, embedDist))
		fmt.Println("xtask: embedded frontend assets into", embedDist)
	case "ensure-bin":
		fatalIf(os.MkdirAll(binDir, 0o755))
	case "clean":
		fatalIf(os.RemoveAll(binDir))
		fatalIf(os.RemoveAll(frontendDist))
		fatalIf(clearDir(embedDist))
		fmt.Println("xtask: cleaned build artifacts")
	default:
		fatal(fmt.Errorf("unknown subcommand %q", os.Args[1]))
	}
}

// clearDir removes everything inside dir except the git placeholder,
// creating the directory if it does not exist yet.
func clearDir(dir string) error {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return err
	}
	for _, e := range entries {
		if e.Name() == keepFile {
			continue
		}
		if err := os.RemoveAll(filepath.Join(dir, e.Name())); err != nil {
			return err
		}
	}
	return nil
}

// copyTree recursively copies src into dst using filepath so it works
// identically on every OS.
func copyTree(src, dst string) error {
	return filepath.WalkDir(src, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}
		target := filepath.Join(dst, rel)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}
		return copyFile(path, target)
	})
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	if _, err := io.Copy(out, in); err != nil {
		out.Close()
		return err
	}
	return out.Close()
}

func fatalIf(err error) {
	if err != nil {
		fatal(err)
	}
}

func fatal(err error) {
	fmt.Fprintln(os.Stderr, "xtask:", err)
	os.Exit(1)
}
