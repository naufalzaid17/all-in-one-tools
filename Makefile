# All-in-One Tools — build automation.
#
# `make build` produces a single self-contained binary in bin/:
#   1. builds the Vue frontend (frontend/dist)
#   2. copies the build output into backend/public/dist (the go:embed root)
#   3. compiles the Go server with the frontend embedded
#
# Recipes must run under both sh (unix) and cmd.exe (Windows), so they
# are limited to "cd <dir> && <go|npm> ..." — anything more complex
# (recursive copy, cleanup) lives in backend/scripts/xtask, a small
# OS-agnostic Go helper.

FRONTEND_DIR := frontend
BACKEND_DIR  := backend
BIN_DIR      := bin

ifeq ($(OS),Windows_NT)
BINEXT := .exe
else
BINEXT :=
endif
BINARY := $(BIN_DIR)/all-in-one-tools$(BINEXT)

.PHONY: build frontend embed backend run dev-backend dev-frontend test clean

## build: full production build (frontend + embed + Go binary)
build: frontend embed backend
	@echo Done. Run the app with: $(BINARY)

## frontend: install dependencies and build the Vue app
frontend:
	cd $(FRONTEND_DIR) && npm install && npm run build

## embed: copy the frontend build output into the go:embed directory
embed:
	cd $(BACKEND_DIR) && go run ./scripts/xtask embed-assets

## backend: compile the Go server (embeds whatever is in backend/public/dist)
backend:
	cd $(BACKEND_DIR) && go run ./scripts/xtask ensure-bin && go build -trimpath -ldflags="-s -w" -o ../$(BINARY) ./cmd/server

## run: build everything and start the server on :8080
run: build
ifeq ($(OS),Windows_NT)
	$(subst /,\,$(BINARY))
else
	./$(BINARY)
endif

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
	cd $(BACKEND_DIR) && go run ./scripts/xtask clean
