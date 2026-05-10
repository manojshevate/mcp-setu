# Homebrew Tap for mcpgo

This is a custom Homebrew tap for distributing mcpgo on macOS.

## Installation

```bash
# Add the tap
brew tap manojshevate/tap

# Install mcpgo
brew install mcpgo

# Verify installation
mcpgo version
```

## Updating

```bash
# Update mcpgo
brew upgrade mcpgo
```

## Repository Structure

This repository should be set up as follows:

```
homebrew-tap/
├── Formula/
│   └── mcpgo.rb          # Main formula file
├── README.md             # This file
└── .gitignore           # Git ignore file
```

## Setup Instructions

To set up this repository:

1. Create a new GitHub repository: `manojshevate/homebrew-tap`
2. Clone this template
3. Push to GitHub
4. Update the GitHub Actions workflow in the main mcpgo repository to use this tap

## Formula Updates

The formula is automatically updated by the GitHub Actions workflow in the main mcpgo repository when a new release is created. The workflow:

1. Builds binaries for all platforms
2. Generates checksums
3. Updates the formula file with new version and SHA256 values
4. Commits and pushes changes to this repository

## Testing Formula

To test the formula locally before release:

```bash
# Install from local formula
brew install --verbose ./Formula/mcpgo.rb

# Uninstall for cleanup
brew uninstall mcpgo
```

## Manual Formula Updates

If needed, manually update the formula by:

1. Editing `Formula/mcpgo.rb`
2. Updating the version number
3. Downloading the binary and calculating SHA256:
   ```bash
   sha256sum mcpgo_v0.1.0_darwin_amd64.tar.gz
   ```
4. Updating the sha256 values in the formula
5. Committing and pushing changes
