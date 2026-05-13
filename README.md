# mcp-setu

[![Tests](https://github.com/manojshevate/mcp-setu/actions/workflows/test.yml/badge.svg)](https://github.com/manojshevate/mcp-setu/actions/workflows/test.yml)
[![Documentation](https://img.shields.io/badge/docs-GitHub%20Pages-blue)](https://manojshevate.github.io/mcp-setu/)

A Go CLI that bridges [Ollama](https://ollama.com) to [MCP](https://modelcontextprotocol.io) servers, giving local language models tool-calling abilities in a modern interactive chat TUI. Drop-in compatible with your existing Claude Desktop `mcp.json`.

```
┌──────────────┐                  ┌──────────────┐  JSON-RPC stdio  ┌──────────────┐
│    Ollama    │  /api/chat HTTP  │   mcp-setu   │ ◄──────────────► │  MCP Server  │
│  (gemma2,    │ ◄──────────────► │   (bridge)   │                  │ (filesystem, │
│  qwen2.5, …) │                  │              │                  │  sqlite, …)  │
└──────────────┘                  └──────────────┘                  └──────────────┘
```

## Install

```bash
go install github.com/manojshevate/mcp-setu/cmd/mcp-setu@latest
# or
brew tap manojshevate/mcp-setu
brew install mcp-setu
# or
git clone https://github.com/manojshevate/mcp-setu && cd mcp-setu && make install
```

## Quick start

```bash
ollama serve                                  # in another terminal
ollama pull gemma2:latest                     # any tool-capable model
mcp-setu chat                                 # opens the TUI
```

A minimal `mcp.json` (Claude Desktop-compatible):

```json
{
  "ollama": { "model": "gemma2:latest" },
  "mcpServers": {
    "memory": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-memory"]
    }
  }
}
```

Use a custom file with `--config /path/to/mcp.json`. To reuse your Claude Desktop config:

```bash
mcp-setu --config ~/Library/Application\ Support/Claude/claude_desktop_config.json chat
```

## Chat TUI

The interactive chat is a full-screen TUI with the input pinned at the bottom in a bordered box and a scrolling output area on top. The currently active model is shown as a badge in the bottom-right corner.

### Input Box

- The input field starts with the cursor at the beginning (like Claude Code)
- As you type, text appears after the cursor, which moves naturally through your input
- Use `↑` and `↓` to navigate through your command history
- The input box automatically wraps long text across multiple lines while staying pinned at the bottom

### Navigation & Input

| Keys           | What it does                              |
|----------------|-------------------------------------------|
| `↑` / `↓`      | Cycle through previous inputs (or navigate options in menus) |
| `/`            | Trigger slash-command autocomplete (appears above input) |
| `Tab`          | Accept the selected suggestion            |
| `Esc`          | Dismiss autocomplete or exit model picker  |
| `Ctrl+C` / `Ctrl+D` | Exit                                 |

### Slash Commands

| Command          | What it does                       |
|------------------|------------------------------------|
| `/tools`         | List available tools               |
| `/servers`       | Show connected MCP servers         |
| `/model`         | **Interactive model picker** (↑↓ to navigate, Enter to select) |
| `/model <space>` | **Autocomplete model names** (Tab to select) |
| `/model <name>`  | Switch to specific model directly  |
| `/stats`         | Performance metrics                |
| `/clear`         | Reset conversation history         |
| `/help`          | Show all commands                  |
| `/quit`, `/exit` | Exit                               |

**Model Selection Examples:**
- Type `/model` + Enter → Opens interactive picker where you navigate with arrow keys and confirm with Enter
- Type `/model ` (with space) → Shows autocomplete suggestions for available model names (use Tab to complete)
- Type `/model gemma2:latest` → Directly switch to the specified model

### Processing Status

While the LLM is processing your request, a live elapsed time counter is displayed above the input:

```
⟳ thinking… 2s
⟳ thinking… 45s
⟳ thinking… 2m 15s
```

The timer updates every second, making it easy to see how long the model has been thinking.

Pass `--verbose` to see tool calls and LLM iterations inside the TUI:

```bash
mcp-setu chat --verbose
```

## Other commands

```bash
mcp-setu tools       # list tools and exit
mcp-setu models      # list local Ollama models and tool support
mcp-setu validate    # validate config and connectivity
mcp-setu version
```

## Recommended models

`gemma3`, `qwen2.5` / `qwen3`, `llama3.2` / `llama3.3`, `mistral-nemo`, `command-r`, `phi4`, `deepseek-r1`.

Any locally installed Ollama model works — these have stronger tool-calling. Switch mid-chat with `/model <name>`.

## Documentation

- [Getting started](https://manojshevate.github.io/mcp-setu/getting-started)
- [Configuration](https://manojshevate.github.io/mcp-setu/configuration) — transports (stdio, HTTP streamable, HTTP/SSE), OAuth 2.1, env vars
- [Concepts](https://manojshevate.github.io/mcp-setu/concepts) — agentic loop, MCP, tool definitions
- [Examples](https://manojshevate.github.io/mcp-setu/examples)
- [Troubleshooting](https://manojshevate.github.io/mcp-setu/troubleshooting)
- [CLI reference](https://manojshevate.github.io/mcp-setu/cli/)

### Building docs locally

The documentation site (under `docs/`) uses [VitePress](https://vitepress.dev) and requires Node.js:

```bash
npm install     # first time only
make docs-dev   # start dev server at http://localhost:5173
```

## Development

```bash
make build      # bin/mcp-setu
make test       # run all tests
make vet
```

See [CONTRIBUTING.md](CONTRIBUTING.md) and [docs/development.md](docs/development.md) for the architecture, package layout, and release process.

## License

MIT — see [LICENSE](LICENSE).
