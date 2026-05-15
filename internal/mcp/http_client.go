package mcp

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/manojshevate/mcp-setu/internal/version"
)

// HTTPStreamableClient implements MCP communication over HTTP with streaming support.
// This is the modern MCP transport standard for remote servers.
// Supports Bearer token authentication as per MCP specification.
type HTTPStreamableClient struct {
	name          string
	url           string
	httpClient    *http.Client
	tokenProvider TokenProvider
	tools         map[string]*Tool
	mu            sync.Mutex
	nextID        int
}

// NewHTTPStreamableClient creates a new HTTP Streamable client with optional authentication.
func NewHTTPStreamableClient(ctx context.Context, name string, url string, tokenProvider TokenProvider) (*HTTPStreamableClient, error) {
	if tokenProvider == nil {
		tokenProvider = &NoAuthProvider{}
	}

	c := &HTTPStreamableClient{
		name: name,
		url:  strings.TrimSuffix(url, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		tokenProvider: tokenProvider,
		tools:         make(map[string]*Tool),
		nextID:        1,
	}

	// Initialize the server
	if err := c.initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize HTTP Streamable server %q: %w", name, err)
	}

	return c, nil
}

// initialize sends the initialize request and processes the response.
func (c *HTTPStreamableClient) initialize(ctx context.Context) error {
	initReq := InitializeRequest{
		ProtocolVersion: MCPProtocolVersion,
		Capabilities: map[string]any{
			"tools": map[string]any{},
		},
		ClientInfo: ClientInfo{
			Name:    "mcp-setu",
			Version: version.Version,
		},
	}

	resp, err := c.sendRequest(ctx, "initialize", initReq)
	if err != nil {
		return fmt.Errorf("initialize request failed: %w", err)
	}
	if resp.Error != nil {
		return fmt.Errorf("initialize error: %s", resp.Error.Message)
	}

	// Send initialized notification (MUST include Authorization header per MCP spec).
	notif := InitializedNotification{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
		Params:  map[string]any{},
	}
	notifData, err := json.Marshal(notif)
	if err != nil {
		return fmt.Errorf("failed to marshal initialized notification: %w", err)
	}
	req, err := http.NewRequest("POST", c.url, bytes.NewReader(notifData))
	if err != nil {
		return fmt.Errorf("failed to create initialized notification request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Add Bearer token if available
	if c.tokenProvider != nil {
		if token, err := c.tokenProvider.GetToken(ctx, c.url); err == nil && token != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		}
	}
	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send initialized notification: %w", err)
	}
	defer httpResp.Body.Close()
	io.Copy(io.Discard, httpResp.Body)

	// List tools.
	toolResp, err := c.sendRequest(ctx, "tools/list", map[string]any{})
	if err != nil {
		return fmt.Errorf("tools/list request failed: %w", err)
	}
	if toolResp.Error != nil {
		return fmt.Errorf("tools/list error: %s", toolResp.Error.Message)
	}

	// Parse tools from result.
	if toolsData, ok := toolResp.Result["tools"]; ok {
		if toolsSlice, ok := toolsData.([]any); ok {
			for _, t := range toolsSlice {
				if toolMap, ok := t.(map[string]any); ok {
					var tool Tool
					toolJSON, err := json.Marshal(toolMap)
					if err != nil {
						fmt.Fprintf(os.Stderr, "warning: failed to marshal tool definition: %v\n", err)
						continue
					}
					if err := json.Unmarshal(toolJSON, &tool); err != nil {
						continue
					}
					c.tools[tool.Name] = &tool
				}
			}
		}
	}

	return nil
}

// GetTools returns all tools available from this server.
func (c *HTTPStreamableClient) GetTools() map[string]*Tool {
	c.mu.Lock()
	defer c.mu.Unlock()
	toolsCopy := make(map[string]*Tool)
	for k, v := range c.tools {
		toolsCopy[k] = v
	}
	return toolsCopy
}

// CallTool executes a tool and returns the result.
func (c *HTTPStreamableClient) CallTool(ctx context.Context, name string, arguments map[string]any) (string, error) {
	callReq := CallToolRequest{
		Name:      name,
		Arguments: arguments,
	}

	resp, err := c.sendRequest(ctx, "tools/call", callReq)
	if err != nil {
		return "", fmt.Errorf("tool call failed: %w", err)
	}
	if resp.Error != nil {
		return "", fmt.Errorf("tool error: %s", resp.Error.Message)
	}

	// Parse result.
	if resultData, ok := resp.Result["content"]; ok {
		if resultSlice, ok := resultData.([]any); ok && len(resultSlice) > 0 {
			if resultMap, ok := resultSlice[0].(map[string]any); ok {
				if content, ok := resultMap["text"].(string); ok {
					return content, nil
				}
			}
		}
	}

	return "", fmt.Errorf("unexpected tool result format")
}

// sendRequest sends a JSON-RPC request via HTTP POST with Bearer token authorization.
// Handles 401 responses by extracting scope requirements from WWW-Authenticate header.
func (c *HTTPStreamableClient) sendRequest(ctx context.Context, method string, params any) (*JSONRPCResponse, error) {
	c.mu.Lock()
	id := c.nextID
	c.nextID++
	c.mu.Unlock()

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      &id,
		Method:  method,
	}

	// Marshal params to map.
	if params != nil {
		paramJSON, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request params: %w", err)
		}
		if err := json.Unmarshal(paramJSON, &req.Params); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to unmarshal request params: %v\n", err)
		}
	}

	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON-RPC request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.url, bytes.NewReader(reqData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Add Bearer token if available (MCP spec: RFC 8707 Resource Indicators + OAuth 2.1)
	if c.tokenProvider != nil {
		token, err := c.tokenProvider.GetToken(ctx, c.url)
		if err != nil {
			// Fail loudly if auth provider is configured but cannot provide token
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
		if token != "" {
			httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		}
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Handle 401 Unauthorized with scope challenge (MCP spec: RFC 6750 + RFC 9728)
	if httpResp.StatusCode == http.StatusUnauthorized {
		wwwAuth := httpResp.Header.Get("WWW-Authenticate")
		if wwwAuth != "" {
			authParams := ParseWWWAuthenticate(wwwAuth)
			resourceMetadata := authParams["resource_metadata"]
			requiredScopes := authParams["scope"]
			if resourceMetadata != "" || requiredScopes != "" {
				errMsg := fmt.Sprintf("HTTP 401: Authorization required for %q", c.name)
				if resourceMetadata != "" {
					errMsg += fmt.Sprintf("\nResource metadata: %s", resourceMetadata)
				}
				if requiredScopes != "" {
					errMsg += fmt.Sprintf("\nRequired scopes: %s", requiredScopes)
				}
				return nil, fmt.Errorf("%s", errMsg)
			}
		}
		return nil, fmt.Errorf("HTTP 401: Unauthorized - authentication required")
	}

	if httpResp.StatusCode == http.StatusForbidden {
		// 403 Forbidden - may indicate insufficient scope (MCP spec: RFC 6750)
		wwwAuth := httpResp.Header.Get("WWW-Authenticate")
		if wwwAuth != "" && strings.Contains(wwwAuth, "insufficient_scope") {
			authParams := ParseWWWAuthenticate(wwwAuth)
			requiredScopes := authParams["scope"]
			return nil, fmt.Errorf("HTTP 403: Insufficient scope - required scopes: %s", requiredScopes)
		}
		return nil, fmt.Errorf("HTTP 403: Forbidden - insufficient permissions")
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request returned status %d", httpResp.StatusCode)
	}

	var resp JSONRPCResponse
	if err := json.NewDecoder(httpResp.Body).Decode(&resp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &resp, nil
}

// Close closes the HTTP client (no-op for HTTP).
func (c *HTTPStreamableClient) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}

// ============================================================================
// HTTP/SSE Client (Legacy/Deprecated but still supported)
// ============================================================================

// HTTPSSEClient implements MCP communication over HTTP with Server-Sent Events.
// This transport is deprecated in favor of HTTP Streamable but still supported for compatibility.
// Supports Bearer token authentication as per MCP specification.
type HTTPSSEClient struct {
	name          string
	url           string
	httpClient    *http.Client
	tokenProvider TokenProvider
	tools         map[string]*Tool
	mu            sync.Mutex
	nextID        int
	lastEventID   string
}

// NewHTTPSSEClient creates a new HTTP/SSE client with optional authentication.
func NewHTTPSSEClient(ctx context.Context, name string, url string, tokenProvider TokenProvider) (*HTTPSSEClient, error) {
	if tokenProvider == nil {
		tokenProvider = &NoAuthProvider{}
	}

	c := &HTTPSSEClient{
		name: name,
		url:  strings.TrimSuffix(url, "/"),
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		tokenProvider: tokenProvider,
		tools:         make(map[string]*Tool),
		nextID:        1,
	}

	// Initialize the server
	if err := c.initialize(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize HTTP/SSE server %q: %w", name, err)
	}

	return c, nil
}

// initialize sends the initialize request and processes the response.
func (c *HTTPSSEClient) initialize(ctx context.Context) error {
	initReq := InitializeRequest{
		ProtocolVersion: MCPProtocolVersion,
		Capabilities: map[string]any{
			"tools": map[string]any{},
		},
		ClientInfo: ClientInfo{
			Name:    "mcp-setu",
			Version: version.Version,
		},
	}

	resp, err := c.sendRequest(ctx, "initialize", initReq)
	if err != nil {
		return fmt.Errorf("initialize request failed: %w", err)
	}
	if resp.Error != nil {
		return fmt.Errorf("initialize error: %s", resp.Error.Message)
	}

	// Send initialized notification via POST (MUST include Authorization header per MCP spec).
	notif := InitializedNotification{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
		Params:  map[string]any{},
	}
	notifData, err := json.Marshal(notif)
	if err != nil {
		return fmt.Errorf("failed to marshal initialized notification: %w", err)
	}
	req, err := http.NewRequest("POST", c.url, bytes.NewReader(notifData))
	if err != nil {
		return fmt.Errorf("failed to create initialized notification request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	// Add Bearer token if available
	if c.tokenProvider != nil {
		if token, err := c.tokenProvider.GetToken(ctx, c.url); err == nil && token != "" {
			req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		}
	}
	httpResp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send initialized notification: %w", err)
	}
	defer httpResp.Body.Close()
	io.Copy(io.Discard, httpResp.Body)

	// List tools.
	toolResp, err := c.sendRequest(ctx, "tools/list", map[string]any{})
	if err != nil {
		return fmt.Errorf("tools/list request failed: %w", err)
	}
	if toolResp.Error != nil {
		return fmt.Errorf("tools/list error: %s", toolResp.Error.Message)
	}

	// Parse tools from result.
	if toolsData, ok := toolResp.Result["tools"]; ok {
		if toolsSlice, ok := toolsData.([]any); ok {
			for _, t := range toolsSlice {
				if toolMap, ok := t.(map[string]any); ok {
					var tool Tool
					toolJSON, err := json.Marshal(toolMap)
					if err != nil {
						fmt.Fprintf(os.Stderr, "warning: failed to marshal tool definition: %v\n", err)
						continue
					}
					if err := json.Unmarshal(toolJSON, &tool); err != nil {
						continue
					}
					c.tools[tool.Name] = &tool
				}
			}
		}
	}

	return nil
}

// GetTools returns all tools available from this server.
func (c *HTTPSSEClient) GetTools() map[string]*Tool {
	c.mu.Lock()
	defer c.mu.Unlock()
	toolsCopy := make(map[string]*Tool)
	for k, v := range c.tools {
		toolsCopy[k] = v
	}
	return toolsCopy
}

// CallTool executes a tool and returns the result.
func (c *HTTPSSEClient) CallTool(ctx context.Context, name string, arguments map[string]any) (string, error) {
	callReq := CallToolRequest{
		Name:      name,
		Arguments: arguments,
	}

	resp, err := c.sendRequest(ctx, "tools/call", callReq)
	if err != nil {
		return "", fmt.Errorf("tool call failed: %w", err)
	}
	if resp.Error != nil {
		return "", fmt.Errorf("tool error: %s", resp.Error.Message)
	}

	// Parse result.
	if resultData, ok := resp.Result["content"]; ok {
		if resultSlice, ok := resultData.([]any); ok && len(resultSlice) > 0 {
			if resultMap, ok := resultSlice[0].(map[string]any); ok {
				if content, ok := resultMap["text"].(string); ok {
					return content, nil
				}
			}
		}
	}

	return "", fmt.Errorf("unexpected tool result format")
}

// sendRequest sends a JSON-RPC request via HTTP POST with Bearer token and reads response from SSE stream.
func (c *HTTPSSEClient) sendRequest(ctx context.Context, method string, params any) (*JSONRPCResponse, error) {
	c.mu.Lock()
	id := c.nextID
	c.nextID++
	c.mu.Unlock()

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      &id,
		Method:  method,
	}

	// Marshal params to map.
	if params != nil {
		paramJSON, err := json.Marshal(params)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request params: %w", err)
		}
		if err := json.Unmarshal(paramJSON, &req.Params); err != nil {
			fmt.Fprintf(os.Stderr, "warning: failed to unmarshal request params: %v\n", err)
		}
	}

	reqData, err := json.Marshal(req)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JSON-RPC request: %w", err)
	}
	httpReq, err := http.NewRequestWithContext(ctx, "POST", c.url, bytes.NewReader(reqData))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")

	// Add Bearer token if available (MCP spec: OAuth 2.1)
	if c.tokenProvider != nil {
		token, err := c.tokenProvider.GetToken(ctx, c.url)
		if err != nil {
			// Fail loudly if auth provider is configured but cannot provide token
			return nil, fmt.Errorf("authentication failed: %w", err)
		}
		if token != "" {
			httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		}
	}

	// Support connection resumption with Last-Event-ID (MCP spec: RFC 9728)
	if c.lastEventID != "" {
		httpReq.Header.Set("Last-Event-ID", c.lastEventID)
	}

	httpResp, err := c.httpClient.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request failed: %w", err)
	}
	defer httpResp.Body.Close()

	// Handle 401 Unauthorized
	if httpResp.StatusCode == http.StatusUnauthorized {
		wwwAuth := httpResp.Header.Get("WWW-Authenticate")
		if wwwAuth != "" {
			authParams := ParseWWWAuthenticate(wwwAuth)
			resourceMetadata := authParams["resource_metadata"]
			requiredScopes := authParams["scope"]
			if resourceMetadata != "" || requiredScopes != "" {
				errMsg := fmt.Sprintf("HTTP 401: Authorization required for %q", c.name)
				if resourceMetadata != "" {
					errMsg += fmt.Sprintf("\nResource metadata: %s", resourceMetadata)
				}
				if requiredScopes != "" {
					errMsg += fmt.Sprintf("\nRequired scopes: %s", requiredScopes)
				}
				return nil, fmt.Errorf("%s", errMsg)
			}
		}
		return nil, fmt.Errorf("HTTP 401: Unauthorized - authentication required")
	}

	if httpResp.StatusCode == http.StatusForbidden {
		wwwAuth := httpResp.Header.Get("WWW-Authenticate")
		if wwwAuth != "" && strings.Contains(wwwAuth, "insufficient_scope") {
			authParams := ParseWWWAuthenticate(wwwAuth)
			requiredScopes := authParams["scope"]
			return nil, fmt.Errorf("HTTP 403: Insufficient scope - required scopes: %s", requiredScopes)
		}
		return nil, fmt.Errorf("HTTP 403: Forbidden - insufficient permissions")
	}

	if httpResp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("HTTP request returned status %d", httpResp.StatusCode)
	}

	// Read SSE stream
	return c.readSSEResponse(ctx, httpResp.Body, id)
}

// readSSEResponse reads a JSON-RPC response from an SSE stream.
func (c *HTTPSSEClient) readSSEResponse(ctx context.Context, body io.Reader, expectedID int) (*JSONRPCResponse, error) {
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Text()

		// Parse SSE event ID
		if strings.HasPrefix(line, "id:") {
			c.lastEventID = strings.TrimSpace(strings.TrimPrefix(line, "id:"))
		}

		// Parse SSE data
		if strings.HasPrefix(line, "data:") {
			jsonData := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
			var resp JSONRPCResponse
			if err := json.Unmarshal([]byte(jsonData), &resp); err != nil {
				continue
			}

			// Match response ID
			if resp.ID != nil && *resp.ID == expectedID {
				return &resp, nil
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanner error: %w", err)
	}

	return nil, fmt.Errorf("unexpected EOF while reading SSE response")
}

// Close closes the HTTP client (no-op for HTTP).
func (c *HTTPSSEClient) Close() error {
	c.httpClient.CloseIdleConnections()
	return nil
}
