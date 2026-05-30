package session

import (
	"testing"
)

// --- generateToken ---

func TestGenerateToken_length(t *testing.T) {
	token, err := generateToken()
	if err != nil {
		t.Fatalf("generateToken error=%v", err)
	}
	// 32 bytes = 64 hex chars
	if len(token) != 64 {
		t.Errorf("token length=%d, want 64 (32 bytes hex)", len(token))
	}
}

func TestGenerateToken_uniqueness(t *testing.T) {
	tokens := make(map[string]bool)
	for i := 0; i < 100; i++ {
		token, err := generateToken()
		if err != nil {
			t.Fatalf("generateToken error on iteration %d: %v", i, err)
		}
		if tokens[token] {
			t.Errorf("duplicate token generated on iteration %d", i)
		}
		tokens[token] = true
	}
}

