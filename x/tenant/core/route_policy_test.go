package tenant

import (
	"context"
	"testing"
	"time"
)

func TestInMemoryRoutePolicyStore(t *testing.T) {
	store := NewInMemoryRoutePolicyStore()
	ctx := context.Background()

	_, err := store.RoutePolicy(ctx, "missing")
	if err != ErrRoutePolicyNotFound {
		t.Fatalf("expected ErrRoutePolicyNotFound, got %v", err)
	}

	policy := RoutePolicy{
		TenantID: "t-1",
		Strategy: "weighted",
		Payload:  []byte(`{"rules":[]}`),
	}
	if err := store.SetRoutePolicy(ctx, policy); err != nil {
		t.Fatalf("set failed: %v", err)
	}

	got, err := store.RoutePolicy(ctx, "t-1")
	if err != nil {
		t.Fatalf("get failed: %v", err)
	}
	if got.TenantID != "t-1" || got.Strategy != "weighted" {
		t.Fatalf("unexpected policy: %+v", got)
	}
}

func TestCachedRoutePolicyProvider(t *testing.T) {
	ctx := context.Background()
	provider := &testRoutePolicyProvider{}
	cache := NewInMemoryRoutePolicyCache(10, 50*time.Millisecond)
	cached := NewCachedRoutePolicyProvider(provider, cache)

	_, err := cached.RoutePolicy(ctx, "t-1")
	if err != nil {
		t.Fatalf("first call failed: %v", err)
	}
	_, err = cached.RoutePolicy(ctx, "t-1")
	if err != nil {
		t.Fatalf("second call failed: %v", err)
	}
	if provider.calls != 1 {
		t.Fatalf("expected 1 provider call, got %d", provider.calls)
	}

	if err := cached.Invalidate(ctx, "t-1"); err != nil {
		t.Fatalf("invalidate failed: %v", err)
	}
	_, _ = cached.RoutePolicy(ctx, "t-1")
	if provider.calls != 2 {
		t.Fatalf("expected 2 provider calls after invalidate, got %d", provider.calls)
	}

	time.Sleep(60 * time.Millisecond)
	_, _ = cached.RoutePolicy(ctx, "t-1")
	if provider.calls != 3 {
		t.Fatalf("expected cache refresh after ttl, got %d", provider.calls)
	}
}

func TestInMemoryRoutePolicyCache_EvictExpiredFirst(t *testing.T) {
	ctx := context.Background()
	maxSize := 3
	// Use a tiny TTL so the pre-filled entries expire immediately.
	cache := NewInMemoryRoutePolicyCache(maxSize, 1*time.Millisecond)
	policy := RoutePolicy{TenantID: "x", Strategy: "weighted"}

	for i := 0; i < maxSize; i++ {
		id := string(rune('a' + i))
		_ = cache.Set(ctx, id, RoutePolicy{TenantID: id})
	}

	// Wait for all entries to expire.
	time.Sleep(5 * time.Millisecond)

	// Adding a new entry should evict an expired one, not fail.
	if err := cache.Set(ctx, "new", policy); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	cache.mu.RLock()
	size := len(cache.entries)
	_, hasNew := cache.entries["new"]
	cache.mu.RUnlock()

	if !hasNew {
		t.Error("new entry should be present after eviction")
	}
	if size > maxSize {
		t.Errorf("cache size %d exceeds maxSize %d", size, maxSize)
	}
}

func TestInMemoryRoutePolicyCache_EvictArbitraryWhenNoneExpired(t *testing.T) {
	ctx := context.Background()
	maxSize := 3
	cache := NewInMemoryRoutePolicyCache(maxSize, 1*time.Hour)

	for i := 0; i < maxSize; i++ {
		id := string(rune('a' + i))
		_ = cache.Set(ctx, id, RoutePolicy{TenantID: id})
	}

	if err := cache.Set(ctx, "new", RoutePolicy{TenantID: "new"}); err != nil {
		t.Fatalf("Set failed: %v", err)
	}
	cache.mu.RLock()
	size := len(cache.entries)
	_, hasNew := cache.entries["new"]
	cache.mu.RUnlock()

	if !hasNew {
		t.Error("new entry should be present")
	}
	if size > maxSize {
		t.Errorf("cache size %d exceeds maxSize %d", size, maxSize)
	}
}

type testRoutePolicyProvider struct {
	calls int
}

func (p *testRoutePolicyProvider) RoutePolicy(ctx context.Context, tenantID string) (RoutePolicy, error) {
	p.calls++
	return RoutePolicy{
		TenantID:  tenantID,
		Strategy:  "weighted",
		Payload:   []byte(`{"rules":[]}`),
		UpdatedAt: time.Now().UTC(),
	}, nil
}
