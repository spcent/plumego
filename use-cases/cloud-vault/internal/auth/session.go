package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
)

const sessionTokenLen = 32

// GenerateSessionToken creates a cryptographically secure random session token.
func GenerateSessionToken() (string, error) {
	b := make([]byte, sessionTokenLen)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("auth: generate session token: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// HashSessionToken generates a SHA-256 hash of a session token for storage.
func HashSessionToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return base64.StdEncoding.EncodeToString(hash[:])
}
