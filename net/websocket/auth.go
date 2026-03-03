package websocket

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/spcent/plumego/security/password"
)

// RoomAuthenticator is the interface for WebSocket room authentication.
//
// Implement this interface to provide custom authentication logic, or use
// NewSimpleRoomAuth / NewSecureRoomAuth for the built-in implementations.
//
// Example:
//
//	auth := websocket.NewSimpleRoomAuth(secret)
//	auth.SetRoomPassword("chat", "s3cr3t")
//
//	cfg := websocket.ServerConfig{
//	    Hub:  hub,
//	    Auth: auth, // *simpleRoomAuth satisfies RoomAuthenticator
//	}
type RoomAuthenticator interface {
	CheckRoomPassword(room, provided string) bool
	VerifyJWT(token string) (map[string]any, error)
}

// SimpleRoomAuth stores metadata about rooms (password).
//
// This is a simple room authentication implementation that uses password-based
// authentication for room access. It stores hashed passwords for each room.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	auth := websocket.NewSimpleRoomAuth(secret)
//	auth.SetRoomPassword("chat-room", "my-secret-password")
//
//	// Check if password is correct
//	if auth.CheckRoomPassword("chat-room", "my-secret-password") {
//		// Allow access
//	}
type SimpleRoomAuth struct {
	roomPasswords map[string]string // room -> password (hashed)
	mu            sync.RWMutex
	jwtSecret     []byte // HMAC secret for HS256
}

// NewSimpleRoomAuth creates a new simple room authentication instance.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	secret := []byte("my-jwt-secret")
//	auth := websocket.NewSimpleRoomAuth(secret)
func NewSimpleRoomAuth(secret []byte) *SimpleRoomAuth {
	return &SimpleRoomAuth{
		roomPasswords: make(map[string]string),
		jwtSecret:     secret,
	}
}

// SetRoomPassword sets the password for a room.
// The password is hashed before storing for security.
//
// Example:
//
//	import "github.com/spcent/plumego/net/websocket"
//
//	auth := websocket.NewSimpleRoomAuth(secret)
//	auth.SetRoomPassword("chat-room", "my-secret-password")
func (s *SimpleRoomAuth) SetRoomPassword(room, pwd string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	hashed, err := password.HashPassword(pwd)
	if err != nil {
		log.Printf("Error hashing room password: %v", err)
		return
	}
	s.roomPasswords[room] = string(hashed)
}

func (s *SimpleRoomAuth) CheckRoomPassword(room, provided string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if hashed, ok := s.roomPasswords[room]; ok {
		err := password.CheckPassword(hashed, provided)
		return err == nil
	}
	// no password set => allowed
	return true
}

// VerifyJWT verifies an HS256 token and returns the payload map.
func (s *SimpleRoomAuth) VerifyJWT(token string) (map[string]any, error) {
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
	mac := hmac.New(sha256.New, s.jwtSecret)
	// "header.payload" is the prefix of the original token string up to the
	// second dot. Slicing avoids the extra allocation from parts[0]+"."+parts[1].
	signingLen := len(parts[0]) + 1 + len(parts[1])
	mac.Write([]byte(token[:signingLen]))
	expected := mac.Sum(nil)
	if !hmac.Equal(expected, sig) {
		return nil, ErrInvalidToken
	}
	// Verify exp if present
	if expv, ok := payload["exp"]; ok {
		switch t := expv.(type) {
		case float64:
			if time.Now().Unix() > int64(t) {
				return nil, ErrTokenExpired
			}
		case int64:
			if time.Now().Unix() > t {
				return nil, ErrTokenExpired
			}
		}
	}
	return payload, nil
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
	case []interface{}:
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
