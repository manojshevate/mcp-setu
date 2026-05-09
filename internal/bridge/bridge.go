package bridge

import (
	"context"
	"fmt"

	"github.com/manojshevate/mcpgo/internal/mcp"
	"github.com/manojshevate/mcpgo/internal/ollama"
	"github.com/manojshevate/mcpgo/internal/ui"
)

// Bridge orchestrates the agentic loop between Ollama and MCP servers.
// It maintains the conversation state and coordinates tool calls between the
// language model and MCP-connected tools, with a safety limit of 20 iterations.
type Bridge struct {
	ollamaClient *ollama.Client
	mcpClient    *mcp.MultiClient
	model        string
	temperature  float64
	printer      *ui.Printer
}

// NewBridge creates a new Bridge.
func NewBridge(ollamaClient *ollama.Client, mcpClient *mcp.MultiClient, model string, temperature float64, printer *ui.Printer) *Bridge {
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

		// Execute tool calls.
		for _, call := range resp.ToolCalls {
			b.printer.PrintToolCall(call.Name, call.Arguments)

			result, err := b.mcpClient.CallTool(ctx, call.Name, call.Arguments)
			if err != nil {
				result = fmt.Sprintf("error: %v", err)
				b.printer.PrintWarning(fmt.Sprintf("Tool %q failed: %v", call.Name, err))
			}

			b.printer.PrintToolResult(call.Name, result, len(result) > 120)

			// Add tool result as a message.
			messages = append(messages, ollama.Message{
				Role:    "user",
				Content: fmt.Sprintf("Tool %q result: %s", call.Name, result),
			})
		}
	}

	return "", fmt.Errorf("agent loop exceeded %d iterations — possible infinite tool loop", maxIterations)
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
