.PHONY: help build run install clean lint test vet coverage docs-dev docs-build docs-preview docs-generate

help:
	@echo "mcp-setu - MCP Bridge for Ollama"
	@echo ""
	@echo "Build targets:"
	@echo "  make build       - Build the mcp-setu binary"
	@echo "  make install     - Install mcp-setu to GOPATH/bin"
	@echo "  make run         - Run mcp-setu in interactive chat mode"
	@echo "  make run-tools   - List all available tools"
	@echo "  make run-models  - List available Ollama models"
	@echo "  make run-validate - Validate config and test connectivity"
	@echo ""
	@echo "Test targets:"
	@echo "  make test        - Run all unit tests"
	@echo "  make test-v      - Run tests with verbose output"
	@echo "  make coverage    - Run tests with coverage report"
	@echo "  make vet         - Run go vet checks"
	@echo "  make lint        - Run golangci-lint (requires installation)"
	@echo ""
	@echo "Documentation targets:"
	@echo "  make docs-dev    - Run docs dev server (http://localhost:5173)"
	@echo "  make docs-build  - Build docs for production"
	@echo "  make docs-preview - Preview built docs locally"
	@echo "  make docs-generate - Generate CLI reference docs"
	@echo ""
	@echo "Other targets:"
	@echo "  make clean       - Remove built binaries"
	@echo ""
	@echo "Quick start:"
	@echo "  1. make install"
	@echo "  2. ollama pull gemma4:e4b"
	@echo "  3. mcp-setu chat"
	@echo ""
	@echo "Documentation quick start:"
	@echo "  1. npm install (first time only)"
	@echo "  2. make docs-dev"
	@echo "  3. Open http://localhost:5173"

VERSION ?= v0.1.0-dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -ldflags="-X github.com/manojshevate/mcp-setu/internal/version.Version=$(VERSION) -X github.com/manojshevate/mcp-setu/internal/version.Commit=$(COMMIT) -X github.com/manojshevate/mcp-setu/internal/version.BuildDate=$(BUILD_DATE)"

build:
	@echo "Building mcp-setu ($(VERSION))..."
	@go build $(LDFLAGS) -o bin/mcp-setu ./cmd/mcp-setu
	@echo "Built: bin/mcp-setu"

install:
	@echo "Installing mcp-setu ($(VERSION))..."
	@go install $(LDFLAGS) ./cmd/mcp-setu/...
	@echo "Installed to $(GOPATH)/bin/mcp-setu"

release-build:
	@echo "Building release binaries..."
	@mkdir -p bin/releases
	@for os in linux darwin windows; do \
		for arch in amd64 arm64; do \
			if [ "$$os" = "windows" ]; then \
				GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o bin/releases/mcp-setu_$(VERSION)_$${os}_$${arch}.exe ./cmd/mcp-setu; \
			else \
				GOOS=$$os GOARCH=$$arch go build $(LDFLAGS) -o bin/releases/mcp-setu_$(VERSION)_$${os}_$${arch} ./cmd/mcp-setu; \
			fi; \
		done; \
	done
	@echo "Release binaries built in bin/releases/"

run: build
	@./bin/mcp-setu chat

run-verbose: build
	@./bin/mcp-setu chat --verbose

run-tools: build
	@./bin/mcp-setu tools

run-models: build
	@./bin/mcp-setu models

run-validate: build
	@./bin/mcp-setu validate

vet:
	@echo "Running go vet..."
	@go vet ./...
	@echo "go vet passed"

lint:
	@echo "Running golangci-lint..."
	@golangci-lint run ./...

test:
	@echo "Running tests..."
	@go test ./...

test-v:
	@echo "Running tests (verbose)..."
	@go test ./... -v

coverage:
	@echo "Running tests with coverage..."
	@go test ./internal/bridge -cover

clean:
	@echo "Cleaning..."
	@rm -f bin/mcp-setu
	@go clean
	@echo "Cleaned"

deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated"

docs-dev:
	@echo "Starting docs dev server..."
	@npm run docs:dev

docs-build:
	@echo "Building docs..."
	@npm run docs:prepare

docs-preview:
	@echo "Previewing built docs..."
	@npm run docs:preview

docs-generate:
	@echo "Generating CLI reference docs..."
	@npm run docs:generate-cli
