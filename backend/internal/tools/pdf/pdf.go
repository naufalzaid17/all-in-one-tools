// Package pdf implements the PDF Toolkit tool: merge, compress,
// encrypt (password-protect) and decrypt operations backed by pdfcpu.
//
// PDFs can be large, so every operation is file-to-file inside a
// per-request temp workspace — uploads are never buffered in RAM.
package pdf

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"github.com/pdfcpu/pdfcpu/pkg/api"
	"github.com/pdfcpu/pdfcpu/pkg/pdfcpu/model"

	"github.com/naufalzaid17/all-in-one-tools/backend/internal/core"
	"github.com/naufalzaid17/all-in-one-tools/backend/internal/fileio"
)

const contentTypePDF = "application/pdf"

var uploadOpts = fileio.ParseOptions{AllowedExtensions: []string{".pdf"}}

// Tool is the PDF Toolkit tool module.
type Tool struct{}

// New returns the PDF Toolkit tool. pdfcpu's config directory is
// disabled so the server never writes outside its temp workspaces.
func New() *Tool {
	api.DisableConfigDir()
	return &Tool{}
}

func (t *Tool) ID() string              { return "pdf" }
func (t *Tool) Name() string            { return "PDF Toolkit" }
func (t *Tool) Category() core.Category { return core.CategoryDocument }
func (t *Tool) Description() string {
	return "Merge, compress, password-protect and unlock PDF files."
}

// RegisterRoutes mounts the tool endpoints under /api/v1/tools/pdf.
func (t *Tool) RegisterRoutes(r chi.Router) {
	r.Post("/merge", t.handleMerge)
	r.Post("/compress", t.handleCompress)
	r.Post("/encrypt", t.handleEncrypt)
	r.Post("/decrypt", t.handleDecrypt)
}

func (t *Tool) handleMerge(w http.ResponseWriter, r *http.Request) {
	ws, form, ok := parseUpload(w, r)
	if !ok {
		return
	}
	defer ws.Close()

	if len(form.Files) < 2 {
		core.RespondError(w, http.StatusBadRequest, core.CodeInvalidInput, "merging requires at least two PDF files")
		return
	}

	inFiles := make([]string, 0, len(form.Files))
	for _, f := range form.Files {
		inFiles = append(inFiles, f.Path)
	}
	outPath := ws.Path("merged.pdf")
	if err := api.MergeCreateFile(inFiles, outPath, false, nil); err != nil {
		core.RespondError(w, http.StatusUnprocessableEntity, core.CodeInvalidInput, "merge failed: "+pdfErr(err))
		return
	}
	core.RespondFileDownload(w, outPath, "merged.pdf", contentTypePDF)
}

func (t *Tool) handleCompress(w http.ResponseWriter, r *http.Request) {
	ws, form, ok := parseUpload(w, r)
	if !ok {
		return
	}
	defer ws.Close()

	in, ok := singleFile(w, form)
	if !ok {
		return
	}
	outPath := ws.Path("compressed.pdf")
	if err := api.OptimizeFile(in.Path, outPath, nil); err != nil {
		core.RespondError(w, http.StatusUnprocessableEntity, core.CodeInvalidInput, "compression failed: "+pdfErr(err))
		return
	}
	core.RespondFileDownload(w, outPath, deriveName(in.Name, "-compressed", ".pdf"), contentTypePDF)
}

func (t *Tool) handleEncrypt(w http.ResponseWriter, r *http.Request) {
	ws, form, ok := parseUpload(w, r)
	if !ok {
		return
	}
	defer ws.Close()

	in, ok := singleFile(w, form)
	if !ok {
		return
	}
	password := form.Value("password")
	if len(password) < 4 {
		core.RespondError(w, http.StatusBadRequest, core.CodeInvalidInput, "password must be at least 4 characters")
		return
	}

	// AES-256; the password acts as both user and owner password.
	conf := model.NewAESConfiguration(password, password, 256)
	outPath := ws.Path("protected.pdf")
	if err := api.EncryptFile(in.Path, outPath, conf); err != nil {
		core.RespondError(w, http.StatusUnprocessableEntity, core.CodeInvalidInput, "encryption failed: "+pdfErr(err))
		return
	}
	core.RespondFileDownload(w, outPath, deriveName(in.Name, "-protected", ".pdf"), contentTypePDF)
}

func (t *Tool) handleDecrypt(w http.ResponseWriter, r *http.Request) {
	ws, form, ok := parseUpload(w, r)
	if !ok {
		return
	}
	defer ws.Close()

	in, ok := singleFile(w, form)
	if !ok {
		return
	}
	password := form.Value("password")
	if password == "" {
		core.RespondError(w, http.StatusBadRequest, core.CodeInvalidInput, "password is required")
		return
	}

	conf := model.NewAESConfiguration(password, password, 256)
	outPath := ws.Path("unlocked.pdf")
	if err := api.DecryptFile(in.Path, outPath, conf); err != nil {
		core.RespondError(w, http.StatusUnprocessableEntity, core.CodeInvalidInput, "decryption failed (wrong password?): "+pdfErr(err))
		return
	}
	core.RespondFileDownload(w, outPath, deriveName(in.Name, "-unlocked", ".pdf"), contentTypePDF)
}

func parseUpload(w http.ResponseWriter, r *http.Request) (*fileio.Workspace, *fileio.Form, bool) {
	return fileio.ParseRequest(w, r, uploadOpts)
}

func singleFile(w http.ResponseWriter, form *fileio.Form) (fileio.UploadedFile, bool) {
	if len(form.Files) != 1 {
		core.RespondError(w, http.StatusBadRequest, core.CodeInvalidInput, "exactly one PDF file is required")
		return fileio.UploadedFile{}, false
	}
	return form.Files[0], true
}

// deriveName turns "report.pdf" + "-compressed" into
// "report-compressed.pdf".
func deriveName(original, suffix, ext string) string {
	base := strings.TrimSuffix(original, ext)
	if base == "" {
		base = "output"
	}
	return base + suffix + ext
}

// pdfErr trims pdfcpu's sometimes multi-line errors down to the first
// line so API error messages stay single-line and user-presentable.
func pdfErr(err error) string {
	msg := err.Error()
	if i := strings.IndexByte(msg, '\n'); i > 0 {
		msg = msg[:i]
	}
	return msg
}
