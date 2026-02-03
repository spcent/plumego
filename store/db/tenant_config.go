package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"sync"
	"time"

	"github.com/spcent/plumego/tenant"
)

// DBTenantConfigManager loads tenant configurations from a database.
// It implements tenant.ConfigManager with optional LRU caching.
type DBTenantConfigManager struct {
	db    *sql.DB
	cache *tenantCache
}

// DBTenantConfigOption configures DBTenantConfigManager.
type DBTenantConfigOption func(*DBTenantConfigManager)

// WithTenantCache enables LRU caching with specified size and TTL.
func WithTenantCache(size int, ttl time.Duration) DBTenantConfigOption {
	return func(m *DBTenantConfigManager) {
		m.cache = newTenantCache(size, ttl)
	}
}

// NewDBTenantConfigManager creates a database-backed tenant config manager.
//
// Example:
//
//	db, _ := sql.Open("postgres", "...")
//	manager := NewDBTenantConfigManager(db,
//	    WithTenantCache(1000, 5*time.Minute),
//	)
func NewDBTenantConfigManager(db *sql.DB, options ...DBTenantConfigOption) *DBTenantConfigManager {
	m := &DBTenantConfigManager{
		db: db,
	}

	for _, opt := range options {
		opt(m)
	}

	return m
}

// GetTenantConfig retrieves tenant configuration from database (with optional caching).
func (m *DBTenantConfigManager) GetTenantConfig(ctx context.Context, tenantID string) (tenant.Config, error) {
	if m == nil || m.db == nil {
		return tenant.Config{}, tenant.ErrTenantNotFound
	}

	// Check cache first
	if m.cache != nil {
		if cfg, ok := m.cache.Get(tenantID); ok {
			return cfg, nil
		}
	}

	// Query database
	query := `
		SELECT id, quota_requests_per_minute, quota_tokens_per_minute,
		       allowed_models, allowed_tools, metadata, updated_at
		FROM tenants
		WHERE id = $1
	`

	var cfg tenant.Config
	var allowedModelsJSON, allowedToolsJSON, metadataJSON sql.NullString
	var updatedAt time.Time

	err := m.db.QueryRowContext(ctx, query, tenantID).Scan(
		&cfg.TenantID,
		&cfg.Quota.RequestsPerMinute,
		&cfg.Quota.TokensPerMinute,
		&allowedModelsJSON,
		&allowedToolsJSON,
		&metadataJSON,
		&updatedAt,
	)

	if err == sql.ErrNoRows {
		return tenant.Config{}, tenant.ErrTenantNotFound
	}
	if err != nil {
		return tenant.Config{}, err
	}

	// Parse JSON fields
	if allowedModelsJSON.Valid {
		if err := json.Unmarshal([]byte(allowedModelsJSON.String), &cfg.Policy.AllowedModels); err != nil {
			return tenant.Config{}, err
		}
	}
	if allowedToolsJSON.Valid {
		if err := json.Unmarshal([]byte(allowedToolsJSON.String), &cfg.Policy.AllowedTools); err != nil {
			return tenant.Config{}, err
		}
	}
	if metadataJSON.Valid {
		if err := json.Unmarshal([]byte(metadataJSON.String), &cfg.Metadata); err != nil {
			return tenant.Config{}, err
		}
	}

	cfg.UpdatedAt = updatedAt

	// Cache the result
	if m.cache != nil {
		m.cache.Set(tenantID, cfg)
	}

	return cfg, nil
}

// SetTenantConfig inserts or updates tenant configuration in database.
func (m *DBTenantConfigManager) SetTenantConfig(ctx context.Context, cfg tenant.Config) error {
	if m == nil || m.db == nil {
		return errors.New("database connection not available")
	}

	// Marshal JSON fields
	allowedModels, err := json.Marshal(cfg.Policy.AllowedModels)
	if err != nil {
		return err
	}
	allowedTools, err := json.Marshal(cfg.Policy.AllowedTools)
	if err != nil {
		return err
	}
	metadata, err := json.Marshal(cfg.Metadata)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO tenants (
			id, quota_requests_per_minute, quota_tokens_per_minute,
			allowed_models, allowed_tools, metadata, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (id)
		DO UPDATE SET
			quota_requests_per_minute = EXCLUDED.quota_requests_per_minute,
			quota_tokens_per_minute = EXCLUDED.quota_tokens_per_minute,
			allowed_models = EXCLUDED.allowed_models,
			allowed_tools = EXCLUDED.allowed_tools,
			metadata = EXCLUDED.metadata,
			updated_at = EXCLUDED.updated_at
	`

	_, err = m.db.ExecContext(ctx, query,
		cfg.TenantID,
		cfg.Quota.RequestsPerMinute,
		cfg.Quota.TokensPerMinute,
		string(allowedModels),
		string(allowedTools),
		string(metadata),
		time.Now().UTC(),
	)

	if err != nil {
		return err
	}

	// Invalidate cache
	if m.cache != nil {
		m.cache.Delete(cfg.TenantID)
	}

	return nil
}

// DeleteTenantConfig removes a tenant configuration from database.
func (m *DBTenantConfigManager) DeleteTenantConfig(ctx context.Context, tenantID string) error {
	if m == nil || m.db == nil {
		return errors.New("database connection not available")
	}

	result, err := m.db.ExecContext(ctx, "DELETE FROM tenants WHERE id = $1", tenantID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return tenant.ErrTenantNotFound
	}

	// Invalidate cache
	if m.cache != nil {
		m.cache.Delete(tenantID)
	}

	return nil
}

// ListTenants returns a paginated list of tenant configurations.
func (m *DBTenantConfigManager) ListTenants(ctx context.Context, limit, offset int) ([]tenant.Config, error) {
	if m == nil || m.db == nil {
		return nil, errors.New("database connection not available")
	}

	if limit <= 0 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	query := `
		SELECT id, quota_requests_per_minute, quota_tokens_per_minute,
		       allowed_models, allowed_tools, metadata, updated_at
		FROM tenants
		ORDER BY id
		LIMIT $1 OFFSET $2
	`

	rows, err := m.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var configs []tenant.Config
	for rows.Next() {
		var cfg tenant.Config
		var allowedModelsJSON, allowedToolsJSON, metadataJSON sql.NullString
		var updatedAt time.Time

		err := rows.Scan(
			&cfg.TenantID,
			&cfg.Quota.RequestsPerMinute,
			&cfg.Quota.TokensPerMinute,
			&allowedModelsJSON,
			&allowedToolsJSON,
			&metadataJSON,
			&updatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Parse JSON fields
		if allowedModelsJSON.Valid {
			json.Unmarshal([]byte(allowedModelsJSON.String), &cfg.Policy.AllowedModels)
		}
		if allowedToolsJSON.Valid {
			json.Unmarshal([]byte(allowedToolsJSON.String), &cfg.Policy.AllowedTools)
		}
		if metadataJSON.Valid {
			json.Unmarshal([]byte(metadataJSON.String), &cfg.Metadata)
		}

		cfg.UpdatedAt = updatedAt
		configs = append(configs, cfg)
	}

	return configs, rows.Err()
}

// QuotaConfig implements tenant.QuotaConfigProvider.
func (m *DBTenantConfigManager) QuotaConfig(ctx context.Context, tenantID string) (tenant.QuotaConfig, error) {
	cfg, err := m.GetTenantConfig(ctx, tenantID)
	if err != nil {
		return tenant.QuotaConfig{}, err
	}
	return cfg.Quota, nil
}

// PolicyConfig implements tenant.PolicyConfigProvider.
func (m *DBTenantConfigManager) PolicyConfig(ctx context.Context, tenantID string) (tenant.PolicyConfig, error) {
	cfg, err := m.GetTenantConfig(ctx, tenantID)
	if err != nil {
		return tenant.PolicyConfig{}, err
	}
	return cfg.Policy, nil
}

// tenantCache is a simple LRU cache for tenant configurations.
type tenantCache struct {
	mu      sync.RWMutex
	entries map[string]*cacheEntry
	maxSize int
	ttl     time.Duration
}

type cacheEntry struct {
	config    tenant.Config
	expiresAt time.Time
}

func newTenantCache(maxSize int, ttl time.Duration) *tenantCache {
	return &tenantCache{
		entries: make(map[string]*cacheEntry),
		maxSize: maxSize,
		ttl:     ttl,
	}
}

func (c *tenantCache) Get(tenantID string) (tenant.Config, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, ok := c.entries[tenantID]
	if !ok {
		return tenant.Config{}, false
	}

	// Check expiration
	if time.Now().After(entry.expiresAt) {
		return tenant.Config{}, false
	}

	return entry.config, true
}

func (c *tenantCache) Set(tenantID string, config tenant.Config) {
	c.mu.Lock()
	defer c.mu.Unlock()

	// Evict if at max size
	if len(c.entries) >= c.maxSize {
		// Simple eviction: remove first entry (should use LRU in production)
		for k := range c.entries {
			delete(c.entries, k)
			break
		}
	}

	c.entries[tenantID] = &cacheEntry{
		config:    config,
		expiresAt: time.Now().Add(c.ttl),
	}
}

func (c *tenantCache) Delete(tenantID string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.entries, tenantID)
}
