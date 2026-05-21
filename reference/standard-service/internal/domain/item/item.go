// Package item contains the reference application's item domain model and store.
package item

import (
	"fmt"
	"sync"
	"time"
)

// Item is the canonical item resource in this reference application.
type Item struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// MemoryStore is a thread-safe in-memory item repository.
type MemoryStore struct {
	mu    sync.RWMutex
	items map[string]Item
	next  int
}

// NewMemoryStore returns a ready-to-use in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: make(map[string]Item)}
}

// Create stores an item with the provided name.
func (s *MemoryStore) Create(name string) Item {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.next++
	item := Item{
		ID:        fmt.Sprintf("item-%d", s.next),
		Name:      name,
		CreatedAt: time.Now().UTC().Format(time.RFC3339),
	}
	s.items[item.ID] = item
	return item
}

// Get returns an item by id.
func (s *MemoryStore) Get(id string) (Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[id]
	return item, ok
}
