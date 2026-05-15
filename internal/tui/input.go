package tui

import (
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

type InputModel struct {
	value        string
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
						m.value = m.history[len(m.history)-1-m.historyIdx]
						m.cursor = len(m.value)
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
					m.value = m.history[len(m.history)-1-m.historyIdx]
					m.cursor = len(m.value)
				} else if m.historyIdx == 0 {
					m.historyIdx = -1
					m.value = ""
					m.cursor = 0
				}
			}

		case tea.KeyBackspace:
			if m.mode == modeModelSelect {
				m.ExitModelSelect()
			} else if m.cursor > 0 {
				m.value = m.value[:m.cursor-1] + m.value[m.cursor:]
				m.cursor--
				m.updateAC()
			}

		case tea.KeyDelete:
			if m.mode == modeModelSelect {
				m.ExitModelSelect()
			} else if m.cursor < len(m.value) {
				m.value = m.value[:m.cursor] + m.value[m.cursor+1:]
				m.updateAC()
			}

		case tea.KeyTab:
			if m.mode == modeAutocomplete && len(m.autocomplete) > 0 {
				m.value = m.autocomplete[m.selectedAC]
				m.cursor = len(m.value)
				m.mode = modeNormal
			}

		case tea.KeySpace:
			if m.mode == modeModelSelect {
				// Ignore space in picker mode.
			} else {
				if m.cursor < 0 || m.cursor > len(m.value) {
					m.cursor = len(m.value)
				}
				m.value = m.value[:m.cursor] + " " + m.value[m.cursor:]
				m.cursor++
				m.updateAC()
			}

		case tea.KeyRunes:
			if m.mode == modeModelSelect {
				// Any typing exits the picker and starts fresh.
				m.ExitModelSelect()
				for _, r := range msg.Runes {
					if m.cursor < 0 || m.cursor > len(m.value) {
						m.cursor = len(m.value)
					}
					m.value = m.value[:m.cursor] + string(r) + m.value[m.cursor:]
					m.cursor++
				}
				m.updateAC()
			} else {
				for _, r := range msg.Runes {
					if m.cursor < 0 || m.cursor > len(m.value) {
						m.cursor = len(m.value)
					}
					m.value = m.value[:m.cursor] + string(r) + m.value[m.cursor:]
					m.cursor++
				}
				m.updateAC()
			}
		}
	}
	return m, nil
}

func (m *InputModel) updateAC() {
	if strings.HasPrefix(m.value, "/model ") {
		// Subcommand autocomplete: match model names after "/model ".
		prefix := strings.TrimPrefix(m.value, "/model ")
		m.autocomplete = getMatchingModels(prefix, m.models)
		if len(m.autocomplete) > 0 {
			m.mode = modeAutocomplete
		} else {
			m.mode = modeNormal
		}
		m.selectedAC = 0
		return
	}
	if strings.HasPrefix(m.value, "/") {
		m.autocomplete = getMatchingCommands(m.value)
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

// RenderLine returns just the input line (prompt cursor, placeholder, etc.)
// without any autocomplete or picker overlay.
func (m InputModel) RenderLine() string {
	if m.mode == modeModelSelect {
		return lipgloss.NewStyle().Faint(true).Render("↑/↓ select model · enter confirm · esc cancel")
	}

	cursorBlock := lipgloss.NewStyle().Background(lipgloss.Color("240")).Render(" ")

	if m.value == "" {
		// Empty input: show cursor at start with placeholder
		return cursorBlock + lipgloss.NewStyle().Faint(true).Render("type message...")
	}

	// Non-empty input: position cursor at current location.
	// Clamp cursor to valid bounds to prevent out-of-bounds panics.
	cur := m.cursor
	if cur < 0 {
		cur = 0
	}
	if cur > len(m.value) {
		cur = len(m.value)
	}
	before := m.value[:cur]
	after := m.value[cur:]
	return before + cursorBlock + after
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
	m.value = ""
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

func (m InputModel) GetValue() string {
	return strings.TrimSpace(m.value)
}

func (m *InputModel) Clear() {
	m.value = ""
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
