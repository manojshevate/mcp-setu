package bridge

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/manojshevate/mcp-setu/internal/content"
	"github.com/manojshevate/mcp-setu/internal/logger"
	"github.com/manojshevate/mcp-setu/internal/mcp"
	"github.com/manojshevate/mcp-setu/internal/ollama"
)

// Printer is the minimal output interface needed by the agentic loop.
// Both *ui.Printer and *tui.Printer satisfy it.
type Printer interface {
	PrintLLMProcessing(iteration int)
	PrintWarning(msg string)
	PrintResponseStart()
	PrintResponseChunk(chunk string)
	PrintResponseEnd()
	PrintToolCall(name string, args map[string]any)
	PrintToolResult(name string, result string, truncated bool)
	PrintStructuredContent(sc *content.StructuredContent)
}

// OllamaClient defines the interface for Ollama API interactions.
type OllamaClient interface {
	Chat(ctx context.Context, model string, messages []ollama.Message, tools []ollama.Tool, temperature float64, contextLength int) (*ollama.Message, error)
	ChatStream(ctx context.Context, model string, messages []ollama.Message, tools []ollama.Tool, temperature float64, contextLength int) (<-chan ollama.StreamEvent, error)
	EnsureModelExists(ctx context.Context, model string) error
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
	MessageCount     int           // Total messages sent
	ToolCallCount    int           // Total tool calls made
	TotalDuration    time.Duration // Total time spent processing
	IterationCount   int           // Total iterations across all messages
	LastResponseTime time.Duration // Last message response time
	AverageLoopTime  time.Duration // Average time per iteration
}

// Bridge orchestrates the agentic loop between Ollama and MCP servers.
// It maintains the conversation state and coordinates tool calls between the
// language model and MCP-connected tools, with a safety limit of 20 iterations.
type Bridge struct {
	ollamaClient  OllamaClient
	mcpClient     MCPClient
	model         string
	temperature   float64
	contextLength int
	printer       Printer
	logger        *logger.Logger
	stats         Stats
	startTime     time.Time
	mu            sync.RWMutex // protects model and stats
}

// NewBridge creates a new Bridge.
func NewBridge(ollamaClient OllamaClient, mcpClient MCPClient, model string, temperature float64, contextLength int, printer Printer, log *logger.Logger) *Bridge {
	return &Bridge{
		ollamaClient:  ollamaClient,
		mcpClient:     mcpClient,
		model:         model,
		temperature:   temperature,
		contextLength: contextLength,
		printer:       printer,
		logger:        log,
		startTime:     time.Now(),
	}
}

// SetModel changes the model and validates tool support.
// The existence check is performed outside any lock because it is a network
// call that must not block concurrent readers. After the check succeeds the
// write lock is held for the entire read-compare-write sequence so the
// transition is genuinely atomic: no other goroutine can observe a
// half-updated model or interleave a concurrent SetModel between the read
// and the write.
func (b *Bridge) SetModel(ctx context.Context, model string) error {
	// Cheap early-exit: read current model without a lock using a snapshot
	// taken under the read lock. If the model is already set we skip the
	// expensive network round-trip entirely.
	b.mu.RLock()
	same := b.model == model
	b.mu.RUnlock()
	if same {
		return nil
	}

	// Network call outside any lock — intentional; at worst two concurrent
	// callers both validate successfully and then race to the write lock
	// below, which is resolved atomically.
	if err := b.ollamaClient.EnsureModelExists(ctx, model); err != nil {
		return err
	}

	// Hold the write lock for the ENTIRE read-compare-write so there is no
	// window between checking and assigning. This is the single authoritative
	// critical section for b.model updates.
	b.mu.Lock()
	if b.model != model {
		b.model = model
	}
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

// SetPrinter updates the printer (used by TUI to inject its own printer).
func (b *Bridge) SetPrinter(p Printer) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.printer = p
}

// ProcessMessage runs the agentic loop for a user message.
// A 5-minute timeout is applied to prevent indefinite hangs if the Ollama API stalls.
func (b *Bridge) ProcessMessage(ctx context.Context, messages []ollama.Message) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	start := time.Now()

	// Log user message (last message in history before processing).
	if len(messages) > 0 {
		lastMsg := messages[len(messages)-1]
		if lastMsg.Role == "user" {
			b.logger.LogUserMessage(lastMsg.Content)
		}
	}

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
		contextLength := b.contextLength
		b.mu.RUnlock()

		// For the final response (no tool calls expected), use streaming.
		// For intermediate iterations with tool calls, we need the full response first.
		var resp *ollama.Message
		var err error

		// Check if this is potentially the final response by trying streaming first
		// If we get tool calls, we already have them from the streaming response
		resp, err = b.processMessageWithStreaming(ctx, model, messages, tools, temperature, contextLength)
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
			b.logger.LogLLMResponse(resp.Content)
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

// processMessageWithStreaming handles streaming response from Ollama and collects tool calls.
func (b *Bridge) processMessageWithStreaming(ctx context.Context, model string, messages []ollama.Message, tools []ollama.Tool, temperature float64, contextLength int) (*ollama.Message, error) {
	streamChan, err := b.ollamaClient.ChatStream(ctx, model, messages, tools, temperature, contextLength)
	if err != nil {
		// Fall back to non-streaming if streaming fails
		b.printer.PrintWarning(fmt.Sprintf("Streaming failed, falling back to non-streaming mode: %v", err))
		msg, chatErr := b.ollamaClient.Chat(ctx, model, messages, tools, temperature, contextLength)
		if chatErr != nil {
			return nil, chatErr
		}
		// CRITICAL: print the fallback response — REPL no longer does this
		if msg != nil && msg.Content != "" {
			b.printer.PrintResponseStart()
			b.printer.PrintResponseChunk(msg.Content)
			b.printer.PrintResponseEnd()
		}
		return msg, nil
	}

	var fullContent strings.Builder
	var allToolCalls []ollama.ToolCall
	seenToolCalls := make(map[string]bool) // track tool calls already seen by name+args
	started := false

	// Stream chunks as they arrive in real-time
	for event := range streamChan {
		// Check for stream errors
		if event.Err != nil {
			b.printer.PrintWarning(fmt.Sprintf("Stream error encountered: %v", event.Err))
			continue
		}

		// Only start printing when we get actual content (avoid empty frames on tool-only iterations)
		if event.Content != "" {
			if !started {
				b.printer.PrintResponseStart()
				started = true
			}
			fullContent.WriteString(event.Content)
			b.printer.PrintResponseChunk(event.Content)
		}

		// Collect tool calls from the event, deduplicating by full identity
		// (name + arguments) so the same tool can still be invoked in parallel
		// with different arguments, but identical repeats across stream events
		// are filtered out.
		if len(event.ToolCalls) > 0 {
			for _, tc := range event.ToolCalls {
				name, args := tc.NormalizeToolCall()
				argsJSON, _ := json.Marshal(args)
				key := name + "\x00" + string(argsJSON)
				if !seenToolCalls[key] {
					allToolCalls = append(allToolCalls, tc)
					seenToolCalls[key] = true
				}
			}
		}
	}

	// Only print end if we actually started printing
	if started {
		b.printer.PrintResponseEnd()
	}

	msg := &ollama.Message{
		Role:      "assistant",
		Content:   fullContent.String(),
		ToolCalls: allToolCalls,
	}

	// Log response with tool calls if present
	if len(allToolCalls) > 0 && fullContent.Len() > 0 {
		b.logger.LogInfo(fmt.Sprintf("LLM response with %d tool calls: %s", len(allToolCalls), fullContent.String()))
	}

	return msg, nil
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
			b.logger.LogToolCall(toolName, toolArgs)

			result, err := b.mcpClient.CallTool(ctx, toolName, toolArgs)
			success := err == nil
			if err != nil {
				result = fmt.Sprintf("error: %v", err)
				b.printer.PrintWarning(fmt.Sprintf("Tool %q failed: %v", toolName, err))
			}

			b.printer.PrintToolResult(toolName, result, len(result) > 120)
			b.logger.LogToolResult(toolName, result, success)
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

	// Sort tools by name for deterministic output
	sort.Slice(tools, func(i, j int) bool {
		return tools[i].Function.Name < tools[j].Function.Name
	})

	return tools
}

// IsStructuredContent checks if the response is JSON-wrapped structured content
func (b *Bridge) IsStructuredContent(response string) bool {
	return content.IsStructuredContent(response)
}

// ParseStructuredContent parses a structured content response
func (b *Bridge) ParseStructuredContent(response string) (*content.StructuredContent, error) {
	return content.ParseStructuredContent([]byte(response))
}
