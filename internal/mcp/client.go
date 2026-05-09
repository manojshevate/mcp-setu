package mcp

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/manojshevate/mcpgo/internal/config"
)

// Client manages a connection to a single MCP server via stdio.
type Client struct {
	name    string
	cmd     *exec.Cmd
	stdin   *bufio.Writer
	stdout  *bufio.Reader
	process *os.Process
	tools   map[string]*Tool
	mu      sync.Mutex
	nextID  int
}

// NewClient creates a new MCP client and spawns the server process.
func NewClient(ctx context.Context, name string, cfg config.ServerConfig) (*Client, error) {
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
		ProtocolVersion: "2024-11-05",
		Capabilities: map[string]any{
			"tools": map[string]any{},
		},
		ClientInfo: ClientInfo{
			Name:    "mcpgo",
			Version: "0.1.0",
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
					_ = json.Unmarshal(toolJSON, &tool)
					c.tools[tool.Name] = &tool
				}
			}
		}
	}

	return nil
}

// GetTools returns all tools available from this server.
func (c *Client) GetTools() map[string]*Tool {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.tools
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
func (c *Client) sendRequest(ctx context.Context, method string, params any) (*JSONRPCResponse, error) {
	c.mu.Lock()
	id := c.nextID
	c.nextID++
	c.mu.Unlock()

	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      id,
		Method:  method,
	}

	// Marshal params to map.
	if params != nil {
		paramJSON, _ := json.Marshal(params)
		_ = json.Unmarshal(paramJSON, &req.Params)
	}

	reqData, _ := json.Marshal(req)
	if _, err := c.stdin.WriteString(string(reqData) + "\n"); err != nil {
		return nil, fmt.Errorf("failed to write request: %w", err)
	}
	if err := c.stdin.Flush(); err != nil {
		return nil, fmt.Errorf("failed to flush request: %w", err)
	}

	// Read response.
	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
		}

		line, err := c.stdout.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		var resp JSONRPCResponse
		if err := json.Unmarshal([]byte(line), &resp); err != nil {
			continue
		}

		// Match response ID.
		if resp.ID == id {
			return &resp, nil
		}
	}
}

// Close closes the connection to the MCP server.
func (c *Client) Close() error {
	if c.process == nil {
		return nil
	}

	// Try SIGTERM first.
	_ = c.process.Signal(os.Interrupt)

	// Wait a bit for graceful shutdown.
	time.Sleep(3 * time.Second)

	// Check if still running.
	if c.cmd.ProcessState == nil || !c.cmd.ProcessState.Exited() {
		_ = c.process.Kill()
	}

	return c.cmd.Wait()
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
	clients map[string]*Client
	tools   map[string]string // tool name -> server name
	mu      sync.Mutex
}

// NewMultiClient creates a new multi-client manager.
func NewMultiClient() *MultiClient {
	return &MultiClient{
		clients: make(map[string]*Client),
		tools:   make(map[string]string),
	}
}

// Add adds a new server connection.
func (mc *MultiClient) Add(name string, client *Client) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.clients[name] = client
	for toolName := range client.GetTools() {
		mc.tools[toolName] = name
	}
}

// GetServer returns the client for a given server name.
func (mc *MultiClient) GetServer(name string) *Client {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	return mc.clients[name]
}

// GetServerForTool returns the client for the server that owns a given tool.
func (mc *MultiClient) GetServerForTool(toolName string) *Client {
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
