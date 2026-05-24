package auth

import (
	"os"
	"path/filepath"
	"testing"
)

func TestFileStorage(t *testing.T) {
	// Create temp dir for test
	tmpDir := t.TempDir()

	storage := &FileStorage{dir: tmpDir}

	token := &StoredToken{
		AccessToken:  "test_access_token",
		RefreshToken: "test_refresh_token",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	}

	// Test Store
	if err := storage.Store("test-server", token); err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// Verify file permissions
	path := storage.tokenFile("test-server")
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Failed to stat token file: %v", err)
	}

	// Check that file exists and has mode 0600
	perm := info.Mode().Perm()
	if perm != 0600 {
		t.Errorf("token file permissions wrong: got %o, expected 0600", perm)
	}

	// Test Retrieve
	retrieved, err := storage.Retrieve("test-server")
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("retrieved token is nil")
	}

	if retrieved.AccessToken != token.AccessToken {
		t.Errorf("access token mismatch: got %q, expected %q", retrieved.AccessToken, token.AccessToken)
	}

	if retrieved.RefreshToken != token.RefreshToken {
		t.Errorf("refresh token mismatch: got %q, expected %q", retrieved.RefreshToken, token.RefreshToken)
	}

	// Test Delete
	if err := storage.Delete("test-server"); err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify deletion
	retrieved, err = storage.Retrieve("test-server")
	if err != nil {
		t.Fatalf("Retrieve after delete failed: %v", err)
	}

	if retrieved != nil {
		t.Error("token should be nil after delete")
	}
}

func TestFileStorageNotFound(t *testing.T) {
	tmpDir := t.TempDir()
	storage := &FileStorage{dir: tmpDir}

	retrieved, err := storage.Retrieve("nonexistent-server")
	if err != nil {
		t.Fatalf("Retrieve for nonexistent server failed: %v", err)
	}

	if retrieved != nil {
		t.Error("expected nil for nonexistent token")
	}
}

func TestFileStorageSanitizeFilename(t *testing.T) {
	tmpDir := t.TempDir()
	storage := &FileStorage{dir: tmpDir}

	// Test with special characters in server name
	serverName := "my-special:server@host"
	path := storage.tokenFile(serverName)

	// Path should be safe (no special chars in filename)
	filename := filepath.Base(path)
	if filename == "my-special:server@host.json" {
		t.Error("filename was not sanitized")
	}

	// Should be able to store and retrieve
	token := &StoredToken{AccessToken: "test", TokenType: "Bearer", ExpiresIn: 3600}
	if err := storage.Store(serverName, token); err != nil {
		t.Fatalf("Store with special chars failed: %v", err)
	}

	retrieved, err := storage.Retrieve(serverName)
	if err != nil {
		t.Fatalf("Retrieve with special chars failed: %v", err)
	}

	if retrieved == nil || retrieved.AccessToken != "test" {
		t.Error("failed to retrieve token with special char server name")
	}
}

func TestChainedStorage(t *testing.T) {
	tmpDir := t.TempDir()

	storage1 := &FileStorage{dir: tmpDir}
	storage2 := &FileStorage{dir: tmpDir}

	chained := NewChainedStorage(storage1, storage2)

	token := &StoredToken{
		AccessToken:  "test_access",
		RefreshToken: "test_refresh",
		TokenType:    "Bearer",
		ExpiresIn:    3600,
	}

	// Store should use first storage
	if err := chained.Store("test-server", token); err != nil {
		t.Fatalf("Store failed: %v", err)
	}

	// Retrieve should find in first storage
	retrieved, err := chained.Retrieve("test-server")
	if err != nil {
		t.Fatalf("Retrieve failed: %v", err)
	}

	if retrieved == nil {
		t.Fatal("retrieved token is nil")
	}

	if retrieved.AccessToken != token.AccessToken {
		t.Errorf("access token mismatch: got %q, expected %q", retrieved.AccessToken, token.AccessToken)
	}
}
