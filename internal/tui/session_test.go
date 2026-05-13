package tui

import (
	"context"
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
