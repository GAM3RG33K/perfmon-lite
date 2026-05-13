.PHONY: build run mock test lint clean help

APP_NAME := perfmon
GO_FLAGS := -ldflags="-s -w"
COVERAGE_FILE := coverage.out

help: ## Show this help
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | \
		awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build: ## Build the binary
	go build -o $(APP_NAME) $(GO_FLAGS) ./cmd/perfmon/

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
	rm -rf dist/ build/ release/

tidy: ## Tidy Go modules
	go mod tidy

dev: mock ## Alias for mock mode (default dev workflow)

all: tidy fmt build test ## Run all checks
