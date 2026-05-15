package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// inputMode controls what the input widget is doing.
type inputMode int

const (
	modeNormal       inputMode = iota // normal text entry
	modeAutocomplete                  // slash-command autocomplete dropdown visible
	modeModelSelect                   // interactive model picker (no text entry)
)

// SendMsg is a message fired when the user confirms their input for sending
// (Ctrl+Enter or Cmd+Enter in multiline mode, or Enter in single-line mode
// when multiline is disabled).
type SendMsg struct{}

type InputModel struct {
	// valueRunes holds the input text as a rune slice so that all cursor
	// arithmetic operates on Unicode code-points, not bytes. This prevents
	// corruption of any multibyte character (é, →, 世界, emoji, …).
	// External callers always receive a plain string via GetValue().
	valueRunes   []rune
	cursor       int
	width        int
	history      []string
	historyIdx   int
	autocomplete []string
	selectedAC   int
	mode         inputMode

	// model picker state
	models       []string // names of locally available models
	currentModel string   // the currently active model name
	pickerIdx    int      // selected row in picker
}

func NewInputModel() InputModel {
	return InputModel{
		valueRunes: []rune{},
		history:    make([]string, 0),
		historyIdx: -1,
	}
}

func (m InputModel) Init() tea.Cmd {
	return nil
}

func (m InputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyUp:
			switch m.mode {
			case modeAutocomplete:
				if m.selectedAC > 0 {
					m.selectedAC--
				}
			case modeModelSelect:
				if m.pickerIdx > 0 {
					m.pickerIdx--
				}
			default:
				if m.historyIdx < len(m.history)-1 {
					m.historyIdx++
					if m.historyIdx < len(m.history) {
						m.valueRunes = []rune(m.history[len(m.history)-1-m.historyIdx])
						m.cursor = len(m.valueRunes)
					}
				}
			}

		case tea.KeyDown:
			switch m.mode {
			case modeAutocomplete:
				if m.selectedAC < len(m.autocomplete)-1 {
					m.selectedAC++
				}
			case modeModelSelect:
				if m.pickerIdx < len(m.models)-1 {
					m.pickerIdx++
				}
			default:
				if m.historyIdx > 0 {
					m.historyIdx--
					m.valueRunes = []rune(m.history[len(m.history)-1-m.historyIdx])
					m.cursor = len(m.valueRunes)
				} else if m.historyIdx == 0 {
					m.historyIdx = -1
					m.valueRunes = []rune{}
					m.cursor = 0
				}
			}

		case tea.KeyEnter:
			// Plain Enter sends the message (default chat behaviour).
			// Alt+Enter (Option+Enter on macOS) inserts a newline for multiline input.
			// In model picker mode this case is never reached because session.go
			// handles tea.KeyEnter directly before delegating to InputModel.Update.
			if m.mode == modeModelSelect {
				// Guard: shouldn't happen, but be safe.
			} else if msg.Alt {
				// Alt+Enter → insert newline for multiline input.
				if m.cursor < 0 || m.cursor > len(m.valueRunes) {
					m.cursor = len(m.valueRunes)
				}
				newRunes := make([]rune, 0, len(m.valueRunes)+1)
				newRunes = append(newRunes, m.valueRunes[:m.cursor]...)
				newRunes = append(newRunes, '\n')
				newRunes = append(newRunes, m.valueRunes[m.cursor:]...)
				m.valueRunes = newRunes
				m.cursor++
				// Newlines cancel any active autocomplete.
				m.mode = modeNormal
				m.autocomplete = nil
			} else {
				// Plain Enter → send.
				return m, func() tea.Msg { return SendMsg{} }
			}

		case tea.KeyBackspace:
			if m.mode == modeModelSelect {
				m.ExitModelSelect()
			} else if m.cursor > 0 {
				newRunes := make([]rune, 0, len(m.valueRunes)-1)
				newRunes = append(newRunes, m.valueRunes[:m.cursor-1]...)
				newRunes = append(newRunes, m.valueRunes[m.cursor:]...)
				m.valueRunes = newRunes
				m.cursor--
				m.updateAC()
			}

		case tea.KeyDelete:
			if m.mode == modeModelSelect {
				m.ExitModelSelect()
			} else if m.cursor < len(m.valueRunes) {
				newRunes := make([]rune, 0, len(m.valueRunes)-1)
				newRunes = append(newRunes, m.valueRunes[:m.cursor]...)
				newRunes = append(newRunes, m.valueRunes[m.cursor+1:]...)
				m.valueRunes = newRunes
				m.updateAC()
			}

		case tea.KeyTab:
			if m.mode == modeAutocomplete && len(m.autocomplete) > 0 {
				m.valueRunes = []rune(m.autocomplete[m.selectedAC])
				m.cursor = len(m.valueRunes)
				m.mode = modeNormal
			}

		case tea.KeySpace:
			if m.mode == modeModelSelect {
				// Ignore space in picker mode.
			} else {
				if m.cursor < 0 || m.cursor > len(m.valueRunes) {
					m.cursor = len(m.valueRunes)
				}
				newRunes := make([]rune, 0, len(m.valueRunes)+1)
				newRunes = append(newRunes, m.valueRunes[:m.cursor]...)
				newRunes = append(newRunes, ' ')
				newRunes = append(newRunes, m.valueRunes[m.cursor:]...)
				m.valueRunes = newRunes
				m.cursor++
				m.updateAC()
			}

		case tea.KeyRunes:
			if m.mode == modeModelSelect {
				// Any typing exits the picker and starts fresh.
				m.ExitModelSelect()
				for _, r := range msg.Runes {
					if m.cursor < 0 || m.cursor > len(m.valueRunes) {
						m.cursor = len(m.valueRunes)
					}
					newRunes := make([]rune, 0, len(m.valueRunes)+1)
					newRunes = append(newRunes, m.valueRunes[:m.cursor]...)
					newRunes = append(newRunes, r)
					newRunes = append(newRunes, m.valueRunes[m.cursor:]...)
					m.valueRunes = newRunes
					m.cursor++
				}
				m.updateAC()
			} else {
				for _, r := range msg.Runes {
					if m.cursor < 0 || m.cursor > len(m.valueRunes) {
						m.cursor = len(m.valueRunes)
					}
					newRunes := make([]rune, 0, len(m.valueRunes)+1)
					newRunes = append(newRunes, m.valueRunes[:m.cursor]...)
					newRunes = append(newRunes, r)
					newRunes = append(newRunes, m.valueRunes[m.cursor:]...)
					m.valueRunes = newRunes
					m.cursor++
				}
				m.updateAC()
			}
		}
	}
	return m, nil
}

func (m *InputModel) updateAC() {
	val := string(m.valueRunes)
	if strings.HasPrefix(val, "/model ") {
		// Subcommand autocomplete: match model names after "/model ".
		prefix := strings.TrimPrefix(val, "/model ")
		m.autocomplete = getMatchingModels(prefix, m.models)
		if len(m.autocomplete) > 0 {
			m.mode = modeAutocomplete
		} else {
			m.mode = modeNormal
		}
		m.selectedAC = 0
		return
	}
	if strings.HasPrefix(val, "/") {
		m.autocomplete = getMatchingCommands(val)
		if len(m.autocomplete) > 0 {
			m.mode = modeAutocomplete
		} else {
			m.mode = modeNormal
		}
		m.selectedAC = 0
	} else {
		m.mode = modeNormal
		m.autocomplete = nil
	}
}

// LineCount returns the number of lines in the current value (1 for single-line).
func (m InputModel) LineCount() int {
	if len(m.valueRunes) == 0 {
		return 1
	}
	count := 1
	for _, r := range m.valueRunes {
		if r == '\n' {
			count++
		}
	}
	return count
}

// RenderLine returns just the input line (prompt cursor, placeholder, etc.)
// without any autocomplete or picker overlay.
//
// For multiline content, it renders the first line of text and appends a muted
// "(N lines)" indicator so the user knows there are additional lines.
func (m InputModel) RenderLine() string {
	if m.mode == modeModelSelect {
		return lipgloss.NewStyle().Faint(true).Render("↑/↓ select model · enter confirm · esc cancel")
	}

	cursorBlock := lipgloss.NewStyle().Background(lipgloss.Color("240")).Render(" ")

	if len(m.valueRunes) == 0 {
		// Empty input: show cursor at start with placeholder.
		hint := lipgloss.NewStyle().Faint(true).Render("type message… enter to send · alt+enter for newline")
		return cursorBlock + hint
	}

	// Clamp cursor to valid bounds to prevent out-of-bounds panics.
	cur := m.cursor
	if cur < 0 {
		cur = 0
	}
	if cur > len(m.valueRunes) {
		cur = len(m.valueRunes)
	}

	lineCount := m.LineCount()
	if lineCount <= 1 {
		// Single-line: show the full value with cursor.
		before := string(m.valueRunes[:cur])
		after := string(m.valueRunes[cur:])
		return before + cursorBlock + after
	}

	// Multiline: show only the first line and a "(N lines · alt+enter to send)" hint.
	// Find the first newline in rune space.
	firstLineEnd := len(m.valueRunes)
	for i, r := range m.valueRunes {
		if r == '\n' {
			firstLineEnd = i
			break
		}
	}
	firstLineRunes := m.valueRunes[:firstLineEnd]
	indicator := lipgloss.NewStyle().Faint(true).Render(
		fmt.Sprintf(" (%s · alt+enter for newline)", pluralLines(lineCount)),
	)

	// Render cursor within the first line if it falls there, otherwise at the end.
	if cur <= firstLineEnd {
		before := string(firstLineRunes[:cur])
		after := string(firstLineRunes[cur:])
		return before + cursorBlock + after + indicator
	}
	return string(firstLineRunes) + indicator
}

// pluralLines returns "N line" or "N lines".
func pluralLines(n int) string {
	if n == 1 {
		return "1 line"
	}
	return fmt.Sprintf("%d lines", n)
}

// RenderAutocomplete returns the autocomplete/picker overlay lines (without a
// trailing newline). Returns an empty string when nothing should be shown.
func (m InputModel) RenderAutocomplete() string {
	switch m.mode {
	case modeAutocomplete:
		if len(m.autocomplete) == 0 {
			return ""
		}
		var acs []string
		for i, cmd := range m.autocomplete {
			if i == m.selectedAC {
				acs = append(acs, lipgloss.NewStyle().
					Background(lipgloss.Color("63")).
					Render("  "+cmd+"  "))
			} else {
				acs = append(acs, "  "+cmd)
			}
		}
		return strings.Join(acs, "\n")

	case modeModelSelect:
		if len(m.models) == 0 {
			return lipgloss.NewStyle().Faint(true).Render("  (no local models found)")
		}
		var rows []string
		for i, name := range m.models {
			label := "  " + name
			if name == m.currentModel {
				label += lipgloss.NewStyle().Faint(true).Render("  (current)")
			}
			if i == m.pickerIdx {
				rows = append(rows, lipgloss.NewStyle().
					Background(lipgloss.Color("63")).
					Render(label+"  "))
			} else {
				rows = append(rows, label)
			}
		}
		return strings.Join(rows, "\n")
	}
	return ""
}

// View satisfies tea.Model. It combines RenderLine and RenderAutocomplete
// for backwards compatibility and ensures the input model is a valid Bubble Tea model.
func (m InputModel) View() string {
	line := m.RenderLine()
	ac := m.RenderAutocomplete()
	if ac != "" {
		return line + "\n" + ac
	}
	return line
}

// Mode returns the current input mode.
func (m InputModel) Mode() inputMode {
	return m.mode
}

// DismissAutocomplete dismisses the autocomplete dropdown without clearing the input.
func (m *InputModel) DismissAutocomplete() {
	if m.mode == modeAutocomplete {
		m.mode = modeNormal
	}
}

// EnterModelSelect switches into the interactive model picker.
func (m *InputModel) EnterModelSelect() {
	m.mode = modeModelSelect
	m.pickerIdx = 0
	// Pre-select the current model.
	for i, name := range m.models {
		if name == m.currentModel {
			m.pickerIdx = i
			break
		}
	}
}

// ExitModelSelect cancels the picker and returns to normal mode.
func (m *InputModel) ExitModelSelect() {
	m.mode = modeNormal
	m.valueRunes = []rune{}
	m.cursor = 0
}

// SelectedModel returns the model name currently highlighted in the picker.
// Returns an empty string when not in picker mode.
func (m InputModel) SelectedModel() string {
	if m.mode != modeModelSelect || len(m.models) == 0 {
		return ""
	}
	if m.pickerIdx >= 0 && m.pickerIdx < len(m.models) {
		return m.models[m.pickerIdx]
	}
	return ""
}

// SetModels updates the available model list for autocomplete and the picker.
func (m *InputModel) SetModels(names []string) {
	m.models = names
}

// SetCurrentModel records the currently active model name (for the picker annotation).
func (m *InputModel) SetCurrentModel(name string) {
	m.currentModel = name
}

// GetValue returns the current input value with leading/trailing whitespace trimmed.
// Internal newlines (multiline content) are preserved.
func (m InputModel) GetValue() string {
	return strings.Trim(string(m.valueRunes), " \t\r\n")
}

func (m *InputModel) Clear() {
	m.valueRunes = []rune{}
	m.cursor = 0
	m.historyIdx = -1
	m.mode = modeNormal
	m.autocomplete = nil
	m.selectedAC = 0
	m.pickerIdx = 0
}

func (m *InputModel) SetWidth(w int) {
	m.width = w
}

func (m *InputModel) AddToHistory(s string) {
	if s != "" {
		m.history = append(m.history, s)
		m.historyIdx = -1
	}
}

func getMatchingCommands(prefix string) []string {
	all := []string{"/tools", "/clear", "/model", "/stats", "/servers", "/help", "/quit", "/exit"}
	var matches []string
	for _, cmd := range all {
		if strings.HasPrefix(cmd, prefix) {
			matches = append(matches, cmd)
		}
	}
	return matches
}

// getMatchingModels returns model names that start with the given prefix.
func getMatchingModels(prefix string, models []string) []string {
	var matches []string
	for _, name := range models {
		if strings.HasPrefix(name, prefix) {
			matches = append(matches, "/model "+name)
		}
	}
	return matches
}
