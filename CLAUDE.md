# mcp-setu — Claude Code Context

## Project Overview

`mcp-setu` is a Go CLI that bridges Ollama (local LLMs) to MCP (Model Context Protocol) servers, enabling tool-calling in interactive multi-turn chat sessions.

- **Binary name**: `mcp-setu`
- **Module**: `github.com/manojshevate/mcp-setu`
- **Entry point**: `cmd/mcp-setu/main.go`
- **Go version**: 1.26+

## Architecture

```
cmd/mcp-setu/       CLI entry point (Cobra commands: chat, tools, models, validate, version)
internal/bridge/    Agentic loop — orchestrates Ollama ↔ MCP tool calls
internal/config/    JSON config loading (mcp.json, Claude Desktop compatible)
internal/mcp/       MCP stdio client (JSON-RPC 2.0, spawns server subprocesses)
internal/ollama/    Ollama HTTP API client (/api/chat, /api/show, /api/tags)
internal/ui/        Terminal output formatting (lipgloss)
internal/version/   Version info injected at build time via ldflags
```

## Common Commands

```bash
make build        # Build to bin/mcp-setu
make install      # Install to $GOPATH/bin
make test         # Run all tests
make vet          # go vet checks
make coverage     # Test coverage for bridge package
```

## Key Conventions

- Config file: `mcp.json` (default), `--config` flag to override
- `mcpServers` block is byte-for-byte compatible with Claude Desktop config format
- The `ollama` block is mcp-setu-specific
- All dependencies injected explicitly — no global state
- Errors always wrapped with `fmt.Errorf(...%w...)`
- Tests use `httptest` for HTTP mocking; MCP client tests mock subprocess stdio

## Testing

```bash
go test ./...           # All packages
go test ./internal/bridge -cover
```

Coverage targets: bridge (82%+), config (92%+), ui (65%+).

## Release Process

1. Tag with `git tag -a vX.Y.Z`
2. Push tag — GitHub Actions builds 6 platform binaries and creates a release
3. Homebrew formula auto-updated for non-prerelease tags (requires `HOMEBREW_TAP_TOKEN` secret)

See `RELEASE.md` for full checklist.
