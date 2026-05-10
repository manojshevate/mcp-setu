package bridge

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/manojshevate/mcp-setu/internal/mcp"
	"github.com/manojshevate/mcp-setu/internal/ollama"
	"github.com/manojshevate/mcp-setu/internal/ui"
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
	GetServer(name string) mcp.MCPClientInterface
	GetServerForTool(toolName string) mcp.MCPClientInterface
	CallTool(ctx context.Context, toolName string, arguments map[string]any) (string, error)
	GetAllServerNames() []string
	CloseAll() error
}

// Stats tracks performance metrics for a session.
type Stats struct {
	MessageCount      int           // Total messages sent
	ToolCallCount     int           // Total tool calls made
	TotalDuration     time.Duration // Total time spent processing
	IterationCount    int           // Total iterations across all messages
	LastResponseTime  time.Duration // Last message response time
	AverageLoopTime   time.Duration // Average time per iteration
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
	stats        Stats
	startTime    time.Time
	mu           sync.RWMutex // protects model and stats
}

// NewBridge creates a new Bridge.
func NewBridge(ollamaClient OllamaClient, mcpClient MCPClient, model string, temperature float64, printer *ui.Printer) *Bridge {
	return &Bridge{
		ollamaClient: ollamaClient,
		mcpClient:    mcpClient,
		model:        model,
		temperature:  temperature,
		printer:      printer,
		startTime:    time.Now(),
	}
}

// SetModel changes the model and validates tool support.
func (b *Bridge) SetModel(ctx context.Context, model string) error {
	b.mu.RLock()
	currentModel := b.model
	b.mu.RUnlock()

	if model == currentModel {
		return nil
	}
	if err := b.ollamaClient.CheckToolSupport(ctx, model); err != nil {
		return err
	}
	b.mu.Lock()
	b.model = model
	b.mu.Unlock()
	return nil
}

// GetModel returns the current model.
func (b *Bridge) GetModel() string {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.model
}

// GetStats returns the current stats.
func (b *Bridge) GetStats() Stats {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.stats.TotalDuration = time.Since(b.startTime)
	if b.stats.IterationCount > 0 {
		b.stats.AverageLoopTime = b.stats.TotalDuration / time.Duration(b.stats.IterationCount)
	}
	return b.stats
}

// ProcessMessage runs the agentic loop for a user message.
func (b *Bridge) ProcessMessage(ctx context.Context, messages []ollama.Message) (string, error) {
	start := time.Now()

	// Build tools list for Ollama.
	tools := b.buildToolsList()

	// Agentic loop.
	maxIterations := 20
	iteration := 0
	toolCallCount := 0

	for iteration < maxIterations {
		iteration++

		// Log that LLM is processing (verbose mode).
		b.printer.PrintLLMProcessing(iteration)

		// Call Ollama with tools.
		b.mu.RLock()
		model := b.model
		temperature := b.temperature
		b.mu.RUnlock()
		resp, err := b.ollamaClient.Chat(ctx, model, messages, tools, temperature)
		if err != nil {
			return "", fmt.Errorf("ollama chat failed: %w", err)
		}

		// Add assistant response to history.
		messages = append(messages, *resp)

		// Check for tool calls.
		if len(resp.ToolCalls) == 0 {
			// No tool calls, return the response content.
			duration := time.Since(start)
			b.mu.Lock()
			b.stats.MessageCount++
			b.stats.ToolCallCount += toolCallCount
			b.stats.IterationCount += iteration
			b.stats.LastResponseTime = duration
			b.mu.Unlock()
			return resp.Content, nil
		}

		toolCallCount += len(resp.ToolCalls)

		// Execute tool calls in parallel for independent calls.
		toolResults := b.executeToolsParallel(ctx, resp.ToolCalls)

		// Add tool results to conversation history in order.
		for i := range resp.ToolCalls {
			result := toolResults[i]
			// Add tool result as a message with proper "tool" role.
			messages = append(messages, ollama.Message{
				Role:    "tool",
				Content: result,
			})
		}
	}

	return "", fmt.Errorf("agent loop exceeded %d iterations — possible infinite tool loop", maxIterations)
}

// executeToolsParallel executes multiple tool calls concurrently with a limit of 8 workers.
// Results are returned in the same order as the input tool calls.
func (b *Bridge) executeToolsParallel(ctx context.Context, calls []ollama.ToolCall) []string {
	if len(calls) == 0 {
		return []string{}
	}

	const maxWorkers = 8
	results := make([]string, len(calls))
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, maxWorkers)

	// Execute all tool calls concurrently with bounded workers.
	for i, call := range calls {
		wg.Add(1)
		go func(index int, toolCall ollama.ToolCall) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			// Extract tool name and arguments (handles both old and new formats)
			toolName, toolArgs := toolCall.NormalizeToolCall()

			b.printer.PrintToolCall(toolName, toolArgs)

			result, err := b.mcpClient.CallTool(ctx, toolName, toolArgs)
			if err != nil {
				result = fmt.Sprintf("error: %v", err)
				b.printer.PrintWarning(fmt.Sprintf("Tool %q failed: %v", toolName, err))
			}

			b.printer.PrintToolResult(toolName, result, len(result) > 120)
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
