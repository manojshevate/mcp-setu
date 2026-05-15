# Configuration

## Overview

`mcp-setu` reads configuration from `mcp.json` by default. You can specify a different path with `--config <path>`.

```bash
mcp-setu --config ~/.config/my-mcp.json chat
```

## Configuration File Format

The config has two main sections:

```json
{
  "ollama": {
    "baseUrl": "http://localhost:11434",
    "model": "llama3.2:3b",
    "systemPrompt": "You are an AI assistant with access to tools via the Model Context Protocol. Use available tools to help the user accomplish their tasks. Be helpful, concise, and direct.",
    "temperature": 0.7,
    "contextLength": 4096
  },
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "."]
    }
  }
}
```

### Ollama Section

| Field | Type | Default | Description |
|-------|------|---------|-------------|
| `baseUrl` | string | `http://localhost:11434` | Ollama API endpoint |
| `model` | string | `llama3.2:3b` | Model to use (must support tool calling) |
| `systemPrompt` | string | `"You are an AI assistant with access to tools via the Model Context Protocol..."` | System prompt sent with every message |
| `temperature` | number | `0.7` | Sampling temperature (0–1) |
| `contextLength` | number | `4096` | Max context window in tokens |

### MCP Servers Section

Map of server name → configuration. Each server config depends on its transport type.

## Transport Types

### 1. Stdio (Default)

Local subprocess-based servers. Used by most Node.js MCP servers.

```json
{
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "."],
      "env": {}
    }
  }
}
```

| Field | Type | Description |
|-------|------|-------------|
| `command` | string | Executable to run (must be in PATH) |
| `args` | array | Arguments to pass |
| `env` | object | Optional environment variables |

**Common servers:**

- **Filesystem** — File operations
  ```json
  {"command": "npx", "args": ["-y", "@modelcontextprotocol/server-filesystem", "."]}
  ```

- **SQLite** — Database queries
  ```json
  {"command": "npx", "args": ["-y", "@modelcontextprotocol/server-sqlite", "./db.sqlite"]}
  ```

- **Memory** — Persistent context storage
  ```json
  {"command": "npx", "args": ["-y", "@modelcontextprotocol/server-memory"]}
  ```

### 2. HTTP Streamable (Modern)

Remote servers using HTTP POST with streaming.

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

| Field | Type | Description |
|-------|------|-------------|
| `type` | string | Must be `"http-streamable"` |
| `url` | string | Server URL |
| `auth` | object | Optional authentication (see below) |

### 3. HTTP/SSE (Legacy)

Server-Sent Events (deprecated but still supported).

```json
{
  "mcpServers": {
    "legacy": {
      "type": "http-sse",
      "url": "http://legacy-mcp-server.com/events"
    }
  }
}
```

## Supported Models

| Model | Example | Notes |
|-------|---------|-------|
| **Llama 3.2** | `llama3.2:3b` | ⭐ Recommended—fast, reliable tool calling |
| **Gemma 3** | `gemma3:2b` | Efficient, strong tool use |
| **Qwen** | `qwen2.5:7b` | Excellent tool calling, great for coding |
| **Llama 3.3** | `llama3.3:70b` | Meta's largest, strongest reasoning |
| **Mistral** | `mistral-nemo:12b` | Balanced speed and accuracy |
| **Command R** | `command-r:35b` | Strong multi-tool chaining |
| **Phi 4** | `phi4:14b` | Compact and capable |
| **DeepSeek R1** | `deepseek-r1:7b` | Strong reasoning + tool use |

**Check local models:**

```bash
mcp-setu models
```

This shows which models are installed and which support tool calling.

## Authentication

For HTTP-based servers, you can add authentication:

### Bearer Token

```json
{
  "mcpServers": {
    "protected-api": {
      "type": "http-streamable",
      "url": "https://api.example.com/mcp",
      "auth": {
        "type": "bearer-token",
        "token": "your-api-token"
      }
    }
  }
}
```

**Better practice** — Use environment variable:

```json
{
  "auth": {
    "type": "bearer-token",
    "tokenEnvVar": "MCP_API_TOKEN"
  }
}
```

Then set before running:

```bash
export MCP_API_TOKEN="your-token"
mcp-setu chat
```

### OAuth 2.1

> **⚠ Not yet implemented.** The `oauth2` auth type is planned for a future release. Configuring it currently returns an error: `OAuth2 auth type is not yet implemented`. Use `bearer-token` or `env` auth types instead.

The config schema is reserved for when OAuth 2.1 support ships:

```json
{
  "mcpServers": {
    "oauth-server": {
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

Until OAuth2 is implemented, use `bearer-token` with a pre-obtained token or the `env` type to read a token from an environment variable.

## Claude Desktop Compatibility

Your Claude Desktop `mcp.json` works with `mcp-setu` immediately:

**macOS:**

```bash
mcp-setu --config ~/Library/Application\ Support/Claude/claude_desktop_config.json chat
```

**Windows (PowerShell):**

```bash
mcp-setu --config $env:APPDATA\Claude\claude_desktop_config.json chat
```

**Linux:**

```bash
mcp-setu --config ~/.config/Claude/claude_desktop_config.json chat
```

The only difference is the `ollama` block, which is mcp-setu-specific.

## Example: Complete Config

```json
{
  "ollama": {
    "baseUrl": "http://localhost:11434",
    "model": "qwen2.5:7b",
    "systemPrompt": "You are an expert Python developer. Write clean, efficient, and well-documented code. Help with debugging, optimization, and best practices. Use available tools to analyze files and test solutions.",
    "temperature": 0.5,
    "contextLength": 8192
  },
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/home/user/projects"]
    },
    "sqlite": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-sqlite", "./app.db"]
    },
    "remote-api": {
      "type": "http-streamable",
      "url": "https://my-mcp-server.com/mcp",
      "auth": {
        "type": "bearer-token",
        "tokenEnvVar": "MCP_API_TOKEN"
      }
    }
  }
}
```

## Validation

Check your config before starting:

```bash
mcp-setu validate
```

This verifies:
- Config file syntax
- Ollama connectivity
- Model supports tool calling
- All MCP servers can start
- Each server provides tools

## Next Steps

- **[Examples](./examples.md)** — Real-world usage patterns
- **[Troubleshooting](./troubleshooting.md)** — Solve issues
- **[CLI Reference](./cli/)** — Command documentation
