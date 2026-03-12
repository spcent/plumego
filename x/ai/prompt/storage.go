package prompt

import (
	"context"
	"fmt"
	"sort"
	"sync"
)

// MemoryStorage implements in-memory template storage.
type MemoryStorage struct {
	templates map[string]*Template
	mu        sync.RWMutex
}

// NewMemoryStorage creates a new memory storage.
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		templates: make(map[string]*Template),
	}
}

// Save implements Storage.
func (s *MemoryStorage) Save(ctx context.Context, tmpl *Template) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Deep copy to avoid mutation
	copied := *tmpl
	copied.Variables = make([]Variable, len(tmpl.Variables))
	copy(copied.Variables, tmpl.Variables)

	copied.Tags = make([]string, len(tmpl.Tags))
	copy(copied.Tags, tmpl.Tags)

	if tmpl.Metadata != nil {
		copied.Metadata = make(map[string]string)
		for k, v := range tmpl.Metadata {
			copied.Metadata[k] = v
		}
	}

	s.templates[tmpl.ID] = &copied
	return nil
}

// Load implements Storage.
func (s *MemoryStorage) Load(ctx context.Context, templateID string) (*Template, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	tmpl, ok := s.templates[templateID]
	if !ok {
		return nil, fmt.Errorf("template not found: %s", templateID)
	}

	// Return copy
	copied := *tmpl
	return &copied, nil
}

// Delete implements Storage.
func (s *MemoryStorage) Delete(ctx context.Context, templateID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.templates, templateID)
	return nil
}

// ListVersions implements Storage.
func (s *MemoryStorage) ListVersions(ctx context.Context, name string) ([]*Template, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var versions []*Template
	for _, tmpl := range s.templates {
		if tmpl.Name == name {
			copied := *tmpl
			versions = append(versions, &copied)
		}
	}

	// Sort by version (simple string comparison for now)
	sort.Slice(versions, func(i, j int) bool {
		return versions[i].Version < versions[j].Version
	})

	return versions, nil
}

// ListByTag implements Storage.
func (s *MemoryStorage) ListByTag(ctx context.Context, tag string) ([]*Template, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	var result []*Template
	for _, tmpl := range s.templates {
		for _, t := range tmpl.Tags {
			if t == tag {
				copied := *tmpl
				result = append(result, &copied)
				break
			}
		}
	}

	return result, nil
}

// Count returns the number of templates.
func (s *MemoryStorage) Count() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return len(s.templates)
}

// Clear removes all templates.
func (s *MemoryStorage) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.templates = make(map[string]*Template)
}
