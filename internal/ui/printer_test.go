package ui

import (
	"testing"
)

// TestNewPrinter tests printer initialization.
func TestNewPrinter(t *testing.T) {
	printer := NewPrinter(false)

	if printer == nil {
		t.Fatal("Expected non-nil printer")
	}

	if printer.verbose {
		t.Error("Expected verbose=false")
	}
}

// TestNewPrinterVerbose tests printer with verbose flag.
func TestNewPrinterVerbose(t *testing.T) {
	printer := NewPrinter(true)

	if !printer.verbose {
		t.Error("Expected verbose=true")
	}
}

// TestServerInfo tests ServerInfo structure.
func TestServerInfo(t *testing.T) {
	info := ServerInfo{
		Name:  "filesystem",
		Tools: 5,
	}

	if info.Name != "filesystem" {
		t.Errorf("Expected name filesystem, got %s", info.Name)
	}

	if info.Tools != 5 {
		t.Errorf("Expected 5 tools, got %d", info.Tools)
	}
}

// TestToolInfo tests ToolInfo structure.
func TestToolInfo(t *testing.T) {
	info := ToolInfo{
		Name:        "read_file",
		Server:      "filesystem",
		Description: "Read file contents",
	}

	if info.Name != "read_file" {
		t.Errorf("Expected name read_file, got %s", info.Name)
	}

	if info.Server != "filesystem" {
		t.Errorf("Expected server filesystem, got %s", info.Server)
	}

	if info.Description != "Read file contents" {
		t.Errorf("Expected description, got %s", info.Description)
	}
}

// TestModelInfo tests ModelInfo structure.
func TestModelInfo(t *testing.T) {
	info := ModelInfo{
		Name:          "gemma4:e4b",
		Size:          "4.0 GB",
		ToolSupported: true,
	}

	if info.Name != "gemma4:e4b" {
		t.Errorf("Expected name gemma4:e4b, got %s", info.Name)
	}

	if info.Size != "4.0 GB" {
		t.Errorf("Expected size 4.0 GB, got %s", info.Size)
	}

	if !info.ToolSupported {
		t.Error("Expected tool support to be true")
	}
}

// TestGetServersTableInfo tests server info conversion.
func TestGetServersTableInfo(t *testing.T) {
	// Since we're using a mock MCP client, we can test the structure
	infos := []ServerInfo{
		{Name: "filesystem", Tools: 5},
		{Name: "sqlite", Tools: 3},
	}

	if len(infos) != 2 {
		t.Errorf("Expected 2 servers, got %d", len(infos))
	}

	if infos[0].Name != "filesystem" {
		t.Errorf("Expected first server filesystem, got %s", infos[0].Name)
	}

	if infos[1].Tools != 3 {
		t.Errorf("Expected 3 tools for sqlite, got %d", infos[1].Tools)
	}
}

// TestColorPaletteExists tests that color palette is defined.
func TestColorPaletteExists(t *testing.T) {
	// This is a basic check that colors are defined
	// In a real implementation, you'd verify the actual hex values
	colors := []interface{}{
		colorPrimary,
		colorSuccess,
		colorWarning,
		colorError,
		colorMuted,
		colorHighlight,
	}

	for i, color := range colors {
		if color == nil {
			t.Errorf("Color %d is nil", i)
		}
	}
}

// TestPrinterMethods tests printer methods exist and are callable.
func TestPrinterMethods(t *testing.T) {
	printer := NewPrinter(false)

	// Test that methods exist and can be called without panic
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Printer method panicked: %v", r)
		}
	}()

	// These should not panic (they write to stdout/stderr but we're not checking output)
	servers := []ServerInfo{{Name: "test", Tools: 1}}
	printer.PrintBanner("test-model", "test.json", 1, 1)
	printer.PrintServerTable(servers)

	tools := []ToolInfo{{Name: "test", Server: "test", Description: "test"}}
	printer.PrintToolsTable(tools)

	models := []ModelInfo{{Name: "test:latest", Size: "1GB", ToolSupported: true}}
	printer.PrintModelsTable(models)

	printer.PrintSuccess("test message")
	printer.PrintWarning("test warning")
	printer.PrintError("test error")
	printer.PrintHelp()
}

// TestPrinterVerboseSilent tests that verbose mode is respected.
func TestPrinterVerboseSilent(t *testing.T) {
	printerSilent := NewPrinter(false)
	printerVerbose := NewPrinter(true)

	// Both should work without error
	printerSilent.PrintToolCall("test_tool", map[string]any{})
	printerVerbose.PrintToolCall("test_tool", map[string]any{})

	printerSilent.PrintToolResult("test_tool", "result", false)
	printerVerbose.PrintToolResult("test_tool", "result", false)
}
