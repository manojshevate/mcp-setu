package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
)

// PKCEPair holds the verifier and challenge for PKCE (RFC 7636).
type PKCEPair struct {
	Verifier  string
	Challenge string
}

// GeneratePKCE creates a new PKCE verifier and challenge using S256 method (recommended).
// Uses 32 random bytes, which base64url-encodes to 43 characters (RFC 7636 minimum).
func GeneratePKCE() (*PKCEPair, error) {
	// RFC 7636: verifier is 43-128 characters, unreserved characters only.
	// 32 random bytes → 43 base64url chars (256 bits of entropy, canonical size).
	verifierBytes := make([]byte, 32)
	_, err := rand.Read(verifierBytes)
	if err != nil {
		return nil, err
	}

	verifier := base64.RawURLEncoding.EncodeToString(verifierBytes)

	// S256 challenge: BASE64URL(SHA256(verifier))
	hash := sha256.Sum256([]byte(verifier))
	challenge := base64.RawURLEncoding.EncodeToString(hash[:])

	return &PKCEPair{
		Verifier:  verifier,
		Challenge: challenge,
	}, nil
}
