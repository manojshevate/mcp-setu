# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Modern full-screen chat TUI built on Bubble Tea: input is pinned to the
  bottom, output scrolls above, with `/`-triggered slash-command autocomplete
  (Tab to accept), `↑`/`↓` input history, and `Ctrl+C`/`Ctrl+D` to exit.
- New `/quit` and `/exit` slash commands (in addition to legacy `exit`/`quit`).
- `bridge.Printer` interface so different front-ends (CLI and TUI) can plug
  into the agentic loop without writing directly to stdout/stderr.

### Fixed
- Duplicate/garbled banner and missing input visibility when launching `chat`
  (banner was printed twice before the TUI started); banner is now rendered
  inside the TUI's initial output area.
- `--verbose` is now honored inside the chat TUI: tool calls, LLM iterations,
  and tool results appear inline in the output panel only when the flag is set.

### Changed
- README rewritten to be short and crisp; long-form content now lives in
  `docs/`.

### Deprecated
- TBD

### Removed
- TBD

### Security
- TBD

---

## [0.1.0] - 2026-05-10

### Added
- Initial release of mcp-setu
- Interactive chat interface with Ollama models
- Support for multiple MCP servers
- Tool calling capability
- Model management commands
- Configuration validation
- Release infrastructure:
  - Version management via `internal/version/version.go`
  - `mcp-setu version` command showing version, commit, and build date
  - GitHub Actions workflows for testing and releases
  - Cross-platform binary builds (Linux, macOS, Windows)
  - Homebrew distribution support
  - `go install` compatibility

### Fixed
- N/A (initial release)

### Known Issues
- Homebrew tap requires separate repository setup
- First release only supports macOS via Homebrew (Intel and Apple Silicon)

---

## Format Guidelines

### Version Header
Use format: `## [X.Y.Z] - YYYY-MM-DD`

### Section Headers
- **Added**: New features
- **Fixed**: Bug fixes
- **Changed**: Changes in existing functionality
- **Deprecated**: Soon-to-be removed features
- **Removed**: Removed features
- **Security**: Security fixes

### Examples

```markdown
## [0.2.0] - 2025-06-10

### Added
- New command: `/config` for managing configuration
- Support for persistent session history

### Fixed
- Bug in tool calling with multi-word arguments

### Changed
- Improved error messages for model switching

### Deprecated
- Old config format (support until v1.0.0)
```

## Release Checklist

When releasing a new version:

1. Move content from "Unreleased" section
2. Create new header with version and date
3. Ensure all changes are categorized
4. Update any links/references
5. Commit with message: `docs: update CHANGELOG for vX.Y.Z`
