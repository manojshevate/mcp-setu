.PHONY: help build run install clean lint test vet coverage

help:
	@echo "mcpgo - MCP Bridge for Ollama"
	@echo ""
	@echo "Available targets:"
	@echo "  make build       - Build the mcpgo binary"
	@echo "  make install     - Install mcpgo to GOPATH/bin"
	@echo "  make run         - Run mcpgo in interactive chat mode"
	@echo "  make run-tools   - List all available tools"
	@echo "  make run-models  - List available Ollama models"
	@echo "  make run-validate - Validate config and test connectivity"
	@echo "  make test        - Run all unit tests"
	@echo "  make test-v      - Run tests with verbose output"
	@echo "  make coverage    - Run tests with coverage report"
	@echo "  make vet         - Run go vet checks"
	@echo "  make lint        - Run golangci-lint (requires installation)"
	@echo "  make clean       - Remove built binaries"
	@echo ""
	@echo "Quick start:"
	@echo "  1. make install"
	@echo "  2. ollama pull gemma4:e4b"
	@echo "  3. mcpgo chat"

VERSION ?= v0.1.0-dev
COMMIT ?= $(shell git rev-parse --short HEAD 2>/dev/null || echo "unknown")
BUILD_DATE ?= $(shell date -u +'%Y-%m-%dT%H:%M:%SZ')
LDFLAGS := -ldflags="-X github.com/manojshevate/mcpgo/internal/version.Version=$(VERSION) -X github.com/manojshevate/mcpgo/internal/version.Commit=$(COMMIT) -X github.com/manojshevate/mcpgo/internal/version.BuildDate=$(BUILD_DATE)"

build:
	@echo "Building mcpgo ($(VERSION))..."
	@go build $(LDFLAGS) -o bin/mcpgo ./cmd/mcpgo
	@echo "Built: bin/mcpgo"

install:
	@echo "Installing mcpgo ($(VERSION))..."
	@go install $(LDFLAGS) ./cmd/mcpgo/...
	@echo "Installed to $(GOPATH)/bin/mcpgo"

run: build
	@./bin/mcpgo chat

run-verbose: build
	@./bin/mcpgo chat --verbose

run-tools: build
	@./bin/mcpgo tools

run-models: build
	@./bin/mcpgo models

run-validate: build
	@./bin/mcpgo validate

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
	@rm -f bin/mcpgo
	@go clean
	@echo "Cleaned"

deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated"
