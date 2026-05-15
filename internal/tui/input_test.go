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
				valueRunes: []rune(tt.value),
				width:      tt.width,
				cursor:     tt.cursor,
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
	m.valueRunes = []rune("test")
	if m.GetValue() != "test" {
		t.Errorf("expected 'test', got '%s'", m.GetValue())
	}

	// Test Clear
	m.Clear()
	if len(m.valueRunes) != 0 {
		t.Errorf("expected empty value after clear, got '%s'", string(m.valueRunes))
	}

	// Test cursor position
	m.cursor = 5
	m.valueRunes = []rune("hello world")
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
	m.valueRunes = []rune("hello")
	m.updateAC()
	if m.mode == modeAutocomplete {
		t.Error("expected no autocomplete for non-slash input")
	}

	// Autocomplete with /
	m.valueRunes = []rune("/t")
	m.updateAC()
	if m.mode != modeAutocomplete {
		t.Error("expected autocomplete mode for /t")
	}
	if len(m.autocomplete) == 0 {
		t.Error("expected autocomplete suggestions for /t")
	}

	// No autocomplete for non-matching prefix
	m.valueRunes = []rune("/xyz")
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
	m.valueRunes = []rune("/m")
	m.updateAC()
	if m.mode != modeAutocomplete {
		t.Errorf("expected modeAutocomplete after /m, got %d", m.mode)
	}

	// Clearing the value returns to normal mode.
	m.valueRunes = []rune("hello")
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
	if len(m.valueRunes) != 0 {
		t.Errorf("expected empty value after ExitModelSelect, got %q", string(m.valueRunes))
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
	m.valueRunes = []rune("/model ")
	m.updateAC()
	if m.mode != modeAutocomplete {
		t.Errorf("expected modeAutocomplete after '/model ', got %d", m.mode)
	}
	if len(m.autocomplete) == 0 {
		t.Error("expected autocomplete entries for '/model '")
	}

	// Typing "/model ll" should narrow to llama3, llama2.
	m.valueRunes = []rune("/model ll")
	m.updateAC()
	if m.mode != modeAutocomplete {
		t.Errorf("expected modeAutocomplete after '/model ll', got %d", m.mode)
	}
	if len(m.autocomplete) != 2 {
		t.Errorf("expected 2 autocomplete matches for 'll', got %d", len(m.autocomplete))
	}

	// Typing "/model mis" should narrow to mistral only.
	m.valueRunes = []rune("/model mis")
	m.updateAC()
	if len(m.autocomplete) != 1 {
		t.Errorf("expected 1 autocomplete match for 'mis', got %d", len(m.autocomplete))
	}

	// Typing "/model zzz" (no match) should exit autocomplete.
	m.valueRunes = []rune("/model zzz")
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
	m.valueRunes = []rune("/t")
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
	m.valueRunes = []rune("hello")
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

// TestInputModelEnterSends verifies that plain Enter fires a SendMsg (default chat behaviour).
func TestInputModelEnterSends(t *testing.T) {
	m := NewInputModel()
	m.valueRunes = []rune("hello")
	m.cursor = 5

	// Press plain Enter → SendMsg fired.
	_, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter})
	if cmd == nil {
		t.Fatal("expected non-nil cmd for plain Enter (send)")
	}
	msg := cmd()
	if _, ok := msg.(SendMsg); !ok {
		t.Errorf("expected SendMsg from Enter, got %T", msg)
	}
}

// TestInputModelShiftEnterInsertsNewline verifies that Shift+Enter inserts a newline.
// Note: bubbletea v1.3.10 KeyMsg has no Shift field; Shift+Enter is detected via
// msg.Alt (Alt+Enter escape sequence is the cross-platform terminal trigger).
func TestInputModelShiftEnterInsertsNewline(t *testing.T) {
	m := NewInputModel()
	m.valueRunes = []rune("hello")
	m.cursor = 5

	// Press Shift+Enter (detected as Alt+Enter in terminal) → newline inserted (not SendMsg).
	updated, cmd := m.Update(tea.KeyMsg{Type: tea.KeyEnter, Alt: true})
	m = updated.(InputModel)

	// cmd should be nil for a newline insertion.
	if cmd != nil {
		t.Error("expected nil cmd for Shift+Enter (newline), got non-nil")
	}
	if !strings.Contains(string(m.valueRunes), "\n") {
		t.Errorf("expected newline in value after Shift+Enter, got %q", string(m.valueRunes))
	}
	if m.LineCount() != 2 {
		t.Errorf("expected 2 lines after Shift+Enter, got %d", m.LineCount())
	}
}

// TestInputModelLineCount checks LineCount for various values.
func TestInputModelLineCount(t *testing.T) {
	cases := []struct {
		value string
		want  int
	}{
		{"", 1},
		{"hello", 1},
		{"hello\nworld", 2},
		{"a\nb\nc", 3},
	}
	for _, tc := range cases {
		m := InputModel{valueRunes: []rune(tc.value)}
		if got := m.LineCount(); got != tc.want {
			t.Errorf("LineCount(%q) = %d, want %d", tc.value, got, tc.want)
		}
	}
}

// TestInputModelMultilineRenderLine verifies the "(N lines)" indicator appears.
func TestInputModelMultilineRenderLine(t *testing.T) {
	m := InputModel{valueRunes: []rune("line1\nline2\nline3"), cursor: 17}
	line := m.RenderLine()
	if !strings.Contains(line, "3 lines") {
		t.Errorf("expected '3 lines' indicator in multiline RenderLine, got %q", line)
	}
	if !strings.Contains(line, "line1") {
		t.Errorf("expected first line 'line1' in RenderLine, got %q", line)
	}
	if !strings.Contains(line, "shift+enter for newline") {
		t.Errorf("expected 'shift+enter for newline' hint in multiline RenderLine, got %q", line)
	}
}

// TestInputModelGetValuePreservesNewlines ensures internal newlines survive GetValue.
func TestInputModelGetValuePreservesNewlines(t *testing.T) {
	m := InputModel{valueRunes: []rune("hello\nworld")}
	v := m.GetValue()
	if v != "hello\nworld" {
		t.Errorf("expected GetValue to preserve newlines, got %q", v)
	}
}

// TestPluralLines verifies the pluralLines helper.
func TestPluralLines(t *testing.T) {
	if got := pluralLines(1); got != "1 line" {
		t.Errorf("pluralLines(1) = %q, want '1 line'", got)
	}
	if got := pluralLines(3); got != "3 lines" {
		t.Errorf("pluralLines(3) = %q, want '3 lines'", got)
	}
}

// TestInputModelNonASCII verifies that multibyte Unicode characters (accented
// letters, arrows, CJK characters, emoji) are not corrupted when typed, edited
// via Backspace, and retrieved via GetValue.
func TestInputModelNonASCII(t *testing.T) {
	// Simulate typing "héllo→世界" character by character.
	input := []rune("héllo→世界")
	m := NewInputModel()
	for _, r := range input {
		updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{r}})
		m = updated.(InputModel)
	}

	got := m.GetValue()
	want := "héllo→世界"
	if got != want {
		t.Errorf("GetValue after typing non-ASCII: got %q, want %q", got, want)
	}
	if m.cursor != len(input) {
		t.Errorf("cursor should be at rune position %d, got %d", len(input), m.cursor)
	}

	// Backspace should remove the last rune ('界') cleanly.
	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyBackspace})
	m = updated.(InputModel)
	got = m.GetValue()
	if got != "héllo→世" {
		t.Errorf("GetValue after Backspace: got %q, want %q", "héllo→世", got)
	}
	if m.cursor != len(input)-1 {
		t.Errorf("cursor after Backspace should be %d, got %d", len(input)-1, m.cursor)
	}

	// Insert 'Z' in the middle (after 'é', at rune index 2).
	m.cursor = 2
	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'Z'}})
	m = updated.(InputModel)
	got = m.GetValue()
	if got != "héZllo→世" {
		t.Errorf("GetValue after mid-insert: got %q, want %q", "héZllo→世", got)
	}

	// RenderLine must not panic and must contain non-ASCII text.
	line := m.RenderLine()
	if line == "" {
		t.Error("RenderLine returned empty string for non-ASCII input")
	}
}
