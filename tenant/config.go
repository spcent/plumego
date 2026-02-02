// Package tenant provides multi-tenancy infrastructure (EXPERIMENTAL).
//
// ⚠️  EXPERIMENTAL: This package's API may change in minor versions.
//
// The tenant package enables multi-tenancy support with per-tenant
// configuration, quota enforcement, and policy management. However,
// it is not yet feature-complete for production use.
//
// Current limitations:
//   - No HTTP middleware for automatic tenant extraction
//   - No integration with core.App configuration options
//   - Limited test coverage (not production-tested)
//   - No comprehensive examples or documentation
//
// For production multi-tenancy, we recommend:
//   - Implementing custom tenant extraction middleware
//   - Using this package's types as a reference foundation
//   - Thoroughly testing your tenant isolation
//
// Planned for future versions:
//   - Complete middleware integration
//   - Production-ready test suite
//   - Example multi-tenant applications
//   - API stability guarantees
package tenant

import (
	"context"
	"errors"
	"sync"
	"time"
)

var ErrTenantNotFound = errors.New("tenant not found")

// Config defines tenant-level policies and quotas.
type Config struct {
	TenantID  string
	Quota     QuotaConfig
	Policy    PolicyConfig
	Metadata  map[string]string
	UpdatedAt time.Time
}

// ConfigManager loads tenant configuration snapshots.
type ConfigManager interface {
	GetTenantConfig(ctx context.Context, tenantID string) (Config, error)
}

// QuotaConfigProvider loads quota configuration for a tenant.
type QuotaConfigProvider interface {
	QuotaConfig(ctx context.Context, tenantID string) (QuotaConfig, error)
}

// PolicyConfigProvider loads policy configuration for a tenant.
type PolicyConfigProvider interface {
	PolicyConfig(ctx context.Context, tenantID string) (PolicyConfig, error)
}

// InMemoryConfigManager stores tenant configs in memory.
type InMemoryConfigManager struct {
	mu      sync.RWMutex
	configs map[string]Config
}

// NewInMemoryConfigManager creates an in-memory config manager.
func NewInMemoryConfigManager() *InMemoryConfigManager {
	return &InMemoryConfigManager{
		configs: make(map[string]Config),
	}
}

// SetTenantConfig inserts or replaces a tenant configuration.
func (m *InMemoryConfigManager) SetTenantConfig(cfg Config) {
	if m == nil {
		return
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	cfg.UpdatedAt = time.Now().UTC()
	m.configs[cfg.TenantID] = cfg
}

// GetTenantConfig fetches the tenant configuration.
func (m *InMemoryConfigManager) GetTenantConfig(ctx context.Context, tenantID string) (Config, error) {
	if m == nil {
		return Config{}, ErrTenantNotFound
	}

	m.mu.RLock()
	defer m.mu.RUnlock()

	cfg, ok := m.configs[tenantID]
	if !ok {
		return Config{}, ErrTenantNotFound
	}
	return cfg, nil
}

// QuotaConfig returns the quota config for a tenant.
func (m *InMemoryConfigManager) QuotaConfig(ctx context.Context, tenantID string) (QuotaConfig, error) {
	cfg, err := m.GetTenantConfig(ctx, tenantID)
	if err != nil {
		return QuotaConfig{}, err
	}
	return cfg.Quota, nil
}

// PolicyConfig returns the policy config for a tenant.
func (m *InMemoryConfigManager) PolicyConfig(ctx context.Context, tenantID string) (PolicyConfig, error) {
	cfg, err := m.GetTenantConfig(ctx, tenantID)
	if err != nil {
		return PolicyConfig{}, err
	}
	return cfg.Policy, nil
}
