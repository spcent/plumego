package db

import (
	"context"
	"database/sql"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
	"github.com/spcent/plumego/tenant"
)

func setupTestDB(t *testing.T) *sql.DB {
	db, err := sql.Open("sqlite3", ":memory:")
	if err != nil {
		t.Fatalf("failed to open test database: %v", err)
	}

	// Create table (SQLite version)
	schema := `
		CREATE TABLE tenants (
			id TEXT PRIMARY KEY,
			quota_requests_per_minute INTEGER NOT NULL DEFAULT 0,
			quota_tokens_per_minute INTEGER NOT NULL DEFAULT 0,
			allowed_models TEXT,
			allowed_tools TEXT,
			metadata TEXT,
			created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
			updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
		CREATE INDEX idx_tenants_updated_at ON tenants(updated_at);
	`

	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create schema: %v", err)
	}

	return db
}

func TestNewDBTenantConfigManager(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewDBTenantConfigManager(db)
	if manager == nil {
		t.Fatal("expected non-nil manager")
	}
	if manager.db != db {
		t.Error("expected db to be set")
	}
}

func TestNewDBTenantConfigManager_WithCache(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewDBTenantConfigManager(db,
		WithTenantCache(100, 5*time.Minute),
	)

	if manager.cache == nil {
		t.Error("expected cache to be configured")
	}
	if manager.cache.maxSize != 100 {
		t.Errorf("expected cache size 100, got %d", manager.cache.maxSize)
	}
	if manager.cache.ttl != 5*time.Minute {
		t.Errorf("expected cache TTL 5m, got %v", manager.cache.ttl)
	}
}

func TestDBTenantConfigManager_SetGet(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewDBTenantConfigManager(db)
	ctx := context.Background()

	cfg := tenant.Config{
		TenantID: "test-tenant",
		Quota: tenant.QuotaConfig{
			RequestsPerMinute: 100,
			TokensPerMinute:   5000,
		},
		Policy: tenant.PolicyConfig{
			AllowedModels: []string{"gpt-4", "gpt-3.5-turbo"},
			AllowedTools:  []string{"search", "calculator"},
		},
		Metadata: map[string]string{
			"name":  "Test Tenant",
			"tier":  "premium",
			"email": "admin@test.com",
		},
	}

	// Set config
	if err := manager.SetTenantConfig(ctx, cfg); err != nil {
		t.Fatalf("SetTenantConfig failed: %v", err)
	}

	// Get config
	retrieved, err := manager.GetTenantConfig(ctx, "test-tenant")
	if err != nil {
		t.Fatalf("GetTenantConfig failed: %v", err)
	}

	// Verify fields
	if retrieved.TenantID != cfg.TenantID {
		t.Errorf("expected tenant ID %s, got %s", cfg.TenantID, retrieved.TenantID)
	}
	if retrieved.Quota.RequestsPerMinute != cfg.Quota.RequestsPerMinute {
		t.Errorf("expected requests %d, got %d", cfg.Quota.RequestsPerMinute, retrieved.Quota.RequestsPerMinute)
	}
	if retrieved.Quota.TokensPerMinute != cfg.Quota.TokensPerMinute {
		t.Errorf("expected tokens %d, got %d", cfg.Quota.TokensPerMinute, retrieved.Quota.TokensPerMinute)
	}
	if len(retrieved.Policy.AllowedModels) != len(cfg.Policy.AllowedModels) {
		t.Errorf("expected %d models, got %d", len(cfg.Policy.AllowedModels), len(retrieved.Policy.AllowedModels))
	}
	if len(retrieved.Policy.AllowedTools) != len(cfg.Policy.AllowedTools) {
		t.Errorf("expected %d tools, got %d", len(cfg.Policy.AllowedTools), len(retrieved.Policy.AllowedTools))
	}
	if len(retrieved.Metadata) != len(cfg.Metadata) {
		t.Errorf("expected %d metadata entries, got %d", len(cfg.Metadata), len(retrieved.Metadata))
	}
}

func TestDBTenantConfigManager_Update(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewDBTenantConfigManager(db)
	ctx := context.Background()

	// Initial config
	cfg1 := tenant.Config{
		TenantID: "update-test",
		Quota: tenant.QuotaConfig{
			RequestsPerMinute: 10,
		},
	}
	manager.SetTenantConfig(ctx, cfg1)

	// Update config
	cfg2 := tenant.Config{
		TenantID: "update-test",
		Quota: tenant.QuotaConfig{
			RequestsPerMinute: 20,
			TokensPerMinute:   1000,
		},
	}
	manager.SetTenantConfig(ctx, cfg2)

	// Verify updated config
	retrieved, err := manager.GetTenantConfig(ctx, "update-test")
	if err != nil {
		t.Fatalf("GetTenantConfig failed: %v", err)
	}
	if retrieved.Quota.RequestsPerMinute != 20 {
		t.Errorf("expected requests 20, got %d", retrieved.Quota.RequestsPerMinute)
	}
	if retrieved.Quota.TokensPerMinute != 1000 {
		t.Errorf("expected tokens 1000, got %d", retrieved.Quota.TokensPerMinute)
	}
}

func TestDBTenantConfigManager_Delete(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewDBTenantConfigManager(db)
	ctx := context.Background()

	// Create tenant
	cfg := tenant.Config{
		TenantID: "delete-test",
	}
	manager.SetTenantConfig(ctx, cfg)

	// Delete tenant
	if err := manager.DeleteTenantConfig(ctx, "delete-test"); err != nil {
		t.Fatalf("DeleteTenantConfig failed: %v", err)
	}

	// Verify deleted
	_, err := manager.GetTenantConfig(ctx, "delete-test")
	if err != tenant.ErrTenantNotFound {
		t.Errorf("expected ErrTenantNotFound, got %v", err)
	}
}

func TestDBTenantConfigManager_DeleteNonExistent(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewDBTenantConfigManager(db)
	ctx := context.Background()

	err := manager.DeleteTenantConfig(ctx, "non-existent")
	if err != tenant.ErrTenantNotFound {
		t.Errorf("expected ErrTenantNotFound, got %v", err)
	}
}

func TestDBTenantConfigManager_GetNotFound(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewDBTenantConfigManager(db)
	ctx := context.Background()

	_, err := manager.GetTenantConfig(ctx, "non-existent")
	if err != tenant.ErrTenantNotFound {
		t.Errorf("expected ErrTenantNotFound, got %v", err)
	}
}

func TestDBTenantConfigManager_ListTenants(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewDBTenantConfigManager(db)
	ctx := context.Background()

	// Create multiple tenants
	for i := 1; i <= 5; i++ {
		cfg := tenant.Config{
			TenantID: "tenant-" + string(rune('0'+i)),
			Quota: tenant.QuotaConfig{
				RequestsPerMinute: i * 10,
			},
		}
		manager.SetTenantConfig(ctx, cfg)
	}

	// List all
	configs, err := manager.ListTenants(ctx, 10, 0)
	if err != nil {
		t.Fatalf("ListTenants failed: %v", err)
	}
	if len(configs) != 5 {
		t.Errorf("expected 5 tenants, got %d", len(configs))
	}

	// List with limit
	configs, err = manager.ListTenants(ctx, 2, 0)
	if err != nil {
		t.Fatalf("ListTenants failed: %v", err)
	}
	if len(configs) != 2 {
		t.Errorf("expected 2 tenants, got %d", len(configs))
	}

	// List with offset
	configs, err = manager.ListTenants(ctx, 10, 3)
	if err != nil {
		t.Fatalf("ListTenants failed: %v", err)
	}
	if len(configs) != 2 {
		t.Errorf("expected 2 tenants (5-3), got %d", len(configs))
	}
}

func TestDBTenantConfigManager_QuotaProvider(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewDBTenantConfigManager(db)
	ctx := context.Background()

	cfg := tenant.Config{
		TenantID: "quota-test",
		Quota: tenant.QuotaConfig{
			RequestsPerMinute: 50,
			TokensPerMinute:   2500,
		},
	}
	manager.SetTenantConfig(ctx, cfg)

	// Test QuotaConfigProvider interface
	quota, err := manager.QuotaConfig(ctx, "quota-test")
	if err != nil {
		t.Fatalf("QuotaConfig failed: %v", err)
	}
	if quota.RequestsPerMinute != 50 {
		t.Errorf("expected 50 requests, got %d", quota.RequestsPerMinute)
	}
	if quota.TokensPerMinute != 2500 {
		t.Errorf("expected 2500 tokens, got %d", quota.TokensPerMinute)
	}
}

func TestDBTenantConfigManager_PolicyProvider(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewDBTenantConfigManager(db)
	ctx := context.Background()

	cfg := tenant.Config{
		TenantID: "policy-test",
		Policy: tenant.PolicyConfig{
			AllowedModels: []string{"gpt-4"},
			AllowedTools:  []string{"search"},
		},
	}
	manager.SetTenantConfig(ctx, cfg)

	// Test PolicyConfigProvider interface
	policy, err := manager.PolicyConfig(ctx, "policy-test")
	if err != nil {
		t.Fatalf("PolicyConfig failed: %v", err)
	}
	if len(policy.AllowedModels) != 1 {
		t.Errorf("expected 1 model, got %d", len(policy.AllowedModels))
	}
	if len(policy.AllowedTools) != 1 {
		t.Errorf("expected 1 tool, got %d", len(policy.AllowedTools))
	}
}

func TestDBTenantConfigManager_CacheHit(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewDBTenantConfigManager(db,
		WithTenantCache(10, 1*time.Minute),
	)
	ctx := context.Background()

	cfg := tenant.Config{
		TenantID: "cache-test",
		Quota: tenant.QuotaConfig{
			RequestsPerMinute: 100,
		},
	}
	manager.SetTenantConfig(ctx, cfg)

	// First Get - cache miss, loads from DB
	_, err := manager.GetTenantConfig(ctx, "cache-test")
	if err != nil {
		t.Fatalf("first Get failed: %v", err)
	}

	// Delete from DB (but still in cache)
	db.Exec("DELETE FROM tenants WHERE id = 'cache-test'")

	// Second Get - should return from cache (not fail)
	cached, err := manager.GetTenantConfig(ctx, "cache-test")
	if err != nil {
		t.Fatalf("second Get failed: %v", err)
	}
	if cached.TenantID != "cache-test" {
		t.Error("expected cached result")
	}
}

func TestDBTenantConfigManager_CacheInvalidation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	manager := NewDBTenantConfigManager(db,
		WithTenantCache(10, 1*time.Minute),
	)
	ctx := context.Background()

	cfg := tenant.Config{
		TenantID: "invalidate-test",
		Quota: tenant.QuotaConfig{
			RequestsPerMinute: 100,
		},
	}
	manager.SetTenantConfig(ctx, cfg)

	// Get and cache
	manager.GetTenantConfig(ctx, "invalidate-test")

	// Update should invalidate cache
	cfg.Quota.RequestsPerMinute = 200
	manager.SetTenantConfig(ctx, cfg)

	// Get again - should fetch fresh from DB
	updated, err := manager.GetTenantConfig(ctx, "invalidate-test")
	if err != nil {
		t.Fatalf("Get after update failed: %v", err)
	}
	if updated.Quota.RequestsPerMinute != 200 {
		t.Errorf("expected updated value 200, got %d (cache not invalidated?)", updated.Quota.RequestsPerMinute)
	}
}

func TestDBTenantConfigManager_NilDB(t *testing.T) {
	var manager *DBTenantConfigManager
	ctx := context.Background()

	_, err := manager.GetTenantConfig(ctx, "test")
	if err != tenant.ErrTenantNotFound {
		t.Errorf("expected ErrTenantNotFound for nil manager, got %v", err)
	}
}

func TestTenantCache_Expiration(t *testing.T) {
	cache := newTenantCache(10, 50*time.Millisecond)

	cfg := tenant.Config{TenantID: "expire-test"}
	cache.Set("expire-test", cfg)

	// Immediate get - should succeed
	_, ok := cache.Get("expire-test")
	if !ok {
		t.Error("expected cache hit immediately after set")
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Get after expiration - should miss
	_, ok = cache.Get("expire-test")
	if ok {
		t.Error("expected cache miss after expiration")
	}
}

func TestTenantCache_Eviction(t *testing.T) {
	cache := newTenantCache(2, 1*time.Minute)

	// Fill cache
	cache.Set("tenant-1", tenant.Config{TenantID: "tenant-1"})
	cache.Set("tenant-2", tenant.Config{TenantID: "tenant-2"})

	// Add third (should evict one)
	cache.Set("tenant-3", tenant.Config{TenantID: "tenant-3"})

	// Cache should have exactly 2 entries
	count := 0
	for _, id := range []string{"tenant-1", "tenant-2", "tenant-3"} {
		if _, ok := cache.Get(id); ok {
			count++
		}
	}
	if count != 2 {
		t.Errorf("expected 2 entries after eviction, got %d", count)
	}
}
