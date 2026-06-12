package tenantspace

import (
	"context"
	"sort"
	"strings"
	"sync"
)

// Repository is the storage contract for tenants and memberships.
// Membership reads and writes are tenant-scoped: callers always pass the
// tenant ID explicitly so cross-tenant access is impossible by construction.
type Repository interface {
	CreateTenant(ctx context.Context, t Tenant) error
	TenantByID(ctx context.Context, id string) (Tenant, bool, error)
	TenantBySlug(ctx context.Context, slug string) (Tenant, bool, error)
	UpdateTenant(ctx context.Context, t Tenant) error

	AddMembership(ctx context.Context, m Membership) error
	Memberships(ctx context.Context, tenantID string) ([]Membership, error)
	MembershipByID(ctx context.Context, tenantID, id string) (Membership, bool, error)
	MembershipByUser(ctx context.Context, tenantID, userID string) (Membership, bool, error)
	MembershipsForUser(ctx context.Context, userID string) ([]Membership, error)
	UpdateMembership(ctx context.Context, m Membership) error
	RemoveMembership(ctx context.Context, tenantID, id string) error
}

// MemoryStore is a mutex-guarded in-memory Repository.
type MemoryStore struct {
	mu          sync.RWMutex
	tenants     map[string]Tenant
	slugs       map[string]string                // lowercased slug → tenant ID
	memberships map[string]map[string]Membership // tenant ID → membership ID → Membership
}

// NewMemoryStore returns an empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		tenants:     make(map[string]Tenant),
		slugs:       make(map[string]string),
		memberships: make(map[string]map[string]Membership),
	}
}

func (s *MemoryStore) CreateTenant(_ context.Context, t Tenant) error {
	key := strings.ToLower(t.Slug)
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, taken := s.slugs[key]; taken {
		return ErrSlugTaken
	}
	s.tenants[t.ID] = t
	s.slugs[key] = t.ID
	s.memberships[t.ID] = make(map[string]Membership)
	return nil
}

func (s *MemoryStore) TenantByID(_ context.Context, id string) (Tenant, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	t, ok := s.tenants[id]
	return t, ok, nil
}

func (s *MemoryStore) TenantBySlug(_ context.Context, slug string) (Tenant, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.slugs[strings.ToLower(slug)]
	if !ok {
		return Tenant{}, false, nil
	}
	t, ok := s.tenants[id]
	return t, ok, nil
}

func (s *MemoryStore) UpdateTenant(_ context.Context, t Tenant) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tenants[t.ID]; !ok {
		return ErrNotFound
	}
	s.tenants[t.ID] = t
	return nil
}

func (s *MemoryStore) AddMembership(_ context.Context, m Membership) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	members, ok := s.memberships[m.TenantID]
	if !ok {
		return ErrNotFound
	}
	for _, existing := range members {
		if existing.UserID == m.UserID {
			return ErrAlreadyMember
		}
	}
	members[m.ID] = m
	return nil
}

func (s *MemoryStore) Memberships(_ context.Context, tenantID string) ([]Membership, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	members, ok := s.memberships[tenantID]
	if !ok {
		return nil, ErrNotFound
	}
	out := make([]Membership, 0, len(members))
	for _, m := range members {
		out = append(out, m)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (s *MemoryStore) MembershipByID(_ context.Context, tenantID, id string) (Membership, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	members, ok := s.memberships[tenantID]
	if !ok {
		return Membership{}, false, nil
	}
	m, ok := members[id]
	return m, ok, nil
}

func (s *MemoryStore) MembershipByUser(_ context.Context, tenantID, userID string) (Membership, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	members, ok := s.memberships[tenantID]
	if !ok {
		return Membership{}, false, nil
	}
	for _, m := range members {
		if m.UserID == userID {
			return m, true, nil
		}
	}
	return Membership{}, false, nil
}

func (s *MemoryStore) MembershipsForUser(_ context.Context, userID string) ([]Membership, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []Membership
	for _, members := range s.memberships {
		for _, m := range members {
			if m.UserID == userID {
				out = append(out, m)
			}
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (s *MemoryStore) UpdateMembership(_ context.Context, m Membership) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	members, ok := s.memberships[m.TenantID]
	if !ok {
		return ErrNotFound
	}
	if _, ok := members[m.ID]; !ok {
		return ErrNotFound
	}
	members[m.ID] = m
	return nil
}

func (s *MemoryStore) RemoveMembership(_ context.Context, tenantID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	members, ok := s.memberships[tenantID]
	if !ok {
		return ErrNotFound
	}
	if _, ok := members[id]; !ok {
		return ErrNotFound
	}
	delete(members, id)
	return nil
}
