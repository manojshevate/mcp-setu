package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Print the full view for a 40x12 terminal, so we can eyeball the layout.
func TestSessionViewNarrowSnapshot(t *testing.T) {
	m := newTestSession(t)
	mUpdated, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 12})
	m = mUpdated.(*SessionModel)
	out := m.View()
	lines := strings.Split(out, "\n")
	for i, l := range lines {
		t.Logf("L%02d (cols=%d): %q", i, len([]rune(l)), l)
	}
}

// Empty model: footer should be suppressed.
func TestSessionViewEmptyModel(t *testing.T) {
	m := newTestSession(t)
	m.model = ""
	mUpdated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = mUpdated.(*SessionModel)
	out := m.View()
	lines := strings.Split(out, "\n")
	if len(lines) != 24 {
		t.Errorf("got %d lines, want 24", len(lines))
	}
	// Last line should NOT contain the model badge glyph.
	if strings.Contains(lines[len(lines)-1], "◆") {
		t.Errorf("expected no model badge when model is empty, got: %q", lines[len(lines)-1])
	}
}
