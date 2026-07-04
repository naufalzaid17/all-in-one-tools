package fileio

import (
	"errors"
	"fmt"
	"io"
	"net/http"
	"path/filepath"
	"strings"
)

// Upload size limits. Individual tools can tighten these via
// ParseOptions but never exceed them.
const (
	DefaultMaxFileSize  = 200 << 20 // 200 MiB per file
	DefaultMaxTotalSize = 512 << 20 // 512 MiB per request
	maxFieldSize        = 1 << 20   // 1 MiB per text field
	maxParts            = 64
)

// ErrTooLarge is returned when an upload exceeds the configured limits.
var ErrTooLarge = errors.New("fileio: upload exceeds the size limit")

// UploadedFile describes one file part that was streamed to disk.
type UploadedFile struct {
	// FieldName is the multipart field the file arrived under.
	FieldName string
	// Name is the sanitised original file name (base name only).
	Name string
	// Path is the absolute location of the staged file inside the
	// request workspace.
	Path string
	// Size is the file size in bytes.
	Size int64
}

// Form is the result of parsing a multipart request: file parts staged
// in the workspace and regular text fields collected in memory.
type Form struct {
	Files  []UploadedFile
	Values map[string]string
}

// Value returns the first value for a text field ("" when absent).
func (f *Form) Value(name string) string { return f.Values[name] }

// FilesFor returns the uploads submitted under the given field name.
func (f *Form) FilesFor(field string) []UploadedFile {
	var out []UploadedFile
	for _, uf := range f.Files {
		if uf.FieldName == field {
			out = append(out, uf)
		}
	}
	return out
}

// ParseOptions bounds a multipart parse.
type ParseOptions struct {
	// MaxFileSize caps each individual file (default DefaultMaxFileSize).
	MaxFileSize int64
	// MaxTotalSize caps the sum of all file parts (default DefaultMaxTotalSize).
	MaxTotalSize int64
	// AllowedExtensions optionally restricts uploads to the given
	// lowercase extensions including the dot (e.g. ".pdf"). Empty means
	// any extension is accepted.
	AllowedExtensions []string
}

// ParseMultipart streams every part of a multipart/form-data request:
// file parts go straight to the workspace (never fully into memory),
// text fields are collected into Form.Values. Parsing is strict about
// limits so a single request cannot exhaust disk or RAM.
func ParseMultipart(r *http.Request, ws *Workspace, opts ParseOptions) (*Form, error) {
	if opts.MaxFileSize <= 0 || opts.MaxFileSize > DefaultMaxFileSize {
		opts.MaxFileSize = DefaultMaxFileSize
	}
	if opts.MaxTotalSize <= 0 || opts.MaxTotalSize > DefaultMaxTotalSize {
		opts.MaxTotalSize = DefaultMaxTotalSize
	}

	mr, err := r.MultipartReader()
	if err != nil {
		return nil, fmt.Errorf("expected multipart/form-data: %w", err)
	}

	form := &Form{Values: make(map[string]string)}
	var total int64

	for parts := 0; ; parts++ {
		if parts >= maxParts {
			return nil, errors.New("too many parts in request")
		}
		part, err := mr.NextPart()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("read multipart body: %w", err)
		}

		if part.FileName() == "" {
			// Text field: small, read into memory.
			data, err := io.ReadAll(io.LimitReader(part, maxFieldSize+1))
			part.Close()
			if err != nil {
				return nil, fmt.Errorf("read field %q: %w", part.FormName(), err)
			}
			if int64(len(data)) > maxFieldSize {
				return nil, ErrTooLarge
			}
			form.Values[part.FormName()] = string(data)
			continue
		}

		if err := checkExtension(part.FileName(), opts.AllowedExtensions); err != nil {
			part.Close()
			return nil, err
		}

		remaining := opts.MaxTotalSize - total
		limit := min(opts.MaxFileSize, remaining)
		if limit <= 0 {
			part.Close()
			return nil, ErrTooLarge
		}

		path, n, err := ws.SaveReader(part.FileName(), part, limit)
		part.Close()
		if err != nil {
			return nil, err
		}
		total += n

		name := SafeBaseName(part.FileName())
		if name == "" {
			name = filepath.Base(path)
		}
		form.Files = append(form.Files, UploadedFile{
			FieldName: part.FormName(),
			Name:      name,
			Path:      path,
			Size:      n,
		})
	}

	return form, nil
}

func checkExtension(name string, allowed []string) error {
	if len(allowed) == 0 {
		return nil
	}
	ext := strings.ToLower(filepath.Ext(name))
	for _, a := range allowed {
		if ext == a {
			return nil
		}
	}
	return fmt.Errorf("file %q has an unsupported type (expected %s)", name, strings.Join(allowed, ", "))
}
