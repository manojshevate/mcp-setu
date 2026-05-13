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
internal/ui/        Terminal output formatting (lipgloss, static output)
internal/tui/       Interactive terminal UI (Bubble Tea, input/autocomplete/picker)
internal/version/   Version info injected at build time via ldflags
```

### TUI (Terminal UI) Features

The `internal/tui/` package provides an interactive chat session with modern UI design and advanced input handling:

#### Visual Design

- **Bordered Input Box**: Messages are typed in a styled box with rounded corners (`╭`, `╮`, `╰`, `╯`) that spans the terminal width with 1-char padding on each side
- **Dynamic Height**: Input box automatically wraps long text across multiple lines while maintaining the fixed input position at the bottom
- **Model Indicator Badge**: Currently active model displayed in the bottom-right corner (e.g., `◆ gemma2:latest`), with automatic truncation for long model names
- **Cursor at Start**: Cursor appears at the beginning of input (like Claude Code) and moves naturally as user types
- **Upward Autocomplete**: Slash-command suggestions rendered *above* the input field for better visibility

#### Input Model

Manages text input, command history, and three distinct modes:
- `modeNormal`: Standard text entry with history navigation (↑↓), cursor tracking
- `modeAutocomplete`: Slash-command autocomplete dropdown visible above input
- `modeModelSelect`: Interactive model picker for `/model` command selection

Key components:
- `RenderLine()`: Returns styled input line with cursor at current position (before placeholder when empty, between text when typing)
- `RenderAutocomplete()`: Returns autocomplete dropdown or picker overlay (empty string when nothing shown)
- `updateAC()`: Handles two autocomplete modes:
  - Command autocomplete: `/tools`, `/clear`, `/model`, etc.
  - Model name autocomplete: `/model <space>` triggers matching against available models

#### Slash Commands

`/tools`, `/clear`, `/model`, `/stats`, `/servers`, `/help`, `/quit`, `/exit`

#### Model Selection

- `/model <Enter>`: Open interactive picker (navigate with ↑↓, select with Enter, cancel with Esc)
- `/model <space>`: Autocomplete model names (Tab to select; matches against eagerly-fetched model list)
- `/model <name>`: Direct switch to specific model
- Models list fetched from Ollama on startup for instant autocomplete availability

#### Key Behaviors

- Esc in picker mode: Cancels and returns to chat input with cleared text
- Esc in autocomplete: Dismisses dropdown without clearing typed text
- Typing while in picker: Exits picker and starts fresh text entry
- Up/Down arrows: Navigate history in normal mode, navigate options in picker/autocomplete

#### Processing Status

- While LLM processes a request, elapsed time displays above input: `⟳ thinking… 2s`, `⟳ thinking… 1m 30s`
- Timer updates every second with clean integer formatting (e.g., "2m 3s" for durations >= 60s)
- Calculated lazily in View() using `time.Since()` with existing 100ms event loop; no separate goroutines
- `formatElapsed()` helper provides consistent formatting across the TUI

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
