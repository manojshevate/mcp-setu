# Troubleshooting

Solutions to common issues.

## Connection Issues

### "Connection refused" or "Ollama is not running"

**Problem:** mcp-setu can't reach Ollama.

**Solution:**

1. Start Ollama in a separate terminal:
   ```bash
   ollama serve
   ```

2. Verify it's running:
   ```bash
   curl http://localhost:11434/api/tags
   ```

3. Check your config points to the right URL:
   ```json
   {
     "ollama": {
       "baseUrl": "http://localhost:11434"
     }
   }
   ```

4. If Ollama is on a different machine, update `baseUrl`:
   ```json
   {
     "ollama": {
       "baseUrl": "http://192.168.1.100:11434"
     }
   }
   ```

### "Connection timeout" or "Ollama taking too long"

**Problem:** Ollama is slow to respond.

**Possible causes:**
- Model is too large for available RAM
- Ollama is running on CPU instead of GPU
- Context length is too large

**Solution:**

1. Use a smaller model:
   ```bash
   ollama pull llama3.2:3b
   mcp-setu chat --model llama3.2:3b
   ```

2. Reduce context length in config:
   ```json
   {
     "ollama": {
       "contextLength": 2048
     }
   }
   ```

3. Enable GPU in Ollama (see [Ollama docs](https://ollama.com/blog/gpu-acceleration))

## Model Issues

### "Model X not found locally"

**Problem:** Model isn't installed.

**Solution:**

```bash
# List what you have
mcp-setu models

# Pull a model
ollama pull llama3.2:3b

# Or if you want a different size
ollama pull qwen2.5:7b
ollama pull llama3.2:1b
```

### "Model X does not support tool calling"

**Problem:** You picked a model that can't call tools.

**Solution:**

1. Check which models support tools:
   ```bash
   mcp-setu models
   ```

2. Switch to a supported model:
   ```bash
   mcp-setu chat --model llama3.2:3b
   ```

3. Or update your config:
   ```json
   {
     "ollama": {
       "model": "llama3.2:3b"
     }
   }
   ```

**Supported models:**
- ✅ Gemma 4, 3
- ✅ Qwen 2.5, 3
- ✅ Llama 3.2, 3.3
- ✅ Mistral Nemo
- ✅ Command R
- ✅ Phi 4
- ✅ DeepSeek R1

See [Configuration](./configuration.md#supported-models) for details.

## MCP Server Issues

### "Failed to start MCP server" or "Command not found"

**Problem:** An MCP server can't start.

**Solution:**

1. **Check the command exists:**
   ```bash
   # If using npx
   which npx
   npx --version
   
   # If using custom command
   which <your-command>
   ```

2. **Update your config** to use the correct path:
   ```json
   {
     "mcpServers": {
       "filesystem": {
         "command": "/usr/local/bin/npx",
         "args": ["-y", "@modelcontextprotocol/server-filesystem", "."]
       }
     }
   }
   ```

3. **Install Node.js if using npx:**
   ```bash
   node --version
   npm --version
   ```

   If not installed, get it from [nodejs.org](https://nodejs.org).

### "MCP server connection timeout"

**Problem:** Server starts but doesn't respond.

**Possible causes:**
- Server is hung or crashed
- Malformed arguments
- Invalid working directory

**Solution:**

1. Test with a simpler server first:
   ```bash
   # Remove problematic servers temporarily
   # Keep just memory server which is most reliable
   ```

2. Check the server's requirements (e.g., SQLite needs a valid database file)

3. Use verbose mode to see more details:
   ```bash
   mcp-setu chat --verbose
   ```

### "Tool execution failed"

**Problem:** Tool ran but errored.

**Solution:**

1. **Check arguments** — Make sure you're using the tool correctly:
   ```
   /tools   # See tool definitions
   ```

2. **Verbose mode** shows the exact error:
   ```bash
   mcp-setu chat --verbose
   ```

3. **Check permissions** — For filesystem tools, ensure you have read/write access:
   ```bash
   ls -la /path/to/directory
   ```

4. **Check database** — For SQLite, ensure the database file exists:
   ```bash
   ls -la ./app.db
   ```

## Installation Issues

### "mcp-setu: command not found"

**Problem:** Binary isn't in PATH.

**Solution:**

1. Check GOPATH is set:
   ```bash
   echo $GOPATH
   ```

2. Check `$GOPATH/bin` is in PATH:
   ```bash
   echo $PATH
   ```

3. Add it to your shell profile if missing (`~/.bashrc`, `~/.zshrc`):
   ```bash
   export PATH="$HOME/go/bin:$PATH"
   source ~/.bashrc
   ```

4. Verify installation:
   ```bash
   mcp-setu version
   ```

### "go: unknown version" (building from source)

**Problem:** Go version is too old.

**Solution:**

```bash
go version  # Check current version

# Update Go (macOS)
brew upgrade go

# Or download from https://golang.org/dl/
```

Requires **Go 1.26+**.

## Configuration Issues

### "Invalid config file"

**Problem:** Config file has syntax errors.

**Solution:**

1. Validate your config:
   ```bash
   mcp-setu validate
   ```

2. Check JSON syntax (use an online JSON validator or your editor)

3. Common mistakes:
   - Missing commas between fields
   - Trailing commas in arrays/objects
   - Unquoted strings
   - Invalid escape sequences

### "Config file not found"

**Problem:** mcp-setu can't find `mcp.json`.

**Solution:**

1. Check the file exists:
   ```bash
   ls -la mcp.json
   ```

2. Specify the full path:
   ```bash
   mcp-setu --config /full/path/to/mcp.json chat
   ```

3. Or scaffold a starter config:
   ```bash
   mcp-setu init
   ```

## Debugging & Getting Help

### Enable verbose mode

```bash
mcp-setu chat --verbose
```

In the chat TUI, this surfaces additional lines inline in the output panel:
- `💭 Processing…` lines per LLM iteration
- `⚙ <tool-name>` for each tool call
- `↳ <result>` (truncated) for each tool result
- Warnings (e.g., streaming fallback) prefixed with `⚠`

### Run validation

```bash
mcp-setu validate
```

Tests:
- Config file syntax
- Ollama connectivity
- Model tool support
- Each MCP server
- All discovered tools

### List available tools

```bash
mcp-setu tools
```

Shows all tools from all connected servers.

### Check model status

```bash
mcp-setu models
```

Shows installed models and which support tool calling.

### Get help with a command

```bash
mcp-setu --help
mcp-setu chat --help
mcp-setu validate --help
```

## Still Stuck?

1. **Check the README** — https://github.com/manojshevate/mcp-setu/blob/main/README.md
2. **Open an issue** — https://github.com/manojshevate/mcp-setu/issues
3. **Start a discussion** — https://github.com/manojshevate/mcp-setu/discussions
4. **Check existing issues** — Your issue might already have a solution

When reporting issues, include:
- Output of `mcp-setu version`
- Your config file (with secrets removed)
- Output of `mcp-setu validate`
- Verbose output: `mcp-setu chat --verbose`
- What you were trying to do
- What happened instead
