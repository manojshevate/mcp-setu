package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// DeviceAuthorizationResponse represents the response from a device authorization endpoint (RFC 8628).
type DeviceAuthorizationResponse struct {
	DeviceCode      string `json:"device_code"`
	UserCode        string `json:"user_code"`
	VerificationURI string `json:"verification_uri"`
	ExpiresIn       int    `json:"expires_in"`
	Interval        int    `json:"interval"`
}

// RequestDeviceAuthorization initiates the device authorization flow.
func RequestDeviceAuthorization(ctx context.Context, tokenEndpoint, clientID string, scopes []string) (*DeviceAuthorizationResponse, error) {
	data := url.Values{}
	data.Set("client_id", clientID)
	if len(scopes) > 0 {
		data.Set("scope", stringSliceToSpace(scopes))
	}

	resp, err := http.PostForm(tokenEndpoint, data)
	if err != nil {
		return nil, fmt.Errorf("device authorization request failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read device authorization response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device authorization failed with status %d: %s", resp.StatusCode, string(body))
	}

	var devResp DeviceAuthorizationResponse
	if err := json.Unmarshal(body, &devResp); err != nil {
		return nil, fmt.Errorf("failed to parse device authorization response: %w", err)
	}

	return &devResp, nil
}

// DevicePoller polls the token endpoint for the user's authorization decision.
type DevicePoller struct {
	tokenEndpoint string
	clientID      string
	deviceCode    string
	interval      time.Duration
	timeout       time.Duration
}

// NewDevicePoller creates a new device flow poller.
func NewDevicePoller(tokenEndpoint, clientID, deviceCode string, intervalSecs, expiresSecs int) *DevicePoller {
	interval := time.Duration(intervalSecs) * time.Second
	if interval < 5*time.Second {
		interval = 5 * time.Second
	}
	timeout := time.Duration(expiresSecs) * time.Second

	return &DevicePoller{
		tokenEndpoint: tokenEndpoint,
		clientID:      clientID,
		deviceCode:    deviceCode,
		interval:      interval,
		timeout:       timeout,
	}
}

// TokenResponse represents a successful token response from the token endpoint.
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	RefreshToken string `json:"refresh_token,omitempty"`
	Error        string `json:"error,omitempty"`
	ErrorDesc    string `json:"error_description,omitempty"`
}

// Poll attempts to retrieve the token, retrying with the specified interval.
func (dp *DevicePoller) Poll(ctx context.Context) (*StoredToken, error) {
	deadline := time.Now().Add(dp.timeout)
	ticker := time.NewTicker(dp.interval)
	defer ticker.Stop()

	for {
		select {
		case <-time.After(time.Until(deadline)):
			return nil, fmt.Errorf("device authorization timeout")
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			token, err := dp.pollOnce()
			if err == nil {
				return token, nil
			}
			// Check if error is retryable
			if !strings.Contains(err.Error(), "authorization_pending") && !strings.Contains(err.Error(), "slow_down") {
				return nil, err
			}
		}
	}
}

// pollOnce attempts one token request.
func (dp *DevicePoller) pollOnce() (*StoredToken, error) {
	data := url.Values{}
	data.Set("grant_type", "urn:ietf:params:oauth:grant-type:device_code")
	data.Set("device_code", dp.deviceCode)
	data.Set("client_id", dp.clientID)

	resp, err := http.PostForm(dp.tokenEndpoint, data)
	if err != nil {
		return nil, fmt.Errorf("device token poll failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read device token response: %w", err)
	}

	var tokenResp TokenResponse
	if err := json.Unmarshal(body, &tokenResp); err != nil {
		return nil, fmt.Errorf("failed to parse device token response: %w", err)
	}

	if tokenResp.Error != "" {
		return nil, fmt.Errorf("device token error: %s - %s", tokenResp.Error, tokenResp.ErrorDesc)
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("device token request failed with status %d", resp.StatusCode)
	}

	return &StoredToken{
		AccessToken:  tokenResp.AccessToken,
		RefreshToken: tokenResp.RefreshToken,
		TokenType:    tokenResp.TokenType,
		ExpiresIn:    tokenResp.ExpiresIn,
	}, nil
}
