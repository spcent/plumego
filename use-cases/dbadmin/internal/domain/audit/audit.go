// Package audit stores app-local audit events.
package audit

import (
	"encoding/json"
	"errors"
	"time"

	kvstore "github.com/spcent/plumego/store/kv"
)

const (
	auditKey = "audit:events"
	maxAudit = 500
)

// Event records an auditable control-plane action.
type Event struct {
	ID        string    `json:"id"`
	User      string    `json:"user"`
	Action    string    `json:"action"`
	Method    string    `json:"method"`
	Path      string    `json:"path"`
	Status    int       `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// Store persists audit events.
type Store struct {
	kv *kvstore.KVStore
}

// NewStore creates an audit store.
func NewStore(kv *kvstore.KVStore) *Store { return &Store{kv: kv} }

// Add records an event.
func (s *Store) Add(e Event) error {
	events, _ := s.List()
	events = append([]Event{e}, events...)
	if len(events) > maxAudit {
		events = events[:maxAudit]
	}
	data, err := json.Marshal(events)
	if err != nil {
		return err
	}
	return s.kv.Set(auditKey, data, 90*24*time.Hour)
}

// List returns recent events, newest first.
func (s *Store) List() ([]Event, error) {
	data, err := s.kv.Get(auditKey)
	if err != nil {
		if errors.Is(err, kvstore.ErrKeyNotFound) || errors.Is(err, kvstore.ErrKeyExpired) {
			return nil, nil
		}
		return nil, err
	}
	var events []Event
	if err := json.Unmarshal(data, &events); err != nil {
		return nil, err
	}
	return events, nil
}
