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
// Only the "ollama" block is mcp-setu-specific.

// Config represents the mcp-setu configuration file.
type Config struct {
	Ollama    OllamaConfig            `json:"ollama"`
	MCPServers map[string]ServerConfig `json:"mcpServers"`
}

// OllamaConfig represents Ollama-specific configuration.
type OllamaConfig struct {
	BaseURL       string   `json:"baseUrl"`
	Model         string   `json:"model"`
	SystemPrompt  string   `json:"systemPrompt"`
	Temperature   *float64 `json:"temperature"`
	ContextLength *int     `json:"contextLength"`
}

// ServerConfig represents the configuration for a single MCP server.
// Supports multiple transport types: stdio (default), http-streamable, and http-sse.
type ServerConfig struct {
	// Transport type: "stdio" (default), "http-streamable", or "http-sse"
	Type string `json:"type,omitempty"`

	// Stdio transport fields
	Command string            `json:"command,omitempty"`
	Args    []string          `json:"args,omitempty"`
	Env     map[string]string `json:"env,omitempty"`

	// HTTP transport fields
	URL string `json:"url,omitempty"`

	// Authentication configuration (optional)
	Auth *AuthConfig `json:"auth,omitempty"`
}

// AuthConfig represents authentication configuration for MCP servers.
type AuthConfig struct {
	// Auth type: "none" (default), "oauth2", "bearer-token", or "env"
	Type string `json:"type,omitempty"`

	// OAuth 2.1 configuration
	AuthorizationServerURL string   `json:"authorizationServerUrl,omitempty"`
	ClientID               string   `json:"clientId,omitempty"`
	ClientSecret           string   `json:"clientSecret,omitempty"` // Should be retrieved from secure storage in production
	Scopes                 []string `json:"scopes,omitempty"`       // OAuth scopes to request

	// Bearer token (for simple token-based auth)
	Token string `json:"token,omitempty"` // Should be retrieved from secure storage

	// Environment variable names for auth
	TokenEnvVar               string `json:"tokenEnvVar,omitempty"`               // Env var containing bearer token
	AuthorizationServerEnvVar string `json:"authorizationServerEnvVar,omitempty"` // Env var for auth server URL
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
	if c.Ollama.Temperature == nil {
		defaultTemp := 0.7
		c.Ollama.Temperature = &defaultTemp
	}
	if c.Ollama.ContextLength == nil {
		defaultCtx := 4096
		c.Ollama.ContextLength = &defaultCtx
	}
	if len(c.MCPServers) == 0 {
		return fmt.Errorf("mcpServers must contain at least one server")
	}
	return nil
}

// ExampleConfig returns a minimal valid example configuration.
func ExampleConfig() string {
	defaultTemp := 0.7
	defaultCtx := 4096
	example := Config{
		Ollama: OllamaConfig{
			BaseURL:       "http://localhost:11434",
			Model:         "gemma2:latest",
			SystemPrompt:  "You are an AI agent that helps users by making calls to the configured MCP tools. Use the available tools to fulfill user requests accurately and efficiently.",
			Temperature:   &defaultTemp,
			ContextLength: &defaultCtx,
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
