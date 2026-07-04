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

Discovery: `GET /api/v1/tools` · Health: `GET /api/v1/health`

All responses share one envelope:

```json
{ "success": true,  "data": { } }
{ "success": false, "error": { "code": "invalid_input", "message": "..." } }
```

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
  cmd/server/        # entrypoint: wiring, graceful shutdown
  internal/core/     # Tool interface, registry, response envelope
  internal/api/      # chi router, SPA handler (vue-router fallback)
  internal/tools/    # one package per tool (jsonutils, security, qrcode)
  public/            # go:embed root; Makefile copies frontend/dist here
frontend/            # Vue 3 + Vite + TypeScript + TailwindCSS v4 + shadcn-vue
```

The backend follows a clean, interface-driven design: each tool implements
`core.Tool` and registers its own routes; the API layer depends only on the
`core` contracts. Adding a tool means adding one package and one
`registry.MustRegister(...)` line in `main.go`.
