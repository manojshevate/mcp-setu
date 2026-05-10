# Release Process for mcpgo

This document describes how to create and publish a new release of mcpgo.

## Release Types

- **MAJOR** (e.g., v1.0.0): Breaking changes or significant feature additions
- **MINOR** (e.g., v0.1.0): Backward-compatible feature additions
- **PATCH** (e.g., v0.1.1): Bug fixes and maintenance updates

## Installation Methods (For Users)

### Via go install
```bash
# Latest version
go install github.com/manojshevate/mcpgo/cmd/mcpgo@latest

# Specific version
go install github.com/manojshevate/mcpgo/cmd/mcpgo@v0.1.0
```

### Via Homebrew (macOS)
```bash
# Add tap
brew tap manojshevate/tap

# Install
brew install mcpgo

# Verify
mcpgo version

# Upgrade
brew upgrade mcpgo
```

### Manual Download
Download binaries from GitHub Releases: https://github.com/manojshevate/mcpgo/releases

## How to Release (For Maintainers)

### Pre-Release Checklist

1-2 days before release:
- [ ] All PRs merged to main
- [ ] All tests pass: `go test ./...`
- [ ] Build works: `make build`
- [ ] Test the binary: `./bin/mcpgo version`

### Release Steps

**Step 1: Update Version**

Edit `internal/version/version.go`:
```go
var (
	Version   = "0.1.0"  // Change from "0.1.0-dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)
```

Commit:
```bash
git add internal/version/version.go
git commit -m "chore: bump version to v0.1.0"
git push origin main
```

**Step 2: Create Git Tag**

```bash
git tag -a v0.1.0 -m "Release v0.1.0 - Description of major changes"
git push origin v0.1.0
```

**Step 3: GitHub Actions Automatic Release**

The workflow automatically:
1. ✅ Builds binaries for all platforms (Linux, macOS, Windows)
2. ✅ Creates GitHub Release with artifacts
3. ✅ Generates SHA256 checksums
4. ✅ Updates Homebrew formula

Monitor at: https://github.com/manojshevate/mcpgo/actions

**Step 4: Verify Release**

Check at: https://github.com/manojshevate/mcpgo/releases/tag/v0.1.0

Verify:
- [ ] All platform binaries present
- [ ] checksums.txt file included
- [ ] Homebrew formula updated (after ~5-10 minutes)

**Step 5: Post-Release**

Update development version:
```bash
# Edit internal/version/version.go
# Change to: Version = "0.2.0-dev"
git add internal/version/version.go
git commit -m "chore: prepare v0.2.0-dev"
git push origin main
```

## Before First Release

### Setup Homebrew Tap

1. Create new repository: `github.com/manojshevate/homebrew-tap`
2. Copy this template structure:
   ```
   Formula/
   └── mcpgo.rb
   README.md
   .gitignore
   ```
3. Create GitHub secret `HOMEBREW_TAP_TOKEN` with repo access

### Template Formula

Create `Formula/mcpgo.rb`:
```ruby
class Mcpgo < Formula
  desc "MCP bridge for Ollama"
  homepage "https://github.com/manojshevate/mcpgo"

  on_macos do
    on_arm do
      url "https://github.com/manojshevate/mcpgo/releases/download/v0.1.0/mcpgo_v0.1.0_darwin_arm64.tar.gz"
      sha256 "PLACEHOLDER_ARM64_SHA256"
    end
    on_intel do
      url "https://github.com/manojshevate/mcpgo/releases/download/v0.1.0/mcpgo_v0.1.0_darwin_amd64.tar.gz"
      sha256 "PLACEHOLDER_AMD64_SHA256"
    end
  end

  version "0.1.0"
  license "MIT"

  def install
    bin.install "mcpgo"
  end

  def test
    assert_match "mcpgo version", shell_output("#{bin}/mcpgo version")
  end
end
```

The workflow updates this automatically on each release.

## Complete Release Flow Summary

```
1. Update version in code
   ↓
2. Commit version change
   ↓
3. Push git tag (v0.1.0)
   ↓
4. GitHub Actions automatically:
   - Builds all binaries
   - Creates GitHub Release
   - Updates Homebrew formula
   ↓
5. Users can install via:
   - go install
   - brew install
   - Manual download from GitHub
```

## Troubleshooting

**Build fails?**
- Check GitHub Actions logs
- Verify locally: `make build`

**Homebrew formula doesn't update?**
- Ensure HOMEBREW_TAP_TOKEN secret is set
- Check workflow logs in Actions tab

**Release doesn't appear?**
- Wait ~2 minutes for Actions to trigger
- Check that tag format is `v*` (e.g., v0.1.0)

## Testing Release Locally

To test the release process without publishing:

```bash
# Test building with custom version
VERSION=v1.0.0-test make build
./bin/mcpgo version

# Test formula locally
brew install --verbose ./homebrew-tap/Formula/mcpgo.rb
```
