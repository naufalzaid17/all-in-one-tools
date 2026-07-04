# All-in-One Tools — build automation.
#
# `make build` produces a single self-contained binary at bin/all-in-one-tools:
#   1. builds the Vue frontend (frontend/dist)
#   2. copies the build output into backend/public/dist (the go:embed root)
#   3. compiles the Go server with the frontend embedded

FRONTEND_DIR := frontend
BACKEND_DIR  := backend
EMBED_DIR    := $(BACKEND_DIR)/public/dist
BIN_DIR      := bin
BINARY       := $(BIN_DIR)/all-in-one-tools

.PHONY: build frontend embed backend run dev-backend dev-frontend test clean

## build: full production build (frontend + embed + Go binary)
build: frontend embed backend
	@echo "✓ Built $(BINARY) — run it with: ./$(BINARY)"

## frontend: install dependencies and build the Vue app
frontend:
	cd $(FRONTEND_DIR) && npm install && npm run build

## embed: copy the frontend build output into the go:embed directory
embed:
	@mkdir -p $(EMBED_DIR)
	@find $(EMBED_DIR) -mindepth 1 ! -name '.gitkeep' -delete
	cp -R $(FRONTEND_DIR)/dist/. $(EMBED_DIR)/

## backend: compile the Go server (embeds whatever is in backend/public/dist)
backend:
	@mkdir -p $(BIN_DIR)
	cd $(BACKEND_DIR) && go build -trimpath -ldflags="-s -w" -o ../$(BINARY) ./cmd/server

## run: build everything and start the server on :8080
run: build
	./$(BINARY)

## dev-backend: run the Go API with live code (no embedded frontend needed)
dev-backend:
	cd $(BACKEND_DIR) && go run ./cmd/server

## dev-frontend: run the Vite dev server (proxies /api to :8080)
dev-frontend:
	cd $(FRONTEND_DIR) && npm run dev

## test: run Go tests
test:
	cd $(BACKEND_DIR) && go test ./...

## clean: remove build artifacts
clean:
	rm -rf $(BIN_DIR) $(FRONTEND_DIR)/dist
	@find $(EMBED_DIR) -mindepth 1 ! -name '.gitkeep' -delete 2>/dev/null || true
