// Package item contains the reference application's item domain model and store.
package item

import (
	"fmt"
	"sync"
	"time"
)

// Item is the canonical item resource in this reference application.
type Item struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
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

// Create stores an item with the provided name and returns the new item.
func (s *MemoryStore) Create(name string) Item {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.next++
	item := Item{
		ID:        fmt.Sprintf("item-%d", s.next),
		Name:      name,
		CreatedAt: time.Now().UTC(),
	}
	s.items[item.ID] = item
	s.ids = append(s.ids, item.ID)
	return item
}

// Get returns an item by id.
func (s *MemoryStore) Get(id string) (Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	item, ok := s.items[id]
	return item, ok
}

// List returns all stored items in creation order.
func (s *MemoryStore) List() []Item {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]Item, 0, len(s.ids))
	for _, id := range s.ids {
		result = append(result, s.items[id])
	}
	return result
}

// Update replaces the name of an existing item and returns the updated item.
// It reports false when no item with that id exists.
func (s *MemoryStore) Update(id, name string) (Item, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[id]
	if !ok {
		return Item{}, false
	}
	item.Name = name
	s.items[id] = item
	return item, true
}

// Delete removes an item by id and reports whether it existed.
func (s *MemoryStore) Delete(id string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[id]; !ok {
		return false
	}
	delete(s.items, id)
	// Remove id from the ordered slice while preserving relative order.
	for i, v := range s.ids {
		if v == id {
			s.ids = append(s.ids[:i], s.ids[i+1:]...)
			break
		}
	}
	return true
}
