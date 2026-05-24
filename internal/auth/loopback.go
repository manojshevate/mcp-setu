package auth

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

// LoopbackServer handles the OAuth callback for the authorization code grant (RFC 8252).
type LoopbackServer struct {
	listener net.Listener
	server   *http.Server
	code     chan string
	err      chan error
	timeout  time.Duration
}

// NewLoopbackServer creates a new loopback HTTP server listening on 127.0.0.1 with an ephemeral port.
func NewLoopbackServer(timeout time.Duration) (*LoopbackServer, error) {
	if timeout == 0 {
		timeout = 120 * time.Second
	}

	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, fmt.Errorf("failed to create loopback listener: %w", err)
	}

	ls := &LoopbackServer{
		listener: listener,
		code:     make(chan string, 1),
		err:      make(chan error, 1),
		timeout:  timeout,
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/callback", ls.callbackHandler)

	ls.server = &http.Server{
		Handler:     mux,
		IdleTimeout: timeout,
	}

	return ls, nil
}

// GetRedirectURI returns the redirect URI for this loopback server.
func (ls *LoopbackServer) GetRedirectURI() string {
	addr := ls.listener.Addr().String()
	return "http://" + addr + "/callback"
}

// Start begins listening for OAuth callbacks.
func (ls *LoopbackServer) Start(ctx context.Context) error {
	go func() {
		err := ls.server.Serve(ls.listener)
		if err != nil && err != http.ErrServerClosed {
			select {
			case ls.err <- err:
			default:
			}
		}
	}()

	// Set up context timeout
	if ls.timeout > 0 {
		go func() {
			time.Sleep(ls.timeout)
			select {
			case ls.err <- fmt.Errorf("oauth callback timeout"):
			default:
			}
		}()
	}

	return nil
}

// WaitForCode blocks until the authorization code is received or timeout occurs.
func (ls *LoopbackServer) WaitForCode(ctx context.Context) (string, error) {
	select {
	case code := <-ls.code:
		return code, nil
	case err := <-ls.err:
		return "", err
	case <-ctx.Done():
		return "", ctx.Err()
	}
}

// Close shuts down the loopback server.
func (ls *LoopbackServer) Close() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	return ls.server.Shutdown(ctx)
}

// callbackHandler handles the OAuth callback with the authorization code.
func (ls *LoopbackServer) callbackHandler(w http.ResponseWriter, r *http.Request) {
	// Verify it's a GET request
	if r.Method != http.MethodGet {
		http.Error(w, "only GET allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse the query parameters
	params := r.URL.Query()

	// Check for error response
	if errCode := params.Get("error"); errCode != "" {
		errDesc := params.Get("error_description")
		msg := fmt.Sprintf("OAuth error: %s", errCode)
		if errDesc != "" {
			msg += " - " + errDesc
		}
		select {
		case ls.err <- fmt.Errorf("%s", msg):
		default:
		}
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// Extract authorization code
	code := params.Get("code")
	if code == "" {
		msg := "missing authorization code in callback"
		select {
		case ls.err <- fmt.Errorf("%s", msg):
		default:
		}
		http.Error(w, msg, http.StatusBadRequest)
		return
	}

	// Send success response
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	fmt.Fprint(w, "Authorization successful! You can close this window and return to your terminal.")

	// Send code to channel
	select {
	case ls.code <- code:
	default:
	}
}

// BuildAuthorizationURL constructs the OAuth authorization endpoint URL with PKCE.
func BuildAuthorizationURL(authServerURL, clientID string, scopes []string, redirectURI string, pkce *PKCEPair, state string) (string, error) {
	u, err := url.Parse(authServerURL)
	if err != nil {
		return "", fmt.Errorf("invalid auth server URL: %w", err)
	}

	q := u.Query()
	q.Set("response_type", "code")
	q.Set("client_id", clientID)
	q.Set("redirect_uri", redirectURI)
	q.Set("scope", "openid profile email")
	if len(scopes) > 0 {
		q.Set("scope", stringSliceToSpace(scopes))
	}
	q.Set("code_challenge", pkce.Challenge)
	q.Set("code_challenge_method", "S256")
	q.Set("state", state)

	u.RawQuery = q.Encode()
	return u.String(), nil
}

// stringSliceToSpace joins a slice of strings with spaces.
func stringSliceToSpace(strs []string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += " "
		}
		result += s
	}
	return result
}
