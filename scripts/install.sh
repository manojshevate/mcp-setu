#!/bin/sh
# mcp-setu installer for macOS and Linux
# Usage: curl -fsSL https://raw.githubusercontent.com/manojshevate/mcp-setu/main/scripts/install.sh | sh
# Or with version pinning: MCP_SETU_VERSION=v0.1.2 curl -fsSL ... | sh

# Wrap script in main function so a truncated download doesn't execute partial code
main() {
	set -eu

	# Detect OS and architecture
	OS="$(uname -s)"
	ARCH="$(uname -m)"

	# Map architecture names
	case "$ARCH" in
		x86_64)
			ARCH="amd64"
			;;
		aarch64 | arm64)
			ARCH="arm64"
			;;
		*)
			error "Unsupported architecture: $ARCH"
			;;
	esac

	# Check for Windows (MINGW/MSYS) and redirect to PowerShell
	case "$OS" in
		MINGW* | MSYS*)
			warning "Detected Windows (MINGW/MSYS). Please use the PowerShell installer:"
			echo "irm https://raw.githubusercontent.com/manojshevate/mcp-setu/main/scripts/install.ps1 | iex"
			exit 1
			;;
	esac

	# Validate OS
	case "$OS" in
		Darwin | Linux)
			;;
		*)
			error "Unsupported OS: $OS. Please use macOS, Linux, or see manual installation options."
			;;
	esac

	# Setup color output (gracefully degrade if tput not available)
	if [ -t 2 ]; then
		RED="$(tput bold 2>/dev/null || true; tput setaf 1 2>/dev/null || true)"
		GREEN="$(tput setaf 2 2>/dev/null || true)"
		PLAIN="$(tput sgr0 2>/dev/null || true)"
	else
		RED=""
		GREEN=""
		PLAIN=""
	fi

	# Helper functions
	status() {
		echo "${GREEN}>>>${PLAIN} $*" >&2
	}

	error() {
		echo "${RED}ERROR:${PLAIN} $*" >&2
		exit 1
	}

	warning() {
		echo "${RED}WARNING:${PLAIN} $*" >&2
	}

	# Create temporary directory and cleanup on exit
	TEMP_DIR=$(mktemp -d)
	cleanup() {
		rm -rf "$TEMP_DIR"
	}
	trap cleanup EXIT

	# Resolve version
	VERSION="${MCP_SETU_VERSION:-}"
	if [ -z "$VERSION" ]; then
		status "Fetching latest release version..."
		# Try to fetch from GitHub API
		if VERSION=$(curl -fsSL "https://api.github.com/repos/manojshevate/mcp-setu/releases/latest" 2>/dev/null | grep -o '"tag_name":"[^"]*' | cut -d'"' -f4 | head -1); then
			if [ -z "$VERSION" ]; then
				# Fallback version if API parsing fails
				VERSION="v0.1.1"
				warning "Could not fetch latest version, using fallback: $VERSION"
			fi
		else
			VERSION="v0.1.1"
			warning "Could not reach GitHub API, using fallback version: $VERSION"
		fi
	fi

	status "Installing mcp-setu $VERSION for $OS/$ARCH"

	# Resolve install directory
	INSTALL_DIR="${MCP_SETU_INSTALL_DIR:-}"

	if [ -z "$INSTALL_DIR" ]; then
		# Check for ~/.local/bin (common on Linux)
		if [ -d "$HOME/.local/bin" ] && echo "$PATH" | grep -q "$HOME/.local/bin"; then
			INSTALL_DIR="$HOME/.local/bin"
		else
			INSTALL_DIR="$HOME/.mcp-setu/bin"
		fi
	fi

	# Create install directory if it doesn't exist
	if [ ! -d "$INSTALL_DIR" ]; then
		status "Creating installation directory: $INSTALL_DIR"
		mkdir -p "$INSTALL_DIR"
	fi

	# Check if binary already exists
	if [ -f "$INSTALL_DIR/mcp-setu" ]; then
		if [ -z "${MCP_SETU_FORCE:-}" ] && [ -t 0 ]; then
			# Interactive mode - prompt user
			printf "mcp-setu already exists at $INSTALL_DIR/mcp-setu. Overwrite? [y/N] " >&2
			read -r RESPONSE
			case "$RESPONSE" in
				y | Y)
					;;
				*)
					error "Installation cancelled"
					;;
			esac
		elif [ -z "${MCP_SETU_FORCE:-}" ]; then
			# Non-interactive mode without force flag - skip silently
			status "mcp-setu already installed, skipping"
			exit 0
		fi
	fi

	# Download binary and checksums
	BINARY_NAME="mcp-setu_${VERSION}_${OS}_${ARCH}.tar.gz"
	DOWNLOAD_URL="https://github.com/manojshevate/mcp-setu/releases/download/${VERSION}/${BINARY_NAME}"
	CHECKSUMS_URL="https://github.com/manojshevate/mcp-setu/releases/download/${VERSION}/checksums.txt"

	status "Downloading $BINARY_NAME..."
	if ! curl -fsSL "$DOWNLOAD_URL" -o "$TEMP_DIR/$BINARY_NAME"; then
		error "Failed to download $DOWNLOAD_URL"
	fi

	status "Downloading checksums..."
	if ! curl -fsSL "$CHECKSUMS_URL" -o "$TEMP_DIR/checksums.txt"; then
		error "Failed to download checksums from $CHECKSUMS_URL"
	fi

	# Verify checksum
	status "Verifying checksum..."
	CHECKSUM_CMD=""
	HASH=""
	EXPECTED_HASH=""

	if command -v sha256sum >/dev/null 2>&1; then
		HASH=$(sha256sum "$TEMP_DIR/$BINARY_NAME" | awk '{print $1}')
		CHECKSUM_CMD="sha256sum"
	elif command -v shasum >/dev/null 2>&1; then
		HASH=$(shasum -a 256 "$TEMP_DIR/$BINARY_NAME" | awk '{print $1}')
		CHECKSUM_CMD="shasum -a 256"
	else
		warning "Neither sha256sum nor shasum found. Skipping checksum verification."
		HASH="skipped"
	fi

	if [ "$HASH" != "skipped" ]; then
		EXPECTED_HASH=$(grep "$BINARY_NAME" "$TEMP_DIR/checksums.txt" | awk '{print $1}')

		if [ -z "$EXPECTED_HASH" ]; then
			error "Checksum not found for $BINARY_NAME in checksums.txt"
		fi

		if [ "$HASH" != "$EXPECTED_HASH" ]; then
			error "Checksum verification failed!\nExpected: $EXPECTED_HASH\nActual:   $HASH"
		fi

		status "${GREEN}✓${PLAIN} Checksum verified"
	fi

	# Extract binary
	status "Extracting binary..."
	if ! tar -xzf "$TEMP_DIR/$BINARY_NAME" -C "$TEMP_DIR"; then
		error "Failed to extract $BINARY_NAME"
	fi

	# Move binary to install directory
	if [ ! -f "$TEMP_DIR/mcp-setu" ]; then
		error "Binary not found in extracted archive"
	fi

	mv "$TEMP_DIR/mcp-setu" "$INSTALL_DIR/mcp-setu"
	chmod +x "$INSTALL_DIR/mcp-setu"

	# Verify installation with smoke test
	status "Verifying installation..."
	if ! VERSION_OUTPUT=$("$INSTALL_DIR/mcp-setu" version 2>&1); then
		error "Installation verification failed. Binary exists but version command failed."
	fi

	status "${GREEN}✓${PLAIN} mcp-setu installed successfully at $INSTALL_DIR/mcp-setu"
	echo "" >&2

	# Check if install directory is in PATH
	if ! echo "$PATH" | grep -q "$INSTALL_DIR"; then
		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
		warning "$INSTALL_DIR is not in your PATH"
		echo "" >&2

		# Try to detect shell and suggest adding to rc file
		if [ -n "${SHELL:-}" ]; then
			SHELL_NAME=$(basename "$SHELL")
			RC_FILE=""

			case "$SHELL_NAME" in
				bash)
					RC_FILE="$HOME/.bashrc"
					;;
				zsh)
					RC_FILE="$HOME/.zshrc"
					;;
				fish)
					RC_FILE="$HOME/.config/fish/config.fish"
					EXPORT_CMD="set -gx PATH $INSTALL_DIR \$PATH"
					;;
				*)
					RC_FILE="$HOME/.bashrc"
					;;
			esac

			if [ -z "${EXPORT_CMD:-}" ]; then
				EXPORT_CMD="export PATH=\"$INSTALL_DIR:\$PATH\""
			fi

			if [ -z "$CI" ] && [ -t 1 ]; then
				# Interactive mode and not in CI - offer to modify rc file
				if ! grep -q "MCP_SETU_INSTALL_PATH" "$RC_FILE" 2>/dev/null; then
					echo "Add the following to $RC_FILE:" >&2
					echo "" >&2
					echo "  $EXPORT_CMD # Added by mcp-setu installer" >&2
					echo "" >&2
					echo "Then run: source $RC_FILE" >&2
				fi
			else
				# Non-interactive or CI mode
				echo "Add the following to $RC_FILE:" >&2
				echo "" >&2
				echo "  $EXPORT_CMD" >&2
				echo "" >&2
				echo "Or run: source $RC_FILE" >&2
			fi
		else
			echo "Please add $INSTALL_DIR to your PATH" >&2
		fi

		echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" >&2
	fi

	echo "" >&2
	status "Run '${GREEN}mcp-setu chat${PLAIN}' to get started!"
	echo "" >&2
}

# Call main function with all arguments
main "$@"
