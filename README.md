# mcp-setu

A Go CLI that bridges [Ollama](https://ollama.com) to [MCP](https://modelcontextprotocol.io) servers, giving local language models tool-calling abilities in a modern interactive chat TUI. Drop-in compatible with your existing Claude Desktop `mcp.json`.

```
┌─────────┐    JSON-RPC stdio    ┌─────────────┐
│ mcp-setu │ ◄──────────────────► │  MCP Server │
│ (bridge)│                      │ (filesystem,│
│         │   /api/chat HTTP     │  sqlite, …) │
└─────────┘ ◄──────────────────►
   Ollama (gemma3, qwen2.5, llama3.2 …)
```

## Install

```bash
go install github.com/manojshevate/mcp-setu/cmd/mcp-setu@latest
# or
brew tap manojshevate/tap && brew install mcp-setu
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

The interactive chat is a full-screen TUI with the input pinned at the bottom and a scrolling output area on top.

| Keys           | What it does                              |
|----------------|-------------------------------------------|
| `↑` / `↓`      | Cycle through previous inputs             |
| `/`            | Open slash-command autocomplete           |
| `Tab`          | Accept the selected suggestion            |
| `Ctrl+C` / `Ctrl+D` | Exit                                 |

| Command          | What it does                       |
|------------------|------------------------------------|
| `/tools`         | List available tools               |
| `/servers`       | Show connected MCP servers         |
| `/model [name]`  | Show or switch model               |
| `/stats`         | Performance metrics                |
| `/clear`         | Reset conversation history         |
| `/help`          | Show all commands                  |
| `/quit`, `/exit` | Exit                               |

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

## Development

```bash
make build      # bin/mcp-setu
make test       # run all tests
make vet
```

See [CONTRIBUTING.md](CONTRIBUTING.md) and [docs/development.md](docs/development.md) for the architecture, package layout, and release process.

## License

MIT — see [LICENSE](LICENSE).
