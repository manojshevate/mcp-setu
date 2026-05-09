# mcpgo - MCP Bridge for Ollama

**mcpgo** is a Go-native CLI application that bridges Ollama to MCP (Model Context Protocol) servers for interactive, multi-turn chat with local language models. It's Claude Desktop config-compatible, meaning you can reuse your existing MCP configuration without any changes. Fully open-source, no cloud dependency, runs entirely on your machine.

## What is mcpgo?

mcpgo connects your local Ollama instance to MCP servers, enabling models like Gemma, Qwen, and Llama to call tools (like filesystem access, database queries, etc.) during interactive conversations. The CLI is inspired by modern tools like Ghostty and Claude Code, with clean terminal UI and helpful error messages. Your existing Claude Desktop `mcp.json` works immediately—just point mcpgo at it.

## Prerequisites

- **Go 1.22+** — [Download](https://golang.org/dl/)
- **Ollama** — [https://ollama.com](https://ollama.com) (run `ollama serve` before starting mcpgo)
- **Node.js 18+** — For npx-based MCP servers
- **A tool-calling capable model** — See [Supported Models](#supported-models) for examples

## Quick Start

```bash
# Clone and build
git clone https://github.com/manojshevate/mcpgo && cd mcpgo
make install

# Pull a tool-supporting model
ollama pull gemma4:e4b

# Run interactive chat
mcpgo chat
```

That's it! You'll see a banner with connected servers and tools, then a prompt to type. Type naturally, and the model will call tools as needed.

## Claude Desktop Config Compatibility

mcpgo uses the **exact same** `mcpServers` configuration format as Claude Desktop. Your existing Claude Desktop MCP setup works immediately without modification.

**Where to find your Claude Desktop config:**

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

**To use your Claude Desktop MCP setup with mcpgo:**

```bash
# macOS
mcpgo --config ~/Library/Application\ Support/Claude/claude_desktop_config.json chat

# Windows (PowerShell)
mcpgo --config $env:APPDATA\Claude\claude_desktop_config.json chat

# Linux
mcpgo --config ~/.config/Claude/claude_desktop_config.json chat
```

The only difference is the `ollama` block, which is mcpgo-specific and not present in Claude Desktop configs. mcpgo will use sensible defaults if the block is missing.

## Configuration

mcpgo reads its configuration from `mcp.json` by default. Use `--config <path>` to specify a different file.

**Full annotated example (`mcp.json`):**

```json
{
  "ollama": {
    "baseUrl": "http://localhost:11434",
    "model": "gemma4:e4b",
    "systemPrompt": "You are a helpful assistant.",
    "temperature": 0.7,
    "contextLength": 4096
  },
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "."],
      "env": {}
    },
    "sqlite": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-sqlite", "./dev.db"],
      "env": {}
    }
  }
}
```

**Field descriptions:**

- **ollama.baseUrl** — Ollama HTTP endpoint (default: `http://localhost:11434`)
- **ollama.model** — Model to use (must support tool calling; see [Supported Models](#supported-models))
- **ollama.systemPrompt** — System prompt sent with every request (default: generic assistant prompt)
- **ollama.temperature** — Sampling temperature 0–1 (default: `0.7`)
- **ollama.contextLength** — Max context window size in tokens (default: `4096`)
- **mcpServers** — Map of MCP server configs (byte-for-byte compatible with Claude Desktop)
  - **command** — Executable to run
  - **args** — Arguments to pass
  - **env** — Optional environment variables

## Supported Models

| Model           | Example tag      | Notes                                    |
|-----------------|------------------|------------------------------------------|
| gemma4          | gemma4:e4b       | ⭐ **Recommended** — Best on-device tool calling |
| gemma3          | gemma3:2b        | Efficient, strong tool use               |
| qwen2.5 / qwen3 | qwen2.5:7b       | Excellent tool calling, great for coding |
| llama3.2        | llama3.2:3b      | Fast, reliable, good for most tasks      |
| llama3.3        | llama3.3:70b     | Meta's largest, strongest reasoning      |
| mistral-nemo    | mistral-nemo:12b | Balanced speed and accuracy              |
| command-r       | command-r:35b    | Strong multi-tool chaining               |
| phi4            | phi4:14b         | Compact and capable                      |
| deepseek-r1     | deepseek-r1:7b   | Strong reasoning + tool use              |

> **mcpgo checks tool support at startup** and exits with a clear error if your model doesn't support tool calling. Run `mcpgo models` to see what's available locally and which models are compatible.

## CLI Reference

### Root flags

```
--config string    Path to config file (default: mcp.json)
--verbose           Print tool calls and results to stderr
```

### Commands

| Command            | Description                                           |
|--------------------|-------------------------------------------------------|
| `mcpgo chat`       | Start interactive chat session (default)              |
| `mcpgo tools`      | List all tools from configured MCP servers and exit   |
| `mcpgo models`     | List local Ollama models and tool support status      |
| `mcpgo validate`   | Validate config and test MCP server connectivity      |

### Chat flags

```
--model string     Override model from config (e.g., --model qwen2.5:7b)
--system string    Override system prompt from config
```

**Example:**

```bash
mcpgo chat --verbose --model gemma4:e4b --system "You are a Python expert."
```

## REPL Commands

Once in the interactive chat loop, these commands are available:

| Command      | Description                    |
|--------------|--------------------------------|
| `/tools`     | List all available tools       |
| `/clear`     | Clear conversation history     |
| `/model`     | Show current model             |
| `/servers`   | Show connected MCP servers     |
| `/help`      | Show this help                 |
| `exit`/`quit`| Quit mcpgo                     |

## Example Session

Here's a realistic multi-turn session using `gemma4:e4b` with the filesystem MCP server:

```
╭─────────────────────────────╮
│  mcpgo  v0.1.0              │
│  MCP Bridge for Ollama      │
╰─────────────────────────────╯
  Model      gemma4:e4b
  Config     mcp.json
  Servers    1 connected
  Tools      5 available

┌────────────────┬─────────┬───────────┐
│ Server         │ Status  │ Tools     │
├────────────────┼─────────┼───────────┤
│ filesystem     │ ✓ ready │ 5 tools   │
└────────────────┴─────────┴───────────┘

❯ what files are in this project?

┃
assistant: Let me check the directory structure for you.

The project contains several key files and directories:

- cmd/mcpgo/main.go - Main entry point for the CLI application
- internal/ - Internal packages for MCP, Ollama, configuration, and UI
- go.mod and go.sum - Go module files
- Makefile - Build automation
- README.md - Project documentation
- mcp.json - Example configuration file

This is a Go CLI tool that bridges Ollama to MCP servers.
┃

❯ show me the main function

⚙  mcp › read_file  {"path": "cmd/mcpgo/main.go"}
↳  read_file  package main

import (
    "bufio"
    "context"
    ...
┃
assistant: The main function sets up the Cobra CLI framework with several subcommands:

1. **chat** - Interactive chat session (the default command)
2. **tools** - List all available tools
3. **models** - List Ollama models
4. **validate** - Validate config and test connectivity

The chat loop uses a REPL (read-eval-print loop) that:
- Loads configuration from mcp.json
- Validates the model supports tool calling
- Connects to MCP servers
- Maintains conversation history
- Processes user input and executes tool calls as needed

Special commands like `/tools`, `/clear`, `/model`, etc., are handled within the loop.
┃

❯ /clear
✓ Conversation cleared.

❯ exit
✓ Goodbye!
```

## How It Works

```
User input
    │
    ▼
┌─────────┐    JSON-RPC stdio    ┌─────────────┐
│  mcpgo  │ ◄──────────────────► │  MCP Server │
│ (bridge)│                      │ (filesystem,│
│         │    /api/chat HTTP     │  sqlite...) │
│         │ ◄──────────────────► └─────────────┘
└─────────┘
    │         Ollama
    ▼       (gemma4, qwen2.5...)
Assistant response
```

### Architecture layers:

1. **Config** — Loads and validates `mcp.json`, supporting Claude Desktop format
2. **MCP Client** — Spawns MCP servers as subprocesses, communicates via JSON-RPC 2.0 over stdio
3. **Ollama Client** — HTTP client for `/api/chat` and model validation
4. **Bridge** — Agentic loop: sends user message + tools to Ollama, executes returned tool calls, loops until done
5. **REPL** — Interactive prompt, special commands, graceful shutdown

Each layer is independent and testable. No global state; all dependencies are passed explicitly.

## Troubleshooting

**"connection refused" or "Ollama is not running"**

```bash
ollama serve
```

Then run mcpgo in another terminal.

**"model X not found locally"**

```bash
ollama pull gemma4:e4b
```

See [Supported Models](#supported-models) for recommended models.

**"model X does not support tool calling"**

Your model must support tool calling. Run:

```bash
mcpgo models
```

Look for models with a ✓ in the "Tool Support" column. Then switch your model in `mcp.json` or use:

```bash
mcpgo chat --model gemma4:e4b
```

**"failed to start MCP server"**

Run `mcpgo validate` for detailed diagnostics. Common issues:

- MCP command not found (e.g., `npx` not in PATH)
- Invalid `command` or `args` in `mcp.json`
- Required environment not available

**Enable verbose output**

```bash
mcpgo chat --verbose
```

This prints all tool calls and results to stderr so you can see what the model is doing.

## Development

### Build from source

```bash
make build
./bin/mcpgo chat
```

### Check code quality

```bash
make vet        # Go vet checks
make lint       # golangci-lint (if installed)
```

### Install development version

```bash
make install
mcpgo chat
```

## Contributing

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/my-feature`
3. Make your changes and commit: `git commit -am 'Add my feature'`
4. Push to your fork: `git push origin feature/my-feature`
5. Open a pull request

Before submitting:

```bash
mcpgo validate    # Ensure config is valid
go vet ./...      # Run code checks
go test ./...     # Run tests (when available)
```

If you have `golangci-lint` installed, also run:

```bash
golangci-lint run ./...
```

## License

MIT — See LICENSE file for details.

---

**Have questions?** Open an issue on [GitHub](https://github.com/manojshevate/mcpgo/issues).