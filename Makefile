.PHONY: help build run install clean lint test vet

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
	@echo "  make vet         - Run go vet checks"
	@echo "  make lint        - Run golangci-lint (requires installation)"
	@echo "  make test        - Run tests (when available)"
	@echo "  make clean       - Remove built binaries"
	@echo ""
	@echo "Quick start:"
	@echo "  1. make install"
	@echo "  2. ollama pull gemma4:e4b"
	@echo "  3. mcpgo chat"

build:
	@echo "Building mcpgo..."
	@go build -o bin/mcpgo ./cmd/mcpgo
	@echo "Built: bin/mcpgo"

install:
	@echo "Installing mcpgo..."
	@go install ./cmd/mcpgo/...
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
