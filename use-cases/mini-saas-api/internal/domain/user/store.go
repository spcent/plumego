package user

import (
	"context"
	"strings"
	"sync"
)

// Repository is the storage contract for user accounts.
type Repository interface {
	Create(ctx context.Context, u User) error
	ByID(ctx context.Context, id string) (User, bool, error)
	ByEmail(ctx context.Context, email string) (User, bool, error)
}

// MemoryStore is a mutex-guarded in-memory Repository.
// Replace with a database-backed Repository for durable deployments.
type MemoryStore struct {
	mu      sync.RWMutex
	byID    map[string]User
	byEmail map[string]string // lowercased email → user ID
}

// NewMemoryStore returns an empty MemoryStore.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		byID:    make(map[string]User),
		byEmail: make(map[string]string),
	}
}

func (s *MemoryStore) Create(_ context.Context, u User) error {
	key := strings.ToLower(u.Email)
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, taken := s.byEmail[key]; taken {
		return ErrEmailTaken
	}
	s.byID[u.ID] = u
	s.byEmail[key] = u.ID
	return nil
}

func (s *MemoryStore) ByID(_ context.Context, id string) (User, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	u, ok := s.byID[id]
	return u, ok, nil
}

func (s *MemoryStore) ByEmail(_ context.Context, email string) (User, bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	id, ok := s.byEmail[strings.ToLower(email)]
	if !ok {
		return User{}, false, nil
	}
	u, ok := s.byID[id]
	return u, ok, nil
}
