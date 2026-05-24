package auth

import (
	"crypto/sha256"
	"encoding/base64"
	"testing"
)

func TestGeneratePKCE(t *testing.T) {
	pkce, err := GeneratePKCE()
	if err != nil {
		t.Fatalf("GeneratePKCE failed: %v", err)
	}

	if pkce.Verifier == "" {
		t.Error("verifier is empty")
	}

	if pkce.Challenge == "" {
		t.Error("challenge is empty")
	}

	// Verify challenge is correct SHA256 hash of verifier
	hash := sha256.Sum256([]byte(pkce.Verifier))
	expectedChallenge := base64.RawURLEncoding.EncodeToString(hash[:])

	if pkce.Challenge != expectedChallenge {
		t.Errorf("challenge mismatch: got %q, expected %q", pkce.Challenge, expectedChallenge)
	}

	// Verify verifier length is appropriate (43-128 chars)
	if len(pkce.Verifier) < 43 || len(pkce.Verifier) > 128 {
		t.Errorf("verifier length out of range: %d", len(pkce.Verifier))
	}
}

func TestGeneratePKCEUniqueness(t *testing.T) {
	pkce1, err := GeneratePKCE()
	if err != nil {
		t.Fatalf("First GeneratePKCE failed: %v", err)
	}

	pkce2, err := GeneratePKCE()
	if err != nil {
		t.Fatalf("Second GeneratePKCE failed: %v", err)
	}

	if pkce1.Verifier == pkce2.Verifier {
		t.Error("generated verifiers are not unique")
	}

	if pkce1.Challenge == pkce2.Challenge {
		t.Error("generated challenges are not unique")
	}
}
