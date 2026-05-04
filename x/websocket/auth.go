package websocket

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"
	"strings"
	"sync"
	"time"

	"github.com/spcent/plumego/security/password"
)

const (
	minJWTSecretLength = 32
	maxJWTNumericDate  = float64(1<<63 - 1)
)

// RoomAuthorizer authorizes access to a websocket room.
//
// Implement this interface to provide custom room policy, or use
// NewSimpleRoomAuth / NewSecureRoomAuth for the built-in implementations.
type RoomAuthorizer interface {
	AuthorizeRoom(room, provided string) bool
}

// TokenAuthenticator authenticates a bearer token and returns token claims.
type TokenAuthenticator interface {
	AuthenticateToken(token string) (map[string]any, error)
}

// SimpleRoomAuth stores metadata about rooms (password).
//
// This is a simple room authentication implementation that uses password-based
// authentication for room access. It stores hashed passwords for each room.
//
// Example:
//
//	import "github.com/spcent/plumego/x/websocket"
//
//	auth := websocket.NewSimpleRoomAuth()
//	if err := auth.SetRoomPassword("chat-room", "my-secret-password"); err != nil {
//		panic(err)
//	}
//
//	if auth.AuthorizeRoom("chat-room", "my-secret-password") {
//		// Allow access
//	}
type SimpleRoomAuth struct {
	roomPasswords map[string]string // room -> password (hashed)
	mu            sync.RWMutex
}

// NewSimpleRoomAuth creates a new simple room authentication instance.
//
// Example:
//
//	import "github.com/spcent/plumego/x/websocket"
//
//	auth := websocket.NewSimpleRoomAuth()
func NewSimpleRoomAuth() *SimpleRoomAuth {
	return &SimpleRoomAuth{
		roomPasswords: make(map[string]string),
	}
}

// SetRoomPassword sets the password for a room.
// The password is hashed before storing for security.
//
// Example:
//
//	import "github.com/spcent/plumego/x/websocket"
//
//	auth := websocket.NewSimpleRoomAuth()
//	if err := auth.SetRoomPassword("chat-room", "my-secret-password"); err != nil {
//		panic(err)
//	}
func (s *SimpleRoomAuth) SetRoomPassword(room, pwd string) error {
	hashed, err := password.HashPassword(pwd)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	s.roomPasswords[room] = string(hashed)
	return nil
}

func (s *SimpleRoomAuth) AuthorizeRoom(room, provided string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if hashed, ok := s.roomPasswords[room]; ok {
		err := password.CheckPassword(hashed, provided)
		return err == nil
	}
	// no password set => allowed
	return true
}

// HS256TokenAuth authenticates compact JWT-like HS256 bearer tokens.
//
// This helper verifies the token signature and optional exp claim. It is not a
// full OIDC policy engine and does not validate issuer, audience, nbf, or iat.
type HS256TokenAuth struct {
	secret []byte
}

// NewHS256TokenAuth creates a token authenticator for HS256 bearer tokens.
func NewHS256TokenAuth(secret []byte) (*HS256TokenAuth, error) {
	if err := validateJWTSecret(secret, minJWTSecretLength); err != nil {
		return nil, err
	}
	cloned := append([]byte(nil), secret...)
	return &HS256TokenAuth{secret: cloned}, nil
}

// AuthenticateToken verifies an HS256 token and returns the payload map.
func (s *HS256TokenAuth) AuthenticateToken(token string) (map[string]any, error) {
	if s == nil || len(s.secret) == 0 {
		return nil, ErrInvalidToken
	}
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, ErrInvalidToken
	}
	hdrb, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var hdr map[string]any
	if err = json.Unmarshal(hdrb, &hdr); err != nil {
		return nil, ErrInvalidToken
	}
	if alg, ok := hdr["alg"].(string); !ok || alg != "HS256" {
		return nil, ErrInvalidToken
	}
	payloadb, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, ErrInvalidToken
	}
	var payload map[string]any
	if err = json.Unmarshal(payloadb, &payload); err != nil {
		return nil, ErrInvalidToken
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, ErrInvalidToken
	}
	mac := hmac.New(sha256.New, s.secret)
	// "header.payload" is the prefix of the original token string up to the
	// second dot. Slicing avoids the extra allocation from parts[0]+"."+parts[1].
	signingLen := len(parts[0]) + 1 + len(parts[1])
	mac.Write([]byte(token[:signingLen]))
	expected := mac.Sum(nil)
	if !hmac.Equal(expected, sig) {
		return nil, ErrInvalidToken
	}
	// Verify exp if present. This helper intentionally does not process nbf,
	// iat, issuer, or audience; applications needing full JWT/OIDC policy
	// should inject their own TokenAuthenticator.
	if expv, ok := payload["exp"]; ok {
		exp, err := jwtNumericDate(expv)
		if err != nil {
			return nil, err
		}
		if time.Now().Unix() > exp {
			return nil, ErrTokenExpired
		}
	}
	return payload, nil
}

func validateJWTSecret(secret []byte, minLen int) error {
	if minLen <= 0 {
		minLen = minJWTSecretLength
	}
	if len(secret) < minLen {
		return fmt.Errorf("%w: got %d bytes, minimum %d bytes required",
			ErrWeakJWTSecret, len(secret), minLen)
	}
	return nil
}

func jwtNumericDate(v any) (int64, error) {
	switch t := v.(type) {
	case float64:
		if math.IsNaN(t) || math.IsInf(t, 0) || t < 0 || t > maxJWTNumericDate || t != math.Trunc(t) {
			return 0, ErrInvalidToken
		}
		return int64(t), nil
	case json.Number:
		i, err := t.Int64()
		if err != nil || i < 0 {
			return 0, ErrInvalidToken
		}
		return i, nil
	case int64:
		if t < 0 {
			return 0, ErrInvalidToken
		}
		return t, nil
	case int:
		if t < 0 {
			return 0, ErrInvalidToken
		}
		return int64(t), nil
	default:
		return 0, ErrInvalidToken
	}
}

// ExtractUserInfo extracts UserInfo from JWT payload.
//
// The "roles" claim is parsed correctly regardless of whether the JWT was
// decoded into map[string]any (where JSON arrays become []interface{}) or
// directly into a typed struct.
func ExtractUserInfo(payload map[string]any) *UserInfo {
	userInfo := &UserInfo{
		Claims: payload,
	}
	// Extract common claims if present
	if id, ok := payload["sub"].(string); ok {
		userInfo.ID = id
	}
	if name, ok := payload["name"].(string); ok {
		userInfo.Name = name
	}
	if email, ok := payload["email"].(string); ok {
		userInfo.Email = email
	}
	// json.Unmarshal into map[string]any always produces []interface{} for JSON
	// arrays, never []string. Handle both representations for forward compat.
	switch v := payload["roles"].(type) {
	case []any:
		for _, r := range v {
			if s, ok := r.(string); ok {
				userInfo.Roles = append(userInfo.Roles, s)
			}
		}
	case []string:
		userInfo.Roles = v
	}
	return userInfo
}
