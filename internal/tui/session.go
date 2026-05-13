package tui

import (
	"context"
	"fmt"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/manojshevate/mcp-setu/internal/bridge"
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
)

// modelsLoadedMsg is fired when the background model fetch completes.
type modelsLoadedMsg struct {
	names []string
}

// SessionModel is the root Bubble Tea model. The layout is strictly:
//
//	┌────────────────────────────────────┐
//	│ output area (scrolling)            │
//	│ ...                                │
//	├────────────────────────────────────┤
//	│ status (only when processing)      │
//	│ [autocomplete / picker overlay]    │
//	│ ❯ input                            │
//	└────────────────────────────────────┘
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
	output      []string // logical lines (already split on '\n')
	input       InputModel
	respBuf     strings.Builder // accumulates streaming chunks
	respActive  bool
	status      string
	processing  bool
	eventCh     chan Event
	printerWire bool // set true after the TUI printer is installed on the bridge
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

// appendBanner adds the welcome banner and connected-server list to the output.
func (m *SessionModel) appendBanner() {
	servers := ui.GetServersTableInfo(m.mcpClient)
	serverCount := len(servers)
	toolCount := len(m.mcpClient.GetAllTools())

	m.output = append(m.output,
		stylePrompt.Render("Welcome to mcp-setu"),
		"",
		fmt.Sprintf("  %s  %s", styleMuted.Render("Model    "), m.model),
		fmt.Sprintf("  %s  %d connected · %d tools", styleMuted.Render("Servers  "), serverCount, toolCount),
		"",
	)
	if serverCount > 0 {
		m.output = append(m.output, styleMuted.Render("  Servers:"))
		for _, s := range servers {
			m.output = append(m.output, fmt.Sprintf("    • %s (%d tools)", s.Name, s.Tools))
		}
		m.output = append(m.output, "")
	}
	m.output = append(m.output,
		styleMuted.Render("  Type /help for commands, /quit to exit. ↑/↓ for history."),
		strings.Repeat("─", 40),
	)
}

// Init fires the event-poller and eagerly fetches the local model list so
// that autocomplete and the picker are populated from the very first keystroke.
func (m SessionModel) Init() tea.Cmd {
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
			return modelsLoadedMsg{names: nil}
		}
		names := make([]string, 0, len(models))
		for _, mod := range models {
			names = append(names, mod.Name)
		}
		return modelsLoadedMsg{names: names}
	}
}

func (m SessionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.input.SetWidth(m.width - 4)
		return m, nil

	case modelsLoadedMsg:
		m.input.SetModels(msg.names)
		return m, nil

	case tea.KeyMsg:
		// Always allow Ctrl+C / Ctrl+D to exit.
		if msg.Type == tea.KeyCtrlC || msg.Type == tea.KeyCtrlD {
			return m, tea.Quit
		}
		// Ignore keystrokes while a request is in flight (except quit).
		if m.processing {
			return m, nil
		}

		// Escape: exit picker or dismiss autocomplete.
		if msg.Type == tea.KeyEsc {
			if m.input.mode == modeModelSelect || m.input.mode == modeAutocomplete {
				m.input.ExitModelSelect()
				m.input.mode = modeNormal
			}
			return m, nil
		}

		if msg.Type == tea.KeyEnter {
			// In model picker mode, confirm selection.
			if m.input.mode == modeModelSelect {
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

			value := m.input.GetValue()
			if value == "" {
				return m, nil
			}
			m.input.AddToHistory(value)
			m.appendUser(value)
			m.input.Clear()

			// Handle "/model" (bare, no name) → enter picker.
			if value == "/model" {
				m.input.EnterModelSelect()
				return m, nil
			}

			if cmd, handled := m.handleSlashCommand(value); handled {
				return m, cmd
			}

			// Install the TUI printer on the bridge exactly once, just-in-time.
			if !m.printerWire {
				m.br.SetPrinter(NewPrinter(m.eventCh, m.verbose))
				m.printerWire = true
			}

			m.history = append(m.history, ollama.Message{Role: "user", Content: value})
			m.processing = true
			m.status = "⟳ thinking…"
			return m, tea.Batch(m.runBridge(value), m.waitForEvent())
		}
		updated, _ := m.input.Update(msg)
		m.input = updated.(InputModel)
		return m, nil

	case eventMsg:
		m.handleEvent(msg.ev)
		return m, m.waitForEvent()

	case tickMsg:
		return m, m.waitForEvent()

	case bridgeDoneMsg:
		m.processing = false
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
//	output area  (scrolling)
//	separator
//	status line  (only when processing)
//	autocomplete / picker overlay  (above the input line)
//	❯ input line
func (m SessionModel) View() string {
	if m.width == 0 || m.height == 0 {
		return ""
	}

	// Bottom-pinned chrome.
	separator := styleSeparator.Render(strings.Repeat("─", m.width))
	inputLine := stylePrompt.Render("❯ ") + m.input.RenderLine()
	statusLine := ""
	if m.status != "" {
		statusLine = styleMuted.Render(m.status)
	}
	acBlock := m.input.RenderAutocomplete()

	// Calculate how many rows the bottom chrome consumes.
	// separator (1) + optional status (1) + optional AC/picker + input (1).
	acLines := 0
	if acBlock != "" {
		acLines = strings.Count(acBlock, "\n") + 1
	}
	chromeHeight := 1 + 1 + acLines // separator + input + ac
	if statusLine != "" {
		chromeHeight++
	}
	outputHeight := m.height - chromeHeight
	if outputHeight < 1 {
		outputHeight = 1
	}

	// Build visible output.
	visual := m.renderOutputLines(outputHeight)

	// Pad the top so output hugs the bottom (against the separator).
	if pad := outputHeight - len(visual); pad > 0 {
		visual = append(make([]string, pad), visual...)
	}

	parts := []string{strings.Join(visual, "\n"), separator}
	if statusLine != "" {
		parts = append(parts, statusLine)
	}
	if acBlock != "" {
		parts = append(parts, acBlock)
	}
	parts = append(parts, inputLine)
	return strings.Join(parts, "\n")
}

// renderOutputLines word-wraps all output (including the in-progress response
// buffer) and returns the trailing `cap` visual rows.
func (m SessionModel) renderOutputLines(cap int) []string {
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

	if len(visual) > cap {
		visual = visual[len(visual)-cap:]
	}
	return visual
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
		m.output = append(m.output, styleMuted.Render("  "+ev.Text))
	case EventWarning:
		m.output = append(m.output, styleWarning.Render("⚠ "+ev.Text))
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
	}
}

func (m *SessionModel) appendUser(text string) {
	prefix := styleUser.Render("you   ")
	m.output = append(m.output, prefix+text)
}

func (m *SessionModel) appendAssistant(text string) {
	prefix := styleAssistant.Render("setu  ")
	// Preserve internal newlines.
	lines := strings.Split(strings.TrimRight(text, "\n"), "\n")
	for i, line := range lines {
		if i == 0 {
			m.output = append(m.output, prefix+line)
		} else {
			m.output = append(m.output, "      "+line)
		}
	}
	m.output = append(m.output, "")
}

func (m *SessionModel) appendError(msg string) {
	m.output = append(m.output, styleError.Render("error: ")+msg)
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
	for _, line := range lines {
		m.output = append(m.output, line)
	}
	m.output = append(m.output, "")
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
