package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

// Verify what happens when the user types more text than fits in one row of
// the input box. The box may grow vertically, breaking the chromeHeight
// invariant (1 + 3 + acLines + footer).
func TestSessionViewLongInput(t *testing.T) {
	m := newTestSession(t)
	mUpdated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = mUpdated.(*SessionModel)

	// Type a really long string (90 chars).
	for _, r := range strings.Repeat("x", 90) {
		mu, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = mu.(*SessionModel)
	}

	out := m.View()
	lines := strings.Split(out, "\n")
	t.Logf("total lines=%d (want 24)", len(lines))
	for i := len(lines) - 8; i < len(lines); i++ {
		if i >= 0 {
			t.Logf("  L%d: %q", i, lines[i])
		}
	}
	if len(lines) != 24 {
		t.Errorf("invariant broken: total lines=%d, want 24", len(lines))
	}
}

// Check the box width at empty input — does it fill the terminal width?
func TestSessionViewBoxFillsWidth(t *testing.T) {
	m := newTestSession(t)
	mUpdated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = mUpdated.(*SessionModel)

	out := m.View()
	lines := strings.Split(out, "\n")
	for i := len(lines) - 5; i < len(lines); i++ {
		if strings.Contains(lines[i], "╭") || strings.Contains(lines[i], "╰") {
			t.Logf("border line L%d (rune count=%d): %q", i, len([]rune(lines[i])), lines[i])
		}
	}
}

// Long model name should be truncated.
func TestTruncateModelLong(t *testing.T) {
	got := truncateModel("very-long-model-name-that-should-be-truncated-aaaaaaaaaa", 80)
	if !strings.Contains(got, "…") {
		t.Errorf("expected ellipsis in truncated name, got %q", got)
	}
	if len(got) > 30 {
		t.Errorf("expected truncation to ~1/3 of width (~26 chars), got %d chars: %q", len(got), got)
	}
}
