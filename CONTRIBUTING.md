# Contributing to mcp-setu

Thank you for considering contributing to mcp-setu! This document provides guidelines and instructions for contributing.

## Code of Conduct

- Be respectful and inclusive
- Focus on constructive feedback
- Help create a welcoming community for all skill levels

## Getting Started

### Prerequisites

- Go 1.26+ ([download](https://golang.org/dl/))
- Ollama ([download](https://ollama.com))
- Node.js 18+ (for testing MCP servers)
- Git

### Setup Development Environment

```bash
# Clone the repository
git clone https://github.com/manojshevate/mcp-setu
cd mcp-setu

# Build the binary
make build

# Run tests and checks
make vet
make test  # when available
```

## Development Workflow

### Making Changes

1. **Create a feature branch**
   ```bash
   git checkout -b feature/your-feature-name
   ```

2. **Make your changes**
   - Follow the existing code style
   - Keep commits atomic and logical
   - Write clear commit messages

3. **Run quality checks**
   ```bash
   go vet ./...
   go test ./...  # when available
   make build
   ```

4. **Validate configuration**
   ```bash
   ./bin/mcp-setu validate
   ```

5. **Test your changes**
   ```bash
   # Test chat functionality
   ./bin/mcp-setu chat
   
   # Test with verbose output
   ./bin/mcp-setu chat --verbose
   
   # Test tool discovery
   ./bin/mcp-setu tools
   ```

### Code Style

- Follow standard Go conventions (formatted with `gofmt`)
- Keep functions small and focused
- Use meaningful variable and function names
- Add godoc comments for exported types and functions
- Wrap errors with context using `fmt.Errorf`

### Error Handling

Always wrap errors to provide context:

```go
// Good
return fmt.Errorf("failed to load config: %w", err)

// Avoid
return err
```

### Concurrency

- Use `sync.Mutex` for protecting shared state
- Always unlock with `defer`
- Use `context.Context` for cancellation and timeouts
- Avoid global state

## Testing

When adding features or fixing bugs:

1. Write tests for new functionality
2. Test edge cases (empty input, large payloads, timeouts, etc.)
3. Ensure `go vet ./...` passes
4. Test with actual Ollama and MCP servers if possible

## Commit Messages

Write clear, descriptive commit messages:

```
feat: Add new feature description

Detailed explanation of what this change does and why.
Include any relevant context or links to issues.

Fixes #123
```

### Commit Message Format

- `feat:` - New feature
- `fix:` - Bug fix
- `docs:` - Documentation only
- `refactor:` - Code reorganization (no functional change)
- `perf:` - Performance improvement
- `test:` - Test additions or fixes
- `ci:` - CI/CD changes
- `chore:` - Build system, dependencies, etc.

## Pull Request Process

1. **Update documentation** if needed (README.md, inline comments)
2. **Test thoroughly** with `make vet` and `make build`
3. **Keep commits clean** - squash WIP commits before submitting
4. **Write a clear PR description**
   - Explain the problem being solved
   - Describe the solution
   - List any breaking changes
   - Reference related issues

5. **Be responsive to feedback** - engage with reviewers constructively

## Areas to Contribute

### High Priority

- **Bug fixes** — Check [open issues](https://github.com/manojshevate/mcp-setu/issues)
- **Performance** — Profile and optimize hot paths
- **Documentation** — Expand examples, clarify complex sections
- **Error messages** — Make them more helpful and actionable

### Good First Issues

- Adding more model prefixes to `KnownToolSupportedModels`
- Improving error messages
- Expanding documentation
- Adding configuration examples

### Advanced Contributions

- Streaming response support
- Parallel tool execution
- Conversation history persistence
- Additional MCP protocol features

## Testing Checklist

Before submitting a PR:

- [ ] Code builds without errors: `make build`
- [ ] Vet checks pass: `make vet`
- [ ] Tested with Ollama running locally
- [ ] Tested with at least one MCP server
- [ ] Verbose mode shows expected output: `./bin/mcp-setu chat --verbose`
- [ ] Validation works: `./bin/mcp-setu validate`
- [ ] Help text is readable: `./bin/mcp-setu --help`

## Documentation Guidelines

### README.md

- Keep it concise but comprehensive
- Include code examples
- Update when changing user-facing behavior
- Keep prerequisites and installation current

### Code Comments

- Explain *why*, not *what*
- For godoc: one-line summary starting with the name
- Use inline comments sparingly for non-obvious logic
- Link to relevant issues or RFCs when appropriate

Example:

```go
// SendMessage sends a message with retries on transient failures.
func (c *Client) SendMessage(ctx context.Context, msg string) error {
```

## Reporting Issues

### Bug Reports

Include:
- Go version: `go version`
- OS and architecture
- Steps to reproduce
- Expected vs actual behavior
- Relevant error messages or logs
- Output of `mcp-setu validate`

### Feature Requests

Include:
- Clear problem statement
- Proposed solution
- Alternative approaches considered
- Example use cases

## Questions?

- Check existing issues and PRs
- Open a discussion on GitHub
- Read the code — it's well-commented!

## License

By contributing, you agree that your contributions will be licensed under the MIT License (see [LICENSE](LICENSE)).

---

**Thank you for contributing to mcp-setu!** 🎉
