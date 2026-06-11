// Package ident generates random identifiers for domain entities.
package ident

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
)

// New returns a random 128-bit identifier in lowercase hex.
func New() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate id: %w", err)
	}
	return hex.EncodeToString(b[:]), nil
}
