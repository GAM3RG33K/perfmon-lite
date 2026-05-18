.PHONY: build cross-build run mock test test-short test-adb lint fmt clean tidy dev all release github-release hooks cut-release retag install update

APP_NAME := perfmon
COVERAGE_FILE := coverage.out
DIST_DIR := dist

# Read version from VERSION file; override with: make build VERSION=1.0.0
VERSION ?= $(shell cat VERSION 2>/dev/null || echo "dev")
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "dev")

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	go build -o $(APP_NAME) -ldflags="-s -w -X main.version=$(VERSION) -X main.commit=$(COMMIT)" ./cmd/perfmon/

cross-build: ## Build binaries for all platforms (linux/darwin/windows x amd64/arm64)
	@mkdir -p $(DIST_DIR)
	@echo "=== Cross-platform build ==="
	@for goos in linux darwin windows; do \
	  for goarch in amd64 arm64; do \
	    ext=""; \
	    if [ "$$goos" = "windows" ]; then ext=".exe"; fi; \
	    out="$(DIST_DIR)/$(APP_NAME)-$$goos-$$goarch$$ext"; \
	    echo "  building $$goos/$$goarch -> $$out"; \
	    GOOS=$$goos GOARCH=$$goarch go build -o $$out $(GO_FLAGS) ./cmd/perfmon/; \
	    ls -lh "$$out" | awk '{print "    " $$5}'; \
	  done; \
	done
	@echo "=== Done - $(DIST_DIR)/ contents ==="
	@ls -lh $(DIST_DIR)/

release: cross-build ## Build all binaries and create a GitHub Release (requires gh CLI)
	@echo ""
	@echo "============================================"
	@echo "  Creating GitHub Release v$(VERSION)"
	@echo "============================================"
	@if ! command -v gh >/dev/null 2>&1; then \
	  echo "ERROR: GitHub CLI (gh) is required."; \
	  echo "  Install: brew install gh   (macOS)"; \
	  echo "  Or:      sudo apt install gh (Linux)"; \
	  exit 1; \
	fi
	@if ! gh auth status >/dev/null 2>&1; then \
	  echo "ERROR: Not authenticated with GitHub CLI."; \
	  echo "  Run: gh auth login"; \
	  exit 1; \
	fi
	@if [ -z "$$(git tag -l 'v$(VERSION)')" ]; then \
	  echo "  Creating git tag v$(VERSION)..."; \
	  git tag -a "v$(VERSION)" -m "Release v$(VERSION)"; \
	  git push origin "v$(VERSION)"; \
	  echo "  Tag pushed."; \
	else \
	  echo "  Tag v$(VERSION) already exists. Use --edit to update the release."; \
	fi
	gh release create "v$(VERSION)" $(DIST_DIR)/* \
	  --title "v$(VERSION)" \
	  --generate-notes
	@echo ""
	@echo " Release v$(VERSION) created:"
	@gh release view "v$(VERSION)" --json url -q .url

github-release: release ## Alias for release

cut-release: ## Create and push a release tag from the current version
	@scripts/release.sh

retag: ## Delete existing tag and re-tag to trigger a new CI release build
	@scripts/release.sh --retag

install: ## Download latest release binary for your OS/arch and install it
	@scripts/install.sh

update: ## Check for newer version and upgrade if available
	@scripts/update.sh

uninstall: ## Remove perfmon from common install locations
	@scripts/uninstall.sh

run: ## Run without mock mode
	go run ./cmd/perfmon/

mock: ## Run with mock telemetry data
	go run ./cmd/perfmon/ --mock

test: ## Run all tests
	go test -v -race -coverprofile=$(COVERAGE_FILE) ./...

test-short: ## Run tests without race detector (faster)
	go test -v -short ./...

test-adb: ## Run ADB integration tests (requires real device/emulator)
	@echo "=== ADB Integration Tests ==="
	go test -tags=adb_test -v -race -count=1 ./internal/platform/android/

lint: ## Run golangci-lint
	golangci-lint run ./... 2>/dev/null || echo "golangci-lint not installed, running go vet instead" && go vet ./...

fmt: ## Format all Go code
	go fmt ./...

clean: ## Remove build artifacts
	rm -f $(APP_NAME)
	rm -f $(COVERAGE_FILE)
	rm -rf $(DIST_DIR)/ build/ release/

tidy: ## Tidy Go modules
	go mod tidy

hooks: ## Install git hooks
	@echo "Installing pre-commit hook..."
	git config core.hooksPath .githooks
	@echo "✅ Git hooks installed from .githooks/"
	@echo "   (pre-commit: validates docs match the codebase)"

dev: mock ## Alias for mock mode (default dev workflow)

all: tidy fmt hooks build test ## Run all checks (includes hook setup)
