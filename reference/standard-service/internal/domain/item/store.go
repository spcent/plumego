package item

import (
	"context"
	"fmt"
	"math/rand/v2"
	"sync"
	"time"
)

// Repository is the persistence contract for Item storage.
// All methods accept a context so callers can propagate request deadlines and
// cancellation to real storage backends. The error return on mutating methods
// distinguishes storage failures (non-nil error) from "not found" (bool false,
// nil error) so callers can return 500 vs 404. Real backends propagate ctx to
// the underlying driver; the in-memory implementation checks ctx.Err() before
// acquiring its lock so the contract holds for both.
type Repository interface {
	Create(ctx context.Context, name, description string) (Item, error)
	Get(ctx context.Context, id string) (Item, bool)
	List(ctx context.Context, offset, limit int) ([]Item, int, error)
	Update(ctx context.Context, id, name, description string) (Item, bool, error)
	Patch(ctx context.Context, id, name, description string) (Item, bool, error)
	Delete(ctx context.Context, id string) (bool, error)
}

// generateID returns a random 16-character hex string suitable for use as an
// opaque resource identifier. Using random IDs rather than sequential integers
// avoids leaking the store's internal state through the API.
func generateID() string {
	return fmt.Sprintf("%016x", rand.Uint64())
}

// MemoryStore is a thread-safe in-memory item repository.
// items is the lookup table; ids tracks insertion order so List is stable.
type MemoryStore struct {
	mu    sync.RWMutex
	items map[string]Item
	ids   []string
}

// NewMemoryStore returns a ready-to-use in-memory store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{items: make(map[string]Item)}
}

// Create stores an item with the provided name and description and returns the new item.
// Returns (Item{}, ctx.Err()) immediately when ctx is already cancelled so handlers
// can detect deadline expiry without acquiring the store lock. Real storage backends
// propagate ctx to the database call and return any storage error; the context error
// path here mirrors that contract so callers are never surprised when the interface
// implementation changes.
func (s *MemoryStore) Create(ctx context.Context, name, description string) (Item, error) {
	if err := ctx.Err(); err != nil {
		return Item{}, err
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
	return item, nil
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

// List returns a paginated slice of items in creation order, the total item count,
// and any error. offset is the zero-based index of the first item to return; limit
// caps the number of items returned. An offset beyond the total returns an empty
// slice; total is always the full count regardless of offset/limit. This signature
// lets real storage backends issue a single LIMIT/OFFSET query rather than loading
// every row into memory.
// Returns (nil, 0, ctx.Err()) immediately when ctx is already cancelled.
func (s *MemoryStore) List(ctx context.Context, offset, limit int) ([]Item, int, error) {
	if err := ctx.Err(); err != nil {
		return nil, 0, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	all := make([]Item, 0, len(s.ids))
	for _, id := range s.ids {
		all = append(all, s.items[id])
	}
	total := len(all)

	start := offset
	if start > len(all) {
		start = len(all)
	}
	page := all[start:]
	if len(page) > limit {
		page = page[:limit]
	}
	return page, total, nil
}

// Update replaces the name and description of an existing item and returns the updated item.
// CreatedAt and ID are immutable; all other fields are replaced by this operation.
// Returns (Item{}, false, ctx.Err()) immediately when ctx is already cancelled.
// Returns (Item{}, false, nil) when no item with that id exists; the boolean false
// distinguishes "not found" from a storage error so callers can return 404 vs 500.
func (s *MemoryStore) Update(ctx context.Context, id, name, description string) (Item, bool, error) {
	if err := ctx.Err(); err != nil {
		return Item{}, false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[id]
	if !ok {
		return Item{}, false, nil
	}
	item.Name = name
	item.Description = description
	s.items[item.ID] = item
	return item, true, nil
}

// Patch applies a partial update to an existing item and returns the result.
// Only non-empty fields are replaced; empty string values leave the corresponding
// field unchanged. ID and CreatedAt are always immutable.
// Returns (Item{}, false, ctx.Err()) immediately when ctx is already cancelled.
// Returns (Item{}, false, nil) when no item with that id exists.
func (s *MemoryStore) Patch(ctx context.Context, id, name, description string) (Item, bool, error) {
	if err := ctx.Err(); err != nil {
		return Item{}, false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	item, ok := s.items[id]
	if !ok {
		return Item{}, false, nil
	}
	if name != "" {
		item.Name = name
	}
	if description != "" {
		item.Description = description
	}
	s.items[item.ID] = item
	return item, true, nil
}

// Delete removes an item by id and reports whether it existed.
// Returns (false, ctx.Err()) immediately when ctx is already cancelled.
// Returns (false, nil) when no item with that id exists; the boolean false
// distinguishes "not found" from a storage error so callers can return 404 vs 500.
func (s *MemoryStore) Delete(ctx context.Context, id string) (bool, error) {
	if err := ctx.Err(); err != nil {
		return false, err
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	if _, ok := s.items[id]; !ok {
		return false, nil
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
	return true, nil
}
