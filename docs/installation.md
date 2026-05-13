# Installation

Choose your installation method below.

## Recommended: go install

```bash
go install github.com/manojshevate/mcp-setu/cmd/mcp-setu@latest
```

This installs the latest released version to `$GOPATH/bin/mcp-setu` (usually `~/go/bin`).

**Verify:**

```bash
mcp-setu version
```

**For a specific version:**

```bash
go install github.com/manojshevate/mcp-setu/cmd/mcp-setu@v0.1.1
```

## Homebrew (macOS)

```bash
brew tap manojshevate/mcp-setu
brew install mcp-setu
```

**To upgrade:**

```bash
brew upgrade mcp-setu
```

## Building from Source

Requires **Go 1.26+**.

```bash
git clone https://github.com/manojshevate/mcp-setu
cd mcp-setu
make install
```

Or build and run locally without installing:

```bash
make build
./bin/mcp-setu chat
```

## Verify Installation

After any installation method:

```bash
mcp-setu version
```

You should see output like:

```
mcp-setu version v0.1.1
commit: abc1234
build date: 2025-05-10T12:00:00Z
```

## Troubleshooting Installation

**"mcp-setu: command not found"**

1. Check `$GOPATH/bin` is in your `$PATH`:

   ```bash
   echo $PATH
   ```

2. If missing, add it to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.):

   ```bash
   export PATH="$HOME/go/bin:$PATH"
   ```

3. Reload your shell:

   ```bash
   source ~/.bashrc  # or source ~/.zshrc
   ```

**Build from source fails with "go: unknown version"**

Update Go to 1.26+:

```bash
brew install go  # macOS
# or download from https://golang.org/dl/
```

**Homebrew tap not found**

Ensure the tap is added:

```bash
brew tap manojshevate/mcp-setu
brew install mcp-setu
```

## Next Steps

1. **[Getting Started](./getting-started.md)** — Run your first chat
2. **[Configuration](./configuration.md)** — Set up MCP servers
3. **[Examples](./examples.md)** — See real-world usage
