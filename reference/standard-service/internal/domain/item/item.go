// Package item contains the reference application's item domain model and store.
package item

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Item is the canonical item resource in this reference application.
type Item struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// MemoryStore is a thread-safe in-memory item repository.
// items is the lookup table; ids tracks insertion order so List is stable.
type MemoryStore struct {
	mu    sync.RWMutex
	items map[string]Item
	ids   []string
	next  int
}

// NewMemoryStore returns a ready-to-use in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: make(map[string]Item)}
}

// Create stores an item with the provided name and description and returns the new item.
// context.Context is accepted so callers can propagate deadlines to real storage backends;
// the in-memory implementation does not block and therefore does not inspect it.
func (s *MemoryStore) Create(_ context.Context, name, description string) Item {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.next++
	item := Item{
		ID:          fmt.Sprintf("item-%d", s.next),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now().UTC(),
	}
	s.items[item.ID] = item
	s.ids = append(s.ids, item.ID)
	return item
}

// Get returns an item by id.
// context.Context is accepted so callers can propagate deadlines to real storage backends.
func (s *MemoryStore) Get(_ context.Context, id string) (Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[id]
	return item, ok
}

// List returns all stored items in creation order.
// context.Context is accepted so callers can propagate deadlines to real storage backends.
func (s *MemoryStore) List(_ context.Context) []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Item, 0, len(s.ids))
	for _, id := range s.ids {
		result = append(result, s.items[id])
	}
	return result
}

// Update replaces the name of an existing item and returns the updated item.
// Description is immutable after creation; only name is replaced by this operation.
// It reports false when no item with that id exists.
// context.Context is accepted so callers can propagate deadlines to real storage backends.
func (s *MemoryStore) Update(_ context.Context, id, name string) (Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[id]
	if !ok {
		return Item{}, false
	}
	item.Name = name
	s.items[item.ID] = item
	return item, true
}

// Delete removes an item by id and reports whether it existed.
// context.Context is accepted so callers can propagate deadlines to real storage backends.
func (s *MemoryStore) Delete(_ context.Context, id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[id]; !ok {
		return false
	}
	delete(s.items, id)
	// Linear scan to remove the id while preserving insertion order; acceptable at
	// the small scale of a reference implementation.
	for i, v := range s.ids {
		if v == id {
			s.ids = append(s.ids[:i], s.ids[i+1:]...)
			break
		}
	}
	return true
}
