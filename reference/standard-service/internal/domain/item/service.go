package item

import (
	"context"
	"strings"
)

// Service is the business logic contract for items.
// Routes.go wires ItemService as the concrete implementation; handlers declare
// a narrow interface (ItemService in internal/handler) that both this ItemService
// and MemoryStore satisfy, so implementations are interchangeable at the wiring
// site without touching handler code.
//
// Business logic that spans multiple repository operations, applies domain-level
// invariants, coordinates side effects (events, notifications, audit logs), or
// implements caching strategies belongs here — not in the handler or the repository.
type Service interface {
	Create(ctx context.Context, name, description string) (Item, error)
	Get(ctx context.Context, id string) (Item, bool)
	List(ctx context.Context, offset, limit int) ([]Item, int, error)
	Update(ctx context.Context, id, name, description string) (Item, bool, error)
	Patch(ctx context.Context, id, name, description string) (Item, bool, error)
	Delete(ctx context.Context, id string) (bool, error)
}

// ItemService implements Service using a Repository.
// It is the production implementation wired by routes.go.
type ItemService struct {
	repo Repository
}

// NewItemService returns a ready-to-use ItemService backed by repo.
// In production pass a database-backed Repository; pass NewMemoryStore() for
// local development and integration tests.
func NewItemService(repo Repository) *ItemService {
	return &ItemService{repo: repo}
}

// Create normalises name and description by trimming leading/trailing whitespace
// before persisting. Handlers validate that the trimmed fields are non-empty;
// normalisation runs at the domain boundary so stored values are always in
// canonical form regardless of what the transport layer delivers.
func (s *ItemService) Create(ctx context.Context, name, description string) (Item, error) {
	return s.repo.Create(ctx, strings.TrimSpace(name), strings.TrimSpace(description))
}

func (s *ItemService) Get(ctx context.Context, id string) (Item, bool) {
	return s.repo.Get(ctx, id)
}

func (s *ItemService) List(ctx context.Context, offset, limit int) ([]Item, int, error) {
	return s.repo.List(ctx, offset, limit)
}

// Update normalises name and description before persisting, matching Create.
func (s *ItemService) Update(ctx context.Context, id, name, description string) (Item, bool, error) {
	return s.repo.Update(ctx, id, strings.TrimSpace(name), strings.TrimSpace(description))
}

// Patch normalises any non-empty fields before persisting, matching Create.
func (s *ItemService) Patch(ctx context.Context, id, name, description string) (Item, bool, error) {
	return s.repo.Patch(ctx, id, strings.TrimSpace(name), strings.TrimSpace(description))
}

func (s *ItemService) Delete(ctx context.Context, id string) (bool, error) {
	return s.repo.Delete(ctx, id)
}
