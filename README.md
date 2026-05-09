# mcpgo - MCP Bridge for Ollama

**mcpgo** is a Go-native CLI application that bridges Ollama to MCP (Model Context Protocol) servers for interactive, multi-turn chat with local language models. It's Claude Desktop config-compatible, meaning you can reuse your existing MCP configuration without any changes. Fully open-source, no cloud dependency, runs entirely on your machine.

## What is mcpgo?

mcpgo connects your local Ollama instance to MCP servers, enabling models like Gemma, Qwen, and Llama to call tools (like filesystem access, database queries, etc.) during interactive conversations. The CLI is inspired by modern tools like Ghostty and Claude Code, with clean terminal UI and helpful error messages. Your existing Claude Desktop `mcp.json` works immediately—just point mcpgo at it.

## Prerequisites

- **Go 1.26+** — [Download](https://golang.org/dl/) — Latest version recommended for performance
- **Ollama** — [https://ollama.com](https://ollama.com) — Run `ollama serve` before starting mcpgo
- **Node.js 18+** — Required for npx-based MCP servers like the filesystem and SQLite tools
- **A tool-calling capable model** — See [Supported Models](#supported-models) for recommendations

## Quick Start

### 1. Install mcpgo

**Option A: Build from source**
```bash
git clone https://github.com/manojshevate/mcpgo
cd mcpgo
make install
```

**Option B: Build and run locally**
```bash
git clone https://github.com/manojshevate/mcpgo
cd mcpgo
make build
./bin/mcpgo chat
```

### 2. Start Ollama

```bash
ollama serve
```

Run this in a separate terminal. It will listen on `http://localhost:11434`.

### 3. Pull a model

```bash
ollama pull gemma4:e4b
```

See [Supported Models](#supported-models) for recommended models.

### 4. Run mcpgo

```bash
mcpgo chat
```

You'll see a startup banner showing connected MCP servers and available tools. Start typing naturally, and the model will call tools as needed.

**Example conversation:**
```
❯ what files are in this project?
assistant: Let me check the directory structure...
```

---

## Building with Make

The project includes a comprehensive Makefile:

```bash
make build       # Compile binary to bin/mcpgo
make install     # Install to $GOPATH/bin
make run         # Run chat directly
make run-verbose # Run with debug output
make vet         # Run go vet checks
make clean       # Remove build artifacts
```

All targets work with Go 1.26+ and handle dependencies automatically.

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

### Common Issues

**"Connection refused" or "Ollama is not running"**

```bash
# Terminal 1: Start Ollama
ollama serve

# Terminal 2: Run mcpgo
mcpgo chat
```

**"Model X not found locally"**

```bash
ollama pull gemma4:e4b
```

Run `mcpgo models` to see all locally available models.

**"Model X does not support tool calling"**

```bash
# Check which models support tool calling
mcpgo models

# Switch to a supported model
mcpgo chat --model llama3.2:latest

# Or update mcp.json
```

See [Supported Models](#supported-models) section for the full list.

**"Failed to start MCP server"**

```bash
# Run validation to diagnose the issue
mcpgo validate
```

Common causes:
- Command not in PATH (e.g., `npx` not found) — install Node.js 18+
- Invalid `command` or `args` in `mcp.json` — check syntax with `mcpgo validate`
- Directory does not exist — check relative paths in `mcp.json`
- Port already in use — less common with stdio transport but possible with network-based servers

**"Connection timeout" or "Ollama taking too long"**

- Increase model context length: `"contextLength": 8192` in `mcp.json`
- Use a smaller model: `gemma4:e4b` instead of `llama3.3:70b`
- Ensure sufficient RAM: Recommended 8GB+ for most models

**Enable verbose debugging**

```bash
mcpgo chat --verbose
```

This prints all tool calls and results to stderr, showing exactly what the model is doing.

### Performance Tips

1. **Use the right model size** — `gemma4:e4b` (4GB) vs `llama3.3:70b` (40GB) have very different performance
2. **Run Ollama on GPU** — Much faster than CPU: set `GPU_LAYERS=100` or check Ollama docs
3. **Monitor MCP servers** — Long-running tools can slow down the chat loop
4. **Keep context shorter** — Reduce `contextLength` to speed up responses (default 4096)

### Debug Mode

For development or deep troubleshooting:

```bash
# See tool calls as they happen
mcpgo chat --verbose

# Validate everything before starting chat
mcpgo validate

# List all discovered tools
mcpgo tools

# List all local models
mcpgo models
```

## Limitations & Future Work

### Current Limitations

- **No streaming responses** — Responses are waited for completely before displaying (by design for tool calling)
- **Linear tool calls** — Tools execute sequentially, not in parallel
- **No conversation persistence** — History is in-memory only; cleared on exit (use `/save` planned)
- **Single model at a time** — Can't switch models mid-conversation (planned)

### Planned Features

- [ ] Conversation history export (JSON, markdown)
- [ ] Model switching during chat (`/model switch gemma4:e4b`)
- [ ] Tool result filtering and transformation
- [ ] Multi-turn tool validation (prevent infinite loops)
- [ ] Built-in prompt templates
- [ ] Configuration profiles for different workflows
- [ ] Integration with cloud providers (Claude API fallback)

**Want to contribute?** See the [Contributing](#contributing) section.

---

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
make test       # Run tests (when available)
```

### Install development version

```bash
make install
mcpgo chat
```

### Project Structure

```
mcpgo/
├── cmd/mcpgo/          # CLI entry point (main.go)
├── internal/
│   ├── bridge/         # Agentic loop orchestrator
│   ├── config/         # Config loading and validation
│   ├── mcp/            # MCP stdio client (JSON-RPC 2.0)
│   ├── ollama/         # Ollama HTTP API client
│   └── ui/             # Terminal output formatting
├── Makefile            # Build targets
├── go.mod              # Module definition
├── go.sum              # Dependency checksums
├── mcp.json            # Example configuration
├── LICENSE             # MIT License
└── README.md           # This file
```

Each package is self-contained with clear responsibilities and no circular dependencies.

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

**MIT License** — This project is open source and free to use, modify, and distribute.

See the [LICENSE](LICENSE) file for full details. You're free to:
- ✅ Use for personal and commercial projects
- ✅ Modify and distribute
- ✅ Use with no attribution required (but appreciated!)

The only requirement is including the license notice in derivative works.

---

## Support

- **Issues & Bugs**: Open an issue on [GitHub](https://github.com/manojshevate/mcpgo/issues)
- **Questions**: Start a discussion or check existing issues
- **Contributing**: See [Contributing](#contributing) section above

---

## Acknowledgments

- Built with [Cobra](https://cobra.dev/) for CLI and [Charm](https://charm.sh/) for beautiful terminal UI
- Compatible with [Claude Desktop](https://claude.ai/code) MCP configuration format
- Inspired by tools like Ghostty and modern CLI applications