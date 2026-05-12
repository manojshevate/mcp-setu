package ollama

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/manojshevate/mcp-setu/internal/config"
)

// Client is an HTTP client for the Ollama API.
type Client struct {
	baseURL string
	http    *http.Client
}

// NewClient creates a new Ollama client with a default timeout.
func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimSuffix(baseURL, "/"),
		http: &http.Client{
			Timeout: config.DefaultTimeout,
		},
	}
}

// Chat sends a chat request to Ollama and returns the response.
func (c *Client) Chat(ctx context.Context, model string, messages []Message, tools []Tool, temperature float64) (*Message, error) {
	req := ChatRequest{
		Model:       model,
		Messages:    messages,
		Tools:       tools,
		Temperature: temperature,
		Stream:      false,
	}

	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewReader(reqData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("ollama error %d: %s", resp.StatusCode, string(body))
	}

	rawBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var chatResp ChatResponse
	if err := json.NewDecoder(bytes.NewReader(rawBody)).Decode(&chatResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &chatResp.Message, nil
}

// StreamEvent represents a chunk from a streaming chat response.
type StreamEvent struct {
	Content   string
	ToolCalls []ToolCall
	Done      bool
	Err       error
}

// ChatStream sends a streaming chat request to Ollama and returns a channel yielding response events.
// Each event contains content chunks and any tool calls found.
// The channel is closed when the stream is complete.
func (c *Client) ChatStream(ctx context.Context, model string, messages []Message, tools []Tool, temperature float64) (<-chan StreamEvent, error) {
	req := ChatRequest{
		Model:       model,
		Messages:    messages,
		Tools:       tools,
		Temperature: temperature,
		Stream:      true,
	}

	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/chat", bytes.NewReader(reqData))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("ollama error %d: %s", resp.StatusCode, string(body))
	}

	// Create a channel to send events
	events := make(chan StreamEvent, 10)

	// Start a goroutine to read the stream and send events
	go func() {
		defer resp.Body.Close()
		defer close(events)

		// Use bufio.Reader with ReadBytes to avoid 64KiB scanner buffer limit
		// that can silently truncate large tool_calls or metadata
		reader := bufio.NewReader(resp.Body)
		for {
			line, err := reader.ReadBytes('\n')
			if err != nil && err != io.EOF {
				// Send error via event
				select {
				case events <- StreamEvent{Err: fmt.Errorf("failed to read stream: %w", err)}:
				case <-ctx.Done():
				}
				return
			}

			// Trim the trailing newline
			line = bytes.TrimSuffix(line, []byte("\n"))
			if len(line) == 0 {
				if err == io.EOF {
					break
				}
				continue
			}

			var chatResp ChatResponse
			if err := json.Unmarshal(line, &chatResp); err != nil {
				// Skip malformed lines but don't exit
				if err == io.EOF {
					break
				}
				continue
			}

			event := StreamEvent{
				Content:   chatResp.Message.Content,
				ToolCalls: chatResp.Message.ToolCalls,
				Done:      chatResp.Done,
			}

			select {
			case events <- event:
			case <-ctx.Done():
				return
			}

			// If done, we can exit early
			if chatResp.Done {
				break
			}
		}
	}()

	return events, nil
}

// CheckToolSupport verifies that a model exists locally.
func (c *Client) CheckToolSupport(ctx context.Context, model string) error {
	// Check if the model exists locally.
	reqBody, _ := json.Marshal(map[string]string{"name": model})
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.baseURL+"/api/show", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to check model: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf(
			"model %q not found locally\n\n"+
				"→ run: ollama pull %s",
			model, model,
		)
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ollama error %d", resp.StatusCode)
	}

	return nil
}

// ListLocalModels returns all locally available models with their tool support status.
func (c *Client) ListLocalModels(ctx context.Context) ([]ModelInfo, error) {
	httpReq, err := http.NewRequestWithContext(ctx, "GET", c.baseURL+"/api/tags", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := c.http.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("ollama error %d", resp.StatusCode)
	}

	var modelsResp ModelsResponse
	if err := json.NewDecoder(resp.Body).Decode(&modelsResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	var result []ModelInfo
	for _, m := range modelsResp.Models {
		result = append(result, ModelInfo{
			Name:          m.Name,
			Size:          formatBytes(m.Size),
			ToolSupported: true,
		})
	}

	return result, nil
}

// formatBytes converts a byte count to a human-readable string.
func formatBytes(b int64) string {
	const unit = 1024
	if b < unit {
		return fmt.Sprintf("%d B", b)
	}
	div, exp := int64(unit), 0
	for n := b / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(b)/float64(div), "KMGTPE"[exp])
}
