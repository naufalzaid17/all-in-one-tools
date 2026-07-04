package fileio

import (
	"errors"
	"net/http"

	"github.com/naufalzaid17/all-in-one-tools/backend/internal/core"
)

// ParseRequest is the standard entry point for upload handlers: it
// creates the request workspace and streams the multipart body into it.
// On failure it writes the error response, cleans up, and returns
// ok=false. On success the caller owns the workspace and must
// defer ws.Close().
func ParseRequest(w http.ResponseWriter, r *http.Request, opts ParseOptions) (*Workspace, *Form, bool) {
	ws, err := NewWorkspace()
	if err != nil {
		core.RespondError(w, http.StatusInternalServerError, core.CodeInternal, "could not create temp workspace")
		return nil, nil, false
	}
	form, err := ParseMultipart(r, ws, opts)
	if err != nil {
		ws.Close()
		status := http.StatusBadRequest
		if errors.Is(err, ErrTooLarge) {
			status = http.StatusRequestEntityTooLarge
		}
		core.RespondError(w, status, core.CodeInvalidInput, err.Error())
		return nil, nil, false
	}
	return ws, form, true
}
