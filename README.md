# mcp-setu - MCP Bridge for Ollama

**mcp-setu** is a Go-native CLI application that bridges Ollama to MCP (Model Context Protocol) servers for interactive, multi-turn chat with local language models. It's Claude Desktop config-compatible, meaning you can reuse your existing MCP configuration without any changes. Fully open-source, no cloud dependency, runs entirely on your machine.

## What is mcp-setu?

mcp-setu connects your local Ollama instance to MCP servers, enabling models like Gemma, Qwen, and Llama to call tools (like filesystem access, database queries, etc.) during interactive conversations. The CLI is inspired by modern tools like Ghostty and Claude Code, with clean terminal UI and helpful error messages. Your existing Claude Desktop `mcp.json` works immediately—just point mcp-setu at it.

## Prerequisites

- **Go 1.26+** — [Download](https://golang.org/dl/) — Latest version recommended for performance
- **Ollama** — [https://ollama.com](https://ollama.com) — Run `ollama serve` before starting mcp-setu
- **Node.js 18+** — Required for npx-based MCP servers like the filesystem and SQLite tools
- **A tool-calling capable model** — See [Supported Models](#supported-models) for recommendations

## Installation

### Using `go install` (Recommended)

```bash
go install github.com/manojshevate/mcp-setu/cmd/mcp-setu@latest
```

This installs the latest released version directly to `$GOPATH/bin/mcp-setu`.

For a specific version:
```bash
go install github.com/manojshevate/mcp-setu/cmd/mcp-setu@v0.1.0
```

### Using Homebrew (macOS)

```bash
brew tap manojshevate/tap
brew install mcp-setu
```

To upgrade:
```bash
brew upgrade mcp-setu
```

### Building from Source

```bash
git clone https://github.com/manojshevate/mcp-setu
cd mcp-setu
make install
```

Or to build and run locally without installing:
```bash
make build
./bin/mcp-setu chat
```

## Quick Start

### 1. Verify Installation

After installing via any method above, verify the installation:

```bash
mcp-setu version
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

### 4. Run mcp-setu

```bash
mcp-setu chat
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
make build       # Compile binary to bin/mcp-setu
make install     # Install to $GOPATH/bin
make run         # Run chat directly
make run-verbose # Run with debug output
make vet         # Run go vet checks
make clean       # Remove build artifacts
```

All targets work with Go 1.26+ and handle dependencies automatically.

## Claude Desktop Config Compatibility

mcp-setu uses the **exact same** `mcpServers` configuration format as Claude Desktop. Your existing Claude Desktop MCP setup works immediately without modification.

**Where to find your Claude Desktop config:**

- **macOS**: `~/Library/Application Support/Claude/claude_desktop_config.json`
- **Windows**: `%APPDATA%\Claude\claude_desktop_config.json`
- **Linux**: `~/.config/Claude/claude_desktop_config.json`

**To use your Claude Desktop MCP setup with mcp-setu:**

```bash
# macOS
mcp-setu --config ~/Library/Application\ Support/Claude/claude_desktop_config.json chat

# Windows (PowerShell)
mcp-setu --config $env:APPDATA\Claude\claude_desktop_config.json chat

# Linux
mcp-setu --config ~/.config/Claude/claude_desktop_config.json chat
```

The only difference is the `ollama` block, which is mcp-setu-specific and not present in Claude Desktop configs. mcp-setu will use sensible defaults if the block is missing.

## Configuration

mcp-setu reads its configuration from `mcp.json` by default. Use `--config <path>` to specify a different file.

**Current default configuration (`mcp.json`):**

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
    "memory": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-memory"]
    }
  }
}
```

**Other server examples:**

You can customize `mcp.json` to add more servers. mcp-setu supports multiple transport types:

**Stdio Transport (Default - Local Subprocess):**

```json
{
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
    },
    "memory": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-memory"]
    }
  }
}
```

**HTTP Streamable Transport (Remote Servers - Modern Standard):**

```json
{
  "mcpServers": {
    "remote-api": {
      "type": "http-streamable",
      "url": "http://your-mcp-server.com/mcp"
    }
  }
}
```

**HTTP/SSE Transport (Legacy - Deprecated but Still Supported):**

```json
{
  "mcpServers": {
    "legacy-server": {
      "type": "http-sse",
      "url": "http://legacy-mcp-server.com/events"
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
- **mcpServers** — Map of MCP server configs (base format compatible with Claude Desktop)
  - **type** — Transport type: `"stdio"` (default), `"http-streamable"` (modern), or `"http-sse"` (legacy/deprecated)
  - **command** — Executable to run (stdio transport only)
  - **args** — Arguments to pass (stdio transport only)
  - **env** — Optional environment variables (stdio transport only)
  - **url** — Server URL (http-streamable and http-sse transports only)
  - **auth** — Optional authentication configuration (see [MCP Authentication](#mcp-authentication--authorization))
    - **type** — Auth type: `"none"` (default), `"bearer-token"`, `"oauth2"`, or `"env"`
    - **token** — Static bearer token (bearer-token type)
    - **tokenEnvVar** — Environment variable containing bearer token
    - **authorizationServerUrl** — OAuth 2.1 authorization server URL (oauth2 type)
    - **authorizationServerEnvVar** — Env var containing auth server URL (oauth2 type)
    - **clientId** — OAuth 2.1 client ID (oauth2 type)
    - **clientSecret** — OAuth 2.1 client secret (oauth2 type)
    - **scopes** — OAuth 2.1 scopes to request (oauth2 type)

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

> **mcp-setu checks tool support at startup** and exits with a clear error if your model doesn't support tool calling. Run `mcp-setu models` to see what's available locally and which models are compatible.

## CLI Reference

### Root flags

```
--config string    Path to config file (default: mcp.json)
--verbose           Print tool calls and results to stderr
```

### Commands

| Command            | Description                                           |
|--------------------|-------------------------------------------------------|
| `mcp-setu chat`       | Start interactive chat session (default)              |
| `mcp-setu tools`      | List all tools from configured MCP servers and exit   |
| `mcp-setu models`     | List local Ollama models and tool support status      |
| `mcp-setu validate`   | Validate config and test MCP server connectivity      |

### Chat flags

```
--model string     Override model from config (e.g., --model qwen2.5:7b)
--system string    Override system prompt from config
```

**Example:**

```bash
mcp-setu chat --verbose --model gemma4:e4b --system "You are a Python expert."
```

## REPL Commands

Once in the interactive chat loop, these commands are available:

| Command           | Description                               |
|-------------------|-------------------------------------------|
| `/tools`          | List all available tools                  |
| `/clear`          | Clear conversation history                |
| `/model`          | Show current model and available options  |
| `/model <name>`   | Switch to a different model               |
| `/stats`          | Show performance statistics               |
| `/servers`        | Show connected MCP servers                |
| `/help`           | Show this help                            |
| `exit`/`quit`     | Quit mcp-setu                                |

## Example Session

Here's a realistic multi-turn session using `gemma4:e4b` with the memory MCP server:

```
╭─────────────────────────────╮
│  mcp-setu  v0.1.0              │
│  MCP Bridge for Ollama      │
╰─────────────────────────────╯
  Model      gemma4:e4b
  Config     mcp.json
  Servers    1 connected
  Tools      2 available

┌────────────────┬─────────┬───────────┐
│ Server         │ Status  │ Tools     │
├────────────────┼─────────┼───────────┤
│ memory         │ ✓ ready │ 2 tools   │
└────────────────┴─────────┴───────────┘

❯ remember my project is a Go CLI tool for bridging Ollama to MCP servers

⚙  mcp › set_context  {"key": "project", "value": "Go CLI tool for Ollama-MCP bridge"}
↳  set_context  Context stored successfully
┃
assistant: Got it! I'll remember that your project is a Go CLI tool for bridging Ollama to MCP servers. This context will be available for our conversation.
┃

❯ what was our previous conversation about?

⚙  mcp › get_context  {}
↳  get_context  We discussed implementing a new feature for the app.

The model remembers our previous conversation about implementing a new feature. The memory server stores context between messages.
┃

❯ /tools

┌─────────────────────────────┬──────────────┬──────────────────────────────────┐
│ Tool                        │ Server       │ Description                      │
├─────────────────────────────┼──────────────┼──────────────────────────────────┤
│ get_context                 │ memory       │ Retrieve stored conversation... │
│ set_context                 │ memory       │ Store new conversation contex... │
└─────────────────────────────┴──────────────┴──────────────────────────────────┘

❯ /model

Current model: gemma4:e4b
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
│  mcp-setu  │ ◄──────────────────► │  MCP Server │
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

## MCP Transport Mechanisms

mcp-setu supports all three MCP transport mechanisms for maximum flexibility:

### Transport Types Supported

1. **Stdio (JSON-RPC 2.0)** ✅ Default
   - MCP servers spawned as subprocesses
   - Communication via stdin/stdout
   - No network overhead, ideal for local servers
   - Supports Claude Desktop `mcp.json` format directly
   - Best for: Local Node.js servers, development, and reliable connections

2. **HTTP Streamable** ✅ Modern Standard
   - HTTP POST requests with streaming response support
   - Ideal for remote MCP servers
   - Built-in connection resumption via Last-Event-ID headers
   - Best for: Cloud-hosted servers, remote integrations, load balancing

3. **HTTP/SSE** ✅ Legacy (Deprecated but Supported)
   - Server-Sent Events for streaming responses
   - Deprecated in favor of HTTP Streamable
   - Still widely used by existing servers
   - Best for: Legacy server compatibility, gradual migration paths

### Configuration Examples

**Default Stdio Transport:**
```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "."]
    }
  }
}
```

**HTTP Streamable (Modern):**
```json
{
  "mcpServers": {
    "remote-api": {
      "type": "http-streamable",
      "url": "http://your-mcp-server.com/mcp"
    }
  }
}
```

**HTTP/SSE (Legacy):**
```json
{
  "mcpServers": {
    "legacy-server": {
      "type": "http-sse",
      "url": "http://legacy-mcp-server.com/events"
    }
  }
}
```

### Transport Selection Guide

| Use Case | Transport | Why |
|----------|-----------|-----|
| Local Node.js servers | Stdio | Simple, reliable, no network overhead |
| Cloud-hosted servers | HTTP Streamable | Modern standard, built for remote use |
| Legacy servers (pre-SSE deprecation) | HTTP/SSE | Compatibility with existing deployments |
| Mixed environments | Multiple in same config | Mix stdio, HTTP Streamable, and HTTP/SSE as needed |

### MCP Standard Reference

For more information, see the [MCP specification on transports](https://modelcontextprotocol.io/specification/2025-06-18/basic/transports).

> **Note**: HTTP/SSE is deprecated in the MCP standard in favor of HTTP Streamable, but mcp-setu continues to support it for compatibility with existing servers.

## MCP Authentication & Authorization

mcp-setu implements the MCP authorization specification for protecting remote MCP servers. This follows OAuth 2.1 standards with PKCE, Protected Resource Metadata discovery, and scope-based access control.

### Supported Authentication Methods

**For HTTP-based transports** (HTTP Streamable and HTTP/SSE):
- ✅ **Bearer Token** — Simple static token auth (suitable for API tokens)
- ✅ **OAuth 2.1** — Full OAuth 2.1 flow with PKCE, scope challenges, and token refresh
- ✅ **Environment Variables** — Tokens read from environment vars for secure credential handling

**For Stdio transport:**
- ✅ **Environment Variables Only** — Per MCP spec, stdio should not use OAuth; credentials come from env vars

### Authentication Configuration

#### 1. No Authentication (Default)

```json
{
  "mcpServers": {
    "local-server": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-memory"]
    }
  }
}
```

#### 2. Bearer Token Authentication

```json
{
  "mcpServers": {
    "protected-api": {
      "type": "http-streamable",
      "url": "https://api.example.com/mcp",
      "auth": {
        "type": "bearer-token",
        "token": "your-api-token-here"
      }
    }
  }
}
```

**Better practice** — Store token in environment variable:

```json
{
  "mcpServers": {
    "protected-api": {
      "type": "http-streamable",
      "url": "https://api.example.com/mcp",
      "auth": {
        "type": "bearer-token",
        "tokenEnvVar": "MCP_API_TOKEN"
      }
    }
  }
}
```

Then set the environment variable before running:
```bash
export MCP_API_TOKEN="your-api-token"
mcp-setu chat
```

#### 3. OAuth 2.1 Authorization

For protected MCP servers with OAuth 2.1:

```json
{
  "mcpServers": {
    "oauth-protected": {
      "type": "http-streamable",
      "url": "https://api.example.com/mcp",
      "auth": {
        "type": "oauth2",
        "authorizationServerUrl": "https://auth.example.com",
        "clientId": "your-client-id",
        "clientSecret": "your-client-secret",
        "scopes": ["mcp:read", "mcp:write"]
      }
    }
  }
}
```

Or use environment variables for credentials (recommended for production):

```json
{
  "mcpServers": {
    "oauth-protected": {
      "type": "http-streamable",
      "url": "https://api.example.com/mcp",
      "auth": {
        "type": "oauth2",
        "authorizationServerEnvVar": "MCP_AUTH_SERVER",
        "scopes": ["mcp:read", "mcp:write"]
      }
    }
  }
}
```

#### 4. Environment Variable Token (Stdio)

For stdio-based servers, use environment variables:

```json
{
  "mcpServers": {
    "secure-local": {
      "type": "stdio",
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-custom"],
      "auth": {
        "type": "env",
        "tokenEnvVar": "MCP_LOCAL_TOKEN"
      }
    }
  }
}
```

### How MCP Authorization Works

1. **401 Unauthorized Responses**: When an MCP server returns HTTP 401, mcp-setu parses the `WWW-Authenticate` header to discover:
   - Resource metadata URL (where to find scope requirements)
   - Required scopes for the resource
   - Authorization server location

2. **Protected Resource Metadata** (RFC 9728): MCP servers advertise their authorization requirements via:
   - WWW-Authenticate header with `resource_metadata` parameter
   - Well-known URIs: `/.well-known/oauth-protected-resource` or `/.well-known/oauth-protected-resource/{path}`

3. **Bearer Tokens** (RFC 6750): Valid tokens are sent in the Authorization header:
   ```
   Authorization: Bearer <access-token>
   ```

4. **Scope Challenges** (Step-Up Authorization): When a token lacks required scopes, the server returns HTTP 403 with:
   ```
   WWW-Authenticate: Bearer error="insufficient_scope", scope="required:scope1 required:scope2"
   ```

5. **Resource Indicators** (RFC 8707): OAuth requests include the target resource:
   ```
   resource=https://api.example.com/mcp
   ```

### Security Best Practices

✅ **DO:**
- Store sensitive credentials (tokens, secrets) in environment variables
- Use HTTPS/TLS for all HTTP-based MCP servers
- Use Bearer tokens for stateless, simple authentication
- Use OAuth 2.1 for dynamic, revocable access with scopes
- Keep tokens short-lived (authorization servers SHOULD issue short-lived access tokens)
- Validate HTTPS certificates and use strong TLS versions

❌ **DON'T:**
- Store secrets in `mcp.json` config files
- Use unencrypted HTTP for authentication
- Pass tokens in query strings (only use Authorization headers)
- Reuse tokens across different MCP servers
- Cache long-lived tokens without refresh mechanisms

### Environment Variables for Authentication

Common environment variable patterns:

```bash
# Bearer tokens
export MCP_API_TOKEN="sk-..."                    # Static API token
export MCP_CUSTOM_TOKEN="token-..."              # Custom token for specific server

# OAuth 2.1
export MCP_AUTH_SERVER="https://auth.example.com"
export MCP_CLIENT_ID="client-id"
export MCP_CLIENT_SECRET="secret"                # Only in secure environments

# Multiple servers
export MCP_SERVER_1_TOKEN="token-1"
export MCP_SERVER_2_TOKEN="token-2"
```

### Authentication Standard Reference

mcp-setu implements the MCP Authorization specification:
- [MCP Authorization & Authorization Specification](https://modelcontextprotocol.io/specification/draft/basic/authorization)
- [OAuth 2.1 (IETF Draft)](https://datatracker.ietf.org/doc/html/draft-ietf-oauth-v2-1-13)
- [PKCE (RFC 7636)](https://datatracker.ietf.org/doc/html/rfc7636)
- [Protected Resource Metadata (RFC 9728)](https://datatracker.ietf.org/doc/html/rfc9728)
- [Resource Indicators (RFC 8707)](https://www.rfc-editor.org/rfc/rfc8707.html)
- [Bearer Token Usage (RFC 6750)](https://datatracker.ietf.org/doc/html/rfc6750)

> **Note**: Authorization is **OPTIONAL** in MCP. Public servers don't require authentication, while private servers may enforce OAuth 2.1, bearer tokens, or other mechanisms.

## Troubleshooting

### Common Issues

**"Connection refused" or "Ollama is not running"**

```bash
# Terminal 1: Start Ollama
ollama serve

# Terminal 2: Run mcp-setu
mcp-setu chat
```

**"Model X not found locally"**

```bash
ollama pull gemma4:e4b
```

Run `mcp-setu models` to see all locally available models.

**"Model X does not support tool calling"**

```bash
# Check which models support tool calling
mcp-setu models

# Switch to a supported model
mcp-setu chat --model llama3.2:latest

# Or update mcp.json
```

See [Supported Models](#supported-models) section for the full list.

**"Failed to start MCP server"**

```bash
# Run validation to diagnose the issue
mcp-setu validate
```

Common causes:
- Command not in PATH (e.g., `npx` not found) — install Node.js 18+
- Invalid `command` or `args` in `mcp.json` — check syntax with `mcp-setu validate`
- Directory does not exist — check relative paths in `mcp.json`
- Port already in use — less common with stdio transport but possible with network-based servers

**"Connection timeout" or "Ollama taking too long"**

- Increase model context length: `"contextLength": 8192` in `mcp.json`
- Use a smaller model: `gemma4:e4b` instead of `llama3.3:70b`
- Ensure sufficient RAM: Recommended 8GB+ for most models

**Enable verbose debugging**

```bash
mcp-setu chat --verbose
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
mcp-setu chat --verbose

# Validate everything before starting chat
mcp-setu validate

# List all discovered tools
mcp-setu tools

# List all local models
mcp-setu models
```

## Limitations & Future Work

### Current Limitations

- **No streaming responses** — Responses are waited for completely before displaying (by design for tool calling)
- **Linear tool calls** — Tools execute sequentially, not in parallel (FIXED: now parallel for independent calls!)
- **No conversation persistence** — History is in-memory only; cleared on exit
- **Single model at a time** — Can't switch models mid-conversation

### Planned Enhancements

#### ✅ Completed

- [x] **Model switching during chat**
  - Change models mid-conversation with `/model <name>` command
  - Auto-suggests available models when you run `/model` without arguments
  - Validates tool support before switching

- [x] **Performance monitoring**
  - Track response times, tool calls, iterations, and session duration
  - Display stats with `/stats` command
  - Show summary on exit with message count, tools, iterations, and duration

#### 🚀 High Priority

- [ ] **Conversation history export**
  - Save chat history to JSON or Markdown format
  - Load previous conversations
  - Search within history
  - Implementation: Add `SaveHistory(filename string)` and `LoadHistory(filename string)` methods to bridge

- [ ] **Tool result filtering**
  - Filter or transform tool results before sending to model
  - Useful for large outputs (truncate logs, summarize files)
  - Implementation: Add `ResultFilter` interface with default and custom implementations

- [ ] **Streaming responses**
  - Show model output as it arrives (real-time typing effect)
  - Improves UX for long responses
  - Implementation: Use Ollama `/api/chat?stream=true`, parse SSE events, print chunks

#### 📊 Medium Priority

- [ ] **Multi-turn tool validation**
  - Detect and prevent infinite tool loops (currently max 20 iterations)
  - Tool result caching to avoid redundant calls
  - Implementation: Track tool calls per iteration, cache results by (tool_name, args_hash)

- [ ] **Configuration profiles**
  - Multiple named profiles (e.g., "coding", "research", "creative")
  - Switch profiles with `/profile <name>`
  - Implementation: Allow multiple mcp.json sections or separate profile files

- [ ] **Built-in prompt templates**
  - Pre-built system prompts for different use cases
  - `/template <name>` to load templates
  - Examples: "code_expert", "researcher", "creative_writer"
  - Implementation: Embed templates in binary, or load from `~/.mcp-setu/templates/`

- [ ] **Tool authorization**
  - Require user confirmation before executing sensitive tools
  - Tool whitelisting/blacklisting
  - Implementation: Add `RequiresApproval` field to tool metadata

#### 🔮 Nice to Have

- [ ] **Circuit breaker for failing tools**
  - Auto-disable tools that consistently fail
  - Fallback to alternative tools
  - Implementation: Track failure rate per tool, skip if > threshold

- [ ] **Integration with cloud providers**
  - Claude API fallback when Ollama is unavailable
  - OpenAI API support
  - Implementation: Add `--fallback-provider claude|openai` flag

- [ ] **MCP server auto-discovery**
  - Scan common locations for MCP servers
  - Auto-register compatible servers
  - Implementation: Search `/usr/local/bin`, `~/.local/bin`, `node_modules/.bin/`

- [ ] **Performance monitoring**
  - Track response times, token usage, cost estimates
  - Display stats: `/stats`
  - Implementation: Add timing metadata to messages, calculate token counts

- [ ] **Persistent conversation database**
  - SQLite backend for chat history
  - Full-text search
  - Implementation: Use `internal/db` package with migration system

- [ ] **Interactive tool debugging**
  - `/debug <tool_name>` to test tool calls
  - Inspect tool schemas and arguments
  - Implementation: Add interactive REPL mode for tool testing

### Contributing Ideas

Found an enhancement you want to implement?

1. **Create an issue** describing the enhancement
2. **Reference this section** in your PR to link the enhancement
3. **Add tests** for new functionality (we have 40+ tests already!)
4. **Update documentation** if user-facing

See [CONTRIBUTING.md](CONTRIBUTING.md) for full guidelines.

### Test Coverage Summary

Current test suite coverage:

| Package | Coverage | Status |
|---------|----------|--------|
| bridge | 82.2% | ✅ Comprehensive |
| config | 92.6% | ✅ Excellent |
| mcp | 13.5% | ⚠️ Structure tests only (network hard to test) |
| ollama | 28.6% | ⚠️ Structure tests only (HTTP hard to test) |
| ui | 65.2% | ✅ Good |
| **Overall** | **56.4%** | ✅ Strong foundation |

Run tests with: `make test` or `make coverage`

---

## Documentation

### Local Development

The documentation is built with [VitePress](https://vitepress.dev/) and hosted on GitHub Pages.

**First time setup:**

```bash
npm install
```

**Run docs locally:**

```bash
make docs-dev
```

Visit `http://localhost:5173` to see the docs.

**Build for production:**

```bash
make docs-build
```

**Preview built docs:**

```bash
make docs-preview
```

**Generate CLI reference** (auto-run during build):

```bash
make docs:generate-cli
```

### Documentation Structure

- `docs/index.md` — Home page
- `docs/getting-started.md` — Quick start guide
- `docs/installation.md` — Installation methods
- `docs/configuration.md` — Config reference
- `docs/examples.md` — Usage examples
- `docs/troubleshooting.md` — Common issues
- `docs/concepts.md` — How it works
- `docs/development.md` — Contributing guide
- `docs/cli/index.md` — CLI reference
- `docs/.vitepress/config.ts` — VitePress configuration

### Deployment

Documentation is automatically deployed to GitHub Pages via GitHub Actions when you push to `main`. See `.github/workflows/deploy-docs.yml` for details.

Live docs: https://manojshevate.github.io/mcp-setu/

## Development

### Build from source

```bash
make build
./bin/mcp-setu chat
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
mcp-setu chat
```

### Project Structure

```
mcp-setu/
├── cmd/mcp-setu/          # CLI entry point (main.go)
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
mcp-setu validate    # Ensure config is valid
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

- **Issues & Bugs**: Open an issue on [GitHub](https://github.com/manojshevate/mcp-setu/issues)
- **Questions**: Start a discussion or check existing issues
- **Contributing**: See [Contributing](#contributing) section above

---

## Acknowledgments

- Built with [Cobra](https://cobra.dev/) for CLI and [Charm](https://charm.sh/) for beautiful terminal UI
- Compatible with [Claude Desktop](https://claude.ai/code) MCP configuration format
- Inspired by tools like Ghostty and modern CLI applications