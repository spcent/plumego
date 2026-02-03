package tenant

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestInMemoryConfigManager_SetGet(t *testing.T) {
	mgr := NewInMemoryConfigManager()

	cfg := Config{
		TenantID: "test-tenant",
		Quota: QuotaConfig{
			RequestsPerMinute: 100,
			TokensPerMinute:   5000,
		},
		Policy: PolicyConfig{
			AllowedModels: []string{"gpt-4", "gpt-3.5"},
			AllowedTools:  []string{"search", "calculator"},
		},
		Metadata: map[string]string{
			"name":  "Test Tenant",
			"tier":  "premium",
			"email": "admin@test.com",
		},
	}

	// Set config
	mgr.SetTenantConfig(cfg)

	// Get config
	ctx := context.Background()
	retrieved, err := mgr.GetTenantConfig(ctx, "test-tenant")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
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
	if len(retrieved.Metadata) != len(cfg.Metadata) {
		t.Errorf("expected %d metadata entries, got %d", len(cfg.Metadata), len(retrieved.Metadata))
	}
}

func TestInMemoryConfigManager_NotFound(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	ctx := context.Background()

	_, err := mgr.GetTenantConfig(ctx, "non-existent")
	if err != ErrTenantNotFound {
		t.Errorf("expected ErrTenantNotFound, got %v", err)
	}
}

func TestInMemoryConfigManager_UpdatedAt(t *testing.T) {
	mgr := NewInMemoryConfigManager()

	before := time.Now().UTC()
	cfg := Config{
		TenantID: "test-tenant",
	}
	mgr.SetTenantConfig(cfg)
	after := time.Now().UTC()

	ctx := context.Background()
	retrieved, err := mgr.GetTenantConfig(ctx, "test-tenant")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if retrieved.UpdatedAt.Before(before) || retrieved.UpdatedAt.After(after) {
		t.Errorf("UpdatedAt %v not between %v and %v", retrieved.UpdatedAt, before, after)
	}

	// Update config and verify timestamp changes
	time.Sleep(10 * time.Millisecond)
	beforeUpdate := time.Now().UTC()
	mgr.SetTenantConfig(cfg)
	retrieved2, _ := mgr.GetTenantConfig(ctx, "test-tenant")

	if !retrieved2.UpdatedAt.After(retrieved.UpdatedAt) {
		t.Errorf("expected UpdatedAt to increase on update")
	}
	if retrieved2.UpdatedAt.Before(beforeUpdate) {
		t.Errorf("UpdatedAt should be after second update time")
	}
}

func TestInMemoryConfigManager_Concurrent(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	ctx := context.Background()

	// Concurrent writes
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			cfg := Config{
				TenantID: "concurrent-test",
				Quota: QuotaConfig{
					RequestsPerMinute: id * 10,
				},
			}
			mgr.SetTenantConfig(cfg)
		}(i)
	}
	wg.Wait()

	// Should not panic and should have a value
	cfg, err := mgr.GetTenantConfig(ctx, "concurrent-test")
	if err != nil {
		t.Fatalf("expected config to exist after concurrent writes")
	}
	if cfg.TenantID != "concurrent-test" {
		t.Errorf("unexpected tenant ID: %s", cfg.TenantID)
	}

	// Concurrent reads
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_, err := mgr.GetTenantConfig(ctx, "concurrent-test")
			if err != nil {
				t.Errorf("concurrent read failed: %v", err)
			}
		}()
	}
	wg.Wait()
}

func TestInMemoryConfigManager_QuotaProvider(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	ctx := context.Background()

	cfg := Config{
		TenantID: "quota-test",
		Quota: QuotaConfig{
			RequestsPerMinute: 50,
			TokensPerMinute:   2500,
		},
	}
	mgr.SetTenantConfig(cfg)

	// Test QuotaConfigProvider interface
	quota, err := mgr.QuotaConfig(ctx, "quota-test")
	if err != nil {
		t.Fatalf("QuotaConfig failed: %v", err)
	}
	if quota.RequestsPerMinute != 50 {
		t.Errorf("expected 50 requests, got %d", quota.RequestsPerMinute)
	}
	if quota.TokensPerMinute != 2500 {
		t.Errorf("expected 2500 tokens, got %d", quota.TokensPerMinute)
	}

	// Test non-existent tenant
	_, err = mgr.QuotaConfig(ctx, "non-existent")
	if err != ErrTenantNotFound {
		t.Errorf("expected ErrTenantNotFound, got %v", err)
	}
}

func TestInMemoryConfigManager_PolicyProvider(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	ctx := context.Background()

	cfg := Config{
		TenantID: "policy-test",
		Policy: PolicyConfig{
			AllowedModels: []string{"gpt-4"},
			AllowedTools:  []string{"search"},
		},
	}
	mgr.SetTenantConfig(cfg)

	// Test PolicyConfigProvider interface
	policy, err := mgr.PolicyConfig(ctx, "policy-test")
	if err != nil {
		t.Fatalf("PolicyConfig failed: %v", err)
	}
	if len(policy.AllowedModels) != 1 || policy.AllowedModels[0] != "gpt-4" {
		t.Errorf("unexpected allowed models: %v", policy.AllowedModels)
	}
	if len(policy.AllowedTools) != 1 || policy.AllowedTools[0] != "search" {
		t.Errorf("unexpected allowed tools: %v", policy.AllowedTools)
	}

	// Test non-existent tenant
	_, err = mgr.PolicyConfig(ctx, "non-existent")
	if err != ErrTenantNotFound {
		t.Errorf("expected ErrTenantNotFound, got %v", err)
	}
}

func TestInMemoryConfigManager_NilSafety(t *testing.T) {
	var mgr *InMemoryConfigManager

	// SetTenantConfig should not panic
	mgr.SetTenantConfig(Config{TenantID: "test"})

	// GetTenantConfig should return error
	ctx := context.Background()
	_, err := mgr.GetTenantConfig(ctx, "test")
	if err != ErrTenantNotFound {
		t.Errorf("expected ErrTenantNotFound for nil manager, got %v", err)
	}
}

func TestInMemoryConfigManager_MultipleUpdates(t *testing.T) {
	mgr := NewInMemoryConfigManager()
	ctx := context.Background()

	// Initial config
	cfg1 := Config{
		TenantID: "update-test",
		Quota: QuotaConfig{
			RequestsPerMinute: 10,
		},
		Metadata: map[string]string{"version": "1"},
	}
	mgr.SetTenantConfig(cfg1)

	// Update config
	cfg2 := Config{
		TenantID: "update-test",
		Quota: QuotaConfig{
			RequestsPerMinute: 20,
			TokensPerMinute:   1000,
		},
		Metadata: map[string]string{"version": "2"},
	}
	mgr.SetTenantConfig(cfg2)

	// Verify latest config
	retrieved, err := mgr.GetTenantConfig(ctx, "update-test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if retrieved.Quota.RequestsPerMinute != 20 {
		t.Errorf("expected requests 20, got %d", retrieved.Quota.RequestsPerMinute)
	}
	if retrieved.Quota.TokensPerMinute != 1000 {
		t.Errorf("expected tokens 1000, got %d", retrieved.Quota.TokensPerMinute)
	}
	if retrieved.Metadata["version"] != "2" {
		t.Errorf("expected version 2, got %s", retrieved.Metadata["version"])
	}
}
