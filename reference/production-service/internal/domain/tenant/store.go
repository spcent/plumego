package tenant

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
)

type storeFile struct {
	Profiles []Profile `json:"profiles"`
}

// Store is a local JSON-backed repository of tenant profiles.
// It is intentionally lightweight; replace it with an application-owned
// repository backed by a real database in production use.
type Store struct {
	mu       sync.RWMutex
	path     string
	profiles map[string]Profile
}

// NewStore loads profiles from path. When path is empty the store starts with
// the built-in development defaults and holds them in memory only.
func NewStore(path string) (*Store, error) {
	profiles := defaultProfiles()
	if path != "" {
		loaded, err := loadStoreFile(path, profiles)
		if err != nil {
			return nil, err
		}
		profiles = loaded
	}
	return &Store{path: path, profiles: profiles}, nil
}

func defaultProfiles() map[string]Profile {
	profiles := make(map[string]Profile)
	profiles["tenant-a"] = Profile{
		TenantID: "tenant-a",
		Name:     "Tenant A",
		Plan:     "production",
		Features: []string{"api", "ops", "tenant_context"},
	}
	profiles["tenant-b"] = Profile{
		TenantID: "tenant-b",
		Name:     "Tenant B",
		Plan:     "standard",
		Features: []string{"api", "tenant_context"},
	}
	return profiles
}

func loadStoreFile(path string, fallback map[string]Profile) (map[string]Profile, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if err := writeStoreFile(path, fallback); err != nil {
			return nil, err
		}
		return fallback, nil
	}
	if err != nil {
		return nil, err
	}

	var file storeFile
	if err := json.Unmarshal(content, &file); err != nil {
		return nil, fmt.Errorf("decode profile store %s: %w", path, err)
	}
	profiles := make(map[string]Profile, len(file.Profiles))
	for _, profile := range file.Profiles {
		if profile.TenantID == "" {
			return nil, fmt.Errorf("decode profile store %s: tenant_id is required", path)
		}
		profiles[profile.TenantID] = profile
	}
	return profiles, nil
}

func writeStoreFile(path string, profiles map[string]Profile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create profile store directory: %w", err)
	}
	file := storeFile{Profiles: make([]Profile, 0, len(profiles))}
	for _, profile := range profiles {
		file.Profiles = append(file.Profiles, profile)
	}
	content, err := json.MarshalIndent(file, "", "  ")
	if err != nil {
		return fmt.Errorf("encode profile store: %w", err)
	}
	if err := os.WriteFile(path, append(content, '\n'), 0o600); err != nil {
		return fmt.Errorf("write profile store %s: %w", path, err)
	}
	return nil
}

// Get returns the profile for tenantID and whether it was found.
func (s *Store) Get(tenantID string) (Profile, bool) {
	if s == nil {
		return Profile{}, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	profile, ok := s.profiles[tenantID]
	return profile, ok
}

// Kind returns a short label describing the storage backend.
func (s *Store) Kind() string {
	if s == nil || s.path == "" {
		return "app_local_in_memory_reference"
	}
	return "app_local_json_file_reference"
}

// Replacement describes how to swap this store for a production-grade one.
func (s *Store) Replacement() string {
	if s == nil || s.path == "" {
		return "set APP_PROFILE_STORE_PATH or replace Store behind App.Profiles in internal/app"
	}
	return "replace JSON loader behind App.Profiles with an application-owned repository"
}

// Path returns the file path backing this store, or empty string for in-memory.
func (s *Store) Path() string {
	if s == nil {
		return ""
	}
	return s.path
}
