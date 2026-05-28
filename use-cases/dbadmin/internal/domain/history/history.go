// Package history stores SQL query history per connection.
package history

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	kvstore "github.com/spcent/plumego/store/kv"
)

const maxPerConn = 100

// Entry is a single recorded query execution.
type Entry struct {
	ID        string    `json:"id"`
	ConnID    string    `json:"conn_id"`
	Database  string    `json:"database"`
	SQL       string    `json:"sql"`
	Duration  int64     `json:"duration_ms"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Store persists query history in a KV store.
type Store struct {
	kv *kvstore.KVStore
}

// NewStore creates a history Store.
func NewStore(kv *kvstore.KVStore) *Store { return &Store{kv: kv} }

// Add records a new history entry for a connection.
func (s *Store) Add(e *Entry) error {
	entries, _ := s.List(e.ConnID)
	// Prepend and keep at most maxPerConn.
	entries = append([]Entry{*e}, entries...)
	if len(entries) > maxPerConn {
		entries = entries[:maxPerConn]
	}
	return s.save(e.ConnID, entries)
}

// List returns history entries for a connection, most recent first.
func (s *Store) List(connID string) ([]Entry, error) {
	data, err := s.kv.Get(key(connID))
	if err != nil {
		if errors.Is(err, kvstore.ErrKeyNotFound) || errors.Is(err, kvstore.ErrKeyExpired) {
			return nil, nil
		}
		return nil, err
	}
	var entries []Entry
	if err := json.Unmarshal(data, &entries); err != nil {
		return nil, err
	}
	return entries, nil
}

func (s *Store) save(connID string, entries []Entry) error {
	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	return s.kv.Set(key(connID), data, 30*24*time.Hour) // keep 30 days
}

func key(connID string) string { return fmt.Sprintf("history:%s", connID) }
