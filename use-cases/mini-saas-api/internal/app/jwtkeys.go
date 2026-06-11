package app

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/spcent/plumego/security/jwt"
)

// seedJWTKey derives a deterministic HS256 signing key from the configured
// secret and installs it as the active key. Deriving from the secret means
// access tokens stay verifiable across restarts and across instances sharing
// the same APP_JWT_SECRET; changing the secret rotates the key (old tokens
// stop verifying, by design). The key ID embeds a digest prefix so a secret
// change yields a new key ID and previously persisted keys remain inert.
func seedJWTKey(ctx context.Context, store jwt.KeyStore, secret string) error {
	sum := sha256.Sum256([]byte(secret))
	keyID := "seed-" + hex.EncodeToString(sum[:4])
	key := jwt.JWTSigningKey{
		ID:        keyID,
		Algorithm: jwt.AlgorithmHS256,
		Secret:    sum[:], // exactly 32 bytes, as HS256 keys require
		CreatedAt: time.Now().UTC(),
	}
	payload, err := json.Marshal(key)
	if err != nil {
		return fmt.Errorf("marshal signing key: %w", err)
	}
	if err := store.SetContext(ctx, "jwt:keys:"+keyID, payload, 0); err != nil {
		return fmt.Errorf("store signing key: %w", err)
	}
	if err := store.SetContext(ctx, "jwt:active", []byte(keyID), 0); err != nil {
		return fmt.Errorf("activate signing key: %w", err)
	}
	return nil
}
