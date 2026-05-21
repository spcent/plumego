package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/router"
)

// Item is the canonical item resource in this reference application.
type Item struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	CreatedAt string `json:"created_at"`
}

// ItemRepository is the minimal persistence contract that ItemHandler depends on.
// Pass a concrete implementation from routes.go; pass a stub in tests.
type ItemRepository interface {
	Create(name string) Item
	Get(id string) (Item, bool)
}

// ItemHandler demonstrates constructor injection: declare the dependency as an
// interface field and wire the concrete implementation in routes.go.
type ItemHandler struct {
	Repo ItemRepository
}

type createItemReq struct {
	Name string `json:"name"`
}

// Create handles POST /api/v1/items.
// It demonstrates the canonical request body decode path and structured validation errors.
//
//	POST /api/v1/items {"name":"widget"}  → 201 {"id":"item-1","name":"widget",...}
//	POST /api/v1/items {}                 → 400 TypeRequired
//	POST /api/v1/items (bad JSON)         → 400 TypeBadRequest
func (h ItemHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createItemReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeBadRequest).
			Message("request body must be valid JSON").
			Build())
		return
	}
	if req.Name == "" {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeRequired).
			Detail("field", "name").
			Message("name is required").
			Build())
		return
	}
	item := h.Repo.Create(req.Name)
	_ = contract.WriteResponse(w, r, http.StatusCreated, item, nil)
}

// GetByID handles GET /api/v1/items/:id.
// It demonstrates path parameter extraction and the canonical 404 error path.
//
//	GET /api/v1/items/item-1  → 200 item
//	GET /api/v1/items/missing → 404 TypeNotFound
func (h ItemHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id := router.Param(r, "id")
	item, ok := h.Repo.Get(id)
	if !ok {
		_ = contract.WriteError(w, r, contract.NewErrorBuilder().
			Type(contract.TypeNotFound).
			Detail("id", id).
			Message("item not found").
			Build())
		return
	}
	_ = contract.WriteResponse(w, r, http.StatusOK, item, nil)
}

// MemoryItemStore is a thread-safe in-memory ItemRepository.
// It satisfies the ItemRepository interface. Replace with a real
// persistence layer when moving to production.
type MemoryItemStore struct {
	mu    sync.RWMutex
	items map[string]Item
	next  int
}

// NewMemoryItemStore returns a ready-to-use in-memory store.
func NewMemoryItemStore() *MemoryItemStore {
	return &MemoryItemStore{items: make(map[string]Item)}
}

func (s *MemoryItemStore) Create(name string) Item {
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

func (s *MemoryItemStore) Get(id string) (Item, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	item, ok := s.items[id]
	return item, ok
}
