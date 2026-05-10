package mcp

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/manojshevate/mcp-setu/internal/config"
)

// TokenProvider handles authentication and provides access tokens for MCP servers.
type TokenProvider interface {
	// GetToken returns a valid access token for the given resource/scope.
	// Returns the token and any error encountered.
	GetToken(ctx context.Context, resource string) (string, error)

	// Close performs any cleanup needed by the token provider.
	Close() error
}

// NoAuthProvider provides no authentication (suitable for unauthenticated servers).
type NoAuthProvider struct{}

func (p *NoAuthProvider) GetToken(ctx context.Context, resource string) (string, error) {
	return "", nil // No token
}

func (p *NoAuthProvider) Close() error {
	return nil
}

// BearerTokenProvider provides a static bearer token.
type BearerTokenProvider struct {
	token string
}

func NewBearerTokenProvider(token string) *BearerTokenProvider {
	return &BearerTokenProvider{token: token}
}

func (p *BearerTokenProvider) GetToken(ctx context.Context, resource string) (string, error) {
	if p.token == "" {
		return "", fmt.Errorf("bearer token not configured")
	}
	return p.token, nil
}

func (p *BearerTokenProvider) Close() error {
	return nil
}

// EnvVarTokenProvider retrieves token from environment variable.
type EnvVarTokenProvider struct {
	envVar string
}

func NewEnvVarTokenProvider(envVar string) *EnvVarTokenProvider {
	return &EnvVarTokenProvider{envVar: envVar}
}

func (p *EnvVarTokenProvider) GetToken(ctx context.Context, resource string) (string, error) {
	if p.envVar == "" {
		return "", fmt.Errorf("environment variable not configured")
	}
	token := os.Getenv(p.envVar)
	if token == "" {
		return "", fmt.Errorf("environment variable %q not set", p.envVar)
	}
	return token, nil
}

func (p *EnvVarTokenProvider) Close() error {
	return nil
}


// NewTokenProvider creates a TokenProvider based on auth configuration.
// Enforces HTTPS for authorization endpoints per MCP spec.
func NewTokenProvider(authCfg *config.AuthConfig) (TokenProvider, error) {
	if authCfg == nil {
		return &NoAuthProvider{}, nil
	}

	authType := authCfg.Type
	if authType == "" {
		authType = "none"
	}

	switch authType {
	case "none":
		return &NoAuthProvider{}, nil

	case "bearer-token":
		// Try to get token from config first, then from environment variable
		token := authCfg.Token
		if token == "" && authCfg.TokenEnvVar != "" {
			token = os.Getenv(authCfg.TokenEnvVar)
		}
		if token == "" {
			return nil, fmt.Errorf("bearer-token auth type requires 'token' field or '%s' environment variable", authCfg.TokenEnvVar)
		}
		return NewBearerTokenProvider(token), nil

	case "env":
		if authCfg.TokenEnvVar == "" {
			return nil, fmt.Errorf("env auth type requires 'tokenEnvVar' field")
		}
		return NewEnvVarTokenProvider(authCfg.TokenEnvVar), nil

	case "oauth2":
		return nil, fmt.Errorf("oauth2 auth type is not yet supported - use 'bearer-token' or 'env' instead")

	default:
		return nil, fmt.Errorf("unknown auth type: %q", authType)
	}
}

// ParseWWWAuthenticate parses a WWW-Authenticate header and extracts auth parameters (RFC 6750).
// Properly handles quoted values that may contain commas.
// Returns a map of parameter names to values.
func ParseWWWAuthenticate(headerValue string) map[string]string {
	params := make(map[string]string)

	// Remove "Bearer " prefix if present
	headerValue = strings.TrimSpace(headerValue)
	if strings.HasPrefix(headerValue, "Bearer ") {
		headerValue = strings.TrimPrefix(headerValue, "Bearer ")
	}

	// Parse parameters carefully to handle quoted values with commas
	var current strings.Builder
	inQuotes := false
	i := 0

	for i < len(headerValue) {
		ch := headerValue[i]

		// Handle quotes
		if ch == '"' {
			inQuotes = !inQuotes
			current.WriteByte(ch)
			i++
			continue
		}

		// Handle comma (parameter separator) only outside quotes
		if ch == ',' && !inQuotes {
			// Parse the parameter
			param := strings.TrimSpace(current.String())
			if param != "" {
				if idx := strings.Index(param, "="); idx > 0 {
					key := strings.TrimSpace(param[:idx])
					value := strings.TrimSpace(param[idx+1:])
					// Remove quotes if present
					if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
						value = value[1 : len(value)-1]
					}
					params[key] = value
				}
			}
			current.Reset()
			i++
			continue
		}

		current.WriteByte(ch)
		i++
	}

	// Parse the last parameter
	param := strings.TrimSpace(current.String())
	if param != "" {
		if idx := strings.Index(param, "="); idx > 0 {
			key := strings.TrimSpace(param[:idx])
			value := strings.TrimSpace(param[idx+1:])
			// Remove quotes if present
			if len(value) >= 2 && value[0] == '"' && value[len(value)-1] == '"' {
				value = value[1 : len(value)-1]
			}
			params[key] = value
		}
	}

	return params
}
