package bridge

import (
	"context"
	"testing"

	"github.com/manojshevate/mcp-setu/internal/mcp"
	"github.com/manojshevate/mcp-setu/internal/ollama"
	"github.com/manojshevate/mcp-setu/internal/ui"
)

// MockOllamaClient implements a mock Ollama client for testing.
type MockOllamaClient struct {
	responses  []ollama.Message
	callCount  int
	maxCalls   int
}

// NewMockOllamaClient creates a new mock Ollama client.
func NewMockOllamaClient(responses []ollama.Message) *MockOllamaClient {
	return &MockOllamaClient{
		responses: responses,
		maxCalls:  len(responses),
	}
}

// Chat simulates a chat call and returns predefined responses.
func (m *MockOllamaClient) Chat(ctx context.Context, model string, messages []ollama.Message, tools []ollama.Tool, temperature float64) (*ollama.Message, error) {
	if m.callCount >= m.maxCalls {
		return &ollama.Message{
			Role:    "assistant",
			Content: "No more responses",
		}, nil
	}

	resp := m.responses[m.callCount]
	m.callCount++
	return &resp, nil
}

// ChatStream simulates a streaming chat call by returning a channel with events.
func (m *MockOllamaClient) ChatStream(ctx context.Context, model string, messages []ollama.Message, tools []ollama.Tool, temperature float64) (<-chan ollama.StreamEvent, error) {
	// Reuse the Chat method to get the response, then stream it
	resp, err := m.Chat(ctx, model, messages, tools, temperature)
	if err != nil {
		return nil, err
	}

	ch := make(chan ollama.StreamEvent)
	go func() {
		defer close(ch)
		// Stream the content and tool calls as an event
		ch <- ollama.StreamEvent{
			Content:   resp.Content,
			ToolCalls: resp.ToolCalls,
			Done:      true,
		}
	}()

	return ch, nil
}

// CheckToolSupport is a no-op for testing.
func (m *MockOllamaClient) CheckToolSupport(ctx context.Context, model string) error {
	return nil
}

// ListLocalModels is a no-op for testing.
func (m *MockOllamaClient) ListLocalModels(ctx context.Context) ([]ollama.ModelInfo, error) {
	return []ollama.ModelInfo{}, nil
}

// MockMCPClient implements a mock MCP client for testing.
type MockMCPClient struct {
	tools       map[string]string
	toolResults map[string]string
}

// NewMockMCPClient creates a new mock MCP client.
func NewMockMCPClient() *MockMCPClient {
	return &MockMCPClient{
		tools:       make(map[string]string),
		toolResults: make(map[string]string),
	}
}

// AddTool adds a tool to the mock client.
func (m *MockMCPClient) AddTool(name, server string) {
	m.tools[name] = server
}

// SetToolResult sets the result for a tool call.
func (m *MockMCPClient) SetToolResult(name, result string) {
	m.toolResults[name] = result
}

// GetAllTools returns all tools.
func (m *MockMCPClient) GetAllTools() map[string]string {
	return m.tools
}

// GetServer returns nil for mock.
func (m *MockMCPClient) GetServer(name string) mcp.MCPClientInterface {
	return nil
}

// GetServerForTool returns nil for mock.
func (m *MockMCPClient) GetServerForTool(toolName string) mcp.MCPClientInterface {
	return nil
}

// CallTool returns predefined results.
func (m *MockMCPClient) CallTool(ctx context.Context, toolName string, arguments map[string]any) (string, error) {
	if result, ok := m.toolResults[toolName]; ok {
		return result, nil
	}
	return "default result", nil
}

// GetAllServerNames returns empty list for mock.
func (m *MockMCPClient) GetAllServerNames() []string {
	return []string{}
}

// CloseAll is a no-op for mock.
func (m *MockMCPClient) CloseAll() error {
	return nil
}

// TestProcessMessageNoTools tests processing a message without tool calls.
func TestProcessMessageNoTools(t *testing.T) {
	printer := ui.NewPrinter(false)
	ollamaClient := NewMockOllamaClient([]ollama.Message{
		{
			Role:    "assistant",
			Content: "Hello, how can I help?",
		},
	})
	mcpClient := NewMockMCPClient()

	bridge := NewBridge(ollamaClient, mcpClient, "test-model", 0.7, printer)

	messages := []ollama.Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant.",
		},
		{
			Role:    "user",
			Content: "Hello",
		},
	}

	response, err := bridge.ProcessMessage(context.Background(), messages)
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}

	if response != "Hello, how can I help?" {
		t.Errorf("Expected response %q, got %q", "Hello, how can I help?", response)
	}
}

// TestProcessMessageWithToolCall tests processing with a single tool call.
func TestProcessMessageWithToolCall(t *testing.T) {
	printer := ui.NewPrinter(false)

	// First response: assistant calls a tool
	// Second response: assistant provides final answer
	ollamaClient := NewMockOllamaClient([]ollama.Message{
		{
			Role: "assistant",
			ToolCalls: []ollama.ToolCall{
				{
					Name:      "test_tool",
					Arguments: map[string]any{"arg": "value"},
				},
			},
		},
		{
			Role:    "assistant",
			Content: "The tool returned: test result",
		},
	})

	mcpClient := NewMockMCPClient()
	mcpClient.AddTool("test_tool", "test_server")
	mcpClient.SetToolResult("test_tool", "test result")

	bridge := NewBridge(ollamaClient, mcpClient, "test-model", 0.7, printer)

	messages := []ollama.Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant.",
		},
	}

	response, err := bridge.ProcessMessage(context.Background(), messages)
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}

	if response != "The tool returned: test result" {
		t.Errorf("Expected response containing tool result, got %q", response)
	}

	// Verify that the Ollama client was called twice (initial + after tool)
	if ollamaClient.callCount != 2 {
		t.Errorf("Expected 2 Ollama calls, got %d", ollamaClient.callCount)
	}
}

// TestProcessMessageWithMultipleToolCalls tests parallel execution of multiple independent tool calls.
func TestProcessMessageWithMultipleToolCalls(t *testing.T) {
	printer := ui.NewPrinter(false)

	// First response: assistant calls multiple tools
	// Second response: assistant provides final answer
	ollamaClient := NewMockOllamaClient([]ollama.Message{
		{
			Role: "assistant",
			ToolCalls: []ollama.ToolCall{
				{
					Name:      "tool_a",
					Arguments: map[string]any{"arg": "a"},
				},
				{
					Name:      "tool_b",
					Arguments: map[string]any{"arg": "b"},
				},
				{
					Name:      "tool_c",
					Arguments: map[string]any{"arg": "c"},
				},
			},
		},
		{
			Role:    "assistant",
			Content: "Results: A=a_result, B=b_result, C=c_result",
		},
	})

	mcpClient := NewMockMCPClient()
	mcpClient.AddTool("tool_a", "server1")
	mcpClient.AddTool("tool_b", "server1")
	mcpClient.AddTool("tool_c", "server1")
	mcpClient.SetToolResult("tool_a", "a_result")
	mcpClient.SetToolResult("tool_b", "b_result")
	mcpClient.SetToolResult("tool_c", "c_result")

	bridge := NewBridge(ollamaClient, mcpClient, "test-model", 0.7, printer)

	messages := []ollama.Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant.",
		},
	}

	response, err := bridge.ProcessMessage(context.Background(), messages)
	if err != nil {
		t.Fatalf("ProcessMessage failed: %v", err)
	}

	if response != "Results: A=a_result, B=b_result, C=c_result" {
		t.Errorf("Expected final response, got %q", response)
	}
}

// TestProcessMessageMaxIterations tests the safety limit on iterations.
func TestProcessMessageMaxIterations(t *testing.T) {
	printer := ui.NewPrinter(false)

	// Create responses that always request tools (causes loop)
	responses := []ollama.Message{}
	for i := 0; i < 25; i++ {
		responses = append(responses, ollama.Message{
			Role: "assistant",
			ToolCalls: []ollama.ToolCall{
				{
					Name:      "infinite_tool",
					Arguments: map[string]any{},
				},
			},
		})
	}

	ollamaClient := NewMockOllamaClient(responses)
	mcpClient := NewMockMCPClient()
	mcpClient.AddTool("infinite_tool", "server")
	mcpClient.SetToolResult("infinite_tool", "result")

	bridge := NewBridge(ollamaClient, mcpClient, "test-model", 0.7, printer)

	messages := []ollama.Message{
		{
			Role:    "system",
			Content: "You are a helpful assistant.",
		},
	}

	_, err := bridge.ProcessMessage(context.Background(), messages)
	if err == nil {
		t.Fatalf("Expected error for exceeded iterations, got nil")
	}

	if err.Error() != "agent loop exceeded 20 iterations — possible infinite tool loop" {
		t.Errorf("Expected iteration limit error, got: %v", err)
	}
}

// TestExecuteToolsParallel tests that multiple tools are executed concurrently.
func TestExecuteToolsParallel(t *testing.T) {
	printer := ui.NewPrinter(false)
	mcpClient := NewMockMCPClient()
	mcpClient.AddTool("tool_1", "server")
	mcpClient.AddTool("tool_2", "server")
	mcpClient.AddTool("tool_3", "server")
	mcpClient.SetToolResult("tool_1", "result_1")
	mcpClient.SetToolResult("tool_2", "result_2")
	mcpClient.SetToolResult("tool_3", "result_3")

	bridge := NewBridge(nil, mcpClient, "test-model", 0.7, printer)

	toolCalls := []ollama.ToolCall{
		{Name: "tool_1", Arguments: map[string]any{}},
		{Name: "tool_2", Arguments: map[string]any{}},
		{Name: "tool_3", Arguments: map[string]any{}},
	}

	results := bridge.executeToolsParallel(context.Background(), toolCalls)

	if len(results) != 3 {
		t.Errorf("Expected 3 results, got %d", len(results))
	}

	expectedResults := map[int]string{
		0: "result_1",
		1: "result_2",
		2: "result_3",
	}

	for i, expected := range expectedResults {
		if results[i] != expected {
			t.Errorf("Result %d: expected %q, got %q", i, expected, results[i])
		}
	}
}

// TestBuildToolsList tests building the tools list for Ollama.
func TestBuildToolsList(t *testing.T) {
	printer := ui.NewPrinter(false)
	mcpClient := NewMockMCPClient()
	mcpClient.AddTool("read_file", "filesystem")
	mcpClient.AddTool("list_files", "filesystem")
	mcpClient.AddTool("execute_query", "sqlite")

	bridge := NewBridge(nil, mcpClient, "test-model", 0.7, printer)

	tools := bridge.buildToolsList()

	// Since we're using a mock that returns nil for GetServer,
	// the buildToolsList will return an empty slice. That's expected behavior.
	// In a real scenario with actual servers, this would return the tools.
	if len(tools) != 0 {
		t.Errorf("Expected 0 tools from mock (GetServer returns nil), got %d", len(tools))
	}
}

// TestNewBridge tests bridge initialization.
func TestNewBridge(t *testing.T) {
	printer := ui.NewPrinter(false)
	ollamaClient := NewMockOllamaClient([]ollama.Message{})
	mcpClient := NewMockMCPClient()

	bridge := NewBridge(ollamaClient, mcpClient, "gemma4:e4b", 0.7, printer)

	if bridge.model != "gemma4:e4b" {
		t.Errorf("Expected model %q, got %q", "gemma4:e4b", bridge.model)
	}

	if bridge.temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", bridge.temperature)
	}
}
