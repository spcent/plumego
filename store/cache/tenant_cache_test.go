package cache

import (
	"context"
	"testing"
	"time"

	"github.com/spcent/plumego/tenant"
)

func TestTenantCache_BasicOperations(t *testing.T) {
	cache := NewMemoryCache()
	tenantCache := NewTenantCache(cache)

	ctx1 := tenant.ContextWithTenantID(context.Background(), "tenant-1")
	ctx2 := tenant.ContextWithTenantID(context.Background(), "tenant-2")

	// Set value for tenant 1
	err := tenantCache.Set(ctx1, "key1", []byte("value1"), time.Minute)
	if err != nil {
		t.Fatalf("unexpected error setting key: %v", err)
	}

	// Get value for tenant 1
	val, err := tenantCache.Get(ctx1, "key1")
	if err != nil {
		t.Fatalf("unexpected error getting key: %v", err)
	}
	if string(val) != "value1" {
		t.Errorf("expected 'value1', got '%s'", string(val))
	}

	// Tenant 2 should not see tenant 1's data
	_, err = tenantCache.Get(ctx2, "key1")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound for tenant 2, got %v", err)
	}

	// Set value for tenant 2 with same key
	err = tenantCache.Set(ctx2, "key1", []byte("value2"), time.Minute)
	if err != nil {
		t.Fatalf("unexpected error setting key for tenant 2: %v", err)
	}

	// Verify tenant 1's value unchanged
	val, err = tenantCache.Get(ctx1, "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "value1" {
		t.Errorf("tenant 1 value changed unexpectedly: got '%s'", string(val))
	}

	// Verify tenant 2's value
	val, err = tenantCache.Get(ctx2, "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "value2" {
		t.Errorf("expected 'value2' for tenant 2, got '%s'", string(val))
	}
}

func TestTenantCache_NoTenantID(t *testing.T) {
	cache := NewMemoryCache()
	tenantCache := NewTenantCache(cache)

	ctx := context.Background() // No tenant ID

	// Should fail without tenant ID
	err := tenantCache.Set(ctx, "key1", []byte("value1"), time.Minute)
	if err != tenant.ErrTenantNotFound {
		t.Errorf("expected ErrTenantNotFound, got %v", err)
	}

	_, err = tenantCache.Get(ctx, "key1")
	if err != tenant.ErrTenantNotFound {
		t.Errorf("expected ErrTenantNotFound for Get, got %v", err)
	}
}

func TestTenantCache_KeyPrefix(t *testing.T) {
	cache := NewMemoryCache()
	tenantCache := NewTenantCache(cache, WithKeyPrefix("app"), WithSeparator(":"))

	ctx := tenant.ContextWithTenantID(context.Background(), "test-tenant")

	err := tenantCache.Set(ctx, "mykey", []byte("myvalue"), time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Check that the key is stored with correct prefix in underlying cache
	val, err := cache.Get(context.Background(), "app:test-tenant:mykey")
	if err != nil {
		t.Fatalf("key not found with expected prefix: %v", err)
	}
	if string(val) != "myvalue" {
		t.Errorf("expected 'myvalue', got '%s'", string(val))
	}
}

func TestTenantCache_Delete(t *testing.T) {
	cache := NewMemoryCache()
	tenantCache := NewTenantCache(cache)

	ctx := tenant.ContextWithTenantID(context.Background(), "tenant-1")

	// Set and verify
	tenantCache.Set(ctx, "key1", []byte("value1"), time.Minute)
	val, _ := tenantCache.Get(ctx, "key1")
	if string(val) != "value1" {
		t.Fatal("value not set correctly")
	}

	// Delete
	err := tenantCache.Delete(ctx, "key1")
	if err != nil {
		t.Fatalf("unexpected error deleting: %v", err)
	}

	// Verify deleted
	_, err = tenantCache.Get(ctx, "key1")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestTenantCache_Exists(t *testing.T) {
	cache := NewMemoryCache()
	tenantCache := NewTenantCache(cache)

	ctx1 := tenant.ContextWithTenantID(context.Background(), "tenant-1")
	ctx2 := tenant.ContextWithTenantID(context.Background(), "tenant-2")

	// Key doesn't exist initially
	exists, err := tenantCache.Exists(ctx1, "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("key should not exist initially")
	}

	// Set for tenant 1
	tenantCache.Set(ctx1, "key1", []byte("value1"), time.Minute)

	// Should exist for tenant 1
	exists, err = tenantCache.Exists(ctx1, "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !exists {
		t.Error("key should exist for tenant 1")
	}

	// Should not exist for tenant 2
	exists, err = tenantCache.Exists(ctx2, "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Error("key should not exist for tenant 2")
	}
}

func TestTenantCache_Incr(t *testing.T) {
	cache := NewMemoryCache()
	tenantCache := NewTenantCache(cache)

	ctx1 := tenant.ContextWithTenantID(context.Background(), "tenant-1")
	ctx2 := tenant.ContextWithTenantID(context.Background(), "tenant-2")

	// Increment for tenant 1
	val, err := tenantCache.Incr(ctx1, "counter", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 5 {
		t.Errorf("expected 5, got %d", val)
	}

	// Increment again
	val, err = tenantCache.Incr(ctx1, "counter", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 8 {
		t.Errorf("expected 8, got %d", val)
	}

	// Tenant 2's counter should be independent
	val, err = tenantCache.Incr(ctx2, "counter", 10)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 10 {
		t.Errorf("expected 10 for tenant 2, got %d", val)
	}

	// Verify tenant 1's counter unchanged
	val, err = tenantCache.Incr(ctx1, "counter", 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 8 {
		t.Errorf("expected tenant 1 counter to still be 8, got %d", val)
	}
}

func TestTenantCache_Decr(t *testing.T) {
	cache := NewMemoryCache()
	tenantCache := NewTenantCache(cache)

	ctx := tenant.ContextWithTenantID(context.Background(), "tenant-1")

	// Set initial value
	tenantCache.Incr(ctx, "counter", 100)

	// Decrement
	val, err := tenantCache.Decr(ctx, "counter", 30)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val != 70 {
		t.Errorf("expected 70, got %d", val)
	}
}

func TestTenantCache_Append(t *testing.T) {
	cache := NewMemoryCache()
	tenantCache := NewTenantCache(cache)

	ctx := tenant.ContextWithTenantID(context.Background(), "tenant-1")

	// Append to non-existent key
	err := tenantCache.Append(ctx, "data", []byte("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Append more data
	err = tenantCache.Append(ctx, "data", []byte(" world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Verify
	val, err := tenantCache.Get(ctx, "data")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "hello world" {
		t.Errorf("expected 'hello world', got '%s'", string(val))
	}
}

func TestTenantCache_TTLExpiration(t *testing.T) {
	cache := NewMemoryCache()
	tenantCache := NewTenantCache(cache)

	ctx := tenant.ContextWithTenantID(context.Background(), "tenant-1")

	// Set with short TTL
	err := tenantCache.Set(ctx, "key1", []byte("value1"), 100*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should exist immediately
	val, err := tenantCache.Get(ctx, "key1")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "value1" {
		t.Errorf("expected 'value1', got '%s'", string(val))
	}

	// Wait for expiration
	time.Sleep(150 * time.Millisecond)

	// Should be expired
	_, err = tenantCache.Get(ctx, "key1")
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after expiration, got %v", err)
	}
}

func TestTenantCache_RawCache(t *testing.T) {
	cache := NewMemoryCache()
	tenantCache := NewTenantCache(cache)

	// Should return the underlying cache
	raw := tenantCache.RawCache()
	if raw != cache {
		t.Error("RawCache should return underlying cache")
	}

	// Can bypass tenant isolation with raw cache
	ctx := context.Background()
	err := raw.Set(ctx, "global-key", []byte("global-value"), time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := raw.Get(ctx, "global-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "global-value" {
		t.Errorf("expected 'global-value', got '%s'", string(val))
	}
}

func TestTenantCache_IsolationBetweenTenants(t *testing.T) {
	cache := NewMemoryCache()
	tenantCache := NewTenantCache(cache)

	// Create contexts for 3 different tenants
	ctx1 := tenant.ContextWithTenantID(context.Background(), "tenant-1")
	ctx2 := tenant.ContextWithTenantID(context.Background(), "tenant-2")
	ctx3 := tenant.ContextWithTenantID(context.Background(), "tenant-3")

	// Each tenant sets the same key with different values
	tenantCache.Set(ctx1, "config", []byte("config-1"), time.Minute)
	tenantCache.Set(ctx2, "config", []byte("config-2"), time.Minute)
	tenantCache.Set(ctx3, "config", []byte("config-3"), time.Minute)

	// Verify each tenant gets their own value
	val1, _ := tenantCache.Get(ctx1, "config")
	val2, _ := tenantCache.Get(ctx2, "config")
	val3, _ := tenantCache.Get(ctx3, "config")

	if string(val1) != "config-1" {
		t.Errorf("tenant 1: expected 'config-1', got '%s'", string(val1))
	}
	if string(val2) != "config-2" {
		t.Errorf("tenant 2: expected 'config-2', got '%s'", string(val2))
	}
	if string(val3) != "config-3" {
		t.Errorf("tenant 3: expected 'config-3', got '%s'", string(val3))
	}

	// Delete for tenant 2 shouldn't affect others
	tenantCache.Delete(ctx2, "config")

	val1, _ = tenantCache.Get(ctx1, "config")
	if string(val1) != "config-1" {
		t.Error("tenant 1's value should not be affected by tenant 2's delete")
	}

	_, err := tenantCache.Get(ctx2, "config")
	if err != ErrNotFound {
		t.Error("tenant 2's value should be deleted")
	}

	val3, _ = tenantCache.Get(ctx3, "config")
	if string(val3) != "config-3" {
		t.Error("tenant 3's value should not be affected by tenant 2's delete")
	}
}
