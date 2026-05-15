package tui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestInputModelView(t *testing.T) {
	tests := []struct {
		name           string
		value          string
		width          int
		cursor         int
		shouldNotPanic bool
	}{
		{
			name:           "empty input",
			value:          "",
			width:          80,
			cursor:         0,
			shouldNotPanic: true,
		},
		{
			name:           "short input",
			value:          "hello",
			width:          80,
			cursor:         5,
			shouldNotPanic: true,
		},
		{
			name:           "long input with large width",
			value:          "this is a very long input that should fit in the terminal",
			width:          200,
			cursor:         20,
			shouldNotPanic: true,
		},
		{
			name:           "long input with small width",
			value:          "this is a very long input that does not fit in the terminal",
			width:          10,
			cursor:         30,
			shouldNotPanic: true,
		},
		{
			name:           "very small width",
			value:          "hello",
			width:          1,
			cursor:         3,
			shouldNotPanic: true,
		},
		{
			name:           "zero width",
			value:          "hello",
			width:          0,
			cursor:         2,
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
	if m.mode == modeAutocomplete {
		t.Error("expected no autocomplete for non-slash input")
	}

	// Autocomplete with /
	m.value = "/t"
	m.updateAC()
	if m.mode != modeAutocomplete {
		t.Error("expected autocomplete mode for /t")
	}
	if len(m.autocomplete) == 0 {
		t.Error("expected autocomplete suggestions for /t")
	}

	// No autocomplete for non-matching prefix
	m.value = "/xyz"
	m.updateAC()
	if m.mode == modeAutocomplete && len(m.autocomplete) > 0 {
		t.Error("expected no autocomplete for /xyz")
	}
}

func TestInputModelModeEnum(t *testing.T) {
	m := NewInputModel()

	// Starts in normal mode.
	if m.mode != modeNormal {
		t.Errorf("expected modeNormal at start, got %d", m.mode)
	}

	// Typing /m triggers autocomplete mode.
	m.value = "/m"
	m.updateAC()
	if m.mode != modeAutocomplete {
		t.Errorf("expected modeAutocomplete after /m, got %d", m.mode)
	}

	// Clearing the value returns to normal mode.
	m.value = "hello"
	m.updateAC()
	if m.mode != modeNormal {
		t.Errorf("expected modeNormal for non-slash input, got %d", m.mode)
	}
}

func TestInputModelModelSelect(t *testing.T) {
	m := NewInputModel()
	m.SetModels([]string{"llama3", "mistral", "phi3"})
	m.SetCurrentModel("llama3")

	// EnterModelSelect switches to picker mode.
	m.EnterModelSelect()
	if m.mode != modeModelSelect {
		t.Errorf("expected modeModelSelect after EnterModelSelect, got %d", m.mode)
	}
	// pickerIdx should be pre-set to the current model index.
	if m.pickerIdx != 0 {
		t.Errorf("expected pickerIdx=0 (llama3), got %d", m.pickerIdx)
	}

	// Navigate down.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(InputModel)
	if m.pickerIdx != 1 {
		t.Errorf("expected pickerIdx=1 after KeyDown, got %d", m.pickerIdx)
	}

	// Navigate down again.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(InputModel)
	if m.pickerIdx != 2 {
		t.Errorf("expected pickerIdx=2 after second KeyDown, got %d", m.pickerIdx)
	}

	// Navigate past the end — should clamp.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyDown})
	m = updated.(InputModel)
	if m.pickerIdx != 2 {
		t.Errorf("expected pickerIdx to clamp at 2, got %d", m.pickerIdx)
	}

	// SelectedModel returns the highlighted model.
	if got := m.SelectedModel(); got != "phi3" {
		t.Errorf("expected SelectedModel=phi3, got %q", got)
	}

	// Navigate back up.
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyUp})
	m = updated.(InputModel)
	if m.pickerIdx != 1 {
		t.Errorf("expected pickerIdx=1 after KeyUp, got %d", m.pickerIdx)
	}

	// ExitModelSelect resets to normal.
	m.ExitModelSelect()
	if m.mode != modeNormal {
		t.Errorf("expected modeNormal after ExitModelSelect, got %d", m.mode)
	}
	if m.value != "" {
		t.Errorf("expected empty value after ExitModelSelect, got %q", m.value)
	}
}

func TestInputModelSelectedModelEmpty(t *testing.T) {
	m := NewInputModel()
	// Not in picker mode — SelectedModel should return "".
	if got := m.SelectedModel(); got != "" {
		t.Errorf("expected empty SelectedModel when not in picker, got %q", got)
	}

	// Picker mode with no models loaded.
	m.mode = modeModelSelect
	if got := m.SelectedModel(); got != "" {
		t.Errorf("expected empty SelectedModel when models list is empty, got %q", got)
	}
}

func TestInputModelModelAutocomplete(t *testing.T) {
	m := NewInputModel()
	m.SetModels([]string{"llama3", "llama2", "mistral"})

	// Typing "/model " should trigger model-name autocomplete.
	m.value = "/model "
	m.updateAC()
	if m.mode != modeAutocomplete {
		t.Errorf("expected modeAutocomplete after '/model ', got %d", m.mode)
	}
	if len(m.autocomplete) == 0 {
		t.Error("expected autocomplete entries for '/model '")
	}

	// Typing "/model ll" should narrow to llama3, llama2.
	m.value = "/model ll"
	m.updateAC()
	if m.mode != modeAutocomplete {
		t.Errorf("expected modeAutocomplete after '/model ll', got %d", m.mode)
	}
	if len(m.autocomplete) != 2 {
		t.Errorf("expected 2 autocomplete matches for 'll', got %d", len(m.autocomplete))
	}

	// Typing "/model mis" should narrow to mistral only.
	m.value = "/model mis"
	m.updateAC()
	if len(m.autocomplete) != 1 {
		t.Errorf("expected 1 autocomplete match for 'mis', got %d", len(m.autocomplete))
	}

	// Typing "/model zzz" (no match) should exit autocomplete.
	m.value = "/model zzz"
	m.updateAC()
	if m.mode == modeAutocomplete {
		t.Error("expected no autocomplete for '/model zzz' with no matching models")
	}
}

func TestGetMatchingModels(t *testing.T) {
	models := []string{"llama3", "llama2", "mistral", "phi3"}

	tests := []struct {
		prefix  string
		wantLen int
	}{
		{"", 4},
		{"ll", 2},
		{"mi", 1},
		{"phi", 1},
		{"zzz", 0},
	}

	for _, tt := range tests {
		t.Run(tt.prefix, func(t *testing.T) {
			got := getMatchingModels(tt.prefix, models)
			if len(got) != tt.wantLen {
				t.Errorf("prefix %q: expected %d matches, got %d (%v)", tt.prefix, tt.wantLen, len(got), got)
			}
			// Each result should be prefixed with "/model ".
			for _, name := range got {
				if !strings.HasPrefix(name, "/model ") {
					t.Errorf("expected result to start with '/model ', got %q", name)
				}
			}
		})
	}
}

func TestInputModelRenderAutocomplete(t *testing.T) {
	m := NewInputModel()

	// Normal mode — empty overlay.
	if got := m.RenderAutocomplete(); got != "" {
		t.Errorf("expected empty RenderAutocomplete in normal mode, got %q", got)
	}

	// Autocomplete mode — overlay contains suggestions.
	m.value = "/t"
	m.updateAC()
	if m.mode != modeAutocomplete {
		t.Skip("no autocomplete triggered, skipping")
	}
	ac := m.RenderAutocomplete()
	if ac == "" {
		t.Error("expected non-empty RenderAutocomplete in autocomplete mode")
	}

	// Model picker mode with models — overlay contains model names.
	m2 := NewInputModel()
	m2.SetModels([]string{"llama3", "mistral"})
	m2.EnterModelSelect()
	overlay := m2.RenderAutocomplete()
	if !strings.Contains(overlay, "llama3") {
		t.Errorf("expected 'llama3' in picker overlay, got: %q", overlay)
	}

	// Model picker mode with no models — faint placeholder.
	m3 := NewInputModel()
	m3.EnterModelSelect()
	empty := m3.RenderAutocomplete()
	if empty == "" {
		t.Error("expected placeholder text when picker has no models")
	}
}

func TestInputModelRenderLine(t *testing.T) {
	m := NewInputModel()

	// Normal mode, empty value — shows placeholder.
	line := m.RenderLine()
	if line == "" {
		t.Error("expected non-empty RenderLine")
	}

	// Normal mode with value.
	m.value = "hello"
	m.cursor = 5
	line = m.RenderLine()
	if !strings.Contains(line, "hello") {
		t.Errorf("expected 'hello' in RenderLine, got %q", line)
	}

	// Model select mode — shows guidance text instead of input.
	m2 := NewInputModel()
	m2.SetModels([]string{"llama3"})
	m2.EnterModelSelect()
	line2 := m2.RenderLine()
	if strings.Contains(line2, "llama3") {
		t.Errorf("RenderLine should not show model names (those go in RenderAutocomplete), got %q", line2)
	}
	if line2 == "" {
		t.Error("expected guidance text in RenderLine for picker mode")
	}
}
