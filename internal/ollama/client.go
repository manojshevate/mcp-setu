package ollama

import (
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

// CheckToolSupport verifies that a model supports tool calling and exists locally.
func (c *Client) CheckToolSupport(ctx context.Context, model string) error {
	// First, check if the model exists locally.
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

	// Check if the model supports tool calling.
	if !supportsToolCalling(model) {
		supportedList := strings.Join(KnownToolSupportedModels, ", ")
		return fmt.Errorf(
			"model %q does not support tool calling\n\n"+
				"Supported models: %s\n\n"+
				"→ Switch model in mcp.json or use: mcp-setu chat --model gemma4:e4b\n"+
				"→ See all local models: mcp-setu models",
			model, supportedList,
		)
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
			ToolSupported: supportsToolCalling(m.Name),
		})
	}

	return result, nil
}

// supportsToolCalling checks if a model name matches a known tool-supporting model prefix.
func supportsToolCalling(modelName string) bool {
	for _, prefix := range KnownToolSupportedModels {
		if strings.HasPrefix(modelName, prefix) {
			return true
		}
	}
	return false
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
