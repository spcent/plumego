// Package history stores SQL query history per connection.
package history

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
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

// SearchOptions filters the results returned by Search.
type SearchOptions struct {
	Query    string // substring match against Entry.SQL (case-insensitive)
	HasError *bool  // nil = don't filter; true = only entries with Error != ""; false = only entries with Error == ""
	Database string // exact match against Entry.Database, empty = don't filter
	Limit    int    // max entries to return, 0 = use existing maxPerConn default
}

// Search returns history entries for a connection, most recent first,
// filtered according to opts. It builds on List, so it inherits the
// existing per-connection cap of maxPerConn entries.
func (s *Store) Search(connID string, opts SearchOptions) ([]Entry, error) {
	entries, err := s.List(connID)
	if err != nil {
		return nil, err
	}
	query := strings.ToLower(strings.TrimSpace(opts.Query))
	limit := opts.Limit
	if limit <= 0 {
		limit = maxPerConn
	}
	filtered := make([]Entry, 0, len(entries))
	for _, e := range entries {
		if query != "" && !strings.Contains(strings.ToLower(e.SQL), query) {
			continue
		}
		if opts.HasError != nil {
			hasErr := e.Error != ""
			if hasErr != *opts.HasError {
				continue
			}
		}
		if opts.Database != "" && e.Database != opts.Database {
			continue
		}
		filtered = append(filtered, e)
		if len(filtered) >= limit {
			break
		}
	}
	return filtered, nil
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
	return s.kv.Set(key(connID), data, 30*24*time.Hour) // keep 30 days
}

func key(connID string) string { return fmt.Sprintf("history:%s", connID) }
