package tenant

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrRoutePolicyNotFound = errors.New("route policy not found")

// RoutePolicy describes tenant-specific routing policy.
// Payload is application-defined JSON to keep the framework unopinionated.
type RoutePolicy struct {
	TenantID  string
	Strategy  string
	Payload   []byte
	Metadata  map[string]string
	UpdatedAt time.Time
}

// RoutePolicyProvider loads routing policy for a tenant.
type RoutePolicyProvider interface {
	RoutePolicy(ctx context.Context, tenantID string) (RoutePolicy, error)
}

// RoutePolicyStore persists routing policy for a tenant.
type RoutePolicyStore interface {
	RoutePolicyProvider
	SetRoutePolicy(ctx context.Context, policy RoutePolicy) error
	DeleteRoutePolicy(ctx context.Context, tenantID string) error
}

// InMemoryRoutePolicyStore stores route policies in memory.
type InMemoryRoutePolicyStore struct {
	mu      sync.RWMutex
	entries map[string]RoutePolicy
}

// NewInMemoryRoutePolicyStore creates an in-memory route policy store.
func NewInMemoryRoutePolicyStore() *InMemoryRoutePolicyStore {
	return &InMemoryRoutePolicyStore{
		entries: make(map[string]RoutePolicy),
	}
}

// RoutePolicy returns the current policy for a tenant.
func (s *InMemoryRoutePolicyStore) RoutePolicy(ctx context.Context, tenantID string) (RoutePolicy, error) {
	if s == nil {
		return RoutePolicy{}, ErrRoutePolicyNotFound
	}
	s.mu.RLock()
	defer s.mu.RUnlock()

	policy, ok := s.entries[tenantID]
	if !ok {
		return RoutePolicy{}, ErrRoutePolicyNotFound
	}
	return policy, nil
}

// SetRoutePolicy stores policy for a tenant.
func (s *InMemoryRoutePolicyStore) SetRoutePolicy(ctx context.Context, policy RoutePolicy) error {
	if s == nil {
		return nil
	}
	if policy.UpdatedAt.IsZero() {
		policy.UpdatedAt = time.Now().UTC()
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.entries[policy.TenantID] = policy
	return nil
}

// DeleteRoutePolicy removes policy for a tenant.
func (s *InMemoryRoutePolicyStore) DeleteRoutePolicy(ctx context.Context, tenantID string) error {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.entries, tenantID)
	return nil
}
