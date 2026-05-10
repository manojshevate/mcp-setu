package mcp

const (
	// MCPProtocolVersion specifies the MCP protocol version to use.
	// Using the latest stable version (2025-11-25) which includes improved
	// authentication handling, parallel tool calls, and server-side agent loops.
	MCPProtocolVersion = "2025-11-25"
)

// JSONRPCRequest represents a JSON-RPC 2.0 request.
type JSONRPCRequest struct {
	JSONRPC string        `json:"jsonrpc"`
	ID      *int          `json:"id,omitempty"` // pointer to distinguish null from omitted
	Method  string        `json:"method"`
	Params  map[string]any `json:"params,omitempty"`
}

// JSONRPCResponse represents a JSON-RPC 2.0 response.
type JSONRPCResponse struct {
	JSONRPC string      `json:"jsonrpc"`
	ID      *int        `json:"id,omitempty"` // pointer to distinguish null from omitted
	Result  map[string]any `json:"result,omitempty"`
	Error   *JSONRPCError `json:"error,omitempty"`
}

// JSONRPCError represents a JSON-RPC 2.0 error.
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

// InitializeRequest is the parameters for the initialize method.
type InitializeRequest struct {
	ProtocolVersion string `json:"protocolVersion"`
	Capabilities    map[string]any `json:"capabilities"`
	ClientInfo      ClientInfo `json:"clientInfo"`
}

// ClientInfo identifies the client to the MCP server.
type ClientInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// InitializedNotification is sent by the client after initialization.
type InitializedNotification struct {
	JSONRPC string `json:"jsonrpc"`
	Method  string `json:"method"`
	Params  map[string]any `json:"params"`
}

// Tool represents an MCP tool.
type Tool struct {
	Name        string      `json:"name"`
	Description string      `json:"description"`
	InputSchema map[string]any `json:"inputSchema"`
}

// CallToolRequest is the parameters for calling a tool.
type CallToolRequest struct {
	Name      string      `json:"name"`
	Arguments map[string]any `json:"arguments"`
}

