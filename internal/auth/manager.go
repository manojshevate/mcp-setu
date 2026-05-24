package auth

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/manojshevate/mcp-setu/internal/config"
)

// Manager handles credential lookup with three-tier fallback: env > keyring > file > interactive.
type Manager struct {
	storage TokenStorage
}

// NewManager creates a new credential manager with keyring and file storage.
func NewManager() (*Manager, error) {
	fileStorage, err := NewFileStorage()
	if err != nil {
		return nil, fmt.Errorf("failed to create file storage: %w", err)
	}

	// Try keyring first, fall back to file
	storage := NewChainedStorage(
		NewKeyringStorage(),
		fileStorage,
	)

	return &Manager{storage: storage}, nil
}

// GetToken retrieves a token for the given server, implementing the three-tier lookup.
func (m *Manager) GetToken(ctx context.Context, serverName string, authCfg *config.AuthConfig) (string, error) {
	if authCfg == nil {
		return "", fmt.Errorf("no auth config for server")
	}

	// Tier 1: Environment variable (highest priority for CI/CD)
	if authCfg.TokenEnvVar != "" {
		token := os.Getenv(authCfg.TokenEnvVar)
		if token != "" {
			return token, nil
		}
	}

	// Also check convention-based env var: MCPSETU_<SERVERNAME>_TOKEN
	convEnvVar := "MCPSETU_" + strings.ToUpper(strings.ReplaceAll(serverName, "-", "_")) + "_TOKEN"
	if token := os.Getenv(convEnvVar); token != "" {
		return token, nil
	}

	// Tier 2: Stored token (keyring or file)
	storedToken, err := m.storage.Retrieve(serverName)
	if err != nil {
		return "", fmt.Errorf("failed to retrieve stored token: %w", err)
	}

	if storedToken != nil {
		// Check if token is still valid
		if !isTokenExpired(storedToken) {
			return storedToken.AccessToken, nil
		}

		// Try to refresh if we have a refresh token
		if storedToken.RefreshToken != "" && authCfg.Type == "oauth2" {
			// TODO: implement refresh logic
			// For now, fall through to interactive flow
		}
	}

	// Tier 3: Interactive OAuth flow (only if interactive)
	if authCfg.Type == "oauth2" {
		if !CanOpenBrowser() {
			return "", fmt.Errorf(
				"no valid token for server %q and cannot perform interactive OAuth flow\n"+
					"Set %s environment variable or use 'mcp-setu auth login %s'",
				serverName, convEnvVar, serverName)
		}

		// Trigger interactive login
		return "", fmt.Errorf(
			"interactive OAuth flow not yet implemented from this context\n"+
				"Use 'mcp-setu auth login %s' to authenticate", serverName)
	}

	// No token available
	return "", fmt.Errorf("no valid token for server %q; set %s environment variable", serverName, convEnvVar)
}

// Login performs interactive login for a server.
func (m *Manager) Login(ctx context.Context, serverName string, authCfg *config.AuthConfig) error {
	if authCfg == nil || authCfg.Type != "oauth2" {
		return fmt.Errorf("server %q does not support OAuth2", serverName)
	}

	if authCfg.AuthorizationServerURL == "" {
		return fmt.Errorf("authorization server URL not configured for server %q", serverName)
	}

	// TODO: discover token endpoint from authorization server metadata
	tokenEndpoint := authCfg.AuthorizationServerURL + "/token"

	flow := NewOAuth2Flow(
		authCfg.AuthorizationServerURL,
		tokenEndpoint,
		authCfg.ClientID,
		authCfg.ClientSecret,
		authCfg.Scopes,
		m.storage,
	)

	// Check if we can open browser
	if CanOpenBrowser() {
		_, err := flow.LoginInteractive(ctx, serverName)
		return err
	}

	// Fall back to device flow
	_, err := flow.LoginDevice(ctx, serverName)
	return err
}

// Logout removes stored credentials for a server.
func (m *Manager) Logout(ctx context.Context, serverName string) error {
	return m.storage.Delete(serverName)
}

// isTokenExpired checks if a token has expired.
func isTokenExpired(token *StoredToken) bool {
	if token.ExpiresIn <= 0 || token.IssuedAt == 0 {
		return false // No expiration info
	}
	expiresAt := time.Unix(token.IssuedAt, 0).Add(time.Duration(token.ExpiresIn) * time.Second)
	return time.Now().After(expiresAt)
}
