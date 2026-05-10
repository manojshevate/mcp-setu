# Release Process for mcp-setu

This document describes how to create and publish a new release of mcp-setu.

## Release Types

- **MAJOR** (e.g., v1.0.0): Breaking changes or significant feature additions
- **MINOR** (e.g., v0.1.0): Backward-compatible feature additions
- **PATCH** (e.g., v0.1.1): Bug fixes and maintenance updates

## Pre-Release Checklist

### 1-2 Days Before Release

- [ ] Ensure all PRs are merged to main branch
- [ ] Review all commits since last release
- [ ] Check if documentation needs updates
- [ ] Verify all tests pass locally: `make test && make vet`

### Day of Release

- [ ] Test the build: `make build`
- [ ] Test the binary:
  ```bash
  ./bin/mcp-setu --help
  ./bin/mcp-setu version
  ./bin/mcp-setu validate
  ```
- [ ] Review release notes in `CHANGELOG.md`

## Release Steps

### Step 1: Update Version

Edit `internal/version/version.go` and update the `Version` constant:

```go
var (
	Version   = "0.1.0"  // Change from "0.1.0-dev"
	Commit    = "unknown"
	BuildDate = "unknown"
)
```

Commit this change:

```bash
git add internal/version/version.go
git commit -m "chore: bump version to v0.1.0"
git push origin main
```

### Step 2: Create Git Tag

Create an annotated tag with a brief description:

```bash
git tag -a v0.1.0 -m "Release v0.1.0 - Brief description of major features/fixes"
git push origin v0.1.0
```

### Step 3: Monitor GitHub Actions

The release workflow will automatically trigger when the tag is pushed:

1. Go to GitHub Actions: https://github.com/manojshevate/mcp-setu/actions
2. Select the "Release" workflow
3. Monitor the build process (typically 5-10 minutes)
4. Verify that:
   - All platform builds complete successfully
   - GitHub Release is created with artifacts
   - Checksums file is included

### Step 4: Verify Release

After the workflow completes, verify:

#### Check GitHub Release

- [ ] Release page created at: https://github.com/manojshevate/mcp-setu/releases/tag/v0.1.0
- [ ] All binary artifacts present:
  - `mcp-setu_v0.1.0_darwin_amd64.tar.gz`
  - `mcp-setu_v0.1.0_darwin_arm64.tar.gz`
  - `mcp-setu_v0.1.0_linux_amd64.tar.gz`
  - `mcp-setu_v0.1.0_linux_arm64.tar.gz`
  - `mcp-setu_v0.1.0_windows_amd64.zip`
  - `mcp-setu_v0.1.0_windows_arm64.zip`
  - `checksums.txt`

#### Test Installation Methods

```bash
# Test go install
go install github.com/manojshevate/mcp-setu/cmd/mcp-setu@v0.1.0
which mcp-setu
mcp-setu version

# Should output: mcp-setu version v0.1.0
```

```bash
# Test Homebrew (after formula is updated)
brew tap manojshevate/tap
brew install mcp-setu
mcp-setu version

# Uninstall after testing
brew uninstall mcp-setu
```

#### Verify Checksums

Download and verify binaries:

```bash
# Download checksums file
curl -L https://github.com/manojshevate/mcp-setu/releases/download/v0.1.0/checksums.txt -o checksums.txt

# Verify a binary
sha256sum -c checksums.txt | grep darwin_amd64
```

### Step 5: Post-Release

#### Update Development Version

After release, update the version for the next development cycle:

```bash
# Edit internal/version/version.go
# Change version to next development version, e.g., v0.2.0-dev
vim internal/version/version.go

git add internal/version/version.go
git commit -m "chore: prepare v0.2.0-dev"
git push origin main
```

#### Update Documentation

- [ ] Update README.md with the new version in installation instructions
- [ ] Add section to CHANGELOG.md for next development version
- [ ] Close any release-related GitHub issues

#### Announce Release

- [ ] Create a GitHub Discussion announcing the release
- [ ] Include highlights and link to full CHANGELOG
- [ ] Mention installation methods (go install, brew)

## Troubleshooting

### Build Fails in GitHub Actions

1. Check the action logs for specific error
2. Verify the issue locally: `make build`
3. Fix the issue and create a new commit
4. Delete the tag: `git tag -d v0.1.0 && git push origin --delete v0.1.0`
5. Create the tag again once fix is verified

### Homebrew Formula Doesn't Update

The formula update requires a separate token. Ensure:

1. The `homebrew-tap` repository exists at `manojshevate/homebrew-tap`
2. A GitHub token with repo permissions is set as `HOMEBREW_TAP_TOKEN` in secrets
3. The formula file structure matches the workflow expectations

If manual update is needed:

```bash
cd ~/homebrew-tap
# Edit Formula/mcp-setu.rb with new version and SHA256 values
git add Formula/mcp-setu.rb
git commit -m "chore: update mcp-setu formula to v0.1.0"
git push origin main
```

### Verify Workflow Secrets

Ensure the following secrets are configured in GitHub:

- `HOMEBREW_TAP_TOKEN` (optional, uses GITHUB_TOKEN as fallback)

Check at: https://github.com/manojshevate/mcp-setu/settings/secrets

## Release Checklist Template

Create a GitHub issue with this template before releasing:

```markdown
# Release v0.1.0

## Pre-Release
- [ ] All PRs merged to main
- [ ] Tests pass: `make test && make vet`
- [ ] Build successful: `make build`
- [ ] Documentation updated

## Release Day
- [ ] Version updated in `internal/version/version.go`
- [ ] Version commit pushed to main
- [ ] Git tag created and pushed
- [ ] GitHub Actions workflow completes
- [ ] Release artifacts verified
- [ ] Installation methods tested (go install, brew)

## Post-Release
- [ ] Development version bumped
- [ ] Release announced
- [ ] CHANGELOG updated

## Details
- Features/fixes: [List of major changes]
- Breaking changes: [If any]
- Known issues: [If any]
```

## Automation Details

The GitHub Actions workflows handle:

1. **test.yml**: Runs on every push/PR
   - Runs `go test ./...`
   - Runs `go vet ./...`
   - Builds the binary

2. **release.yml**: Triggered on version tags (v*)
   - Builds binaries for 6 platforms
   - Creates tarballs/zips
   - Generates SHA256 checksums
   - Creates GitHub Release with artifacts
   - Updates Homebrew formula (if non-dev release)

Both workflows are triggered automatically—no manual intervention needed beyond pushing the tag.
