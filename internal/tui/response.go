package tui

import (
	"github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type ResponseModel struct {
	content string
	width   int
	height  int
	scroll  int
}

func NewResponseModel() ResponseModel {
	return ResponseModel{}
}

func (m ResponseModel) Init() tea.Cmd {
	return nil
}

func (m ResponseModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	return m, nil
}

func (m ResponseModel) View() string {
	if m.height == 0 || m.width == 0 {
		return ""
	}

	// Render content with word wrap
	style := lipgloss.NewStyle().
		Width(m.width).
		Height(m.height).
		Foreground(lipgloss.Color("255"))

	// For now, just return the content styled
	// A full implementation would handle scrolling and viewport
	return style.Render(m.content)
}

func (m *ResponseModel) AddChunk(chunk string) {
	m.content += chunk
}

func (m *ResponseModel) Clear() {
	m.content = ""
	m.scroll = 0
}

func (m *ResponseModel) SetSize(width, height int) {
	m.width = width
	m.height = height
}

func (m ResponseModel) GetContent() string {
	return m.content
}
