// Package audit stores app-local audit events.
package audit

import (
	"encoding/json"
	"errors"
	"time"

	kvstore "github.com/spcent/plumego/store/kv"
)

const (
	auditKey                = "audit:events"
	DefaultMaxEvents        = 500
	DefaultRetention        = 90 * 24 * time.Hour
	defaultKVExpirationSkew = 24 * time.Hour
)

// Event records an auditable control-plane action.
type Event struct {
	ID           string    `json:"id"`
	User         string    `json:"user"`
	Role         string    `json:"role,omitempty"`
	Action       string    `json:"action"`
	Method       string    `json:"method"`
	Path         string    `json:"path"`
	Status       int       `json:"status"`
	RequestID    string    `json:"request_id,omitempty"`
	RemoteAddr   string    `json:"remote_addr,omitempty"`
	DeniedReason string    `json:"denied_reason,omitempty"`
	CreatedAt    time.Time `json:"created_at"`
}

// Options controls local audit retention.
type Options struct {
	MaxEvents int
	Retention time.Duration
}

// Store persists audit events.
type Store struct {
	kv        *kvstore.KVStore
	maxEvents int
	retention time.Duration
}

// NewStore creates an audit store.
func NewStore(kv *kvstore.KVStore) *Store {
	return NewStoreWithOptions(kv, Options{})
}

// NewStoreWithOptions creates an audit store with explicit retention settings.
func NewStoreWithOptions(kv *kvstore.KVStore, opts Options) *Store {
	maxEvents := opts.MaxEvents
	if maxEvents <= 0 {
		maxEvents = DefaultMaxEvents
	}
	retention := opts.Retention
	if retention <= 0 {
		retention = DefaultRetention
	}
	return &Store{kv: kv, maxEvents: maxEvents, retention: retention}
}

// Add records an event.
func (s *Store) Add(e Event) error {
	events, _ := s.List()
	events = append([]Event{e}, events...)
	events = s.trim(events, time.Now().UTC())
	data, err := json.Marshal(events)
	if err != nil {
		return err
	}
	return s.kv.Set(auditKey, data, s.retention+defaultKVExpirationSkew)
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
	return s.trim(events, time.Now().UTC()), nil
}

func (s *Store) trim(events []Event, now time.Time) []Event {
	if len(events) == 0 {
		return events
	}
	cutoff := now.Add(-s.retention)
	kept := events[:0]
	for _, event := range events {
		if event.CreatedAt.IsZero() || event.CreatedAt.After(cutoff) || event.CreatedAt.Equal(cutoff) {
			kept = append(kept, event)
		}
	}
	if len(kept) > s.maxEvents {
		kept = kept[:s.maxEvents]
	}
	return kept
}
