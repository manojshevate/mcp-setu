package tui

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/manojshevate/mcp-setu/internal/bridge"
	"github.com/manojshevate/mcp-setu/internal/mcp"
)

func newTestSession(t *testing.T) *SessionModel {
	t.Helper()
	mcpClient := mcp.NewMultiClient()
	br := bridge.NewBridge(nil, mcpClient, "test-model", 0.7, 4096, nil)
	m := NewSessionModel(context.Background(), br, mcpClient, nil, "test-model", "system prompt", false)
	return &m
}

func TestSessionViewRendersWithoutSize(t *testing.T) {
	m := newTestSession(t)
	if got := m.View(); got != "" {
		t.Fatalf("expected empty view without window size, got %q", got)
	}
}

func TestSessionViewRendersWithSize(t *testing.T) {
	m := newTestSession(t)
	mUpdated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = mUpdated.(*SessionModel)
	out := m.View()
	if out == "" {
		t.Fatal("expected non-empty view")
	}
	if !strings.Contains(out, "Welcome to mcp-setu") {
		t.Errorf("expected banner in output, got: %s", out)
	}
	if !strings.Contains(out, "❯") {
		t.Errorf("expected input prompt at bottom, got: %s", out)
	}
	lines := strings.Split(out, "\n")
	if len(lines) != 24 {
		t.Errorf("expected exactly 24 lines (terminal height), got %d", len(lines))
	}
	// Input prompt should be on the third-from-last line (inside the border box):
	// lines[len-1] = model footer, lines[len-2] = bottom border, lines[len-3] = input content line,
	// lines[len-4] = top border.
	if !strings.Contains(lines[len(lines)-3], "❯") {
		t.Errorf("expected ❯ prompt on third-from-last line (input box content), got: %q", lines[len(lines)-3])
	}
	// Model footer should appear on the last line.
	if !strings.Contains(lines[len(lines)-1], "◆") {
		t.Errorf("expected model indicator (◆) on last line, got: %q", lines[len(lines)-1])
	}
	// At least one line should contain a border glyph from the rounded border.
	hasBorder := false
	for _, l := range lines {
		if strings.ContainsAny(l, "╭╮╰╯") {
			hasBorder = true
			break
		}
	}
	if !hasBorder {
		t.Errorf("expected at least one line with a rounded border glyph (╭╮╰╯), none found in:\n%s", out)
	}
}

func TestSessionViewNarrowTerminal(t *testing.T) {
	m := newTestSession(t)
	// Very narrow terminal — should not panic or produce wrong line count.
	mUpdated, _ := m.Update(tea.WindowSizeMsg{Width: 40, Height: 12})
	m = mUpdated.(*SessionModel)
	out := m.View()
	if out == "" {
		t.Fatal("expected non-empty view for narrow terminal")
	}
	lines := strings.Split(out, "\n")
	if len(lines) != 12 {
		t.Errorf("expected exactly 12 lines for narrow terminal, got %d", len(lines))
	}
}

func TestSessionEventHandling(t *testing.T) {
	m := newTestSession(t)
	mUpdated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = mUpdated.(*SessionModel)

	// Feed streaming response events.
	m.handleEvent(Event{Kind: EventRespStart})
	m.handleEvent(Event{Kind: EventRespChunk, Text: "Hello "})
	m.handleEvent(Event{Kind: EventRespChunk, Text: "world!"})
	m.handleEvent(Event{Kind: EventRespEnd})

	out := m.View()
	if !strings.Contains(out, "Hello world!") {
		t.Errorf("expected streamed response in output, got:\n%s", out)
	}
}

func TestSessionTickDoesNotPolluteOutput(t *testing.T) {
	m := newTestSession(t)
	mUpdated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = mUpdated.(*SessionModel)
	baseline := len(m.output)
	for i := 0; i < 50; i++ {
		mu, _ := m.Update(tickMsg{})
		m = mu.(*SessionModel)
	}
	if len(m.output) != baseline {
		t.Errorf("ticks polluted output: was %d, now %d", baseline, len(m.output))
	}
}

func TestFormatElapsed(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{0 * time.Second, "0s"},
		{1 * time.Second, "1s"},
		{59 * time.Second, "59s"},
		{60 * time.Second, "1m 0s"},
		{61 * time.Second, "1m 1s"},
		{123 * time.Second, "2m 3s"},
		{3600 * time.Second, "60m 0s"},
		{-1 * time.Second, "0s"},
	}
	for _, tc := range cases {
		got := formatElapsed(tc.d)
		if got != tc.want {
			t.Errorf("formatElapsed(%v) = %q, want %q", tc.d, got, tc.want)
		}
	}
}

func TestSessionViewRendersHeader(t *testing.T) {
	m := newTestSession(t)
	// Terminal tall enough to show the header.
	mUpdated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	m = mUpdated.(*SessionModel)
	out := m.View()
	// The header should contain the ASCII art logo text.
	if !strings.Contains(out, "MCP-SETU") {
		t.Errorf("expected 'MCP-SETU' in header, got:\n%s", out)
	}
}

func TestSessionViewScrollPgUp(t *testing.T) {
	m := newTestSession(t)
	mUpdated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	m = mUpdated.(*SessionModel)

	// Add many output lines to make scrolling meaningful.
	for i := 0; i < 50; i++ {
		m.appendOutput(fmt.Sprintf("line %d", i))
	}

	// PgUp should increase scrollOffset.
	before := m.scrollOffset
	mu, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	m = mu.(*SessionModel)
	if m.scrollOffset <= before {
		t.Errorf("expected scrollOffset to increase after PgUp, got %d (was %d)", m.scrollOffset, before)
	}
}

func TestSessionViewScrollPgDown(t *testing.T) {
	m := newTestSession(t)
	mUpdated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	m = mUpdated.(*SessionModel)
	for i := 0; i < 50; i++ {
		m.appendOutput(fmt.Sprintf("line %d", i))
	}
	// Scroll up first.
	mu, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	m = mu.(*SessionModel)
	mu, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	m = mu.(*SessionModel)
	scrolled := m.scrollOffset

	// PgDown should decrease scrollOffset.
	mu, _ = m.Update(tea.KeyMsg{Type: tea.KeyPgDown})
	m = mu.(*SessionModel)
	if m.scrollOffset >= scrolled {
		t.Errorf("expected scrollOffset to decrease after PgDown, got %d (was %d)", m.scrollOffset, scrolled)
	}
}

func TestSessionViewScrollEnd(t *testing.T) {
	m := newTestSession(t)
	mUpdated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	m = mUpdated.(*SessionModel)
	for i := 0; i < 50; i++ {
		m.appendOutput(fmt.Sprintf("line %d", i))
	}
	// Scroll up.
	mu, _ := m.Update(tea.KeyMsg{Type: tea.KeyPgUp})
	m = mu.(*SessionModel)
	if m.scrollOffset == 0 {
		t.Skip("scrollOffset did not increase, skipping End test")
	}
	// End should reset scroll to 0.
	mu, _ = m.Update(tea.KeyMsg{Type: tea.KeyEnd})
	m = mu.(*SessionModel)
	if m.scrollOffset != 0 {
		t.Errorf("expected scrollOffset=0 after End, got %d", m.scrollOffset)
	}
}

func TestSessionViewScrollHome(t *testing.T) {
	m := newTestSession(t)
	mUpdated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 30})
	m = mUpdated.(*SessionModel)
	for i := 0; i < 50; i++ {
		m.appendOutput(fmt.Sprintf("line %d", i))
	}
	// Home should set scrollOffset to a large value (top of history).
	mu, _ := m.Update(tea.KeyMsg{Type: tea.KeyHome})
	m = mu.(*SessionModel)
	if m.scrollOffset <= 0 {
		t.Errorf("expected scrollOffset > 0 after Home, got %d", m.scrollOffset)
	}
}

func TestSessionSendMsg(t *testing.T) {
	m := newTestSession(t)
	mUpdated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = mUpdated.(*SessionModel)

	// Type some text.
	for _, r := range "hello world" {
		mu, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = mu.(*SessionModel)
	}
	if m.input.GetValue() != "hello world" {
		t.Fatalf("expected 'hello world' in input, got %q", m.input.GetValue())
	}

	// SendMsg should dispatch the input (clears the buffer and queues bridge).
	// We just check the input is cleared; the bridge call errors gracefully.
	mu, _ := m.Update(SendMsg{})
	m = mu.(*SessionModel)
	// After send, input is cleared.
	if m.input.GetValue() != "" {
		t.Errorf("expected empty input after SendMsg, got %q", m.input.GetValue())
	}
}

func TestSessionMultilineInput(t *testing.T) {
	m := newTestSession(t)
	mUpdated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = mUpdated.(*SessionModel)

	// Type "line1", press Enter, type "line2".
	for _, r := range "line1" {
		mu, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = mu.(*SessionModel)
	}
	mu, _ := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	m = mu.(*SessionModel)
	for _, r := range "line2" {
		mu, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = mu.(*SessionModel)
	}

	v := m.input.GetValue()
	if !strings.Contains(v, "\n") {
		t.Errorf("expected newline in multiline input value, got %q", v)
	}
	if m.input.LineCount() != 2 {
		t.Errorf("expected 2 lines in input, got %d", m.input.LineCount())
	}
}

func TestSessionStreamingResponseAccumulates(t *testing.T) {
	m := newTestSession(t)
	mUpdated, _ := m.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	m = mUpdated.(*SessionModel)
	m.handleEvent(Event{Kind: EventRespStart})
	for _, c := range []string{"a", "b", "c"} {
		m.handleEvent(Event{Kind: EventRespChunk, Text: c})
	}
	// Mid-stream: respBuf should contain "abc" and respActive should be true.
	if !m.respActive {
		t.Error("expected respActive true mid-stream")
	}
	if m.respBuf.String() != "abc" {
		t.Errorf("expected respBuf = abc, got %q", m.respBuf.String())
	}
	m.handleEvent(Event{Kind: EventRespEnd})
	if m.respActive {
		t.Error("expected respActive false after EventRespEnd")
	}
}
