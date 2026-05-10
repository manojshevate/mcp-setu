# Homebrew Tap for mcp-setu

This is a custom Homebrew tap for distributing mcp-setu on macOS.

## Installation

```bash
# Add the tap
brew tap manojshevate/tap

# Install mcp-setu
brew install mcp-setu

# Verify installation
mcp-setu version
```

## Updating

```bash
# Update mcp-setu
brew upgrade mcp-setu
```

## Repository Structure

This repository should be set up as follows:

```
homebrew-tap/
├── Formula/
│   └── mcp-setu.rb          # Main formula file
├── README.md             # This file
└── .gitignore           # Git ignore file
```

## Setup Instructions

To set up this repository:

1. Create a new GitHub repository: `manojshevate/homebrew-tap`
2. Clone this template
3. Push to GitHub
4. Update the GitHub Actions workflow in the main mcp-setu repository to use this tap

## Formula Updates

The formula is automatically updated by the GitHub Actions workflow in the main mcp-setu repository when a new release is created. The workflow:

1. Builds binaries for all platforms
2. Generates checksums
3. Updates the formula file with new version and SHA256 values
4. Commits and pushes changes to this repository

## Testing Formula

To test the formula locally before release:

```bash
# Install from local formula
brew install --verbose ./Formula/mcp-setu.rb

# Uninstall for cleanup
brew uninstall mcp-setu
```

## Manual Formula Updates

If needed, manually update the formula by:

1. Editing `Formula/mcp-setu.rb`
2. Updating the version number
3. Downloading the binary and calculating SHA256:
   ```bash
   sha256sum mcp-setu_v0.1.0_darwin_amd64.tar.gz
   ```
4. Updating the sha256 values in the formula
5. Committing and pushing changes
