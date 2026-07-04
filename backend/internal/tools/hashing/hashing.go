// Package hashing implements the Hash Generator tool.
package hashing

import (
	"crypto/md5" //nolint:gosec // MD5 is offered as a checksum utility, not for password storage.
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/http"
	"sort"

	"github.com/go-chi/chi/v5"

	"github.com/naufalzaid17/all-in-one-tools/backend/internal/core"
)

// digestFunc computes a lowercase hex digest of the input.
type digestFunc func(input []byte) string

// algorithms maps the wire name of each supported algorithm to its
// implementation. Adding a new algorithm is a single entry here.
var algorithms = map[string]digestFunc{
	"sha256": func(in []byte) string {
		sum := sha256.Sum256(in)
		return hex.EncodeToString(sum[:])
	},
	"md5": func(in []byte) string {
		sum := md5.Sum(in) //nolint:gosec // see package note above
		return hex.EncodeToString(sum[:])
	},
}

// Tool is the Hash Generator tool module.
type Tool struct{}

// New returns the Hash Generator tool.
func New() *Tool { return &Tool{} }

func (t *Tool) ID() string              { return "hash" }
func (t *Tool) Name() string            { return "Hash Generator" }
func (t *Tool) Category() core.Category { return core.CategoryGenerator }
func (t *Tool) Description() string     { return "Generate SHA-256 and MD5 digests from text." }

// RegisterRoutes mounts the tool endpoints under /api/v1/tools/hash.
func (t *Tool) RegisterRoutes(r chi.Router) {
	r.Post("/generate", t.handleGenerate)
}

type generateRequest struct {
	Input string `json:"input"`
	// Algorithms selects which digests to compute. Empty means all
	// supported algorithms.
	Algorithms []string `json:"algorithms,omitempty"`
}

type generateResponse struct {
	// Hashes maps algorithm name to lowercase hex digest.
	Hashes map[string]string `json:"hashes"`
	// InputBytes is the byte length of the hashed input.
	InputBytes int `json:"inputBytes"`
}

func (t *Tool) handleGenerate(w http.ResponseWriter, r *http.Request) {
	var req generateRequest
	if err := core.DecodeJSON(r, &req); err != nil {
		core.RespondError(w, http.StatusBadRequest, core.CodeInvalidRequest, err.Error())
		return
	}

	selected := req.Algorithms
	if len(selected) == 0 {
		selected = supportedAlgorithms()
	}

	input := []byte(req.Input)
	hashes := make(map[string]string, len(selected))
	for _, name := range selected {
		digest, ok := algorithms[name]
		if !ok {
			core.RespondError(w, http.StatusBadRequest, core.CodeInvalidInput,
				fmt.Sprintf("unsupported algorithm %q (supported: %v)", name, supportedAlgorithms()))
			return
		}
		hashes[name] = digest(input)
	}

	core.RespondOK(w, generateResponse{Hashes: hashes, InputBytes: len(input)})
}

func supportedAlgorithms() []string {
	names := make([]string, 0, len(algorithms))
	for name := range algorithms {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
