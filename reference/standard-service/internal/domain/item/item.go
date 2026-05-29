// Package item contains the reference application's item domain model and store.
package item

import (
	"context"
	"fmt"
	"math/rand/v2"
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
}

// generateID returns a random 16-character hex string suitable for use as an
// opaque resource identifier. Using random IDs rather than sequential integers
// avoids leaking the store's internal state through the API.
func generateID() string {
	return fmt.Sprintf("%016x", rand.Uint64())
}

// NewMemoryStore returns a ready-to-use in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: make(map[string]Item)}
}

// Create stores an item with the provided name and description and returns the new item.
// Returns Item{} immediately when ctx is already cancelled so handlers can detect
// deadline expiry without acquiring the store lock. In a real backend Create would
// return (Item, error) and propagate ctx to the database call; the in-memory
// signature omits the error return for simplicity, so callers distinguish
// cancellation by checking ctx.Err() after the call when needed.
func (s *MemoryStore) Create(ctx context.Context, name, description string) Item {
	if ctx.Err() != nil {
		return Item{}
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	item := Item{
		ID:          generateID(),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now().UTC(),
	}
	s.items[item.ID] = item
	s.ids = append(s.ids, item.ID)
	return item
}

// Get returns an item by id.
// Returns (Item{}, false) immediately when ctx is already cancelled so handlers
// can short-circuit cleanly without acquiring the store lock.
func (s *MemoryStore) Get(ctx context.Context, id string) (Item, bool) {
	if ctx.Err() != nil {
		return Item{}, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[id]
	return item, ok
}

// List returns all stored items in creation order.
// Returns nil immediately when ctx is already cancelled so callers can detect
// deadline expiry without acquiring the store lock.
func (s *MemoryStore) List(ctx context.Context) []Item {
	if ctx.Err() != nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Item, 0, len(s.ids))
	for _, id := range s.ids {
		result = append(result, s.items[id])
	}
	return result
}

// Update replaces the name and description of an existing item and returns the updated item.
// CreatedAt and ID are immutable; all other fields are replaced by this operation.
// Returns (Item{}, false) immediately when ctx is already cancelled.
// It reports false when no item with that id exists.
func (s *MemoryStore) Update(ctx context.Context, id, name, description string) (Item, bool) {
	if ctx.Err() != nil {
		return Item{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[id]
	if !ok {
		return Item{}, false
	}
	item.Name = name
	item.Description = description
	s.items[item.ID] = item
	return item, true
}

// Patch applies a partial update to an existing item and returns the result.
// Only non-empty fields are replaced; empty string values leave the corresponding
// field unchanged. ID and CreatedAt are always immutable.
// It reports false when no item with that id exists or ctx is already cancelled.
func (s *MemoryStore) Patch(ctx context.Context, id, name, description string) (Item, bool) {
	if ctx.Err() != nil {
		return Item{}, false
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[id]
	if !ok {
		return Item{}, false
	}
	if name != "" {
		item.Name = name
	}
	if description != "" {
		item.Description = description
	}
	s.items[item.ID] = item
	return item, true
}

// Delete removes an item by id and reports whether it existed.
// Returns false immediately when ctx is already cancelled.
func (s *MemoryStore) Delete(ctx context.Context, id string) bool {
	if ctx.Err() != nil {
		return false
	}
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
