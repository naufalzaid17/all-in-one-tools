package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

// maxRequestBody caps request payloads (1 MiB) so a single request cannot
// exhaust memory. Large-input tools can override this locally if needed.
const maxRequestBody = 1 << 20

// APIResponse is the standard envelope returned by every endpoint.
//
//	success: {"success": true,  "data": {...}}
//	failure: {"success": false, "error": {"code": "...", "message": "..."}}
type APIResponse struct {
	Success bool      `json:"success"`
	Data    any       `json:"data,omitempty"`
	Error   *APIError `json:"error,omitempty"`
}

// APIError carries a stable, machine-readable code alongside a
// human-readable message.
type APIError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// Well-known error codes shared across tools.
const (
	CodeInvalidRequest = "invalid_request"
	CodeInvalidInput   = "invalid_input"
	CodeInternal       = "internal_error"
	CodeNotFound       = "not_found"
)

// RespondOK writes a success envelope with the given payload.
func RespondOK(w http.ResponseWriter, data any) {
	writeJSON(w, http.StatusOK, APIResponse{Success: true, Data: data})
}

// RespondError writes a failure envelope with the given HTTP status,
// error code and message.
func RespondError(w http.ResponseWriter, status int, code, message string) {
	writeJSON(w, status, APIResponse{
		Success: false,
		Error:   &APIError{Code: code, Message: message},
	})
}

// DecodeJSON strictly decodes the request body into dst. It rejects
// unknown fields, oversized bodies and trailing garbage, returning a
// user-presentable error message on failure.
func DecodeJSON(r *http.Request, dst any) error {
	dec := json.NewDecoder(io.LimitReader(r.Body, maxRequestBody))
	dec.DisallowUnknownFields()
	if err := dec.Decode(dst); err != nil {
		return fmt.Errorf("malformed request body: %w", err)
	}
	if err := dec.Decode(&struct{}{}); !errors.Is(err, io.EOF) {
		return errors.New("request body must contain a single JSON object")
	}
	return nil
}

func writeJSON(w http.ResponseWriter, status int, body APIResponse) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	// Encoding a response envelope cannot realistically fail; if the
	// client disconnected mid-write there is nothing left to do.
	_ = json.NewEncoder(w).Encode(body)
}
