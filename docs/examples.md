# Examples

Real-world usage patterns for `mcp-setu`.

## Example 1: File Analysis with Filesystem Server

Analyze code files in your project:

```bash
# Set up config
cat > mcp.json << 'EOF'
{
  "ollama": {
    "model": "gemma4:e4b",
    "systemPrompt": "You are a code review expert. Use the filesystem to analyze and review code."
  },
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "/path/to/project"]
    }
  }
}
EOF

# Start chat
mcp-setu chat
```

**In the chat:**

```
❯ review the main.go file and suggest improvements

⚙ filesystem › read_file {"path": "main.go"}
↳ read_file [file content...]

assistant: I've reviewed main.go. Here are the suggestions:
1. Extract config loading into a separate function
2. Add error context wrapping
3. Consider using sync.Once for singleton instances
...
```

## Example 2: Database Queries with SQLite Server

Query a database during conversation:

```bash
cat > mcp.json << 'EOF'
{
  "ollama": {
    "model": "qwen2.5:7b",
    "systemPrompt": "You are a data analyst. Use SQLite to query the database."
  },
  "mcpServers": {
    "sqlite": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-sqlite", "./app.db"]
    }
  }
}
EOF

mcp-setu chat
```

**In the chat:**

```
❯ how many users registered in the last 7 days?

⚙ sqlite › query {"sql": "SELECT COUNT(*) FROM users WHERE created_at > datetime('now', '-7 days')"}
↳ query [{"COUNT(*)": 42}]

assistant: There were 42 new user registrations in the last 7 days. This is up from 38 last week.
```

## Example 3: Context Storage with Memory Server

Maintain context across conversations:

```bash
cat > mcp.json << 'EOF'
{
  "ollama": {
    "model": "gemma4:e4b",
    "systemPrompt": "You are a helpful assistant. Use memory to maintain context."
  },
  "mcpServers": {
    "memory": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-memory"]
    }
  }
}
EOF

mcp-setu chat
```

**Session 1:**

```
❯ remember my project is a Go CLI tool

⚙ memory › set_context {"key": "project", "value": "Go CLI tool"}
↳ set_context [success]

❯ exit
```

**Session 2 (same config):**

```
❯ what was my project about?

⚙ memory › get_context {}
↳ get_context [{"key": "project", "value": "Go CLI tool"}]

assistant: Your project is a Go CLI tool.
```

## Example 4: Using Claude Desktop Config

Reuse your existing Claude Desktop setup:

```bash
# macOS
mcp-setu --config ~/Library/Application\ Support/Claude/claude_desktop_config.json chat

# Or add to your shell profile for easy access
alias claude-setu='mcp-setu --config ~/Library/Application\ Support/Claude/claude_desktop_config.json'
```

## Example 5: Verbose Mode for Debugging

See all tool calls and results:

```bash
mcp-setu chat --verbose
```

Output:

```
❯ list files in current directory

Tool call:
{
  "name": "read_directory",
  "arguments": {"path": "."}
}

Tool result:
{
  "contents": [
    {"name": "main.go", "type": "file", "size": 2048},
    {"name": "config.json", "type": "file", "size": 512},
    {"name": "README.md", "type": "file", "size": 4096}
  ]
}

assistant: Here are the files in the current directory:
- main.go (2 KB)
- config.json (512 B)
- README.md (4 KB)
```

## Example 6: Model Switching

Switch between models during chat:

```bash
mcp-setu chat
```

**In chat:**

```
❯ /model

Current model: gemma4:e4b
Available models:
  ✓ gemma4:e4b (tool support)
  ✓ qwen2.5:7b (tool support)
  - llama2:7b (no tool support)
  ✓ mistral-nemo:12b (tool support)

❯ /model qwen2.5:7b

✓ Switched to model: qwen2.5:7b
```

## Example 7: Multi-Server Setup

Use multiple MCP servers together:

```json
{
  "ollama": {
    "model": "qwen2.5:7b"
  },
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "."]
    },
    "sqlite": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-sqlite", "./data.db"]
    },
    "memory": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-memory"]
    }
  }
}
```

**In chat:**

```
❯ read config.json and remember its database path

⚙ filesystem › read_file {"path": "config.json"}
↳ read_file {"database": "./data.db"}

⚙ memory › set_context {"key": "db_path", "value": "./data.db"}
↳ set_context [success]

assistant: I've read your config and saved the database path to memory.

❯ how many records are in the main table?

⚙ memory › get_context {}
↳ get_context [{"key": "db_path", "value": "./data.db"}]

⚙ sqlite › query {"sql": "SELECT COUNT(*) as count FROM main"}
↳ query [{"count": 1523}]

assistant: There are 1,523 records in the main table.
```

## Example 8: Performance Monitoring

Check performance metrics:

```bash
mcp-setu chat
```

**In chat (after several exchanges):**

```
❯ /stats

Performance Statistics
  Messages: 12
  Tool calls: 8
  Iterations: 5
  Session duration: 2m 34s
  Average response: 1.2s
  Longest response: 3.4s
```

## Tips & Tricks

1. **Override model temporarily** without changing config:
   ```bash
   mcp-setu chat --model llama3.3:70b
   ```

2. **Override system prompt**:
   ```bash
   mcp-setu chat --system "You are a Python expert"
   ```

3. **Validate before running**:
   ```bash
   mcp-setu validate
   ```

4. **List all tools available**:
   ```bash
   mcp-setu tools
   ```

5. **Use in scripts** (limited to one message):
   ```bash
   echo "what files are in ." | mcp-setu chat
   ```

## Next Steps

- **[Configuration](./configuration.md)** — More config options
- **[Troubleshooting](./troubleshooting.md)** — Solve issues
- **[CLI Reference](./cli/)** — Full command documentation
