// Package security implements the File Encryption tool: it encrypts and
// decrypts arbitrary files with a user password using streaming
// AES-256-GCM with an Argon2id-derived key (see internal/cryptostream
// for the format).
//
// Files of any size are supported: input and output are staged in the
// request's temp workspace and the cipher streams chunk by chunk, so
// memory use is constant regardless of file size.
package security

import (
	"errors"
	"net/http"
	"os"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/naufalzaid17/all-in-one-tools/backend/internal/core"
	"github.com/naufalzaid17/all-in-one-tools/backend/internal/cryptostream"
	"github.com/naufalzaid17/all-in-one-tools/backend/internal/fileio"
)

const (
	encExtension    = ".enc"
	minPasswordLen  = 6
	contentTypeBlob = "application/octet-stream"
)

// Tool is the File Encryption tool module.
type Tool struct{}

// New returns the File Encryption tool.
func New() *Tool { return &Tool{} }

func (t *Tool) ID() string              { return "security" }
func (t *Tool) Name() string            { return "File Encryption" }
func (t *Tool) Category() core.Category { return core.CategorySecurity }
func (t *Tool) Description() string {
	return "Encrypt and decrypt any file with a password (AES-256-GCM, Argon2id)."
}

// RegisterRoutes mounts the tool endpoints under /api/v1/tools/security.
func (t *Tool) RegisterRoutes(r chi.Router) {
	r.Post("/encrypt", t.handleEncrypt)
	r.Post("/decrypt", t.handleDecrypt)
}

func (t *Tool) handleEncrypt(w http.ResponseWriter, r *http.Request) {
	ws, in, password, ok := parseCryptoRequest(w, r, minPasswordLen)
	if !ok {
		return
	}
	defer ws.Close()

	src, err := openUpload(w, in)
	if err != nil {
		return
	}
	defer src.Close()

	dst, outPath, err := ws.CreateFile("encrypted.bin")
	if err != nil {
		core.RespondError(w, http.StatusInternalServerError, core.CodeInternal, "could not stage output file")
		return
	}
	encErr := cryptostream.Encrypt(dst, src, password)
	dst.Close()
	if encErr != nil {
		core.RespondError(w, http.StatusInternalServerError, core.CodeInternal, "encryption failed")
		return
	}

	core.RespondFileDownload(w, outPath, in.Name+encExtension, contentTypeBlob)
}

func (t *Tool) handleDecrypt(w http.ResponseWriter, r *http.Request) {
	// Decryption accepts any password length; validation happens via
	// authentication of the ciphertext itself.
	ws, in, password, ok := parseCryptoRequest(w, r, 1)
	if !ok {
		return
	}
	defer ws.Close()

	src, err := openUpload(w, in)
	if err != nil {
		return
	}
	defer src.Close()

	dst, outPath, err := ws.CreateFile("decrypted.bin")
	if err != nil {
		core.RespondError(w, http.StatusInternalServerError, core.CodeInternal, "could not stage output file")
		return
	}
	decErr := cryptostream.Decrypt(dst, src, password)
	dst.Close()

	switch {
	case errors.Is(decErr, cryptostream.ErrNotEncrypted):
		core.RespondError(w, http.StatusUnprocessableEntity, core.CodeInvalidInput,
			"this file was not encrypted by this tool")
		return
	case errors.Is(decErr, cryptostream.ErrWrongPasswordOrCorrupt):
		core.RespondError(w, http.StatusUnprocessableEntity, core.CodeInvalidInput,
			"wrong password or corrupted file")
		return
	case decErr != nil:
		core.RespondError(w, http.StatusInternalServerError, core.CodeInternal, "decryption failed")
		return
	}

	core.RespondFileDownload(w, outPath, restoredName(in.Name), contentTypeBlob)
}

// parseCryptoRequest handles the shared multipart plumbing: exactly one
// file of any type plus a password field.
func parseCryptoRequest(w http.ResponseWriter, r *http.Request, minPassword int) (*fileio.Workspace, fileio.UploadedFile, string, bool) {
	ws, form, ok := fileio.ParseRequest(w, r, fileio.ParseOptions{})
	if !ok {
		return nil, fileio.UploadedFile{}, "", false
	}
	if len(form.Files) != 1 {
		ws.Close()
		core.RespondError(w, http.StatusBadRequest, core.CodeInvalidInput, "exactly one file is required")
		return nil, fileio.UploadedFile{}, "", false
	}
	password := form.Value("password")
	if len(password) < minPassword {
		ws.Close()
		if minPassword > 1 {
			core.RespondError(w, http.StatusBadRequest, core.CodeInvalidInput,
				"password must be at least 6 characters")
		} else {
			core.RespondError(w, http.StatusBadRequest, core.CodeInvalidInput, "password is required")
		}
		return nil, fileio.UploadedFile{}, "", false
	}
	return ws, form.Files[0], password, true
}

func openUpload(w http.ResponseWriter, in fileio.UploadedFile) (*os.File, error) {
	f, err := os.Open(in.Path)
	if err != nil {
		core.RespondError(w, http.StatusInternalServerError, core.CodeInternal, "could not read uploaded file")
		return nil, err
	}
	return f, nil
}

// restoredName recovers the original file name from "<name>.enc"; for
// anything else it prefixes "decrypted-" so the download never
// silently overwrites the source name space.
func restoredName(name string) string {
	if strings.HasSuffix(strings.ToLower(name), encExtension) && len(name) > len(encExtension) {
		return name[:len(name)-len(encExtension)]
	}
	return "decrypted-" + name
}
