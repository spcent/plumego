// Package security adapts guardus's basic-auth config to plumego's
// auth/authn primitives.
//
// Password verification accepts both upstream-gatus bcrypt (base64-encoded
// $2y$…) hashes and plumego/security/password PBKDF2-HMAC-SHA512 hashes,
// dispatching on the prefix of the decoded blob.
package security

import (
	"crypto/subtle"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/spcent/plumego/security/password"
	"golang.org/x/crypto/bcrypt"
)

// ErrInvalidPassword is returned when verification fails.
var ErrInvalidPassword = errors.New("invalid password")

// VerifyPassword checks plaintext against a stored hash.
//
// stored may be either:
//   - a base64-encoded bcrypt hash (upstream-gatus convention), or
//   - a plumego PBKDF2 hash of the form "<cost>$<salt-b64>$<hash-b64>".
//
// Comparison is constant-time inside both branches.
func VerifyPassword(stored, plain string) error {
	if stored == "" {
		return ErrInvalidPassword
	}
	if decoded, err := base64.StdEncoding.DecodeString(stored); err == nil && bcryptLooking(decoded) {
		if err := bcrypt.CompareHashAndPassword(decoded, []byte(plain)); err != nil {
			return ErrInvalidPassword
		}
		return nil
	}
	// Fall back to plumego PBKDF2 format.
	if err := password.CheckPassword(stored, plain); err != nil {
		return ErrInvalidPassword
	}
	return nil
}

// EqualConstantTime reports whether a == b in constant time over the
// shorter of the two lengths. Provided for handlers that need a generic
// timing-safe equality check beyond the password path.
func EqualConstantTime(a, b string) bool {
	return subtle.ConstantTimeCompare([]byte(a), []byte(b)) == 1
}

func bcryptLooking(b []byte) bool {
	if len(b) < 4 {
		return false
	}
	prefix := string(b[:4])
	return strings.HasPrefix(prefix, "$2a$") || strings.HasPrefix(prefix, "$2b$") || strings.HasPrefix(prefix, "$2y$")
}
