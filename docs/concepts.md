# Concepts

Understanding how mcp-setu works.

## The Agentic Loop

`mcp-setu` runs an agentic loop that bridges Ollama and MCP:

```
User Input
    ↓
[Bridge Loop]
    ↓
Send message + tool definitions to Ollama
    ↓
Ollama decides: "I'll call this tool"
    ↓
Execute tool via MCP (get result)
    ↓
Send result back to Ollama
    ↓
Repeat until Ollama says "I'm done"
    ↓
Display final response to user
```

The loop exits when:
- Ollama stops requesting tools (final response ready)
- Max iterations reached (default: 20, prevents infinite loops)
- An error occurs

## Models & Tool Calling

Not all language models support tool calling. mcp-setu requires a model that understands the JSON-based tool definition format and can generate structured tool calls.

**Supported models:**
- ✅ Gemma 4 & 3
- ✅ Qwen 2.5 & 3
- ✅ Llama 3.2 & 3.3
- ✅ Mistral Nemo
- ✅ Command R
- ✅ Phi 4
- ✅ DeepSeek R1

**Unsupported models:**
- ❌ Llama 2
- ❌ Older Qwen versions
- ❌ Mistral 7B

When you run mcp-setu, it checks tool support at startup and exits with a clear error if your model doesn't support it.

## MCP Servers

MCP (Model Context Protocol) servers are processes that provide tools to the model.

### How MCP Servers Work

1. **mcp-setu spawns a server process** (or connects to HTTP endpoint)
2. **Server advertises tools** via JSON-RPC (or HTTP API)
3. **Model calls a tool**
4. **mcp-setu routes the call** to the appropriate server
5. **Server executes and returns result**
6. **mcp-setu sends result back** to model
7. **Loop continues** until model is satisfied

### Transport Mechanisms

`mcp-setu` supports three ways to talk to MCP servers:

#### Stdio (JSON-RPC 2.0)

Servers run as subprocesses. Communication via stdin/stdout.

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

**Best for:** Local Node.js servers, development, reliable communication.

#### HTTP Streamable (Modern Standard)

Remote servers accessed via HTTP POST with streaming.

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

**Best for:** Cloud-hosted servers, production, remote integration.

#### HTTP/SSE (Legacy)

Server-Sent Events based communication (deprecated).

```json
{
  "mcpServers": {
    "legacy": {
      "type": "http-sse",
      "url": "http://legacy-server.com/events"
    }
  }
}
```

**Best for:** Legacy server compatibility.

## Configuration Structure

```
mcp.json
├── ollama (Ollama settings)
│   ├── baseUrl        → Where Ollama runs
│   ├── model          → Which model to use
│   ├── systemPrompt   → System message
│   ├── temperature    → Creativity (0=deterministic, 1=creative)
│   └── contextLength  → How much history to keep
│
└── mcpServers (Available tools)
    ├── filesystem     → File operations
    ├── sqlite         → Database queries
    ├── memory         → Persistent context
    └── [custom]       → Your servers
```

## Tool Definition Format

Tools are sent to Ollama in this format:

```json
{
  "name": "read_file",
  "description": "Read the contents of a file",
  "inputSchema": {
    "type": "object",
    "properties": {
      "path": {
        "type": "string",
        "description": "Path to the file"
      }
    },
    "required": ["path"]
  }
}
```

The model sees these definitions and knows:
1. What tools are available
2. What each tool does
3. What arguments each tool takes
4. What types of values are required

## Workflow Example

Here's a real execution flow:

```
User: "How many lines are in main.go?"

┌─ mcp-setu ─────────────────────────────┐
│ Sends to Ollama (tool definitions included):
│ - User message: "How many lines..."
│ - Available tools: read_file, etc.
└──────────────────────────────────────┘
                    ↓
        Ollama processes & decides
                    ↓
        "I'll call read_file with path='main.go'"
                    ↓
┌─ mcp-setu ─────────────────────────────┐
│ Routes to filesystem server:
│ → Call: read_file("main.go")
│ ← Result: [file contents...]
└──────────────────────────────────────┘
                    ↓
        Sends result back to Ollama
                    ↓
        "The file has 477 lines"
                    ↓
┌─ mcp-setu ─────────────────────────────┐
│ Checks: Is model done? Yes.
│ Sends response to user
└──────────────────────────────────────┘
                    ↓
        User sees: "main.go has 477 lines"
```

## Performance Considerations

### Response Times

Three main factors:

1. **Model inference time** — How long the model takes to generate a response (usually the slowest)
2. **Tool execution time** — How long tools take to run (filesystem, database queries)
3. **Network latency** — For HTTP-based MCP servers

### Optimization Tips

1. **Use GPU** — Run Ollama with GPU acceleration (100x faster)
2. **Smaller models** — gemma2:latest is much faster than llama3.3:70b
3. **Shorter context** — Reduce `contextLength` if responses are slow
4. **Cache results** — If tools return the same data, the model can reuse answers
5. **Parallel tools** — Independent tool calls run in parallel

### Monitoring

Use `/stats` in chat to monitor:

```
/stats

Performance Statistics
  Messages: 12
  Tool calls: 8
  Iterations: 4
  Session duration: 2m 34s
  Average response time: 1.2s
```

## Security

`mcp-setu` runs entirely on your machine:

1. **No cloud** — Everything stays local
2. **No data sharing** — Models run locally
3. **Credentials in env vars** — Not stored in config files
4. **OAuth 2.1 support** — For remote servers

### Best Practices

- Store API tokens in environment variables
- Use HTTPS/TLS for remote servers
- Validate tool results before trusting them
- Don't share your mcp.json with secrets

## Next Steps

- **[Configuration](./configuration.md)** — Deep dive into settings
- **[Examples](./examples.md)** — Real-world patterns
- **[Development](./development.md)** — Build on mcp-setu
