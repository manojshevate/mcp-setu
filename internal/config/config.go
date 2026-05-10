package config

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// DefaultTimeout is the default timeout for operations (server startup, API requests).
const DefaultTimeout = 300 * time.Second

// Config is intentionally compatible with Claude Desktop's MCP config format.
// The mcpServers block can be copied directly from:
//   macOS: ~/Library/Application Support/Claude/claude_desktop_config.json
//   Windows: %APPDATA%\Claude\claude_desktop_config.json
// Only the "ollama" block is mcpgo-specific.

// Config represents the mcpgo configuration file.
type Config struct {
	Ollama    OllamaConfig            `json:"ollama"`
	MCPServers map[string]ServerConfig `json:"mcpServers"`
}

// OllamaConfig represents Ollama-specific configuration.
type OllamaConfig struct {
	BaseURL       string  `json:"baseUrl"`
	Model         string  `json:"model"`
	SystemPrompt  string  `json:"systemPrompt"`
	Temperature   float64 `json:"temperature"`
	ContextLength int     `json:"contextLength"`
}

// ServerConfig represents the configuration for a single MCP server.
type ServerConfig struct {
	Command string            `json:"command"`
	Args    []string          `json:"args"`
	Env     map[string]string `json:"env,omitempty"`
}

// Load reads and validates the configuration from a file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("config file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg Config
	if err := json.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return &cfg, nil
}

// Validate checks that the configuration is valid.
func (c *Config) Validate() error {
	if c.Ollama.Model == "" {
		return fmt.Errorf("ollama.model is required")
	}
	if c.Ollama.BaseURL == "" {
		c.Ollama.BaseURL = "http://localhost:11434"
	}
	if c.Ollama.SystemPrompt == "" {
		c.Ollama.SystemPrompt = "You are a helpful assistant."
	}
	if c.Ollama.Temperature == 0 {
		c.Ollama.Temperature = 0.7
	}
	if c.Ollama.ContextLength == 0 {
		c.Ollama.ContextLength = 4096
	}
	if len(c.MCPServers) == 0 {
		return fmt.Errorf("mcpServers must contain at least one server")
	}
	return nil
}

// ExampleConfig returns a minimal valid example configuration.
func ExampleConfig() string {
	example := Config{
		Ollama: OllamaConfig{
			BaseURL:      "http://localhost:11434",
			Model:        "gemma4:e4b",
			SystemPrompt: "You are a helpful assistant.",
			Temperature:  0.7,
			ContextLength: 4096,
		},
		MCPServers: map[string]ServerConfig{
			"filesystem": {
				Command: "npx",
				Args:    []string{"-y", "@modelcontextprotocol/server-filesystem", "."},
			},
		},
	}
	data, _ := json.MarshalIndent(example, "", "  ")
	return string(data)
}
