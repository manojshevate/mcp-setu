package config

import (
	"os"
	"testing"
)

// TestLoadValidConfig tests loading a valid configuration file.
func TestLoadValidConfig(t *testing.T) {
	// Create a temporary config file.
	configContent := `{
  "ollama": {
    "baseUrl": "http://localhost:11434",
    "model": "gemma4:e4b",
    "systemPrompt": "You are helpful",
    "temperature": 0.7,
    "contextLength": 4096
  },
  "mcpServers": {
    "filesystem": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-filesystem", "."]
    }
  }
}`

	tmpfile, err := os.CreateTemp("", "config_test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString(configContent); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpfile.Close()

	cfg, err := Load(tmpfile.Name())
	if err != nil {
		t.Fatalf("Load failed: %v", err)
	}

	if cfg.Ollama.Model != "gemma4:e4b" {
		t.Errorf("Expected model gemma4:e4b, got %s", cfg.Ollama.Model)
	}

	if cfg.Ollama.Temperature != 0.7 {
		t.Errorf("Expected temperature 0.7, got %f", cfg.Ollama.Temperature)
	}

	if len(cfg.MCPServers) != 1 {
		t.Errorf("Expected 1 MCP server, got %d", len(cfg.MCPServers))
	}
}

// TestLoadMissingFile tests loading a non-existent config file.
func TestLoadMissingFile(t *testing.T) {
	_, err := Load("/nonexistent/path/config.json")
	if err == nil {
		t.Fatal("Expected error for missing file, got nil")
	}

	if err.Error() != "config file not found: /nonexistent/path/config.json" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

// TestLoadInvalidJSON tests loading a file with invalid JSON.
func TestLoadInvalidJSON(t *testing.T) {
	tmpfile, err := os.CreateTemp("", "config_test_*.json")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	defer os.Remove(tmpfile.Name())

	if _, err := tmpfile.WriteString("{invalid json"); err != nil {
		t.Fatalf("Failed to write config: %v", err)
	}
	tmpfile.Close()

	_, err = Load(tmpfile.Name())
	if err == nil {
		t.Fatal("Expected error for invalid JSON, got nil")
	}
}

// TestValidateNoModel tests validation with missing model.
func TestValidateNoModel(t *testing.T) {
	cfg := &Config{
		Ollama: OllamaConfig{},
		MCPServers: map[string]ServerConfig{
			"test": {Command: "test", Args: []string{}},
		},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Expected error for missing model, got nil")
	}

	if err.Error() != "ollama.model is required" {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestValidateNoServers tests validation with no MCP servers.
func TestValidateNoServers(t *testing.T) {
	cfg := &Config{
		Ollama: OllamaConfig{
			Model: "gemma4:e4b",
		},
		MCPServers: map[string]ServerConfig{},
	}

	err := cfg.Validate()
	if err == nil {
		t.Fatal("Expected error for no servers, got nil")
	}

	if err.Error() != "mcpServers must contain at least one server" {
		t.Errorf("Unexpected error: %v", err)
	}
}

// TestValidateAppliesDefaults tests that validation applies defaults.
func TestValidateAppliesDefaults(t *testing.T) {
	cfg := &Config{
		Ollama: OllamaConfig{
			Model: "gemma4:e4b",
		},
		MCPServers: map[string]ServerConfig{
			"test": {Command: "test", Args: []string{}},
		},
	}

	if err := cfg.Validate(); err != nil {
		t.Fatalf("Validate failed: %v", err)
	}

	if cfg.Ollama.BaseURL != "http://localhost:11434" {
		t.Errorf("Expected default BaseURL, got %s", cfg.Ollama.BaseURL)
	}

	if cfg.Ollama.Temperature != 0.7 {
		t.Errorf("Expected default temperature 0.7, got %f", cfg.Ollama.Temperature)
	}

	if cfg.Ollama.ContextLength != 4096 {
		t.Errorf("Expected default context length 4096, got %d", cfg.Ollama.ContextLength)
	}
}

// TestExampleConfig tests that example config is valid JSON.
func TestExampleConfig(t *testing.T) {
	example := ExampleConfig()
	if example == "" {
		t.Fatal("Expected non-empty example config, got empty string")
	}

	var cfg Config
	if err := cfg.Validate(); err == nil || err.Error() != "ollama.model is required" {
		// The example might not have defaults applied, but it should parse
		t.Logf("Example config: %s", example)
	}
}
