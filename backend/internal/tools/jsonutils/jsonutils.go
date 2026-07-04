// Package jsonutils implements the JSON Formatter tool: pretty-printing
// and minification of arbitrary JSON documents.
package jsonutils

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/naufalzaid17/all-in-one-tools/backend/internal/core"
)

// maxIndentWidth keeps indentation requests sane.
const maxIndentWidth = 8

// Tool is the JSON Formatter tool module.
type Tool struct{}

// New returns the JSON Formatter tool.
func New() *Tool { return &Tool{} }

func (t *Tool) ID() string              { return "json" }
func (t *Tool) Name() string            { return "JSON Formatter" }
func (t *Tool) Category() core.Category { return core.CategoryFormatter }
func (t *Tool) Description() string     { return "Format, validate and minify JSON documents." }

// RegisterRoutes mounts the tool endpoints under /api/v1/tools/json.
func (t *Tool) RegisterRoutes(r chi.Router) {
	r.Post("/format", t.handleFormat)
	r.Post("/minify", t.handleMinify)
}

type formatRequest struct {
	Input string `json:"input"`
	// Indent is the number of spaces per level (default 2).
	// Ignored when UseTabs is true.
	Indent  int  `json:"indent,omitempty"`
	UseTabs bool `json:"useTabs,omitempty"`
}

type minifyRequest struct {
	Input string `json:"input"`
}

type transformResponse struct {
	Output string `json:"output"`
	// Bytes is the size of the output in bytes, handy for the UI to show
	// size deltas after formatting/minifying.
	Bytes int `json:"bytes"`
}

func (t *Tool) handleFormat(w http.ResponseWriter, r *http.Request) {
	var req formatRequest
	if err := core.DecodeJSON(r, &req); err != nil {
		core.RespondError(w, http.StatusBadRequest, core.CodeInvalidRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.Input) == "" {
		core.RespondError(w, http.StatusBadRequest, core.CodeInvalidInput, "input must not be empty")
		return
	}
	if req.Indent < 0 || req.Indent > maxIndentWidth {
		core.RespondError(w, http.StatusBadRequest, core.CodeInvalidInput, "indent must be between 0 and 8")
		return
	}

	indent := "  "
	switch {
	case req.UseTabs:
		indent = "\t"
	case req.Indent > 0:
		indent = strings.Repeat(" ", req.Indent)
	}

	// json.Indent operates on the raw token stream, so number precision
	// and key order are preserved exactly (unlike Unmarshal/Marshal).
	var buf bytes.Buffer
	if err := json.Indent(&buf, []byte(req.Input), "", indent); err != nil {
		core.RespondError(w, http.StatusUnprocessableEntity, core.CodeInvalidInput, "invalid JSON: "+err.Error())
		return
	}
	core.RespondOK(w, transformResponse{Output: buf.String(), Bytes: buf.Len()})
}

func (t *Tool) handleMinify(w http.ResponseWriter, r *http.Request) {
	var req minifyRequest
	if err := core.DecodeJSON(r, &req); err != nil {
		core.RespondError(w, http.StatusBadRequest, core.CodeInvalidRequest, err.Error())
		return
	}
	if strings.TrimSpace(req.Input) == "" {
		core.RespondError(w, http.StatusBadRequest, core.CodeInvalidInput, "input must not be empty")
		return
	}

	var buf bytes.Buffer
	if err := json.Compact(&buf, []byte(req.Input)); err != nil {
		core.RespondError(w, http.StatusUnprocessableEntity, core.CodeInvalidInput, "invalid JSON: "+err.Error())
		return
	}
	core.RespondOK(w, transformResponse{Output: buf.String(), Bytes: buf.Len()})
}
