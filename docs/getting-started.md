# Getting Started

Welcome to **mcp-setu**! This guide will walk you through the basics and get you running in under 5 minutes.

## What is mcp-setu?

`mcp-setu` is a Go-native CLI that connects your local Ollama instance to MCP (Model Context Protocol) servers. This lets you run conversations with language models like Gemma, Qwen, and Llama while they can call tools—filesystems, databases, APIs—to answer questions accurately.

Think of it as a local alternative to Claude with tool-calling powered entirely by your machine.

## Prerequisites

- **Go 1.26+** — [Download](https://golang.org/dl/)
- **Ollama** — [https://ollama.com](https://ollama.com)
- **Node.js 18+** — Only needed for Node-based MCP servers like filesystem or SQLite
- **A tool-capable model** — Gemma4, Qwen, or Llama (see [Supported Models](./configuration.md#supported-models))

## 5-Minute Quick Start

### 1. Install mcp-setu

**Using `go install` (Recommended)**:

```bash
go install github.com/manojshevate/mcp-setu/cmd/mcp-setu@latest
```

**Using Homebrew (macOS)**:

```bash
brew tap manojshevate/mcp-setu
brew install mcp-setu
```

**From source**:

```bash
git clone https://github.com/manojshevate/mcp-setu
cd mcp-setu
make install
```

Verify installation:

```bash
mcp-setu version
```

### 2. Start Ollama

In a separate terminal:

```bash
ollama serve
```

### 3. Pull a Model

```bash
ollama pull llama3.2:3b
```

`llama3.2:3b` is recommended for tool calling. See [Supported Models](./configuration.md#supported-models) for alternatives.

### 4. Create a config file

If you don't have an `mcp.json` yet, scaffold one:

```bash
mcp-setu init
```

This creates a minimal `mcp.json` with a filesystem server. Edit it to add your servers and model, then validate:

```bash
mcp-setu validate
```

### 5. Run mcp-setu

```bash
mcp-setu chat
```

This opens a full-screen TUI. The input is pinned at the bottom; assistant messages, tool calls, and command output scroll above. Type naturally:

```
❯ what files are in this directory?
```

The model will call tools as needed to answer your question.

## What Just Happened?

```
User Input
    ↓
Ollama (with tool definitions)
    ↓ (returns tool calls)
mcp-setu (executes them)
    ↓ (gets results)
Ollama (generates response)
    ↓
User sees the answer
```

## Chat Interface Details

The chat TUI has a clean, modern layout with an input box pinned at the bottom and scrolling content above. Here's what you'll see:

- **Bordered Input Box**: Your messages are typed in a styled box with rounded corners
- **Active Model Badge**: The currently selected model appears in the bottom-right corner (e.g., "◆ llama3.2:3b")
- **Cursor Position**: The cursor starts at the beginning of the input field (like Claude Code) and moves naturally as you type
- **Autocomplete Dropdown**: When you type `/`, suggestions appear **above** the input field for better visibility

### Navigation

| Keys | What it does |
|------|--------------|
| `/` | Trigger autocomplete dropdown (appears above input) |
| `↑` / `↓` | Navigate history (or menu options when autocompleting) |
| `Tab` | Accept the selected suggestion |
| `Esc` | Dismiss autocomplete or exit model picker |

### Commands

| Command | What it does |
|---------|--------------|
| `/tools` | Show all available tools |
| `/model` | **Interactive model picker** — navigate with ↑↓, select with Enter |
| `/model <space>` | **Autocomplete model names** — use Tab to complete |
| `/model <name>` | Switch to specific model directly |
| `/stats` | Performance metrics |
| `/servers` | Connected MCP servers |
| `/clear` | Reset conversation |
| `/help` | Show all commands |
| `/quit` / `/exit` | Quit (also `Ctrl+C`) |

### Real-Time Feedback

While your model is thinking, a **live elapsed timer** shows above the input:

```
⟳ thinking… 2s
⟳ thinking… 45s
⟳ thinking… 2m 15s
```

This makes it easy to see how long the model has been processing your request.

Try `/tools` to see what MCP servers are connected.

## Next Steps

- **[Installation](./installation.md)** — Detailed install methods
- **[Configuration](./configuration.md)** — Set up custom MCP servers
- **[Examples](./examples.md)** — Real-world usage patterns
- **[Troubleshooting](./troubleshooting.md)** — Solve common issues

## Troubleshooting the Quick Start

**"Connection refused"**

```bash
# Terminal 1
ollama serve

# Terminal 2
mcp-setu chat
```

**"Model not found"**

```bash
ollama pull llama3.2:3b
```

**"Model does not support tool calling"**

```bash
mcp-setu models  # See which models support tools
```

## Need Help?

- **Issues?** Open a GitHub issue: https://github.com/manojshevate/mcp-setu/issues
- **Questions?** Start a discussion
- **Want to contribute?** See [CONTRIBUTING.md](https://github.com/manojshevate/mcp-setu/blob/main/CONTRIBUTING.md)
