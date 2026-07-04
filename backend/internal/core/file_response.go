package core

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"os"
	"strconv"
)

// RespondFileDownload streams a file from the given path as an
// attachment download. It is used by tools whose result is a binary
// file staged in a request workspace; the file is streamed, never
// loaded into memory.
func RespondFileDownload(w http.ResponseWriter, path, downloadName, contentType string) {
	f, err := os.Open(path)
	if err != nil {
		RespondError(w, http.StatusInternalServerError, CodeInternal, "result file missing")
		return
	}
	defer f.Close()

	info, err := f.Stat()
	if err != nil {
		RespondError(w, http.StatusInternalServerError, CodeInternal, "result file unreadable")
		return
	}

	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
	setAttachment(w, downloadName)
	w.WriteHeader(http.StatusOK)
	// The response is already committed; a copy error here means the
	// client disconnected mid-download.
	_, _ = io.Copy(w, f)
}

// RespondBytesDownload sends an in-memory result as an attachment
// download. Only use it for small outputs (markdown, CSV, HTML).
func RespondBytesDownload(w http.ResponseWriter, data []byte, downloadName, contentType string) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Length", strconv.Itoa(len(data)))
	setAttachment(w, downloadName)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// setAttachment writes a Content-Disposition header that survives
// non-ASCII file names (RFC 5987 filename* with an ASCII fallback,
// courtesy of mime.FormatMediaType).
func setAttachment(w http.ResponseWriter, name string) {
	if name == "" {
		name = "download"
	}
	disposition := mime.FormatMediaType("attachment", map[string]string{"filename": name})
	if disposition == "" {
		disposition = fmt.Sprintf("attachment; filename=%q", "download")
	}
	w.Header().Set("Content-Disposition", disposition)
	// Let browser JS read the file name from the header.
	w.Header().Set("Access-Control-Expose-Headers", "Content-Disposition")
}
