package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"production-service/internal/domain/tenant"
)

type profileStoreFile struct {
	Profiles []tenant.Profile `json:"profiles"`
}

type profileStore struct {
	mu       sync.RWMutex
	path     string
	profiles map[string]tenant.Profile
}

func newProfileStore(path string) (*profileStore, error) {
	profiles := defaultProfiles()
	if path != "" {
		loaded, err := loadProfileFile(path, profiles)
		if err != nil {
			return nil, err
		}
		profiles = loaded
	}
	return &profileStore{path: path, profiles: profiles}, nil
}

func defaultProfiles() map[string]tenant.Profile {
	profiles := make(map[string]tenant.Profile)
	profiles["tenant-a"] = tenant.Profile{
		TenantID: "tenant-a",
		Name:     "Tenant A",
		Plan:     "production",
		Features: []string{"api", "ops", "tenant_context"},
	}
	profiles["tenant-b"] = tenant.Profile{
		TenantID: "tenant-b",
		Name:     "Tenant B",
		Plan:     "standard",
		Features: []string{"api", "tenant_context"},
	}
	return profiles
}

func loadProfileFile(path string, fallback map[string]tenant.Profile) (map[string]tenant.Profile, error) {
	content, err := os.ReadFile(path)
	if errors.Is(err, os.ErrNotExist) {
		if err := writeProfileFile(path, fallback); err != nil {
			return nil, err
		}
		return fallback, nil
	}
	if err != nil {
		return nil, err
	}

	var file profileStoreFile
	if err := json.Unmarshal(content, &file); err != nil {
		return nil, fmt.Errorf("decode profile store %s: %w", path, err)
	}
	profiles := make(map[string]tenant.Profile, len(file.Profiles))
	for _, profile := range file.Profiles {
		if profile.TenantID == "" {
			return nil, fmt.Errorf("decode profile store %s: tenant_id is required", path)
		}
		profiles[profile.TenantID] = profile
	}
	return profiles, nil
}

func writeProfileFile(path string, profiles map[string]tenant.Profile) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create profile store directory: %w", err)
	}
	file := profileStoreFile{Profiles: make([]tenant.Profile, 0, len(profiles))}
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

func (s *profileStore) Get(tenantID string) (tenant.Profile, bool) {
	if s == nil {
		return tenant.Profile{}, false
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	profile, ok := s.profiles[tenantID]
	return profile, ok
}

func (s *profileStore) Kind() string {
	if s == nil || s.path == "" {
		return "app_local_in_memory_reference"
	}
	return "app_local_json_file_reference"
}

func (s *profileStore) Replacement() string {
	if s == nil || s.path == "" {
		return "set APP_PROFILE_STORE_PATH or replace profileStore behind App.Profiles in internal/app"
	}
	return "replace JSON loader behind App.Profiles with an application-owned repository"
}

func (s *profileStore) Path() string {
	if s == nil {
		return ""
	}
	return s.path
}
