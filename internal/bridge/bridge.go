package bridge

import (
	"context"
	"fmt"
	"sync"

	"github.com/manojshevate/mcpgo/internal/mcp"
	"github.com/manojshevate/mcpgo/internal/ollama"
	"github.com/manojshevate/mcpgo/internal/ui"
)

// OllamaClient defines the interface for Ollama API interactions.
type OllamaClient interface {
	Chat(ctx context.Context, model string, messages []ollama.Message, tools []ollama.Tool, temperature float64) (*ollama.Message, error)
	CheckToolSupport(ctx context.Context, model string) error
	ListLocalModels(ctx context.Context) ([]ollama.ModelInfo, error)
}

// MCPClient defines the interface for MCP server interactions.
type MCPClient interface {
	GetAllTools() map[string]string
	GetServer(name string) *mcp.Client
	GetServerForTool(toolName string) *mcp.Client
	CallTool(ctx context.Context, toolName string, arguments map[string]any) (string, error)
	GetAllServerNames() []string
	CloseAll() error
}

// Bridge orchestrates the agentic loop between Ollama and MCP servers.
// It maintains the conversation state and coordinates tool calls between the
// language model and MCP-connected tools, with a safety limit of 20 iterations.
type Bridge struct {
	ollamaClient OllamaClient
	mcpClient    MCPClient
	model        string
	temperature  float64
	printer      *ui.Printer
}

// NewBridge creates a new Bridge.
func NewBridge(ollamaClient OllamaClient, mcpClient MCPClient, model string, temperature float64, printer *ui.Printer) *Bridge {
	return &Bridge{
		ollamaClient: ollamaClient,
		mcpClient:    mcpClient,
		model:        model,
		temperature:  temperature,
		printer:      printer,
	}
}

// ProcessMessage runs the agentic loop for a user message.
func (b *Bridge) ProcessMessage(ctx context.Context, messages []ollama.Message) (string, error) {
	// Build tools list for Ollama.
	tools := b.buildToolsList()

	// Agentic loop.
	maxIterations := 20
	iteration := 0

	for iteration < maxIterations {
		iteration++

		// Call Ollama with tools.
		resp, err := b.ollamaClient.Chat(ctx, b.model, messages, tools, b.temperature)
		if err != nil {
			return "", fmt.Errorf("ollama chat failed: %w", err)
		}

		// Add assistant response to history.
		messages = append(messages, *resp)

		// Check for tool calls.
		if len(resp.ToolCalls) == 0 {
			// No tool calls, return the response content.
			return resp.Content, nil
		}

		// Execute tool calls in parallel for independent calls.
		toolResults := b.executeToolsParallel(ctx, resp.ToolCalls)

		// Add tool results to conversation history in order.
		for i, call := range resp.ToolCalls {
			result := toolResults[i]
			// Add tool result as a message.
			messages = append(messages, ollama.Message{
				Role:    "user",
				Content: fmt.Sprintf("Tool %q result: %s", call.Name, result),
			})
		}
	}

	return "", fmt.Errorf("agent loop exceeded %d iterations — possible infinite tool loop", maxIterations)
}

// executeToolsParallel executes multiple tool calls concurrently.
// Results are returned in the same order as the input tool calls.
func (b *Bridge) executeToolsParallel(ctx context.Context, calls []ollama.ToolCall) []string {
	if len(calls) == 0 {
		return []string{}
	}

	results := make([]string, len(calls))
	var wg sync.WaitGroup

	// Execute all tool calls concurrently.
	for i, call := range calls {
		wg.Add(1)
		go func(index int, toolCall ollama.ToolCall) {
			defer wg.Done()

			b.printer.PrintToolCall(toolCall.Name, toolCall.Arguments)

			result, err := b.mcpClient.CallTool(ctx, toolCall.Name, toolCall.Arguments)
			if err != nil {
				result = fmt.Sprintf("error: %v", err)
				b.printer.PrintWarning(fmt.Sprintf("Tool %q failed: %v", toolCall.Name, err))
			}

			b.printer.PrintToolResult(toolCall.Name, result, len(result) > 120)
			results[index] = result
		}(i, call)
	}

	// Wait for all goroutines to complete.
	wg.Wait()
	return results
}

// buildToolsList converts MCP tools to Ollama tool definitions.
func (b *Bridge) buildToolsList() []ollama.Tool {
	var tools []ollama.Tool

	allTools := b.mcpClient.GetAllTools()
	for toolName, serverName := range allTools {
		server := b.mcpClient.GetServer(serverName)
		if server == nil {
			continue
		}

		serverTools := server.GetTools()
		if mcpTool, ok := serverTools[toolName]; ok {
			tool := ollama.Tool{
				Type: "function",
				Function: ollama.Function{
					Name:        mcpTool.Name,
					Description: mcpTool.Description,
					Parameters:  mcpTool.InputSchema,
				},
			}
			tools = append(tools, tool)
		}
	}

	return tools
}
