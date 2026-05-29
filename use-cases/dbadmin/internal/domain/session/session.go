// Package session manages web UI sessions backed by the stable KV store.
package session

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	kvstore "github.com/spcent/plumego/store/kv"
)

const (
	sessionTTL     = 24 * time.Hour
	tokenKeyPrefix = "session:"
)

// ErrNotFound is returned when a session token is not found or has expired.
var ErrNotFound = errors.New("session: not found or expired")

// Session holds the data stored for an authenticated UI session.
type Session struct {
	Token     string    `json:"token"`
	User      string    `json:"user"`
	CreatedAt time.Time `json:"created_at"`
}

// Store persists sessions in a KV store.
type Store struct {
	kv *kvstore.KVStore
}

// NewStore creates a Store backed by the provided KV store.
func NewStore(kv *kvstore.KVStore) *Store {
	return &Store{kv: kv}
}

// Create generates a new session for user, persists it, and returns the token.
func (s *Store) Create(user string) (string, error) {
	token, err := generateToken()
	if err != nil {
		return "", fmt.Errorf("generate session token: %w", err)
	}
	sess := Session{
		Token:     token,
		User:      user,
		CreatedAt: time.Now().UTC(),
	}
	data, err := json.Marshal(sess)
	if err != nil {
		return "", fmt.Errorf("marshal session: %w", err)
	}
	if err := s.kv.Set(tokenKeyPrefix+token, data, sessionTTL); err != nil {
		return "", fmt.Errorf("persist session: %w", err)
	}
	return token, nil
}

// Get retrieves a session by token. Returns ErrNotFound when missing or expired.
func (s *Store) Get(token string) (*Session, error) {
	data, err := s.kv.Get(tokenKeyPrefix + token)
	if err != nil {
		if errors.Is(err, kvstore.ErrKeyNotFound) || errors.Is(err, kvstore.ErrKeyExpired) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("get session: %w", err)
	}
	var sess Session
	if err := json.Unmarshal(data, &sess); err != nil {
		return nil, fmt.Errorf("unmarshal session: %w", err)
	}
	return &sess, nil
}

// Delete removes a session by token.
func (s *Store) Delete(token string) error {
	err := s.kv.Delete(tokenKeyPrefix + token)
	if err != nil && !errors.Is(err, kvstore.ErrKeyNotFound) {
		return fmt.Errorf("delete session: %w", err)
	}
	return nil
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
