package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type InputModel struct {
	value        string
	cursor       int
	width        int
	history      []string
	historyIdx   int
	autocomplete []string
	selectedAC   int
	showAC       bool
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
			if m.showAC && m.selectedAC > 0 {
				m.selectedAC--
			} else if !m.showAC && m.historyIdx < len(m.history)-1 {
				m.historyIdx++
				if m.historyIdx < len(m.history) {
					m.value = m.history[len(m.history)-1-m.historyIdx]
					m.cursor = len(m.value)
				}
			}

		case tea.KeyDown:
			if m.showAC && m.selectedAC < len(m.autocomplete)-1 {
				m.selectedAC++
			} else if !m.showAC && m.historyIdx > 0 {
				m.historyIdx--
				m.value = m.history[len(m.history)-1-m.historyIdx]
				m.cursor = len(m.value)
			} else if m.historyIdx == 0 {
				m.historyIdx = -1
				m.value = ""
				m.cursor = 0
			}

		case tea.KeyBackspace:
			if m.cursor > 0 {
				m.value = m.value[:m.cursor-1] + m.value[m.cursor:]
				m.cursor--
				m.updateAC()
			}

		case tea.KeyDelete:
			if m.cursor < len(m.value) {
				m.value = m.value[:m.cursor] + m.value[m.cursor+1:]
				m.updateAC()
			}

		case tea.KeyTab:
			if m.showAC && len(m.autocomplete) > 0 {
				m.value = m.autocomplete[m.selectedAC]
				m.cursor = len(m.value)
				m.showAC = false
			}

		case tea.KeySpace:
			m.value = m.value[:m.cursor] + " " + m.value[m.cursor:]
			m.cursor++
			m.updateAC()

		case tea.KeyRunes:
			for _, r := range msg.Runes {
				m.value = m.value[:m.cursor] + string(r) + m.value[m.cursor:]
				m.cursor++
			}
			m.updateAC()
		}
	}
	return m, nil
}

func (m *InputModel) updateAC() {
	if strings.HasPrefix(m.value, "/") {
		m.autocomplete = getMatchingCommands(m.value)
		m.showAC = len(m.autocomplete) > 0
		m.selectedAC = 0
	} else {
		m.showAC = false
		m.autocomplete = nil
	}
}

func (m InputModel) View() string {
	display := m.value
	if display == "" {
		display = lipgloss.NewStyle().Faint(true).Render("type message...")
	}

	line := display
	if m.cursor >= len(m.value) {
		line += lipgloss.NewStyle().Background(lipgloss.Color("240")).Render(" ")
	}

	if m.showAC && len(m.autocomplete) > 0 {
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
		return line + "\n" + strings.Join(acs, "\n")
	}
	return line
}

func (m InputModel) GetValue() string {
	return strings.TrimSpace(m.value)
}

func (m *InputModel) Clear() {
	m.value = ""
	m.cursor = 0
	m.historyIdx = -1
	m.showAC = false
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
	all := []string{"/tools", "/clear", "/model", "/stats", "/servers", "/help"}
	var matches []string
	for _, cmd := range all {
		if strings.HasPrefix(cmd, prefix) {
			matches = append(matches, cmd)
		}
	}
	return matches
}
