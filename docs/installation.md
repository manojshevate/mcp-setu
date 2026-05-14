# Installation

Choose your installation method below.

## Quick Install (Recommended)

### macOS / Linux

Copy and paste this command into your terminal:

```bash
curl -fsSL https://raw.githubusercontent.com/manojshevate/mcp-setu/main/scripts/install.sh | sh
```

This downloads the pre-built binary and installs it to `~/.mcp-setu/bin` (or `~/.local/bin` if it exists).

**Version pinning (install a specific release):**

```bash
MCP_SETU_VERSION=v0.1.2 curl -fsSL https://raw.githubusercontent.com/manojshevate/mcp-setu/main/scripts/install.sh | sh
```

### Windows (PowerShell 5.0+)

Open PowerShell and paste:

```powershell
irm https://raw.githubusercontent.com/manojshevate/mcp-setu/main/scripts/install.ps1 | iex
```

This downloads the pre-built binary and installs it to `%LOCALAPPDATA%\mcp-setu\bin`.

**Version pinning (install a specific release):**

```powershell
$env:MCP_SETU_VERSION='v0.1.2'; irm https://raw.githubusercontent.com/manojshevate/mcp-setu/main/scripts/install.ps1 | iex
```

## Installation Directory

The installer automatically detects the best location:

**On Linux/macOS:**
- `~/.local/bin` (if it exists and is in `$PATH`)
- `~/.mcp-setu/bin` (created if needed)

You can override with an environment variable:

```bash
MCP_SETU_INSTALL_DIR=/custom/path curl -fsSL https://raw.githubusercontent.com/manojshevate/mcp-setu/main/scripts/install.sh | sh
```

**On Windows:**
- `%LOCALAPPDATA%\mcp-setu\bin` (typically `C:\Users\<YourUser>\AppData\Local\mcp-setu\bin`)

You can override with an environment variable:

```powershell
$env:MCP_SETU_INSTALL_DIR='C:\custom\path'; irm https://raw.githubusercontent.com/manojshevate/mcp-setu/main/scripts/install.ps1 | iex
```

## Environment Variables

| Variable | Description | Example |
|----------|-------------|---------|
| `MCP_SETU_VERSION` | Pin a specific release version | `v0.1.2` |
| `MCP_SETU_INSTALL_DIR` | Custom installation directory | `~/.local/bin` |
| `MCP_SETU_FORCE` | Skip confirmation, overwrite if already installed | `1` |

## Alternative Installation Methods

### go install

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

### Homebrew (macOS)

```bash
brew tap manojshevate/mcp-setu
brew install mcp-setu
```

**To upgrade:**

```bash
brew upgrade mcp-setu
```

### Building from Source

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

### General

**"mcp-setu: command not found"**

The installation directory is not in your `$PATH`. The installer should print instructions on how to add it.

1. Check what's in your `$PATH`:
   ```bash
   echo $PATH
   ```

2. If you used the install.sh script, add the install directory to your shell profile (`~/.bashrc`, `~/.zshrc`, etc.):
   ```bash
   export PATH="$HOME/.mcp-setu/bin:$PATH"
   ```

3. Reload your shell:
   ```bash
   source ~/.bashrc  # or source ~/.zshrc for zsh
   ```

### install.sh / install.ps1 Specific Issues

**"curl: command not found"**

Ensure curl is installed (should be default on modern macOS/Linux):

```bash
# macOS
brew install curl

# Ubuntu/Debian
sudo apt install curl

# Fedora/RHEL
sudo dnf install curl
```

**"Checksum verification failed"**

Your download may be corrupted. Try again:

```bash
MCP_SETU_FORCE=1 curl -fsSL https://raw.githubusercontent.com/manojshevate/mcp-setu/main/scripts/install.sh | sh
```

**PATH not modified / PATH modification didn't take effect**

The install.sh script only modifies shell rc files in interactive mode (not in CI environments). If running in CI or non-interactive contexts:

1. Manually add the install directory to your `$PATH`
2. Or use `go install` or `brew install` instead

**"PATH not added" after installation**

The installer won't modify rc files in non-interactive environments. You have two options:

1. Add manually to your shell profile:
   ```bash
   export PATH="$HOME/.mcp-setu/bin:$PATH"
   ```

2. Use `go install` instead:
   ```bash
   go install github.com/manojshevate/mcp-setu/cmd/mcp-setu@latest
   ```

**Windows: "cannot be loaded because running scripts is disabled"**

Windows PowerShell may block unsigned scripts. Run:

```powershell
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

Or use PowerShell as Administrator if that doesn't work.

**Windows: "The term 'irm' is not recognized"**

You need PowerShell 5.1+ (Windows 10/11 have this by default). Check your version:

```powershell
$PSVersionTable.PSVersion
```

If you're on an older version, use one of the alternative installation methods.

### go install Issues

**Build from source fails with "go: unknown version"**

Update Go to 1.26+:

```bash
brew install go  # macOS
# or download from https://golang.org/dl/
```

### Homebrew Issues

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
