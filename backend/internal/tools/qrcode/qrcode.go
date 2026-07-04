// Package qrcode implements the QR Generator tool: it encodes text or a
// URL into a QR code and returns it as a base64 data URI.
package qrcode

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	qr "github.com/skip2/go-qrcode"

	"github.com/naufalzaid17/all-in-one-tools/backend/internal/core"
)

const (
	defaultSize = 256
	minSize     = 64
	maxSize     = 1024
	// maxContentLen stays well under the QR spec's byte-mode capacity
	// (~2953 bytes at the lowest error-correction level).
	maxContentLen = 2048
)

// recoveryLevels maps wire names to error-correction levels.
var recoveryLevels = map[string]qr.RecoveryLevel{
	"low":     qr.Low,
	"medium":  qr.Medium,
	"high":    qr.High,
	"highest": qr.Highest,
}

// Tool is the QR Generator tool module.
type Tool struct{}

// New returns the QR Generator tool.
func New() *Tool { return &Tool{} }

func (t *Tool) ID() string              { return "qrcode" }
func (t *Tool) Name() string            { return "QR Generator" }
func (t *Tool) Category() core.Category { return core.CategoryGenerator }
func (t *Tool) Description() string     { return "Encode a URL or text as a QR code image." }

// RegisterRoutes mounts the tool endpoints under /api/v1/tools/qrcode.
func (t *Tool) RegisterRoutes(r chi.Router) {
	r.Post("/generate", t.handleGenerate)
}

type generateRequest struct {
	Content string `json:"content"`
	// Size is the output image edge length in pixels (default 256).
	Size int `json:"size,omitempty"`
	// Recovery selects the error-correction level:
	// low | medium | high | highest (default medium).
	Recovery string `json:"recovery,omitempty"`
}

type generateResponse struct {
	// Image is a ready-to-use data URI (data:image/png;base64,...).
	Image string `json:"image"`
	Size  int    `json:"size"`
}

func (t *Tool) handleGenerate(w http.ResponseWriter, r *http.Request) {
	var req generateRequest
	if err := core.DecodeJSON(r, &req); err != nil {
		core.RespondError(w, http.StatusBadRequest, core.CodeInvalidRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.Content) == "" {
		core.RespondError(w, http.StatusBadRequest, core.CodeInvalidInput, "content must not be empty")
		return
	}
	if len(req.Content) > maxContentLen {
		core.RespondError(w, http.StatusBadRequest, core.CodeInvalidInput,
			fmt.Sprintf("content exceeds the maximum of %d bytes", maxContentLen))
		return
	}

	size := req.Size
	if size == 0 {
		size = defaultSize
	}
	if size < minSize || size > maxSize {
		core.RespondError(w, http.StatusBadRequest, core.CodeInvalidInput,
			fmt.Sprintf("size must be between %d and %d pixels", minSize, maxSize))
		return
	}

	recovery := qr.Medium
	if req.Recovery != "" {
		level, ok := recoveryLevels[req.Recovery]
		if !ok {
			core.RespondError(w, http.StatusBadRequest, core.CodeInvalidInput,
				"recovery must be one of: low, medium, high, highest")
			return
		}
		recovery = level
	}

	png, err := qr.Encode(req.Content, recovery, size)
	if err != nil {
		core.RespondError(w, http.StatusUnprocessableEntity, core.CodeInvalidInput,
			"could not encode QR code: "+err.Error())
		return
	}

	core.RespondOK(w, generateResponse{
		Image: "data:image/png;base64," + base64.StdEncoding.EncodeToString(png),
		Size:  size,
	})
}
