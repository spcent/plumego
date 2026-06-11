package project

import (
	"context"
	"sort"
	"sync"
)

// Repository is the tenant-scoped storage contract for projects.
// Every method takes the tenant ID explicitly: cross-tenant access is
// impossible by construction, not by filtering after the fact.
type Repository interface {
	List(ctx context.Context, tenantID string) ([]Project, error)
	ByID(ctx context.Context, tenantID, id string) (Project, bool, error)
	Count(ctx context.Context, tenantID string) (int, error)
	Create(ctx context.Context, p Project) error
	Update(ctx context.Context, p Project) error
	Delete(ctx context.Context, tenantID, id string) error
}

// MemoryStore is a mutex-guarded in-memory Repository.
type MemoryStore struct {
	mu       sync.RWMutex
	byTenant map[string]map[string]Project // tenant ID → project ID → Project
}

// NewMemoryStore returns an empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{byTenant: make(map[string]map[string]Project)}
}

func (s *MemoryStore) List(_ context.Context, tenantID string) ([]Project, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	projects := s.byTenant[tenantID]
	out := make([]Project, 0, len(projects))
	for _, p := range projects {
		out = append(out, p)
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CreatedAt.Before(out[j].CreatedAt) })
	return out, nil
}

func (s *MemoryStore) ByID(_ context.Context, tenantID, id string) (Project, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.byTenant[tenantID][id]
	return p, ok, nil
}

func (s *MemoryStore) Count(_ context.Context, tenantID string) (int, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.byTenant[tenantID]), nil
}

func (s *MemoryStore) Create(_ context.Context, p Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	projects, ok := s.byTenant[p.TenantID]
	if !ok {
		projects = make(map[string]Project)
		s.byTenant[p.TenantID] = projects
	}
	projects[p.ID] = p
	return nil
}

func (s *MemoryStore) Update(_ context.Context, p Project) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	projects := s.byTenant[p.TenantID]
	if _, ok := projects[p.ID]; !ok {
		return ErrNotFound
	}
	projects[p.ID] = p
	return nil
}

func (s *MemoryStore) Delete(_ context.Context, tenantID, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	projects := s.byTenant[tenantID]
	if _, ok := projects[id]; !ok {
		return ErrNotFound
	}
	delete(projects, id)
	return nil
}
