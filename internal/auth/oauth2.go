package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"runtime"
	"time"

	"github.com/mattn/go-isatty"
)

// OAuth2Flow orchestrates the OAuth 2.1 authorization flow.
type OAuth2Flow struct {
	tokenEndpoint string
	authEndpoint  string
	clientID      string
	clientSecret  string
	scopes        []string
	storage       TokenStorage
}

// NewOAuth2Flow creates a new OAuth2 flow orchestrator.
func NewOAuth2Flow(authEndpoint, tokenEndpoint, clientID, clientSecret string, scopes []string, storage TokenStorage) *OAuth2Flow {
	return &OAuth2Flow{
		authEndpoint:  authEndpoint,
		tokenEndpoint: tokenEndpoint,
		clientID:      clientID,
		clientSecret:  clientSecret,
		scopes:        scopes,
		storage:       storage,
	}
}

// ExchangeCodeForToken exchanges an authorization code for an access token using PKCE (RFC 7636).
func (f *OAuth2Flow) ExchangeCodeForToken(ctx context.Context, code, pkceVerifier, redirectURI string) (*StoredToken, error) {
	data := url.Values{}
	data.Set("grant_type", "authorization_code")
	data.Set("code", code)
	data.Set("redirect_uri", redirectURI)
	data.Set("client_id", f.clientID)
	data.Set("code_verifier", pkceVerifier)
	// Note: client_secret is not included (public client as per RFC 8252)

	resp, err := http.PostForm(f.tokenEndpoint, data)
	if err != nil {
		return nil, fmt.Errorf("token exchange failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read token response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("token error: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	return &StoredToken{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		ExpiresIn:    tokenResp.ExpiresIn,
	}, nil
}

// RefreshAccessToken refreshes an access token using the refresh token.
func (f *OAuth2Flow) RefreshAccessToken(ctx context.Context, refreshToken string) (*StoredToken, error) {
	data := url.Values{}
	data.Set("grant_type", "refresh_token")
	data.Set("refresh_token", refreshToken)
	data.Set("client_id", f.clientID)

	resp, err := http.PostForm(f.tokenEndpoint, data)
	if err != nil {
		return nil, fmt.Errorf("token refresh failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read refresh response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("token refresh failed with status %d: %s", resp.StatusCode, string(body))
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse refresh response: %w", err)
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("refresh error: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	return &StoredToken{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		ExpiresIn:    tokenResp.ExpiresIn,
	}, nil
}

// LoginInteractive performs an interactive OAuth login using the authorization code flow.
func (f *OAuth2Flow) LoginInteractive(ctx context.Context, serverName string) (*StoredToken, error) {
	// Generate PKCE pair
	pkce, err := GeneratePKCE()
	if err != nil {
		return nil, fmt.Errorf("failed to generate PKCE: %w", err)
	}

	// Create loopback server
	loopback, err := NewLoopbackServer(120 * time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to create loopback server: %w", err)
	}
	defer loopback.Close()

	redirectURI := loopback.GetRedirectURI()

	// Build authorization URL
	state := "state" // TODO: generate random state
	authURL, err := BuildAuthorizationURL(f.authEndpoint, f.clientID, f.scopes, redirectURI, pkce, state)
	if err != nil {
		return nil, fmt.Errorf("failed to build authorization URL: %w", err)
	}

	// Start loopback server
	if err := loopback.Start(ctx); err != nil {
		return nil, fmt.Errorf("failed to start loopback server: %w", err)
	}

	// Open browser
	fmt.Printf("Opening browser for authentication...\nIf browser doesn't open, visit: %s\n", authURL)
	if err := openBrowser(authURL); err != nil {
		fmt.Printf("Could not open browser automatically: %v\nPlease visit the URL above manually.\n", err)
	}

	// Wait for callback
	code, err := loopback.WaitForCode(ctx)
	if err != nil {
		return nil, fmt.Errorf("oauth callback failed: %w", err)
	}

	// Exchange code for token
	token, err := f.ExchangeCodeForToken(ctx, code, pkce.Verifier, redirectURI)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %w", err)
	}

	// Store token
	if err := f.storage.Store(serverName, token); err != nil {
		return nil, fmt.Errorf("failed to store token: %w", err)
	}

	return token, nil
}

// LoginDevice performs a device authorization flow for headless environments.
func (f *OAuth2Flow) LoginDevice(ctx context.Context, serverName string) (*StoredToken, error) {
	// Request device authorization
	devResp, err := RequestDeviceAuthorization(ctx, f.tokenEndpoint, f.clientID, f.scopes)
	if err != nil {
		return nil, fmt.Errorf("device authorization request failed: %w", err)
	}

	// Prompt user
	fmt.Printf("\nDevice Authorization Required\n")
	fmt.Printf("========================\n")
	fmt.Printf("Go to: %s\n", devResp.VerificationURI)
	fmt.Printf("Enter code: %s\n\n", devResp.UserCode)
	fmt.Printf("Waiting for authorization...\n")

	// Poll for token
	poller := NewDevicePoller(f.tokenEndpoint, f.clientID, devResp.DeviceCode, devResp.Interval, devResp.ExpiresIn)
	token, err := poller.Poll(ctx)
	if err != nil {
		return nil, fmt.Errorf("device authorization failed: %w", err)
	}

	// Store token
	if err := f.storage.Store(serverName, token); err != nil {
		return nil, fmt.Errorf("failed to store token: %w", err)
	}

	return token, nil
}

// openBrowser attempts to open the given URL in the default browser.
func openBrowser(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Run()
	case "linux":
		return exec.Command("xdg-open", url).Run()
	case "windows":
		return exec.Command("cmd", "/c", "start", url).Run()
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}
}

// isTTY checks if stdout is a terminal.
func isTTY() bool {
	return isatty.IsTerminal(os.Stdout.Fd()) || isatty.IsCygwinTerminal(os.Stdout.Fd())
}

// CanOpenBrowser checks if we can open a browser (interactive environment).
func CanOpenBrowser() bool {
	// Check if stdout is a TTY
	if !isTTY() {
		return false
	}

	// Check if we're in an SSH session
	if os.Getenv("SSH_TTY") != "" {
		return false
	}

	// Check if display is available (X11, Wayland)
	if os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != "" {
		return true
	}

	// macOS and Windows don't need display
	if runtime.GOOS == "darwin" || runtime.GOOS == "windows" {
		return true
	}

	return false
}
