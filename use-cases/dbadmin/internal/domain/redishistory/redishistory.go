// Package redishistory stores Redis command history per connection.
package redishistory

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	kvstore "github.com/spcent/plumego/store/kv"
)

const maxPerConn = 100

// Entry is a single recorded Redis command execution.
type Entry struct {
	ID        string    `json:"id"`
	ConnID    string    `json:"conn_id"`
	DBIndex   int       `json:"db_index"`
	Command   string    `json:"command"`
	Duration  int64     `json:"duration_ms"`
	Success   bool      `json:"success"`
	Error     string    `json:"error,omitempty"`
	CreatedAt time.Time `json:"created_at"`
}

// Store persists Redis command history in a KV store.
type Store struct {
	kv *kvstore.KVStore
}

// NewStore creates a Redis history Store.
func NewStore(kv *kvstore.KVStore) *Store { return &Store{kv: kv} }

// Add records a new history entry for a connection.
func (s *Store) Add(e *Entry) error {
	entries, _ := s.List(e.ConnID)
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

// Delete removes a single history entry by ID for the given connection.
func (s *Store) Delete(connID, entryID string) error {
	entries, err := s.List(connID)
	if err != nil {
		return err
	}
	filtered := entries[:0]
	for _, e := range entries {
		if e.ID != entryID {
			filtered = append(filtered, e)
		}
	}
	return s.save(connID, filtered)
}

// Clear removes all history entries for the given connection.
func (s *Store) Clear(connID string) error {
	return s.kv.Delete(key(connID))
}

func (s *Store) save(connID string, entries []Entry) error {
	data, err := json.Marshal(entries)
	if err != nil {
		return err
	}
	return s.kv.Set(key(connID), data, 30*24*time.Hour)
}

func key(connID string) string { return fmt.Sprintf("redishistory:%s", connID) }
