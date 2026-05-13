package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/manojshevate/mcp-setu/internal/config"
	"github.com/manojshevate/mcp-setu/internal/version"
)

// MCPClientInterface defines the methods all MCP client implementations must support.
type MCPClientInterface interface {
	GetTools() map[string]*Tool
	CallTool(ctx context.Context, name string, arguments map[string]any) (string, error)
	Close() error
}

// Client manages a connection to a single MCP server via stdio.
type Client struct {
	name      string
	cmd       *exec.Cmd
	stdin     *bufio.Writer
	stdout    *bufio.Reader
	process   *os.Process
	tools     map[string]*Tool
	mu        sync.Mutex
	nextID    int
	closeOnce sync.Once
}

// NewClient creates a new MCP client based on the transport type.
// Supports: stdio (default), http-streamable, and http-sse transports.
// Automatically configures authentication based on the auth config.
func NewClient(ctx context.Context, name string, cfg config.ServerConfig) (MCPClientInterface, error) {
	// Determine transport type (default to stdio)
	transportType := cfg.Type
	if transportType == "" {
		transportType = "stdio"
	}

	// Create token provider for HTTP transports
	var tokenProvider TokenProvider
	var err error
	if cfg.Auth != nil && (transportType == "http-streamable" || transportType == "http-sse") {
		tokenProvider, err = NewTokenProvider(cfg.Auth)
		if err != nil {
			return nil, fmt.Errorf("failed to create token provider: %w", err)
		}
	}

	switch transportType {
	case "stdio":
		return NewStdioClient(ctx, name, cfg)
	case "http-streamable":
		if cfg.URL == "" {
			return nil, fmt.Errorf("http-streamable transport requires 'url' field")
		}
		return NewHTTPStreamableClient(ctx, name, cfg.URL, tokenProvider)
	case "http-sse":
		if cfg.URL == "" {
			return nil, fmt.Errorf("http-sse transport requires 'url' field")
		}
		return NewHTTPSSEClient(ctx, name, cfg.URL, tokenProvider)
	default:
		return nil, fmt.Errorf("unknown transport type: %q", transportType)
	}
}

// NewStdioClient creates a new stdio-based MCP client and spawns the server process.
// It validates the command path to prevent command injection vulnerabilities.
func NewStdioClient(ctx context.Context, name string, cfg config.ServerConfig) (*Client, error) {
	// Validate command: must be absolute path or a standard executable in PATH.
	if err := validateCommand(cfg.Command); err != nil {
		return nil, fmt.Errorf("invalid command for server %q: %w", name, err)
	}

	// Add timeout to prevent hanging during server startup.
	ctx, cancel := context.WithTimeout(ctx, config.DefaultTimeout)
	defer cancel()

	cmd := exec.CommandContext(ctx, cfg.Command, cfg.Args...)
	cmd.Env = append(os.Environ(), envMapToSlice(cfg.Env)...)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("failed to create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start MCP server %q: %w", name, err)
	}

	c := &Client{
		name:    name,
		cmd:     cmd,
		stdin:   bufio.NewWriter(stdin),
		stdout:  bufio.NewReader(stdout),
		process: cmd.Process,
		tools:   make(map[string]*Tool),
		nextID:  1,
	}

	// Initialize the server.
	if err := c.initialize(ctx); err != nil {
		_ = c.Close()
		return nil, fmt.Errorf("failed to initialize MCP server %q: %w", name, err)
	}

	return c, nil
}

// initialize sends the initialize request and processes the response.
func (c *Client) initialize(ctx context.Context) error {
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

	// Send initialized notification.
	notif := InitializedNotification{
		JSONRPC: "2.0",
		Method:  "notifications/initialized",
		Params:  map[string]any{},
	}
	notifData, _ := json.Marshal(notif)
	_, _ = c.stdin.WriteString(string(notifData) + "\n")
	_ = c.stdin.Flush()

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
					toolJSON, _ := json.Marshal(toolMap)
					if err := json.Unmarshal(toolJSON, &tool); err != nil {
						// Log but continue to allow partial tool loading.
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
// Returns a defensive copy to prevent concurrent map modification.
func (c *Client) GetTools() map[string]*Tool {
	c.mu.Lock()
	defer c.mu.Unlock()
	toolsCopy := make(map[string]*Tool)
	for k, v := range c.tools {
		toolsCopy[k] = v
	}
	return toolsCopy
}

// CallTool executes a tool and returns the result.
func (c *Client) CallTool(ctx context.Context, name string, arguments map[string]any) (string, error) {
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

// sendRequest sends a JSON-RPC request and waits for a response.
// Note: This serializes all I/O to prevent race conditions on the stdio stream.
// Multiple concurrent tool calls should be serialized at a higher level.
func (c *Client) sendRequest(ctx context.Context, method string, params any) (*JSONRPCResponse, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	id := c.nextID
	c.nextID++

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      &id,
		Method:  method,
	}

	// Marshal params to map.
	if params != nil {
		paramJSON, _ := json.Marshal(params)
		_ = json.Unmarshal(paramJSON, &req.Params)
		// Treat empty params map as nil for clean omitempty marshaling
		if len(req.Params) == 0 {
			req.Params = nil
		}
	}

	reqData, _ := json.Marshal(req)
	if _, err := c.stdin.WriteString(string(reqData) + "\n"); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}
	if err := c.stdin.Flush(); err != nil {
		return nil, fmt.Errorf("failed to flush request: %w", err)
	}

	// Read response using a scanner with size limits to prevent unbounded memory growth.
	scanner := bufio.NewScanner(c.stdout)
	scanner.Buffer(make([]byte, 4096), 1024*1024) // Max 1MB per line

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		if !scanner.Scan() {
			if err := scanner.Err(); err != nil {
				return nil, fmt.Errorf("scanner error: %w", err)
			}
			return nil, fmt.Errorf("unexpected EOF while reading response")
		}

		line := scanner.Bytes()

		var resp JSONRPCResponse
		if err := json.Unmarshal(line, &resp); err != nil {
			continue
		}

		// Match response ID.
		if resp.ID != nil && *resp.ID == id {
			return &resp, nil
		}
	}
}

// Close closes the connection to the MCP server.
func (c *Client) Close() error {
	var closeErr error
	c.closeOnce.Do(func() {
		if c.process == nil {
			return
		}

		// Try SIGTERM first.
		if err := c.process.Signal(os.Interrupt); err != nil {
			// Process may have already exited; try to kill it instead.
			_ = c.process.Kill()
		}

		// Wait for process to exit with a timeout of 3 seconds.
		done := make(chan error, 1)
		go func() {
			done <- c.cmd.Wait()
		}()

		select {
		case err := <-done:
			closeErr = err
		case <-time.After(3 * time.Second):
			// Timeout: force kill the process
			if err := c.process.Kill(); err != nil {
				closeErr = fmt.Errorf("failed to kill process: %w", err)
				return
			}
			// Wait for kill to be processed
			closeErr = <-done
		}
	})
	return closeErr
}

// validateCommand checks that a command is safe to execute.
// It must be either an absolute path or a simple executable name (for PATH lookup).
func validateCommand(cmd string) error {
	// Disallow paths with directory traversal or relative paths.
	if strings.Contains(cmd, "..") {
		return fmt.Errorf("command cannot contain '..' (directory traversal)")
	}

	// If it's an absolute path, it must exist and be executable.
	if filepath.IsAbs(cmd) {
		info, err := os.Stat(cmd)
		if err != nil {
			return fmt.Errorf("command path does not exist: %w", err)
		}
		if info.IsDir() {
			return fmt.Errorf("command path is a directory, not an executable")
		}
		return nil
	}

	// For relative commands (like "npx", "node"), they should be simple names
	// without slashes (except for PATH resolution).
	if strings.Contains(cmd, string(filepath.Separator)) && !filepath.IsAbs(cmd) {
		return fmt.Errorf("relative command paths are not allowed; use absolute path or command name")
	}

	return nil
}

// envMapToSlice converts a map to an environment variable slice.
func envMapToSlice(m map[string]string) []string {
	var result []string
	for k, v := range m {
		result = append(result, fmt.Sprintf("%s=%s", k, v))
	}
	return result
}

// MultiClient manages multiple MCP server connections.
type MultiClient struct {
	clients map[string]MCPClientInterface
	tools   map[string]string // tool name -> server name
	mu      sync.Mutex
}

// NewMultiClient creates a new multi-client manager.
func NewMultiClient() *MultiClient {
	return &MultiClient{
		clients: make(map[string]MCPClientInterface),
		tools:   make(map[string]string),
	}
}

// Add adds a new server connection.
func (mc *MultiClient) Add(name string, client MCPClientInterface) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.clients[name] = client
	for toolName := range client.GetTools() {
		mc.tools[toolName] = name
	}
}

// GetServer returns the client for a given server name.
func (mc *MultiClient) GetServer(name string) MCPClientInterface {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	return mc.clients[name]
}

// GetServerForTool returns the client for the server that owns a given tool.
func (mc *MultiClient) GetServerForTool(toolName string) MCPClientInterface {
	mc.mu.Lock()
	serverName := mc.tools[toolName]
	mc.mu.Unlock()
	if serverName == "" {
		return nil
	}
	return mc.GetServer(serverName)
}

// GetAllTools returns all tools from all servers with their source.
func (mc *MultiClient) GetAllTools() map[string]string {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	toolsCopy := make(map[string]string)
	for k, v := range mc.tools {
		toolsCopy[k] = v
	}
	return toolsCopy
}

// CallTool calls a tool on the appropriate server.
func (mc *MultiClient) CallTool(ctx context.Context, toolName string, arguments map[string]any) (string, error) {
	client := mc.GetServerForTool(toolName)
	if client == nil {
		return "", fmt.Errorf("tool %q not found", toolName)
	}
	return client.CallTool(ctx, toolName, arguments)
}

// GetAllServerNames returns the names of all connected servers.
func (mc *MultiClient) GetAllServerNames() []string {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	var names []string
	for name := range mc.clients {
		names = append(names, name)
	}
	return names
}

// CloseAll closes all server connections.
func (mc *MultiClient) CloseAll() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	var errs []string
	for name, client := range mc.clients {
		if err := client.Close(); err != nil {
			errs = append(errs, fmt.Sprintf("%s: %v", name, err))
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("errors closing servers: %s", strings.Join(errs, "; "))
	}
	return nil
}
