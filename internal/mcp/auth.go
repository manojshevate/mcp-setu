package mcp

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/manojshevate/mcpgo/internal/config"
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

// OAuth2Provider handles OAuth 2.1 authorization flow (simplified for CLI usage).
// This is a basic implementation that supports PKCE and token caching.
// Production use may require more sophisticated token management.
type OAuth2Provider struct {
	authServerURL string
	clientID      string
	clientSecret  string
	scopes        []string
	tokenCache    map[string]*TokenCache
	mu             sync.RWMutex
}

// TokenCache holds a cached token with its expiration time.
type TokenCache struct {
	AccessToken  string
	RefreshToken string
	ExpiresAt    time.Time
}

// NewOAuth2Provider creates a new OAuth2 provider (simplified).
// Note: Full OAuth 2.1 implementation would require browser-based authorization flow.
// This version is suitable for environments where tokens are pre-obtained or provided via other means.
func NewOAuth2Provider(authServerURL, clientID, clientSecret string, scopes []string) *OAuth2Provider {
	return &OAuth2Provider{
		authServerURL: authServerURL,
		clientID:      clientID,
		clientSecret:  clientSecret,
		scopes:        scopes,
		tokenCache:    make(map[string]*TokenCache),
	}
}

// GetToken returns a valid token, using cached token if available and not expired.
func (p *OAuth2Provider) GetToken(ctx context.Context, resource string) (string, error) {
	p.mu.RLock()
	cached, exists := p.tokenCache[resource]
	p.mu.RUnlock()

	// Return cached token if it exists and hasn't expired
	if exists && cached != nil && time.Now().Before(cached.ExpiresAt) {
		return cached.AccessToken, nil
	}

	// For CLI usage without interactive auth, we would need:
	// 1. Pre-obtained tokens (via env var or config)
	// 2. Client credentials grant (if configured)
	// 3. Refresh token (if available)
	// 4. Instructions for user to obtain token manually

	// This is a placeholder that signals OAuth flow would be needed
	return "", fmt.Errorf("OAuth 2.1 authorization required for %q - token not available in cache or environment", resource)
}

// CacheToken stores a token for future use.
func (p *OAuth2Provider) CacheToken(resource string, token, refreshToken string, expiresIn time.Duration) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.tokenCache[resource] = &TokenCache{
		AccessToken:  token,
		RefreshToken: refreshToken,
		ExpiresAt:    time.Now().Add(expiresIn),
	}
}

// Close cleans up the provider.
func (p *OAuth2Provider) Close() error {
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
		if authCfg.AuthorizationServerURL == "" && authCfg.AuthorizationServerEnvVar == "" {
			return nil, fmt.Errorf("oauth2 auth type requires 'authorizationServerUrl' field or 'authorizationServerEnvVar'")
		}
		authServerURL := authCfg.AuthorizationServerURL
		if authServerURL == "" {
			authServerURL = os.Getenv(authCfg.AuthorizationServerEnvVar)
		}
		if authServerURL == "" {
			return nil, fmt.Errorf("OAuth 2.1 authorization server URL not configured")
		}

		// Enforce HTTPS for authorization endpoints (MCP spec: MUST use HTTPS)
		if !strings.HasPrefix(authServerURL, "https://") && !strings.HasPrefix(authServerURL, "http://localhost") {
			return nil, fmt.Errorf("OAuth 2.1 authorization server URL MUST use HTTPS (or http://localhost for development): %s", authServerURL)
		}

		return NewOAuth2Provider(authServerURL, authCfg.ClientID, authCfg.ClientSecret, authCfg.Scopes), nil

	default:
		return nil, fmt.Errorf("unknown auth type: %q", authType)
	}
}

// PKCEChallenge generates PKCE parameters for OAuth 2.1 code flow.
// Returns (codeChallenge, codeVerifier, error)
func PKCEChallenge() (string, string, error) {
	// Generate random code verifier (43-128 characters, unreserved characters only)
	verifierBytes := make([]byte, 32)
	_, err := rand.Read(verifierBytes)
	if err != nil {
		return "", "", fmt.Errorf("failed to generate PKCE verifier: %w", err)
	}

	// Use URL-safe base64 encoding without padding
	verifier := base64.RawURLEncoding.EncodeToString(verifierBytes)

	// Create code challenge using S256 method (required by OAuth 2.1)
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return challenge, verifier, nil
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
