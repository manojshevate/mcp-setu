# Third-Party Licenses

This project includes the following third-party dependencies under their respective licenses.

**Note on dependency scope:** This list includes all dependencies in the Go module graph, both direct dependencies (explicitly required in `go.mod`) and indirect/transitive dependencies (required by direct dependencies). Direct dependencies: `charmbracelet/bubbletea`, `charmbracelet/lipgloss`, `mattn/go-isatty`, `spf13/cobra`, `golang.org/x/term`, `golang.org/x/oauth2`, `zalando/go-keyring`.

## Apache License 2.0

- **spf13/cobra** (v1.8.0) - https://github.com/spf13/cobra
  - CLI framework for building command-line applications
  - License: Apache License 2.0
  - See: https://github.com/spf13/cobra/blob/main/LICENSE.txt

- **inconshreveable/mousetrap** (v1.1.0) - https://github.com/inconshreveable/mousetrap
  - Prevents accidental invocation on Windows for Go CLI tools
  - License: Apache License 2.0
  - See: https://github.com/inconshreveable/mousetrap/blob/master/LICENSE

- **zalando/go-keyring** (v0.2.8) - https://github.com/zalando/go-keyring
  - Go library for secure credential storage using OS keyrings
  - Supports macOS Keychain, Linux Secret Service, Windows Credential Manager
  - License: Apache License 2.0
  - See: https://github.com/zalando/go-keyring/blob/master/LICENSE

> **Note:** Apache 2.0 licensed software requires attribution and distribution of license terms. See `NOTICE` file for additional attribution requirements.

## MIT License

- **charmbracelet/bubbletea** (v1.3.10) - https://github.com/charmbracelet/bubbletea
  - Terminal UI framework (Bubble Tea)
  - License: MIT
  - See: https://github.com/charmbracelet/bubbletea/blob/main/LICENSE

- **charmbracelet/lipgloss** (v1.1.0) - https://github.com/charmbracelet/lipgloss
  - Terminal styling and rendering
  - License: MIT
  - See: https://github.com/charmbracelet/lipgloss/blob/main/LICENSE

- **charmbracelet/colorprofile** (v0.2.3-*) - https://github.com/charmbracelet/x
  - Color profile detection
  - License: MIT

- **charmbracelet/x/ansi** (v0.10.1) - https://github.com/charmbracelet/x
  - ANSI terminal sequences
  - License: MIT

- **charmbracelet/x/cellbuf** (v0.0.13-*) - https://github.com/charmbracelet/x
  - Cell buffer implementation
  - License: MIT

- **charmbracelet/x/term** (v0.2.1) - https://github.com/charmbracelet/x
  - Terminal utilities
  - License: MIT

- **aymanbagabas/go-osc52/v2** (v2.0.1) - https://github.com/aymanbagabas/go-osc52
  - ANSI escape sequence support (OSC 52)
  - License: MIT

- **erikgeiser/coninput** (v0.0.0-20211004153227-1c3628e74d0f) - https://github.com/erikgeiser/coninput
  - Console input for Windows
  - License: MIT

- **muesli/ansi** (v0.0.0-20230316100256-276c6243b2f6) - https://github.com/muesli/ansi
  - ANSI utilities
  - License: MIT

- **muesli/cancelreader** (v0.2.2) - https://github.com/muesli/cancelreader
  - Context-aware reader
  - License: MIT

- **muesli/termenv** (v0.16.0) - https://github.com/muesli/termenv
  - Terminal environment detection
  - License: MIT

- **rivo/uniseg** (v0.4.7) - https://github.com/rivo/uniseg
  - Unicode text segmentation
  - License: MIT

- **xo/terminfo** (v0.0.0-20220910002029-abceb7e1c41e) - https://github.com/xo/terminfo
  - Terminal information (terminfo) parsing
  - License: MIT

- **mattn/go-runewidth** (v0.0.16) - https://github.com/mattn/go-runewidth
  - Rune width calculation
  - License: MIT

- **cpuguy83/go-md2man/v2** (v2.0.3) - https://github.com/cpuguy83/go-md2man
  - Markdown to man page converter
  - License: MIT

- **danieljoos/wincred** (v1.2.3) - https://github.com/danieljoos/wincred
  - Windows credential storage (used by go-keyring on Windows)
  - License: MIT

## BSD License (2-Clause & 3-Clause)

- **golang.org/x/sys** (v0.44.0) - https://github.com/golang/sys
  - System call abstractions
  - License: BSD 3-Clause
  - See: https://github.com/golang/sys/blob/master/LICENSE

- **golang.org/x/term** (v0.43.0) - https://github.com/golang/term
  - Terminal utilities from Go standard library
  - License: BSD 3-Clause
  - See: https://github.com/golang/term/blob/master/LICENSE

- **golang.org/x/text** (v0.3.8) - https://github.com/golang/text
  - Text processing and encoding
  - License: BSD 3-Clause
  - See: https://github.com/golang/text/blob/master/LICENSE

- **golang.org/x/oauth2** (v0.36.0) - https://github.com/golang/oauth2
  - OAuth 2.0 client library
  - License: BSD 3-Clause
  - See: https://github.com/golang/oauth2/blob/master/LICENSE

- **golang.org/x/exp** (v0.0.0-20220909182711-5c715a9e8561) - https://github.com/golang/exp
  - Experimental Go packages
  - License: BSD 3-Clause

- **golang.org/x/mod** (v0.6.0-dev.0.20220419223038-86c51ed26bb4) - https://github.com/golang/mod
  - Go module utilities
  - License: BSD 3-Clause

- **golang.org/x/tools** (v0.1.12) - https://github.com/golang/tools
  - Go development tools
  - License: BSD 3-Clause

- **russross/blackfriday/v2** (v2.1.0) - https://github.com/russross/blackfriday
  - Markdown processor
  - License: BSD 2-Clause
  - See: https://github.com/russross/blackfriday/blob/master/LICENSE.txt

- **godbus/dbus/v5** (v5.2.2) - https://github.com/godbus/dbus
  - D-Bus client library (used by go-keyring on Linux)
  - License: BSD 2-Clause
  - See: https://github.com/godbus/dbus/blob/master/LICENSE

## ISC License

- **gopkg.in/yaml.v3** (v3.0.1) - https://gopkg.in/yaml.v3
  - YAML parser
  - License: ISC & Apache 2.0
  - See: https://github.com/go-yaml/yaml/blob/v3/LICENSE

## Permissive Custom Licenses

- **spf13/pflag** (v1.0.5) - https://github.com/spf13/pflag
  - POSIX flag parsing (fork of standard library)
  - License: BSD 3-Clause
  - See: https://github.com/spf13/pflag/blob/master/LICENSE

- **mattn/go-isatty** (v0.0.20) - https://github.com/mattn/go-isatty
  - Terminal detection
  - License: MIT
  - Author: Yasuhiro MATSUMOTO <mattn.jp@gmail.com>

- **mattn/go-localereader** (v0.0.1) - https://github.com/mattn/go-localereader
  - Locale-aware reader
  - License: MIT
  - Author: Yasuhiro MATSUMOTO

- **lucasb-eyer/go-colorful** (v1.2.0) - https://github.com/lucasb-eyer/go-colorful
  - Color representation and conversion
  - License: MIT
  - Author: Lucas Beyer

- **aymanbagabas/go-udiff** (v0.2.0) - https://github.com/aymanbagabas/go-udiff
  - Unified diff implementation
  - License: BSD 3-Clause (Go Authors)

- **bits-and-blooms/bitset** (v1.22.0) - https://github.com/bits-and-blooms/bitset
  - Bit set implementation
  - License: BSD 3-Clause

- **charmbracelet/x/exp/golden** (v0.0.0-20240806155701-69247e0abc2a) - https://github.com/charmbracelet/x
  - Golden file testing (experimental)
  - License: MIT

- **gopkg.in/check.v1** (v0.0.0-20161208181325-20d25e280405) - https://gopkg.in/check.v1
  - Assertion checking library (test dependency)
  - License: BSD 2-Clause
  - See: https://github.com/go-check/check/blob/master/LICENSE

---

## Summary by License Type

| License Type | Count | Notes |
|--------------|-------|-------|
| MIT | 15 | Most permissive, minimal attribution needed |
| Apache 2.0 | 3 | Requires NOTICE file for attribution |
| BSD 3-Clause | 7 | Permissive, requires license text |
| BSD 2-Clause | 3 | Permissive, requires license text |
| ISC | 1 | Permissive license |

## Compliance Notes

### Apache 2.0 Software
The following software is distributed under the Apache License 2.0:
- `spf13/cobra`
- `inconshreveable/mousetrap`
- `zalando/go-keyring`

When distributing this software, you are required to:
1. Include a copy of the Apache License 2.0
2. Include a NOTICE file listing Apache-licensed dependencies
3. Include any NOTICE files from the original projects

See the `NOTICE` file in this repository for Apache 2.0 compliance details.

### MIT Licensed Software
MIT-licensed dependencies are generally the most permissive. However, when distributing this software, you should still include or reference their license texts.

### Go Standard Library
Many dependencies rely on or extend Go's standard library, which is under the BSD 3-Clause license. The original Go Authors should be attributed.

## How to Include These Licenses

When distributing mcp-setu (either as a binary or source):

1. **Include the main LICENSE file** (mcp-setu's MIT license)
2. **Include or reference THIRD_PARTY_LICENSES.md** (this file)
3. **Include the NOTICE file** (for Apache 2.0 compliance)
4. **Optionally include full license texts** in a `licenses/` directory if distributing source

For binary distributions, at minimum include:
- LICENSE
- NOTICE
- THIRD_PARTY_LICENSES.md (or a reference to it)

## License Text References

Full license texts are available at:
- Apache 2.0: https://www.apache.org/licenses/LICENSE-2.0
- MIT: https://opensource.org/licenses/MIT
- BSD 3-Clause: https://opensource.org/licenses/BSD-3-Clause
- BSD 2-Clause: https://opensource.org/licenses/BSD-2-Clause
- ISC: https://opensource.org/licenses/ISC

---

**Last Updated:** 2026-05-24
