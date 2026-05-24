package auth

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/zalando/go-keyring"
)

// StoredToken represents a stored access/refresh token pair.
type StoredToken struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	TokenType    string `json:"token_type"`
	ExpiresIn    int64  `json:"expires_in"`
	IssuedAt     int64  `json:"issued_at"`  // Unix timestamp for expiry calculation
}

// TokenStorage provides a common interface for token persistence.
type TokenStorage interface {
	// Store saves a token for the given server.
	Store(server string, token *StoredToken) error
	// Retrieve gets a token for the given server, returns nil if not found.
	Retrieve(server string) (*StoredToken, error)
	// Delete removes a token for the given server.
	Delete(server string) error
}

// KeyringStorage uses OS keyring for token storage (macOS Keychain, Linux Secret Service, Windows Credential Manager).
type KeyringStorage struct {
	service string
}

// NewKeyringStorage creates a new keyring-based token storage.
func NewKeyringStorage() *KeyringStorage {
	return &KeyringStorage{service: "mcp-setu"}
}

// Store saves a token to the keyring.
func (ks *KeyringStorage) Store(server string, token *StoredToken) error {
	data, err := json.Marshal(token)
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	account := "server:" + server
	err = keyring.Set(ks.service, account, string(data))
	if err != nil {
		return fmt.Errorf("failed to store token in keyring: %w", err)
	}
	return nil
}

// Retrieve gets a token from the keyring.
func (ks *KeyringStorage) Retrieve(server string) (*StoredToken, error) {
	account := "server:" + server
	tokenStr, err := keyring.Get(ks.service, account)
	if err != nil {
		// keyring.ErrNotFound is returned when key doesn't exist
		if err.Error() == "secret not found in keyring" {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to retrieve token from keyring: %w", err)
	}

	var token StoredToken
	if err := json.Unmarshal([]byte(tokenStr), &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}
	return &token, nil
}

// Delete removes a token from the keyring.
func (ks *KeyringStorage) Delete(server string) error {
	account := "server:" + server
	err := keyring.Delete(ks.service, account)
	if err != nil && err.Error() != "secret not found in keyring" {
		return fmt.Errorf("failed to delete token from keyring: %w", err)
	}
	return nil
}

// FileStorage stores tokens in an encrypted-at-rest file (fallback when keyring unavailable).
// Note: This is stored in plain text with mode 0600. Use only as fallback.
type FileStorage struct {
	dir string
}

// NewFileStorage creates a file-based token storage in ~/.config/mcp-setu/credentials.
func NewFileStorage() (*FileStorage, error) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get config directory: %w", err)
	}

	dir := filepath.Join(configDir, "mcp-setu")
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("failed to create config directory: %w", err)
	}

	return &FileStorage{dir: dir}, nil
}

// tokenFile returns the path to the token file for a server.
func (fs *FileStorage) tokenFile(server string) string {
	// Sanitize server name to be filename-safe
	safe := sanitizeFilename(server)
	return filepath.Join(fs.dir, safe+".json")
}

// Store saves a token to a file.
func (fs *FileStorage) Store(server string, token *StoredToken) error {
	data, err := json.MarshalIndent(token, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal token: %w", err)
	}

	path := fs.tokenFile(server)
	if err := os.WriteFile(path, data, 0600); err != nil {
		return fmt.Errorf("failed to write token file: %w", err)
	}
	return nil
}

// Retrieve gets a token from a file.
func (fs *FileStorage) Retrieve(server string) (*StoredToken, error) {
	path := fs.tokenFile(server)
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to read token file: %w", err)
	}

	var token StoredToken
	if err := json.Unmarshal(data, &token); err != nil {
		return nil, fmt.Errorf("failed to unmarshal token: %w", err)
	}
	return &token, nil
}

// Delete removes a token file.
func (fs *FileStorage) Delete(server string) error {
	path := fs.tokenFile(server)
	err := os.Remove(path)
	if err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete token file: %w", err)
	}
	return nil
}

// sanitizeFilename converts a server name to a safe filename.
func sanitizeFilename(name string) string {
	safe := ""
	for _, ch := range name {
		if (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || (ch >= '0' && ch <= '9') || ch == '-' || ch == '_' {
			safe += string(ch)
		} else {
			safe += "_"
		}
	}
	return safe
}

// ChainedStorage tries multiple storage backends in order.
type ChainedStorage struct {
	storages []TokenStorage
}

// NewChainedStorage creates a storage that tries multiple backends.
func NewChainedStorage(storages ...TokenStorage) *ChainedStorage {
	return &ChainedStorage{storages: storages}
}

// Store tries to store in the first storage backend.
func (cs *ChainedStorage) Store(server string, token *StoredToken) error {
	if len(cs.storages) == 0 {
		return fmt.Errorf("no storage backends available")
	}
	return cs.storages[0].Store(server, token)
}

// Retrieve tries each storage backend in order until one has the token.
// Returns the first token found, or nil if no backend has it.
// Errors from individual backends are silently ignored to maintain fallback semantics.
func (cs *ChainedStorage) Retrieve(server string) (*StoredToken, error) {
	for _, storage := range cs.storages {
		token, err := storage.Retrieve(server)
		if err != nil {
			// Silently continue to next backend on error
			continue
		}
		if token != nil {
			return token, nil
		}
	}
	// All backends checked; no token found
	return nil, nil
}

// Delete deletes from all storage backends.
func (cs *ChainedStorage) Delete(server string) error {
	var lastErr error
	for _, storage := range cs.storages {
		if err := storage.Delete(server); err != nil {
			lastErr = err
		}
	}
	return lastErr
}
