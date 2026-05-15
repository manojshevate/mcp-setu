package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/manojshevate/mcp-setu/internal/bridge"
	"github.com/manojshevate/mcp-setu/internal/content"
	"github.com/manojshevate/mcp-setu/internal/mcp"
	"github.com/manojshevate/mcp-setu/internal/ollama"
	"github.com/manojshevate/mcp-setu/internal/ui"
)

// Styling.
var (
	stylePrompt    = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true)
	styleUser      = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true)
	styleAssistant = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Bold(true)
	styleMuted     = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
	styleWarning   = lipgloss.NewStyle().Foreground(lipgloss.Color("#F59E0B"))
	styleError     = lipgloss.NewStyle().Foreground(lipgloss.Color("#EF4444")).Bold(true)
	styleSeparator = lipgloss.NewStyle().Foreground(lipgloss.Color("#374151"))

	styleInputBox = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#374151")).
			Padding(0, 1)

	styleModelBadge = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#6B7280"))

	styleHeaderArt     = lipgloss.NewStyle().Foreground(lipgloss.Color("#7C3AED")).Bold(true)
	styleHeaderWelcome = lipgloss.NewStyle().Foreground(lipgloss.Color("#10B981")).Bold(true)
	styleHeaderMuted   = lipgloss.NewStyle().Foreground(lipgloss.Color("#6B7280"))
)

// headerHeight is the fixed number of lines the header occupies at the top
// of the terminal. It contains the ASCII art logo and welcome message.
const headerHeight = 8

// modelsLoadedMsg is fired when the background model fetch completes.
type modelsLoadedMsg struct {
	names []string
	err   error
}

// SessionModel is the root Bubble Tea model. The layout is strictly:
//
//	output area (scrolling)
//	────────────────────────── separator
//	⟳ thinking… 5s              (status, only when processing)
//	[autocomplete / picker]     (overlay, if triggered)
//	╭──────────────────────╮    (input box with border)
//	│ ❯ user input         │
//	╰──────────────────────╯
//	              ◆ llama3.2:3b    (model indicator, right-aligned)
type SessionModel struct {
	// terminal
	width, height int

	// deps
	ctx          context.Context
	br           *bridge.Bridge
	mcpClient    *mcp.MultiClient
	ollamaClient *ollama.Client
	model        string
	verbose      bool

	// chat state
	history []ollama.Message

	// rendering
	output          []string // logical lines (already split on '\n')
	input           InputModel
	respBuf         strings.Builder // accumulates streaming chunks
	respActive      bool
	status          string
	processing      bool
	processingStart time.Time
	eventCh         chan Event
	printerWire     bool // set true after the TUI printer is installed on the bridge
	scrollOffset    int  // lines scrolled up from the bottom (0 = show tail)
}

func NewSessionModel(
	ctx context.Context,
	br *bridge.Bridge,
	mcpClient *mcp.MultiClient,
	ollamaClient *ollama.Client,
	model string,
	systemPrompt string,
	verbose bool,
) SessionModel {
	eventCh := make(chan Event, 1024)

	m := SessionModel{
		ctx:          ctx,
		br:           br,
		mcpClient:    mcpClient,
		ollamaClient: ollamaClient,
		model:        model,
		history:      []ollama.Message{{Role: "system", Content: systemPrompt}},
		input:        NewInputModel(),
		eventCh:      eventCh,
	}
	m.input.SetCurrentModel(model)
	// Store the verbose flag so we can pass it to the printer later.
	m.verbose = verbose
	m.appendBanner()
	return m
}

// maxOutputLines is the maximum number of logical output lines retained in memory.
// When exceeded, the oldest lines are discarded using a ring-buffer pattern.
const maxOutputLines = 2000

// appendOutput appends lines to m.output and trims to maxOutputLines if needed.
func (m *SessionModel) appendOutput(lines ...string) {
	m.output = append(m.output, lines...)
	if len(m.output) > maxOutputLines {
		m.output = m.output[len(m.output)-maxOutputLines:]
	}
}

// appendBanner previously added the welcome banner to the output area, but the
// welcome message is now shown exclusively in the header (renderHeader). This
// function is intentionally left empty so that callers (NewSessionModel, /clear)
// do not need further changes.
func (m *SessionModel) appendBanner() {
	// Welcome info is displayed in the header only — no duplicate here.
}

// Init fires the event-poller and eagerly fetches the local model list so
// that autocomplete and the picker are populated from the very first keystroke.
func (m *SessionModel) Init() tea.Cmd {
	return tea.Batch(m.waitForEvent(), m.fetchModels())
}

// fetchModels returns a tea.Cmd that lists local models in the background.
func (m SessionModel) fetchModels() tea.Cmd {
	if m.ollamaClient == nil {
		return nil
	}
	ctx := m.ctx
	client := m.ollamaClient
	return func() tea.Msg {
		models, err := client.ListLocalModels(ctx)
		if err != nil {
			return modelsLoadedMsg{names: nil, err: err}
		}
		names := make([]string, 0, len(models))
		for _, mod := range models {
			names = append(names, mod.Name)
		}
		return modelsLoadedMsg{names: names, err: nil}
	}
}

func (m *SessionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.SetWidth(m.width - 6)
		return m, nil

	case modelsLoadedMsg:
		if msg.err != nil {
			// Log the error but don't block; the app is still usable.
			m.appendOutput(styleMuted.Render("  (warning: could not fetch models: " + msg.err.Error() + ")"))
		}
		m.input.SetModels(msg.names)
		return m, nil

	case tea.MouseMsg:
		// Mouse scroll wheel support for message history.
		allVisual := m.renderOutputLines(0)
		totalLines := len(allVisual)
		switch msg.Button {
		case tea.MouseButtonWheelUp:
			// Scroll up (show older messages).
			m.scrollOffset += 3
			if m.scrollOffset > totalLines-m.middleHeight() {
				m.scrollOffset = totalLines - m.middleHeight()
			}
			if m.scrollOffset < 0 {
				m.scrollOffset = 0
			}
			return m, nil
		case tea.MouseButtonWheelDown:
			// Scroll down (show newer messages).
			m.scrollOffset -= 3
			if m.scrollOffset < 0 {
				m.scrollOffset = 0
			}
			return m, nil
		}
		return m, nil

	case tea.KeyMsg:
		// Always allow Ctrl+C / Ctrl+D to exit.
		if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyCtrlD {
			return m, tea.Quit
		}

		// Page Up / Page Down: scroll the message area.
		// These are always handled, even while processing, so the user can read
		// history while waiting for a response.
		if msg.Type == tea.KeyPgUp {
			m.scrollOffset += m.scrollPageSize()
			// Upper clamp is applied in View() when rendering.
			return m, nil
		}
		if msg.Type == tea.KeyPgDown {
			m.scrollOffset -= m.scrollPageSize()
			if m.scrollOffset < 0 {
				m.scrollOffset = 0
			}
			return m, nil
		}

		// Ignore all other keystrokes while a request is in flight (except quit).
		if m.processing {
			return m, nil
		}
		// Home / End are forwarded to the input model for cursor navigation.

		// Escape: exit picker or dismiss autocomplete.
		if msg.Type == tea.KeyEsc {
			if m.input.Mode() == modeModelSelect {
				m.input.ExitModelSelect()
			} else if m.input.Mode() == modeAutocomplete {
				m.input.DismissAutocomplete()
			}
			return m, nil
		}

		if msg.Type == tea.KeyEnter {
			// In model picker mode, Enter confirms the selection.
			if m.input.Mode() == modeModelSelect {
				selected := m.input.SelectedModel()
				m.input.ExitModelSelect()
				if selected != "" && selected != m.model {
					if err := m.br.SetModel(m.ctx, selected); err != nil {
						m.appendError(err.Error())
					} else {
						m.model = selected
						m.input.SetCurrentModel(selected)
						m.appendInfo([]string{"Switched to model: " + selected})
					}
				}
				return m, nil
			}

			// In all other modes, delegate to InputModel which fires SendMsg.
			updated, inputCmd := m.input.Update(msg)
			m.input = updated.(InputModel)
			return m, inputCmd
		}
		updated, inputCmd := m.input.Update(msg)
		m.input = updated.(InputModel)
		return m, inputCmd

	case SendMsg:
		// Fired by Ctrl+Enter / Cmd+Enter: send the buffered input.
		if m.processing {
			return m, nil
		}
		return m, m.sendCurrentInput()

	case eventMsg:
		m.handleEvent(msg.ev)
		return m, m.waitForEvent()

	case tickMsg:
		return m, m.waitForEvent()

	case bridgeDoneMsg:
		m.processing = false
		m.processingStart = time.Time{}
		m.status = ""
		if msg.err != nil {
			m.appendError(msg.err.Error())
		} else if msg.content != "" && !m.respActive {
			// Non-streaming path returned content but no chunks came through.
			m.appendAssistant(msg.content)
		}
		m.history = append(m.history, ollama.Message{Role: "assistant", Content: msg.content})
		return m, m.waitForEvent()

	case quitMsg:
		return m, tea.Quit
	}

	return m, nil
}

// View renders the screen. The layout from top to bottom is:
//
//	╭──────────────────────────────────────────────╮
//	│  ASCII art logo   │   Welcome / server info  │  ← header (headerHeight lines)
//	├──────────────────────────────────────────────┤
//	│                                              │
//	│           scrollable message area            │  ← middle (variable height)
//	│                                              │
//	├──────────────────────────────────────────────┤  ← separator
//	│  ⟳ thinking…  (status line, when active)    │
//	│  [autocomplete / picker overlay]             │
//	│ ╭──────────────────────────────────────────╮ │
//	│ │ ❯ input                                  │ │  ← input box (border)
//	│ ╰──────────────────────────────────────────╯ │
//	│                              ◆ model-name    │  ← model badge
//	╰──────────────────────────────────────────────╯
func (m *SessionModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	// ── FOOTER (input box + accessories) ──────────────────────────────
	separator := styleSeparator.Render(strings.Repeat("─", m.width))

	// Dynamic input box: grow from 1 content line up to maxInputBoxLines.
	lineCount := m.input.LineCount()
	visibleLines := lineCount
	if visibleLines < 1 {
		visibleLines = 1
	}
	if visibleLines > maxInputBoxLines {
		visibleLines = maxInputBoxLines
	}
	contentLines := m.input.RenderAllLines(visibleLines)
	// Prepend the prompt glyph to the first line.
	if len(contentLines) > 0 {
		contentLines[0] = stylePrompt.Render("❯ ") + contentLines[0]
	}
	inputContent := strings.Join(contentLines, "\n")
	inputBox := styleInputBox.Width(m.width - 2).Render(inputContent)

	modelFooter := ""
	if m.model != "" {
		badge := styleModelBadge.Render("◆ " + truncateModel(m.model, m.width))
		modelFooter = lipgloss.PlaceHorizontal(m.width, lipgloss.Right, badge)
	}

	statusLine := ""
	if m.status != "" {
		s := m.status
		if m.processing && !m.processingStart.IsZero() {
			s = s + " " + formatElapsed(time.Since(m.processingStart))
		}
		statusLine = styleMuted.Render(s)
	}
	acBlock := m.input.RenderAutocomplete()

	// Footer height: separator + optional status + optional ac + inputBox + optional modelFooter.
	inputBoxLines := strings.Count(inputBox, "\n") + 1
	acLines := 0
	if acBlock != "" {
		acLines = strings.Count(acBlock, "\n") + 1
	}
	footerHeight := 1 + inputBoxLines + acLines // separator(1) + inputBox + ac
	if statusLine != "" {
		footerHeight++
	}
	if modelFooter != "" {
		footerHeight++
	}

	// ── MIDDLE (scrollable messages) ──────────────────────────────────
	// Effective header rows: use headerHeight only when terminal is tall enough.
	effectiveHeaderHeight := headerHeight
	if m.height < headerHeight+footerHeight+3 {
		// Tiny terminal: drop the header entirely so messages are still visible.
		effectiveHeaderHeight = 0
	}
	middleHeight := m.height - effectiveHeaderHeight - footerHeight
	if middleHeight < 1 {
		middleHeight = 1
	}

	// Build all visual lines from the output buffer.
	allVisual := m.renderOutputLines(0) // 0 = no cap, get everything
	totalLines := len(allVisual)

	// Clamp scroll offset so we never scroll past the top.
	maxOffset := totalLines - middleHeight
	if maxOffset < 0 {
		maxOffset = 0
	}
	if m.scrollOffset > maxOffset {
		m.scrollOffset = maxOffset
	}
	if m.scrollOffset < 0 {
		m.scrollOffset = 0
	}

	// Slice the window: scrollOffset=0 means bottom (latest), higher = older.
	var visual []string
	if m.scrollOffset == 0 {
		// Normal view: latest lines at the bottom.
		if len(allVisual) > middleHeight {
			visual = allVisual[len(allVisual)-middleHeight:]
		} else {
			visual = allVisual
		}
	} else {
		// Scrolled up: show older content.
		endIdx := totalLines - m.scrollOffset
		if endIdx > totalLines {
			endIdx = totalLines
		}
		startIdx := endIdx - middleHeight
		if startIdx < 0 {
			startIdx = 0
		}
		visual = allVisual[startIdx:endIdx]
	}

	// Pad top with blank lines so content hugs the separator.
	if pad := middleHeight - len(visual); pad > 0 {
		visual = append(make([]string, pad), visual...)
	}

	// Scroll indicator: show "↑ N lines above" when scrolled up.
	// Place it on the first (top) line of the middle area so it acts as a
	// header banner, leaving the bottom content (nearest the input) intact.
	if m.scrollOffset > 0 {
		scrollIndicator := styleMuted.Render(fmt.Sprintf("  ↑ %d lines above  (PgUp/PgDn to scroll)", m.scrollOffset))
		if len(visual) > 0 {
			visual[0] = scrollIndicator
		}
	}

	// ── ASSEMBLE ──────────────────────────────────────────────────────
	var parts []string

	// Header (skipped on tiny terminals).
	if effectiveHeaderHeight > 0 {
		headerLines := m.renderHeader()
		parts = append(parts, strings.Join(headerLines, "\n"))
	}

	// Middle.
	parts = append(parts, strings.Join(visual, "\n"))

	// Footer.
	parts = append(parts, separator)
	if statusLine != "" {
		parts = append(parts, statusLine)
	}
	if acBlock != "" {
		parts = append(parts, acBlock)
	}
	parts = append(parts, inputBox)
	if modelFooter != "" {
		parts = append(parts, modelFooter)
	}

	return strings.Join(parts, "\n")
}

// renderOutputLines word-wraps all output (including the in-progress response
// buffer) and returns all visual rows. The caller is responsible for slicing
// to the desired window size and applying the scroll offset.
func (m SessionModel) renderOutputLines(_ int) []string {
	wrapWidth := m.width
	if wrapWidth < 1 {
		wrapWidth = 80
	}

	var visual []string
	add := func(s string) {
		for _, line := range strings.Split(s, "\n") {
			visual = append(visual, wrapLine(line, wrapWidth)...)
		}
	}
	for _, line := range m.output {
		add(line)
	}
	if m.respActive {
		buf := m.respBuf.String()
		if buf != "" {
			add(styleAssistant.Render("setu  ") + buf)
		}
	}
	return visual
}

// scrollPageSize returns how many lines to scroll per Page Up/Down, based
// on the current middle area height (minimum 1).
func (m SessionModel) scrollPageSize() int {
	ps := m.middleHeight()
	if ps < 1 {
		ps = 1
	}
	return ps
}

// middleHeight computes the height of the scrollable message area based on
// the current terminal dimensions and bottom chrome height.
func (m SessionModel) middleHeight() int {
	if m.height == 0 {
		return 0
	}
	// Reserve headerHeight rows at top and chromeHeight rows at bottom.
	// chromeHeight calculation mirrors View() to stay in sync.
	lineCount := m.input.LineCount()
	visibleLines := lineCount
	if visibleLines < 1 {
		visibleLines = 1
	}
	if visibleLines > maxInputBoxLines {
		visibleLines = maxInputBoxLines
	}
	contentLines := m.input.RenderAllLines(visibleLines)
	if len(contentLines) > 0 {
		contentLines[0] = stylePrompt.Render("❯ ") + contentLines[0]
	}
	inputBox := styleInputBox.Width(m.width - 2).Render(strings.Join(contentLines, "\n"))
	inputBoxLines := strings.Count(inputBox, "\n") + 1
	acBlock := m.input.RenderAutocomplete()
	acLines := 0
	if acBlock != "" {
		acLines = strings.Count(acBlock, "\n") + 1
	}
	chromeHeight := 1 + inputBoxLines + acLines // separator(1) + inputBox + ac
	if m.status != "" {
		chromeHeight++
	}
	if m.model != "" {
		chromeHeight++
	}
	mid := m.height - headerHeight - chromeHeight
	if mid < 1 {
		mid = 1
	}
	return mid
}

// renderHeader returns the fixed 8-line header block containing the ASCII art
// banner on the left and welcome/info on the right.
func (m SessionModel) renderHeader() []string {
	if m.width < 1 {
		return make([]string, headerHeight)
	}

	// ASCII art banner — exactly 8 lines, placed on the left.
	artRaw := []string{
		`'##::::'##::'######::'########::::::::::::::::::'######::'########:'########::'##::::'##:`,
		` ###::'###:'##... ##: ##.... ##::::::::::::::::'##... ##: ##.....::... ##..:: ##:::: ##:`,
		` ####'####: ##:::..:: ##:::: ##:::::::::::::::: ##:::..:: ##:::::::::: ##:::: ##:::: ##:`,
		` ## ### ##: ##::::::: ########:::::'#######::::. ######:: ######:::::: ##:::: ##::::'##:`,
		` ##. #: ##: ##::::::: ##.....::::::........:::::..... ##: ##...::::::: ##:::: ##:::: ##:`,
		` ##:.:: ##: ##::: ##: ##:::::::::::::::::::::::'##::: ##: ##:::::::::: ##:::: ##:::: ##:`,
		` ##:::: ##:. ######:: ##:::::::::::::::::::::::. ######:: ########:::: ##::::. #######::`,
		`..:::::..:::......:::..:::::::::::::::::::::::::......:::........:::::..::::::.......:::`,
	}
	art := make([]string, len(artRaw))
	for i, line := range artRaw {
		art[i] = styleHeaderArt.Render(line)
	}

	// Welcome message lines (right side) — info only, no duplicate welcome text.
	servers := ui.GetServersTableInfo(m.mcpClient)
	serverCount := len(servers)
	toolCount := len(m.mcpClient.GetAllTools())
	welcome := []string{
		styleHeaderWelcome.Render("MCP-SETU"),
		styleHeaderMuted.Render("Model:   " + m.model),
		styleHeaderMuted.Render(fmt.Sprintf("Servers: %d connected · %d tools", serverCount, toolCount)),
		styleHeaderMuted.Render("/help for commands · /quit to exit"),
		styleHeaderMuted.Render("↑/↓ history · PgUp/PgDn scroll"),
	}

	// Determine the width of the right panel. We allocate roughly 1/3 of the
	// terminal for the welcome info, with a minimum of 32 columns.
	rightPanelWidth := m.width / 3
	if rightPanelWidth < 32 {
		rightPanelWidth = 32
	}
	if rightPanelWidth > 48 {
		rightPanelWidth = 48
	}
	separatorStr := "  │  "
	separatorWidth := lipgloss.Width(separatorStr)

	// Build 8 lines. The banner spans all 8 rows on the left.
	// The welcome panel is placed on the right of rows 0-4, right-aligned within
	// the right panel area.
	var lines []string
	for i := 0; i < headerHeight; i++ {
		left := ""
		if i < len(art) {
			left = art[i]
		}

		var right string
		if i < len(welcome) {
			welcomeText := welcome[i]
			textWidth := lipgloss.Width(welcomeText)
			// Right-align the welcome text within rightPanelWidth (excluding separator).
			contentWidth := rightPanelWidth - separatorWidth
			if contentWidth < 1 {
				contentWidth = 1
			}
			pad := contentWidth - textWidth
			if pad < 0 {
				pad = 0
			}
			right = separatorStr + strings.Repeat(" ", pad) + welcomeText
		}

		leftWidth := lipgloss.Width(left)
		rightWidth := lipgloss.Width(right)
		gap := m.width - leftWidth - rightWidth
		if gap < 0 {
			if right != "" {
				// Banner is wider than terminal but there is a right-panel entry:
				// truncate the banner line so the info still appears.
				maxLeft := m.width - rightWidth - 1
				if maxLeft < 0 {
					maxLeft = 0
				}
				truncatedLeft := truncateToWidth(left, maxLeft)
				actualLeft := lipgloss.Width(truncatedLeft)
				pad := m.width - actualLeft - rightWidth
				if pad < 1 {
					pad = 1
				}
				lines = append(lines, truncatedLeft+strings.Repeat(" ", pad)+right)
			} else {
				// No right panel: just render the banner line alone.
				lines = append(lines, left)
			}
			continue
		}
		if gap < 1 {
			gap = 1
		}
		lines = append(lines, left+strings.Repeat(" ", gap)+right)
	}

	// Ensure exactly headerHeight lines.
	for len(lines) < headerHeight {
		lines = append(lines, "")
	}
	return lines[:headerHeight]
}

// wrapLine breaks a single visual line at terminal width, preserving any
// already-styled ANSI sequences as a single chunk where possible.
func wrapLine(s string, width int) []string {
	if width < 1 {
		return []string{s}
	}
	// lipgloss's Width does the right thing for ANSI-styled text.
	w := lipgloss.Width(s)
	if w <= width {
		return []string{s}
	}
	// Fall back to plain-byte chunking for very long output (tool dumps).
	// This is imperfect for styled text but safe.
	var out []string
	for len(s) > 0 {
		if len(s) <= width {
			out = append(out, s)
			break
		}
		out = append(out, s[:width])
		s = s[width:]
	}
	return out
}

// ───────────────────────── event handling ─────────────────────────

func (m *SessionModel) handleEvent(ev Event) {
	switch ev.Kind {
	case EventLog:
		m.appendOutput(styleMuted.Render("  " + ev.Text))
	case EventWarning:
		m.appendOutput(styleWarning.Render("⚠ " + ev.Text))
	case EventError:
		m.appendError(ev.Text)
	case EventRespStart:
		m.respBuf.Reset()
		m.respActive = true
	case EventRespChunk:
		m.respBuf.WriteString(ev.Text)
	case EventRespEnd:
		if m.respActive {
			final := m.respBuf.String()
			m.respBuf.Reset()
			m.respActive = false
			if strings.TrimSpace(final) != "" {
				m.appendAssistant(final)
			}
		}
	case EventStructuredContent:
		if ev.StructuredContent != nil {
			m.appendStructuredContent(ev.StructuredContent)
		}
	}
}

func (m *SessionModel) appendUser(text string) {
	prefix := styleUser.Render("you   ")
	m.appendOutput(prefix + text)
}

func (m *SessionModel) appendAssistant(text string) {
	prefix := styleAssistant.Render("setu  ")
	// Preserve internal newlines.
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	for i, line := range lines {
		if i == 0 {
			m.appendOutput(prefix + line)
		} else {
			m.appendOutput("      " + line)
		}
	}
	m.appendOutput("")
}

func (m *SessionModel) appendError(msg string) {
	m.appendOutput(styleError.Render("error: ") + msg)
}

func (m *SessionModel) appendStructuredContent(sc *content.StructuredContent) {
	// Get terminal width for rendering
	width := m.width - 4 // Account for padding
	if width < 40 {
		width = 40
	}

	renderer := content.NewRenderer(width)
	formatted := renderer.Render(sc)

	// Split formatted output into lines and append
	lines := strings.Split(strings.TrimRight(formatted, "\n"), "\n")
	m.appendOutput(lines...)
	m.appendOutput("")
}

// ───────────────────────── commands ─────────────────────────

func (m *SessionModel) handleSlashCommand(input string) (tea.Cmd, bool) {
	switch {
	case input == "/quit" || input == "/exit" || input == "exit" || input == "quit":
		return func() tea.Msg { return quitMsg{} }, true

	case input == "/help":
		m.appendInfo([]string{
			"Commands:",
			"  /tools          List available tools",
			"  /servers        Show connected MCP servers",
			"  /model [name]   Show or switch model",
			"  /stats          Show performance stats",
			"  /clear          Clear conversation history",
			"  /help           Show this help",
			"  /quit, /exit    Exit",
			"",
			"Key Bindings:",
			"  Enter           Send message",
			"  Shift+Enter     Insert newline (multiline mode)",
			"  Page Up/Down    Scroll message history",
			"  Left/Right/Home/End  Navigate in input",
			"  Ctrl+C          Exit",
		})
		return nil, true

	case input == "/tools":
		tools := ui.GetToolsTableFromMCP(m.mcpClient)
		lines := []string{"Available tools:"}
		for _, t := range tools {
			lines = append(lines, fmt.Sprintf("  • %s  %s  %s", t.Name, styleMuted.Render("("+t.Server+")"), t.Description))
		}
		if len(tools) == 0 {
			lines = append(lines, "  (none)")
		}
		m.appendInfo(lines)
		return nil, true

	case input == "/servers":
		servers := ui.GetServersTableInfo(m.mcpClient)
		lines := []string{"Connected servers:"}
		for _, s := range servers {
			lines = append(lines, fmt.Sprintf("  • %s — %d tools", s.Name, s.Tools))
		}
		m.appendInfo(lines)
		return nil, true

	case input == "/stats":
		st := m.br.GetStats()
		m.appendInfo([]string{
			fmt.Sprintf("Stats: %d messages · %d tool calls · %d iterations · %s total",
				st.MessageCount, st.ToolCallCount, st.IterationCount, st.TotalDuration.Round(time.Millisecond)),
		})
		return nil, true

	case input == "/clear":
		sys := m.history[0]
		m.history = []ollama.Message{sys}
		m.output = m.output[:0]
		m.appendBanner()
		m.appendInfo([]string{"Conversation cleared."})
		return nil, true

	// "/model" (bare) is handled before handleSlashCommand (enters picker).
	// "/model name" sets the model directly.
	case strings.HasPrefix(input, "/model "):
		parts := strings.Fields(input)
		if len(parts) != 2 {
			m.appendError("usage: /model <name>")
			return nil, true
		}
		if err := m.br.SetModel(m.ctx, parts[1]); err != nil {
			m.appendError(err.Error())
			return nil, true
		}
		m.model = parts[1]
		m.input.SetCurrentModel(parts[1])
		m.appendInfo([]string{"Switched to model: " + parts[1]})
		return nil, true
	}
	return nil, false
}

func (m *SessionModel) appendInfo(lines []string) {
	m.appendOutput(lines...)
	m.appendOutput("")
}

// sendCurrentInput reads the current input value and dispatches it as a chat
// message or slash command. Returns a tea.Cmd to execute (may be nil).
func (m *SessionModel) sendCurrentInput() tea.Cmd {
	value := m.input.GetValue()
	if value == "" {
		return nil
	}
	m.input.AddToHistory(value)
	m.appendUser(value)
	m.input.Clear()

	// Handle "/model" (bare, no name) → enter picker.
	if value == "/model" {
		m.input.EnterModelSelect()
		return nil
	}

	if cmd, handled := m.handleSlashCommand(value); handled {
		return cmd
	}

	// Install the TUI printer on the bridge exactly once, just-in-time.
	if !m.printerWire {
		m.br.SetPrinter(NewPrinter(m.eventCh, m.verbose))
		m.printerWire = true
	}

	m.history = append(m.history, ollama.Message{Role: "user", Content: value})
	m.processing = true
	m.processingStart = time.Now()
	m.status = "⟳ thinking…"
	return tea.Batch(m.runBridge(value), m.waitForEvent())
}

// ───────────────────────── async plumbing ─────────────────────────

type eventMsg struct{ ev Event }
type tickMsg struct{} // heartbeat to re-poll the channel without producing output
type bridgeDoneMsg struct {
	content string
	err     error
}
type quitMsg struct{}

// waitForEvent returns a tea.Cmd that blocks (with a short timeout) on the
// event channel. Pairing this with re-issuing after each event delivers a
// continuous stream of TUI updates without busy-polling.
func (m SessionModel) waitForEvent() tea.Cmd {
	ch := m.eventCh
	return func() tea.Msg {
		select {
		case ev := <-ch:
			return eventMsg{ev: ev}
		case <-time.After(100 * time.Millisecond):
			return tickMsg{}
		}
	}
}

// runBridge runs ProcessMessage in a goroutine via a tea.Cmd. The bridge writes
// its own progress events to m.eventCh via the TUI printer installed on it.
func (m SessionModel) runBridge(_ string) tea.Cmd {
	return func() tea.Msg {
		content, err := m.br.ProcessMessage(m.ctx, m.history)
		return bridgeDoneMsg{content: content, err: err}
	}
}

// formatElapsed formats a duration as "Xs" for under 60 seconds, or "Xm Ys"
// for 60 seconds and above. Negative durations are clamped to 0.
func formatElapsed(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	secs := int(d.Truncate(time.Second).Seconds())
	if secs < 60 {
		return fmt.Sprintf("%ds", secs)
	}
	return fmt.Sprintf("%dm %ds", secs/60, secs%60)
}

func listModelInfos(ctx context.Context, client *ollama.Client) ([]ui.ModelInfo, error) {
	models, err := client.ListLocalModels(ctx)
	if err != nil {
		return nil, err
	}
	var infos []ui.ModelInfo
	for _, mod := range models {
		infos = append(infos, ui.ModelInfo{Name: mod.Name, Size: mod.Size})
	}
	return infos, nil
}

// truncateToWidth truncates a plain (or ANSI-styled) string to at most maxWidth
// visible columns. It works rune-by-rune on the raw bytes, which is imperfect
// for styled text but sufficient for the header banner use-case where we just
// need to fit within the available terminal width.
func truncateToWidth(s string, maxWidth int) string {
	if maxWidth <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= maxWidth {
		return s
	}
	// Strip runes from the right until we fit.
	runes := []rune(s)
	for len(runes) > 0 {
		if lipgloss.Width(string(runes)) <= maxWidth {
			return string(runes)
		}
		runes = runes[:len(runes)-1]
	}
	return ""
}

// truncateModel shortens a model name to roughly one-third of the terminal
// width, appending "…" if truncation occurs. It uses lipgloss.Width for
// ANSI-aware measurement. Returns the name unchanged when width < 12.
func truncateModel(name string, termWidth int) string {
	if termWidth < 12 {
		return name
	}
	maxLen := termWidth / 3
	if maxLen < 4 {
		maxLen = 4
	}
	if lipgloss.Width(name) <= maxLen {
		return name
	}
	// Trim rune by rune until we fit (maxLen-1 visible cols + "…").
	runes := []rune(name)
	for len(runes) > 0 {
		candidate := string(runes) + "…"
		if lipgloss.Width(candidate) <= maxLen {
			return candidate
		}
		runes = runes[:len(runes)-1]
	}
	return "…"
}
