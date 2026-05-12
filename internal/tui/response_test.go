package tui

import (
	"testing"
)

func TestResponseModelView(t *testing.T) {
	m := NewResponseModel()
	m.SetSize(80, 20)

	// Empty response
	view := m.View()
	if view == "" {
		t.Error("expected non-empty view")
	}

	// Add content
	m.AddChunk("Hello ")
	m.AddChunk("World")

	content := m.GetContent()
	if content != "Hello World" {
		t.Errorf("expected 'Hello World', got '%s'", content)
	}

	// Clear
	m.Clear()
	if m.GetContent() != "" {
		t.Error("expected empty content after clear")
	}
}

func TestResponseModelZeroSize(t *testing.T) {
	m := NewResponseModel()
	// Don't set size - should handle zero width/height gracefully
	view := m.View()
	if view != "" {
		t.Errorf("expected empty view for zero size, got '%s'", view)
	}
}

func TestResponseModelAddChunks(t *testing.T) {
	m := NewResponseModel()
	m.SetSize(80, 20)

	chunks := []string{"chunk1", "chunk2", "chunk3"}
	expected := "chunk1chunk2chunk3"

	for _, chunk := range chunks {
		m.AddChunk(chunk)
	}

	if m.GetContent() != expected {
		t.Errorf("expected '%s', got '%s'", expected, m.GetContent())
	}
}
