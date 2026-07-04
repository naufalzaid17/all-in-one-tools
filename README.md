# All-in-One Tools

A local-first developer utilities app that ships as a **single Go binary**: the
Vue 3 frontend is compiled and embedded into the server with `go:embed`, so the
REST API and the SPA are served from the same port.

## Tools

| Tool | Category | Endpoints |
| --- | --- | --- |
| JSON Formatter | Formatter | `POST /api/v1/tools/json/format`, `POST /api/v1/tools/json/minify` |
| Hash Generator | Generators | `POST /api/v1/tools/hash/generate` (SHA-256, MD5) |
| QR Generator | Generators | `POST /api/v1/tools/qrcode/generate` (base64 PNG data URI) |
| PDF Toolkit | Document & AI Tools | `POST /api/v1/tools/pdf/{merge,compress,encrypt,decrypt}` |
| Excel Toolkit | Document & AI Tools | `POST /api/v1/tools/excel/{merge,to-csv}` |
| Markdown Toolkit | Document & AI Tools | `POST /api/v1/tools/markdown/{merge,to-html}` |
| File Encryption | File Security | `POST /api/v1/tools/security/{encrypt,decrypt}` |

Discovery: `GET /api/v1/tools` · Health: `GET /api/v1/health`

JSON endpoints share one envelope; file-processing endpoints accept
`multipart/form-data` (files under the `files` field plus text options) and
return the result directly as a download (`Content-Disposition: attachment`),
or the JSON error envelope on failure:

```json
{ "success": true,  "data": { } }
{ "success": false, "error": { "code": "invalid_input", "message": "..." } }
```

### File handling & security notes

- **OS-agnostic temp storage** — uploads are streamed into per-request
  workspaces created with `os.MkdirTemp` (respects `os.TempDir()` on Linux,
  Windows and macOS); no path is ever hardcoded and workspaces are always
  removed after the response.
- **Memory strategy** — large binaries (PDF, XLSX, encryption targets) are
  processed file-to-file; small text (Markdown, CSV output) is processed
  in-memory with `bytes.Buffer`.
- **File encryption format** — Argon2id key derivation + chunked AES-256-GCM
  (STREAM-style: per-chunk counter nonces, final-chunk flag bound as
  associated data), so files of any size encrypt in constant memory and any
  truncation/reordering fails authentication. See `internal/cryptostream`.

## Quick start

```sh
make build      # npm build → copy dist into backend/public/dist → go build
./bin/all-in-one-tools
# open http://localhost:8080  (override the port with PORT=9090)
```

## Development

Run the two halves separately for hot reload:

```sh
make dev-backend    # Go API on :8080
make dev-frontend   # Vite dev server on :5173, proxies /api → :8080
```

## Architecture

```
backend/
  cmd/server/           # entrypoint: wiring, graceful shutdown
  internal/core/        # Tool interface, registry, response envelope + downloads
  internal/api/         # chi router, SPA handler (vue-router fallback)
  internal/fileio/      # temp workspaces (os.MkdirTemp) + streamed multipart parsing
  internal/cryptostream/# chunked AES-256-GCM + Argon2id container format
  internal/tools/       # one package per tool: jsonutils, hashing, qrcode,
                        #   pdf, excel, markdown, security
  public/               # go:embed root; Makefile copies frontend/dist here
frontend/               # Vue 3 + Vite + TypeScript + TailwindCSS v4 + shadcn-vue
```

The backend follows a clean, interface-driven design: each tool implements
`core.Tool` and registers its own routes; the API layer depends only on the
`core` contracts. Adding a tool means adding one package and one
`registry.MustRegister(...)` line in `main.go`.
