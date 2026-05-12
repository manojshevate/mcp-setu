package ollama

import (
	"context"
	"testing"
	"time"

	"github.com/manojshevate/mcp-setu/internal/config"
)

// TestNewClient tests client initialization.
func TestNewClient(t *testing.T) {
	client := NewClient("http://localhost:11434")

	if client.baseURL != "http://localhost:11434" {
		t.Errorf("Expected baseURL http://localhost:11434, got %s", client.baseURL)
	}

	if client.http == nil {
		t.Error("Expected http client to be initialized")
	}

	if client.http.Timeout != config.DefaultTimeout {
		t.Errorf("Expected %v timeout, got %v", config.DefaultTimeout, client.http.Timeout)
	}
}

// TestNewClientTrimsTrailingSlash tests that trailing slashes are trimmed.
func TestNewClientTrimsTrailingSlash(t *testing.T) {
	client := NewClient("http://localhost:11434/")

	if client.baseURL != "http://localhost:11434" {
		t.Errorf("Expected baseURL without trailing slash, got %s", client.baseURL)
	}
}

// TestFormatBytes tests byte formatting.
func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes    int64
		expected string
	}{
		{0, "0 B"},
		{512, "512 B"},
		{1024, "1.0 KB"},
		{1024 * 1024, "1.0 MB"},
		{5 * 1024 * 1024, "5.0 MB"},
		{1024 * 1024 * 1024, "1.0 GB"},
		{2147483648, "2.0 GB"},
	}

	for _, tt := range tests {
		result := formatBytes(tt.bytes)
		if result != tt.expected {
			t.Errorf("formatBytes(%d): expected %s, got %s", tt.bytes, tt.expected, result)
		}
	}
}

// TestChatRequestStructure tests chat request structure.
func TestChatRequestStructure(t *testing.T) {
	messages := []Message{
		{Role: "user", Content: "Hello"},
	}
	tools := []Tool{
		{
			Type: "function",
			Function: Function{
				Name:        "test_tool",
				Description: "A test tool",
				Parameters:  map[string]any{},
			},
		},
	}

	req := ChatRequest{
		Model:       "gemma4:e4b",
		Messages:    messages,
		Tools:       tools,
		Temperature: 0.7,
		Stream:      false,
	}

	if req.Model != "gemma4:e4b" {
		t.Errorf("Expected model gemma4:e4b, got %s", req.Model)
	}

	if len(req.Messages) != 1 {
		t.Errorf("Expected 1 message, got %d", len(req.Messages))
	}

	if len(req.Tools) != 1 {
		t.Errorf("Expected 1 tool, got %d", len(req.Tools))
	}

	if req.Stream {
		t.Error("Expected stream=false")
	}
}

// TestMessageStructure tests message structure.
func TestMessageStructure(t *testing.T) {
	msg := Message{
		Role:    "assistant",
		Content: "Hello",
		ToolCalls: []ToolCall{
			{
				Name:      "test_tool",
				Arguments: map[string]any{"arg": "value"},
			},
		},
	}

	if msg.Role != "assistant" {
		t.Errorf("Expected role assistant, got %s", msg.Role)
	}

	if len(msg.ToolCalls) != 1 {
		t.Errorf("Expected 1 tool call, got %d", len(msg.ToolCalls))
	}

	if msg.ToolCalls[0].Name != "test_tool" {
		t.Errorf("Expected tool name test_tool, got %s", msg.ToolCalls[0].Name)
	}
}

// TestToolStructure tests tool structure.
func TestToolStructure(t *testing.T) {
	tool := Tool{
		Type: "function",
		Function: Function{
			Name:        "read_file",
			Description: "Read a file",
			Parameters: map[string]any{
				"properties": map[string]any{
					"path": map[string]any{"type": "string"},
				},
			},
		},
	}

	if tool.Type != "function" {
		t.Errorf("Expected type function, got %s", tool.Type)
	}

	if tool.Function.Name != "read_file" {
		t.Errorf("Expected name read_file, got %s", tool.Function.Name)
	}
}

// TestCheckToolSupportContextTimeout tests context cancellation in CheckToolSupport.
func TestCheckToolSupportContextTimeout(t *testing.T) {
	client := NewClient("http://localhost:11434")

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Millisecond)
	defer cancel()

	// This will timeout since localhost:11434 is not available
	err := client.CheckToolSupport(ctx, "gemma4:e4b")
	if err == nil {
		t.Fatal("Expected error due to timeout or connection failure")
	}
}

// TestListLocalModelsStructure tests model list structure.
func TestListLocalModelsStructure(t *testing.T) {
	models := []ModelDetail{
		{
			Name:       "gemma4:e4b",
			ModifiedAt: "2024-05-09T20:00:00Z",
			Size:       4294967296,
		},
		{
			Name:       "llama3.2:3b",
			ModifiedAt: "2024-05-09T19:00:00Z",
			Size:       3221225472,
		},
	}

	if len(models) != 2 {
		t.Errorf("Expected 2 models, got %d", len(models))
	}

	if models[0].Name != "gemma4:e4b" {
		t.Errorf("Expected first model gemma4:e4b, got %s", models[0].Name)
	}

	if models[0].Size != 4294967296 {
		t.Errorf("Expected size 4GB, got %d", models[0].Size)
	}
}
