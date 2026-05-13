package tui

import (
	"testing"
)

func TestInputModelView(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		width    int
		cursor   int
		shouldNotPanic bool
	}{
		{
			name:     "empty input",
			value:    "",
			width:    80,
			cursor:   0,
			shouldNotPanic: true,
		},
		{
			name:     "short input",
			value:    "hello",
			width:    80,
			cursor:   5,
			shouldNotPanic: true,
		},
		{
			name:     "long input with large width",
			value:    "this is a very long input that should fit in the terminal",
			width:    200,
			cursor:   20,
			shouldNotPanic: true,
		},
		{
			name:     "long input with small width",
			value:    "this is a very long input that does not fit in the terminal",
			width:    10,
			cursor:   30,
			shouldNotPanic: true,
		},
		{
			name:     "very small width",
			value:    "hello",
			width:    1,
			cursor:   3,
			shouldNotPanic: true,
		},
		{
			name:     "zero width",
			value:    "hello",
			width:    0,
			cursor:   2,
			shouldNotPanic: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := InputModel{
				value:  tt.value,
				width:  tt.width,
				cursor: tt.cursor,
			}

			// This should not panic
			view := m.View()
			if view == "" && tt.value != "" {
				t.Logf("view for input '%s': %s", tt.value, view)
			}
		})
	}
}

func TestInputModelDirectUpdates(t *testing.T) {
	m := NewInputModel()
	m.SetWidth(80)

	// Test GetValue
	m.value = "test"
	if m.GetValue() != "test" {
		t.Errorf("expected 'test', got '%s'", m.GetValue())
	}

	// Test Clear
	m.Clear()
	if m.value != "" {
		t.Errorf("expected empty value after clear, got '%s'", m.value)
	}

	// Test cursor position
	m.cursor = 5
	m.value = "hello world"
	if m.cursor != 5 {
		t.Errorf("expected cursor at 5, got %d", m.cursor)
	}
}

func TestInputModelHistory(t *testing.T) {
	m := NewInputModel()

	// Add to history
	m.AddToHistory("first command")
	m.AddToHistory("second command")

	if len(m.history) != 2 {
		t.Errorf("expected 2 history entries, got %d", len(m.history))
	}

	// Empty string should not be added
	m.AddToHistory("")
	if len(m.history) != 2 {
		t.Errorf("expected 2 history entries, got %d", len(m.history))
	}
}

func TestGetMatchingCommands(t *testing.T) {
	tests := []struct {
		prefix   string
		expected []string
	}{
		{"/", []string{"/tools", "/clear", "/model", "/stats", "/servers", "/help", "/quit", "/exit"}},
		{"/t", []string{"/tools"}},
		{"/m", []string{"/model"}},
		{"/s", []string{"/stats", "/servers"}},
		{"/c", []string{"/clear"}},
		{"/h", []string{"/help"}},
		{"/x", []string{}},
		{"/tool", []string{"/tools"}},
	}

	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			matches := getMatchingCommands(tt.prefix)
			if len(matches) != len(tt.expected) {
				t.Errorf("expected %d matches, got %d for prefix '%s'", len(tt.expected), len(matches), tt.prefix)
			}

			for i, m := range matches {
				if i < len(tt.expected) && m != tt.expected[i] {
					t.Errorf("expected '%s', got '%s'", tt.expected[i], m)
				}
			}
		})
	}
}

func TestInputModelUpdateAutocomplete(t *testing.T) {
	m := NewInputModel()

	// No autocomplete without /
	m.value = "hello"
	m.updateAC()
	if m.showAC {
		t.Error("expected no autocomplete for non-slash input")
	}

	// Autocomplete with /
	m.value = "/t"
	m.updateAC()
	if !m.showAC {
		t.Error("expected autocomplete for /t")
	}

	if len(m.autocomplete) == 0 {
		t.Error("expected autocomplete suggestions for /t")
	}

	// No autocomplete for non-matching prefix
	m.value = "/xyz"
	m.updateAC()
	if m.showAC && len(m.autocomplete) > 0 {
		t.Error("expected no autocomplete for /xyz")
	}
}
