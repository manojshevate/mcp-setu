package tui

import (
	"strings"

	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type InputModel struct {
	value          string
	cursor         int
	width          int
	history        []string
	historyIndex   int
	autocomplete   []string
	selectedAC     int
	showAutocomplete bool
}

func NewInputModel() InputModel {
	return InputModel{
		history:      make([]string, 0),
		autocomplete: make([]string, 0),
		historyIndex: -1,
	}
}

func (m InputModel) Init() tea.Cmd {
	return nil
}

func (m InputModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		return m.handleKey(msg)
	}
	return m, nil
}

func (m InputModel) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyUp:
		// Navigate autocomplete up or history up
		if m.showAutocomplete && m.selectedAC > 0 {
			m.selectedAC--
		} else if !m.showAutocomplete && m.historyIndex < len(m.history)-1 {
			m.historyIndex++
			if m.historyIndex < len(m.history) {
				m.value = m.history[len(m.history)-1-m.historyIndex]
				m.cursor = len(m.value)
			}
		}

	case tea.KeyDown:
		// Navigate autocomplete down or history down
		if m.showAutocomplete && m.selectedAC < len(m.autocomplete)-1 {
			m.selectedAC++
		} else if !m.showAutocomplete {
			if m.historyIndex > 0 {
				m.historyIndex--
				m.value = m.history[len(m.history)-1-m.historyIndex]
				m.cursor = len(m.value)
			} else if m.historyIndex == 0 {
				m.historyIndex = -1
				m.value = ""
				m.cursor = 0
			}
		}

	case tea.KeyLeft:
		if m.cursor > 0 {
			m.cursor--
		}
		m.updateAutocomplete()

	case tea.KeyRight:
		if m.cursor < len(m.value) {
			m.cursor++
		}
		m.updateAutocomplete()

	case tea.KeyCtrlA:
		m.cursor = 0

	case tea.KeyCtrlE:
		m.cursor = len(m.value)

	case tea.KeyCtrlU:
		m.value = ""
		m.cursor = 0
		m.showAutocomplete = false

	case tea.KeyBackspace:
		if m.cursor > 0 {
			m.value = m.value[:m.cursor-1] + m.value[m.cursor:]
			m.cursor--
		}
		m.updateAutocomplete()

	case tea.KeyDelete:
		if m.cursor < len(m.value) {
			m.value = m.value[:m.cursor] + m.value[m.cursor+1:]
		}
		m.updateAutocomplete()

	case tea.KeyTab:
		if m.showAutocomplete && len(m.autocomplete) > 0 {
			m.value = m.autocomplete[m.selectedAC]
			m.cursor = len(m.value)
			m.showAutocomplete = false
		}

	case tea.KeyRunes:
		// Insert characters
		for _, r := range msg.Runes {
			if m.cursor <= len(m.value) {
				m.value = m.value[:m.cursor] + string(r) + m.value[m.cursor:]
				m.cursor++
			}
		}
		m.updateAutocomplete()
	}

	return m, nil
}

func (m *InputModel) updateAutocomplete() {
	// Only show autocomplete for commands starting with /
	if strings.HasPrefix(m.value, "/") {
		m.autocomplete = getMatchingCommands(m.value)
		m.showAutocomplete = len(m.autocomplete) > 0
		m.selectedAC = 0
	} else {
		m.showAutocomplete = false
		m.autocomplete = nil
	}
}

func (m InputModel) View() string {
	// Input prompt
	promptStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("63"))
	inputStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("255"))

	// Visible part of input
	inputDisplay := m.value
	if len(inputDisplay) > m.width-4 {
		startIdx := len(inputDisplay) - (m.width - 5)
		if startIdx < 0 {
			startIdx = 0
		}
		if startIdx > len(inputDisplay) {
			startIdx = len(inputDisplay)
		}
		inputDisplay = "…" + inputDisplay[startIdx:]
	}

	// Build the input line
	line := promptStyle.Render("❯ ") + inputStyle.Render(inputDisplay)

	// Add cursor indicator
	if m.cursor >= len(m.value) {
		line += lipgloss.NewStyle().Background(lipgloss.Color("240")).Render(" ")
	}

	// Add autocomplete suggestions if applicable
	if m.showAutocomplete && len(m.autocomplete) > 0 {
		acView := m.renderAutocomplete()
		return lipgloss.JoinVertical(lipgloss.Top, line, acView)
	}

	return line
}

func (m InputModel) renderAutocomplete() string {
	suggestionStyle := lipgloss.NewStyle().Foreground(lipgloss.Color("240"))
	selectedStyle := lipgloss.NewStyle().Background(lipgloss.Color("63")).Foreground(lipgloss.Color("255"))

	var suggestions []string
	for i, cmd := range m.autocomplete {
		if i == m.selectedAC {
			suggestions = append(suggestions, selectedStyle.Render("  → "+cmd))
		} else {
			suggestions = append(suggestions, suggestionStyle.Render("    "+cmd))
		}
	}

	return lipgloss.JoinVertical(lipgloss.Left, suggestions...)
}

func (m InputModel) SetWidth(width int) {
	m.width = width
}

func (m InputModel) GetValue() string {
	return strings.TrimSpace(m.value)
}

func (m *InputModel) Clear() {
	m.value = ""
	m.cursor = 0
	m.historyIndex = -1
	m.showAutocomplete = false
}

func (m *InputModel) AddToHistory(input string) {
	if input != "" {
		m.history = append(m.history, input)
		m.historyIndex = -1
	}
}

func getMatchingCommands(prefix string) []string {
	commands := []string{
		"/tools",
		"/clear",
		"/model",
		"/stats",
		"/servers",
		"/help",
	}

	var matches []string
	for _, cmd := range commands {
		if strings.HasPrefix(cmd, prefix) {
			matches = append(matches, cmd)
		}
	}
	return matches
}
