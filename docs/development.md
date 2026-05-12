# Development

Contributing to mcp-setu and building on top of it.

## Setup Development Environment

### Prerequisites

- **Go 1.26+** — [Download](https://golang.org/dl/)
- **Git**
- **Node.js 18+** — For testing MCP servers

### Clone & Build

```bash
git clone https://github.com/manojshevate/mcp-setu
cd mcp-setu
make build
./bin/mcp-setu chat
```

## Project Structure

```
mcp-setu/
├── cmd/mcp-setu/           # CLI entry point
│   └── main.go             # Cobra commands
│
├── internal/
│   ├── bridge/             # Agentic loop orchestrator
│   │   ├── bridge.go       # Defines the Printer interface used by the loop
│   │   ├── bridge_test.go
│   │   └── stats.go
│   │
│   ├── config/             # Config loading & validation
│   │   ├── config.go
│   │   └── config_test.go
│   │
│   ├── mcp/                # MCP client (JSON-RPC 2.0)
│   │   ├── client.go
│   │   ├── client_test.go
│   │   ├── types.go
│   │   ├── http_client.go
│   │   └── auth.go
│   │
│   ├── ollama/             # Ollama HTTP client
│   │   ├── client.go
│   │   ├── client_test.go
│   │   └── types.go
│   │
│   ├── tui/                # Bubble Tea chat TUI
│   │   ├── runner.go       # Entry point: builds SessionModel and runs tea
│   │   ├── session.go      # Root model: layout, event loop, slash commands
│   │   ├── input.go        # Input field with history & autocomplete
│   │   └── printer.go      # bridge.Printer impl that emits TUI events
│   │
│   ├── ui/                 # Terminal output formatting (non-TUI commands)
│   │   ├── printer.go
│   │   └── printer_test.go
│   │
│   ├── version/            # Version info
│   │   └── version.go
│   │
│   └── tools/              # Utility tools
│       └── docgen/         # CLI doc generation
│           └── main.go
│
├── Makefile                # Build automation
├── go.mod & go.sum         # Go module
├── mcp.json                # Example config
├── LICENSE                 # MIT License
└── README.md               # User docs
```

## Making Changes

### Code Style

- Follow Go conventions
- Use `go fmt` for formatting
- Run `go vet` before committing
- Aim for 80%+ test coverage

### Running Tests

```bash
make test          # Run all tests
make test-v        # Verbose output
make coverage      # Coverage report for bridge
```

### Building & Testing

```bash
make build         # Build binary
make run           # Run in dev mode
make run-verbose   # Run with debug output
make vet           # Check for issues
make lint          # Lint with golangci-lint (if installed)
```

## Adding Features

### Example: Adding a New Command

1. **Create the command function** in `cmd/mcp-setu/main.go`:

```go
newCmd := &cobra.Command{
    Use:   "newcommand",
    Short: "Description",
    RunE: func(cmd *cobra.Command, args []string) error {
        // Your logic here
        return nil
    },
}
rootCmd.AddCommand(newCmd)
```

2. **Test it:**

```bash
make build
./bin/mcp-setu newcommand
```

3. **Update docs** — Edit relevant markdown files

4. **Add tests** — Create `cmd/mcp-setu/newcommand_test.go`

### Example: Adding MCP Server Support

1. **Update config parsing** in `internal/config/config.go`
2. **Add server initialization** in `internal/mcp/client.go`
3. **Add tests** for new server type
4. **Update configuration docs** in `docs/configuration.md`

## Testing

### Unit Tests

```bash
go test ./internal/bridge
go test ./internal/config
```

### Integration Tests

Test with real Ollama and MCP servers:

```bash
make run-validate  # Tests Ollama connectivity
./bin/mcp-setu tools  # Tests MCP server startup
```

### Test Coverage

```bash
make coverage      # See coverage report
```

Target: **80%+** for critical packages like `bridge` and `config`.

## Documentation

### User Documentation

Located in `docs/`:

- `docs/index.md` — Home page
- `docs/getting-started.md` — Quick start
- `docs/installation.md` — Install instructions
- `docs/configuration.md` — Config reference
- `docs/examples.md` — Usage examples
- `docs/troubleshooting.md` — Common issues
- `docs/concepts.md` — How it works

### CLI Reference

Generated automatically from Cobra commands:

```bash
npm run docs:generate-cli
```

This creates docs under `docs/cli/`.

## Committing Changes

Follow conventional commits:

```bash
git add .
git commit -m "feat: add new command" -m "Description of changes"
```

Types:
- `feat:` — New feature
- `fix:` — Bug fix
- `docs:` — Documentation
- `test:` — Tests
- `refactor:` — Code restructuring
- `perf:` — Performance improvement

## Building & Releasing

### Building from Source

```bash
make build
```

### Release Builds

```bash
make release-build
```

This creates binaries for all platforms in `bin/releases/`.

### Publishing a Release

See [RELEASE.md](../RELEASE.md) for the full release checklist.

## Extending mcp-setu

### Custom MCP Servers

Write your own MCP server and add it to your config:

```json
{
  "mcpServers": {
    "my-server": {
      "command": "/path/to/my-server",
      "args": ["--config", "config.json"]
    }
  }
}
```

Your server must:
1. Speak JSON-RPC 2.0 on stdin/stdout
2. Implement the MCP protocol
3. Return tool definitions and execute tool calls

See [MCP Spec](https://modelcontextprotocol.io/) for details.

### Custom UI Output

Modify `internal/ui/printer.go` to customize terminal output:

```go
// Add a new printer method
func (p *Printer) PrintCustom(msg string) {
    // Use lipgloss for styling
}
```

### Building Tools on Top

Use the internal packages as a library:

```go
package main

import (
    "github.com/manojshevate/mcp-setu/internal/bridge"
    "github.com/manojshevate/mcp-setu/internal/config"
)

func main() {
    cfg, _ := config.Load("mcp.json")
    br := bridge.NewBridge(...)
    // Use bridge for your application
}
```

## Performance Profiling

### CPU Profiling

```bash
go test -cpuprofile=cpu.prof ./internal/bridge
go tool pprof cpu.prof
```

### Memory Profiling

```bash
go test -memprofile=mem.prof ./internal/bridge
go tool pprof mem.prof
```

## Debugging

### Enable Debug Logging

```bash
GOFLAGS="-v" make build
```

### Verbose Mode

```bash
./bin/mcp-setu chat --verbose
```

Shows all tool calls and LLM iterations inline in the chat TUI's output panel.

### Testing with Custom Config

```bash
./bin/mcp-setu --config ./test-config.json validate
```

## Contributing

1. **Fork the repository** — https://github.com/manojshevate/mcp-setu/fork
2. **Create a feature branch** — `git checkout -b feature/my-feature`
3. **Make changes** and commit with conventional commits
4. **Push to your fork** — `git push origin feature/my-feature`
5. **Open a Pull Request** — Describe your changes clearly

### Before Submitting

```bash
make vet                # Check for issues
make test               # Run tests
mcp-setu validate       # Test with real config
```

### PR Guidelines

- Clear description of what & why
- Reference related issues
- Add tests for new features
- Update docs if user-facing
- Keep commits clean and logical

## Resources

- **[MCP Specification](https://modelcontextprotocol.io/)** — Protocol details
- **[Go Documentation](https://golang.org/doc/)** — Language reference
- **[Cobra CLI Framework](https://cobra.dev/)** — Command documentation
- **[Bubble Tea](https://github.com/charmbracelet/bubbletea)** — Chat TUI framework
- **[Lipgloss](https://github.com/charmbracelet/lipgloss)** — Terminal styling

## Getting Help

- **Questions?** Start a discussion on GitHub
- **Found a bug?** Open an issue with details
- **Want to contribute?** See [CONTRIBUTING.md](../CONTRIBUTING.md)

## Future Work

See [README.md](../README.md#limitations--future-work) for planned features and enhancements.
