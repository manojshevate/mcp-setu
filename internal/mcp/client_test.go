package mcp

import (
	"testing"
)

// TestValidateCommand tests command path validation.
func TestValidateCommand(t *testing.T) {
	tests := []struct {
		cmd       string
		shouldErr bool
		message   string
	}{
		{"npx", false, "Simple executable name should be valid"},
		{"node", false, "Simple executable name should be valid"},
		{"../../../etc/passwd", true, "Directory traversal should be invalid"},
		{"./relative/path", true, "Relative path should be invalid"},
		{"relative/path", true, "Path with separator should be invalid"},
		{"command..name", true, "Double dots should be invalid"},
	}

	for _, tt := range tests {
		err := validateCommand(tt.cmd)
		hasErr := err != nil

		if hasErr != tt.shouldErr {
			if tt.shouldErr {
				t.Errorf("%s: %q - expected error, got nil", tt.message, tt.cmd)
			} else {
				t.Errorf("%s: %q - expected nil, got error: %v", tt.message, tt.cmd, err)
			}
		}
	}
}

// TestEnvMapToSlice tests environment variable conversion.
func TestEnvMapToSlice(t *testing.T) {
	envMap := map[string]string{
		"KEY1": "value1",
		"KEY2": "value2",
	}

	result := envMapToSlice(envMap)

	if len(result) != 2 {
		t.Errorf("Expected 2 env vars, got %d", len(result))
	}

	// Check that the format is correct (KEY=value)
	for _, env := range result {
		if len(env) < 3 {
			t.Errorf("Invalid env var format: %s", env)
		}
		// Should be in format "KEY=value"
		found := false
		if env == "KEY1=value1" || env == "KEY2=value2" {
			found = true
		}
		if !found {
			t.Errorf("Unexpected env var: %s", env)
		}
	}
}

// TestEnvMapToSliceEmpty tests empty environment map.
func TestEnvMapToSliceEmpty(t *testing.T) {
	envMap := map[string]string{}

	result := envMapToSlice(envMap)

	if len(result) != 0 {
		t.Errorf("Expected empty slice, got %d items", len(result))
	}
}

// TestJSONRPCRequest tests JSON-RPC request structure.
func TestJSONRPCRequest(t *testing.T) {
	id := 1
	req := JSONRPCRequest{
		JSONRPC: "2.0",
		ID:      &id,
		Method:  "tools/list",
		Params:  map[string]any{},
	}

	if req.JSONRPC != "2.0" {
		t.Errorf("Expected jsonrpc 2.0, got %s", req.JSONRPC)
	}

	if req.ID == nil || *req.ID != 1 {
		t.Errorf("Expected ID 1, got %v", req.ID)
	}

	if req.Method != "tools/list" {
		t.Errorf("Expected method tools/list, got %s", req.Method)
	}
}

// TestJSONRPCResponse tests JSON-RPC response structure.
func TestJSONRPCResponse(t *testing.T) {
	id := 1
	resp := JSONRPCResponse{
		JSONRPC: "2.0",
		ID:      &id,
		Result:  map[string]any{"tools": []any{}},
	}

	if resp.JSONRPC != "2.0" {
		t.Errorf("Expected jsonrpc 2.0, got %s", resp.JSONRPC)
	}

	if resp.ID == nil || *resp.ID != 1 {
		t.Errorf("Expected ID 1, got %v", resp.ID)
	}

	if resp.Error != nil {
		t.Error("Expected no error in successful response")
	}
}

// TestJSONRPCError tests JSON-RPC error structure.
func TestJSONRPCError(t *testing.T) {
	err := JSONRPCError{
		Code:    -32600,
		Message: "Invalid Request",
	}

	if err.Code != -32600 {
		t.Errorf("Expected error code -32600, got %d", err.Code)
	}

	if err.Message != "Invalid Request" {
		t.Errorf("Expected message Invalid Request, got %s", err.Message)
	}
}

// TestToolDefinition tests tool definition structure.
func TestToolDefinition(t *testing.T) {
	tool := Tool{
		Name:        "read_file",
		Description: "Read file contents",
		InputSchema: map[string]any{
			"type": "object",
			"properties": map[string]any{
				"path": map[string]any{
					"type":        "string",
					"description": "File path",
				},
			},
		},
	}

	if tool.Name != "read_file" {
		t.Errorf("Expected name read_file, got %s", tool.Name)
	}

	if tool.Description != "Read file contents" {
		t.Errorf("Expected description, got %s", tool.Description)
	}

	if tool.InputSchema == nil {
		t.Error("Expected InputSchema to be set")
	}
}

// TestInitializeRequest tests initialize request structure.
func TestInitializeRequest(t *testing.T) {
	req := InitializeRequest{
		ProtocolVersion: "2024-11-05",
		Capabilities:    map[string]any{"tools": map[string]any{}},
		ClientInfo: ClientInfo{
			Name:    "mcp-setu",
			Version: "0.1.0",
		},
	}

	if req.ProtocolVersion != "2024-11-05" {
		t.Errorf("Expected protocol version 2024-11-05, got %s", req.ProtocolVersion)
	}

	if req.ClientInfo.Name != "mcp-setu" {
		t.Errorf("Expected client name mcp-setu, got %s", req.ClientInfo.Name)
	}

	if req.ClientInfo.Version != "0.1.0" {
		t.Errorf("Expected client version 0.1.0, got %s", req.ClientInfo.Version)
	}
}

// TestMultiClientInitialization tests MultiClient initialization.
func TestMultiClientInitialization(t *testing.T) {
	mc := NewMultiClient()

	if mc == nil {
		t.Fatal("Expected non-nil MultiClient")
	}

	if mc.clients == nil {
		t.Error("Expected clients map to be initialized")
	}

	if mc.tools == nil {
		t.Error("Expected tools map to be initialized")
	}

	names := mc.GetAllServerNames()
	if len(names) != 0 {
		t.Errorf("Expected no servers initially, got %d", len(names))
	}
}

// TestGetAllTools tests getting all tools.
func TestGetAllTools(t *testing.T) {
	mc := NewMultiClient()

	// Initially empty
	tools := mc.GetAllTools()
	if len(tools) != 0 {
		t.Errorf("Expected no tools initially, got %d", len(tools))
	}
}
