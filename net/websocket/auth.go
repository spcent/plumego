package websocket

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/spcent/plumego/security/password"
)

// simpleRoomAuth stores metadata about rooms (password)
type simpleRoomAuth struct {
	roomPasswords map[string]string // room -> password (hashed)
	mu            sync.RWMutex
	jwtSecret     []byte // HMAC secret for HS256
}

func NewSimpleRoomAuth(secret []byte) *simpleRoomAuth {
	return &simpleRoomAuth{
		roomPasswords: make(map[string]string),
		jwtSecret:     secret,
	}
}

func (s *simpleRoomAuth) SetRoomPassword(room, pwd string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Hash password before storing
	hashed, err := password.HashPassword(pwd)
	if err != nil {
		log.Printf("Error hashing room password: %v", err)
		return
	}
	s.roomPasswords[room] = string(hashed)
}

func (s *simpleRoomAuth) CheckRoomPassword(room, provided string) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if hashed, ok := s.roomPasswords[room]; ok {
		// Verify hash
		err := password.CheckPassword(hashed, provided)
		return err == nil
	}
	// no password set => allowed
	return true
}

// verifyJWTHS256 verifies HS256 token and returns payload map
func (s *simpleRoomAuth) VerifyJWT(token string) (map[string]any, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return nil, errors.New("invalid jwt")
	}
	hdrb, err := base64.RawURLEncoding.DecodeString(parts[0])
	if err != nil {
		return nil, err
	}
	var hdr map[string]any
	if err = json.Unmarshal(hdrb, &hdr); err != nil {
		return nil, err
	}
	if alg, ok := hdr["alg"].(string); !ok || alg != "HS256" {
		return nil, errors.New("unsupported jwt alg")
	}
	payloadb, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, err
	}
	var payload map[string]any
	if err = json.Unmarshal(payloadb, &payload); err != nil {
		return nil, err
	}
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha256.New, s.jwtSecret)
	mac.Write([]byte(parts[0] + "." + parts[1]))
	expected := mac.Sum(nil)
	if !hmac.Equal(expected, sig) {
		return nil, errors.New("invalid jwt signature")
	}
	// Verify exp if present
	if expv, ok := payload["exp"]; ok {
		switch t := expv.(type) {
		case float64:
			if time.Now().Unix() > int64(t) {
				return nil, errors.New("jwt expired")
			}
		case int64:
			if time.Now().Unix() > t {
				return nil, errors.New("jwt expired")
			}
		}
	}
	return payload, nil
}

// ExtractUserInfo extracts UserInfo from JWT payload
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
	if roles, ok := payload["roles"].([]string); ok {
		userInfo.Roles = roles
	}
	return userInfo
}
