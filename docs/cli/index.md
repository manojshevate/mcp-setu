# CLI Reference

Complete command-line reference for `mcp-setu`.

## Command Overview

| Command | Purpose |
|---------|---------|
| `mcp-setu chat` | Start interactive chat session (default) |
| `mcp-setu init` | Scaffold a starter `mcp.json` in the current directory |
| `mcp-setu tools` | List all available tools from MCP servers |
| `mcp-setu models` | List local Ollama models and tool support |
| `mcp-setu validate` | Validate config and test connectivity |
| `mcp-setu version` | Show version information |

## Global Flags

```
--config string    Path to config file (default: mcp.json)
--model string     Override model from config
--system string    Override system prompt from config
--verbose          Show tool calls and LLM iterations in the chat TUI
--version          Show version and exit
--help             Show help for a command
```

### Examples

```bash
# Use custom config
mcp-setu --config /etc/mcp/config.json chat

# Enable verbose output
mcp-setu --verbose chat

# Get help
mcp-setu --help
mcp-setu chat --help
```

## Commands

### chat

Start an interactive multi-turn chat session with Ollama using configured MCP tools. The chat runs in a full-screen TUI with a modern interface:

- **Input Box**: A bordered box at the bottom with rounded corners where you type your messages
- **Active Model Badge**: The currently selected model appears in the bottom-right corner (e.g., `◆ llama3.2:3b`)
- **Cursor Behavior**: The cursor starts at the beginning of the input field (like Claude Code) and moves naturally as you type
- **Scrolling Output**: Assistant responses, tool calls, and command output scroll in the panel above the input
- **Processing Timer**: While the model is thinking, a live elapsed timer shows above the input (e.g., `⟳ thinking… 2s`)

```bash
mcp-setu chat [flags]
```

**Flags:**

```
--model string      Override model from config (e.g., --model qwen2.5:7b)
--system string     Override system prompt from config
```

> **Note:** `--model` and `--system` are global flags and also work at the root level: `mcp-setu --model qwen2.5:7b`

**Key bindings:**

| Keys                | What it does                                          |
|---------------------|-------------------------------------------------------|
| `↑` / `↓`           | Cycle through previous inputs (or navigate menu options) |
| `/`                 | Trigger slash-command autocomplete (appears above input) |
| `Tab`               | Accept the selected suggestion                        |
| `Esc`               | Dismiss autocomplete or exit model picker             |
| `Enter`             | Submit                                                |
| `Ctrl+C` / `Ctrl+D` | Exit                                                  |

**Chat commands** (typed at the prompt):

| Command           | Purpose                                              |
|-------------------|------------------------------------------------------|
| `/tools`          | List all available tools                             |
| `/clear`          | Clear conversation history                           |
| `/model`          | Open interactive model picker (↑↓ navigate, Enter to select) |
| `/model <space>`  | Autocomplete model names (Tab to complete)          |
| `/model <name>`   | Switch to a specific model directly                 |
| `/stats`          | Show performance statistics                          |
| `/servers`        | Show connected MCP servers                           |
| `/help`           | Show chat help                                       |
| `/quit`, `/exit`  | Quit the session                                     |

**Processing Status:**

While your model is thinking, a live elapsed timer appears above the input:

```
⟳ thinking… 2s
⟳ thinking… 1m 30s
```

The timer updates every second and uses the format `Xs` for under 60 seconds and `Xm Ys` for longer durations.

**Examples:**

```bash
# Start chat with default config
mcp-setu chat

# Override model
mcp-setu chat --model llama3.3:70b

# Override system prompt
mcp-setu chat --system "You are a Python expert"

# Both
mcp-setu chat --model qwen2.5:7b --system "Code review expert"

# Verbose: show tool calls and LLM iterations in the TUI
mcp-setu chat --verbose
```

### init

Scaffold a starter `mcp.json` config file in the current directory.

```bash
mcp-setu init
```

Creates `mcp.json` with a minimal working example (filesystem server + Ollama config). Exits with an error if `mcp.json` already exists.

**Example:**

```bash
mkdir my-project && cd my-project
mcp-setu init
# ✓ Created mcp.json — edit it then run: mcp-setu chat
```

After running, open `mcp.json`, set your model and servers, then:

```bash
mcp-setu validate
mcp-setu chat
```

**Flags:**

```
--config string    Output path (default: mcp.json)
```

### tools

List all tools available from connected MCP servers.

```bash
mcp-setu tools
```

**Output Example:**

```
┌──────────────────────────────┬──────────────┬──────────────────────┐
│ Tool                         │ Server       │ Description          │
├──────────────────────────────┼──────────────┼──────────────────────┤
│ read_file                    │ filesystem   │ Read file contents   │
│ write_file                   │ filesystem   │ Write file contents  │
│ read_directory               │ filesystem   │ List directory       │
│ query                        │ sqlite       │ Execute SQL query    │
│ set_context                  │ memory       │ Store context        │
│ get_context                  │ memory       │ Retrieve context     │
└──────────────────────────────┴──────────────┴──────────────────────┘
```

**Flags:**

```
--config string    Path to config file (default: mcp.json)
--verbose           Print tool calls and results
```

### models

List Ollama models available locally and their tool-calling support status.

```bash
mcp-setu models
```

**Output Example:**

```
┌─────────────────────┬────────┬────────────────┐
│ Model               │ Size   │ Tool Support   │
├─────────────────────┼────────┼────────────────┤
│ llama3.2:3b          │ 4 GB   │ ✓ Yes          │
│ qwen2.5:7b          │ 6 GB   │ ✓ Yes          │
│ llama3.2:3b         │ 2 GB   │ ✓ Yes          │
│ mistral-nemo:12b    │ 9 GB   │ ✓ Yes          │
│ llama2:7b           │ 4 GB   │ ✗ No           │
└─────────────────────┴────────┴────────────────┘
```

**Flags:**

```
--config string    Path to config file (default: mcp.json)
```

### validate

Validate your `mcp.json` config file and test connectivity to Ollama and MCP servers.

```bash
mcp-setu validate
```

**What it checks:**

1. Config file syntax is valid JSON
2. Required environment variables are set (for `env`/`bearer-token` auth)
3. Ollama is reachable
4. Specified model exists and supports tool calling
5. All configured MCP servers can start (failures are reported but don't stop other servers from being checked)
6. Each server provides at least one tool

**Output Example (Success):**

```
✓ Config file valid
✓ Model llama3.2:3b found and supports tool calling
✓ MCP server "filesystem" connected with 3 tools
✓ MCP server "memory" connected with 2 tools
✓ All validations passed!
```

**Output Example (Failure):**

```
✗ Model llama2:7b does not support tool calling
Supported models:
  ✓ llama3.2:3b
  ✓ qwen2.5:7b
  ...
```

**Flags:**

```
--config string    Path to config file (default: mcp.json)
```

### version

Show version information.

```bash
mcp-setu version
```

**Output Example:**

```
mcp-setu version v0.1.1
commit: abc1234567890def
build date: 2025-05-10T12:00:00Z
```

## Exit Codes

| Code | Meaning |
|------|---------|
| 0 | Success |
| 1 | General error (validation failed, connection error, etc.) |

## Configuration

All commands respect the `--config` flag. By default, they look for `mcp.json` in the current directory.

**Precedence:**

1. `--config` flag (highest priority)
2. `mcp.json` in current directory
3. Default config if none found

## Flags Summary

| Flag | Type | Default | Usage |
|------|------|---------|-------|
| `--config` | string | `mcp.json` | Path to config file |
| `--verbose` | boolean | `false` | Enable debug output |
| `--model` | string | (config) | Override model |
| `--system` | string | (config) | Override system prompt |
| `--version` | boolean | `false` | Show version and exit |

## Examples

### Quick Examples

```bash
# Start interactive chat
mcp-setu chat

# See what tools are available
mcp-setu tools

# Check setup is working
mcp-setu validate

# Show version
mcp-setu version

# Use custom config
mcp-setu --config ~/my-mcp.json chat

# Enable debugging
mcp-setu --verbose chat

# Switch model
mcp-setu chat --model llama3.3:70b

# Custom system prompt
mcp-setu chat --system "You are a Rust expert"

# Everything together
mcp-setu --config ~/my-mcp.json --verbose chat --model qwen2.5:7b
```

### Real-World Workflows

```bash
# Before doing real work, validate
mcp-setu validate

# Start debugging
mcp-setu --verbose chat

# In chat, check what's available
/tools
/servers

# Switch to a better model for the task
/model llama3.3:70b

# Performance monitoring
/stats
```

## Tips & Tricks

1. **Quick tool check** — `mcp-setu tools` without starting interactive chat
2. **Model discovery** — `mcp-setu models` to see what's available
3. **Config testing** — `mcp-setu validate` before major work
4. **Verbose debugging** — Add `--verbose` when things break
5. **Quick help** — `mcp-setu <command> --help`

## Environment Variables

These are used by MCP servers configured in `mcp.json`:

```bash
# Example: Bearer token for remote MCP server
export MCP_API_TOKEN="sk-..."

# OAuth credentials
export MCP_AUTH_SERVER="https://auth.example.com"
export MCP_CLIENT_ID="client-id"
export MCP_CLIENT_SECRET="secret"
```

See [Configuration](../configuration.md#authentication) for more auth examples.

## Next Steps

- **[Getting Started](../getting-started.md)** — Tutorial
- **[Configuration](../configuration.md)** — Config reference
- **[Examples](../examples.md)** — Usage patterns
- **[Troubleshooting](../troubleshooting.md)** — Common issues
