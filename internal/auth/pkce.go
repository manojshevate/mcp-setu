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
func GeneratePKCE() (*PKCEPair, error) {
	// RFC 7636: verifier is 43-128 characters, unreserved characters only.
	// We use 128 bytes of random data encoded as base64url, which gives ~171 chars.
	verifierBytes := make([]byte, 96)
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
