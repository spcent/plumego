package router

import (
	"fmt"
	"testing"
)

func TestNewMatchCacheZeroCapacity(t *testing.T) {
	cache := newMatchCache(0)
	if cache.capacity != 100 {
		t.Errorf("capacity = %d, want 100 for zero input", cache.capacity)
	}

	cache = newMatchCache(-5)
	if cache.capacity != 100 {
		t.Errorf("capacity = %d, want 100 for negative input", cache.capacity)
	}
}

func TestMatcherCacheClear(t *testing.T) {
	cache := newMatchCache(10)

	for i := 0; i < 5; i++ {
		cache.Set(fmt.Sprintf("key-%d", i), &matchResult{})
	}

	cache.Get("key-0")
	cache.Get("missing")

	if cache.Size() == 0 {
		t.Fatal("cache should have entries before Clear")
	}

	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Size after Clear = %d, want 0", cache.Size())
	}

	stats := cache.Stats()
	if stats.Hits != 0 {
		t.Errorf("Hits after Clear = %d, want 0", stats.Hits)
	}
	if stats.Misses != 0 {
		t.Errorf("Misses after Clear = %d, want 0", stats.Misses)
	}
	if stats.ExactEntries != 0 {
		t.Errorf("ExactEntries after Clear = %d, want 0", stats.ExactEntries)
	}
}

func TestMatcherCacheClearEmptyCache(t *testing.T) {
	cache := newMatchCache(10)
	cache.Clear()

	if size := cache.Size(); size != 0 {
		t.Errorf("size after Clear of empty = %d, want 0", size)
	}
}

func TestMatcherStatsInitial(t *testing.T) {
	cache := newMatchCache(50)
	stats := cache.Stats()

	if stats.Capacity != 50 {
		t.Errorf("Capacity = %d, want 50", stats.Capacity)
	}
	if stats.Hits != 0 {
		t.Errorf("Hits = %d, want 0", stats.Hits)
	}
	if stats.Misses != 0 {
		t.Errorf("Misses = %d, want 0", stats.Misses)
	}
	if stats.HitRate != 0 {
		t.Errorf("HitRate = %f, want 0", stats.HitRate)
	}
	if stats.ExactEntries != 0 {
		t.Errorf("ExactEntries = %d, want 0", stats.ExactEntries)
	}
}

func TestMatcherStatsAfterOperations(t *testing.T) {
	cache := newMatchCache(10)
	cache.Set("a", &matchResult{})
	cache.Set("b", &matchResult{})

	cache.Get("a")
	cache.Get("a")
	cache.Get("miss")

	stats := cache.Stats()
	if stats.ExactEntries != 2 {
		t.Errorf("ExactEntries = %d, want 2", stats.ExactEntries)
	}
	if stats.Hits != 2 {
		t.Errorf("Hits = %d, want 2", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Misses = %d, want 1", stats.Misses)
	}
	if stats.HitRate < 0.66 || stats.HitRate > 0.68 {
		t.Errorf("HitRate = %f, want ~0.667", stats.HitRate)
	}
}

func TestMatcherCacheLookupExactHit(t *testing.T) {
	cache := newMatchCache(10)
	mr := &matchResult{RoutePattern: "/health", RouteMethod: "GET"}
	cache.Set("GET:/health", mr)

	result, params, found := cache.Lookup("GET:/health")
	if !found {
		t.Fatal("expected hit on Lookup")
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if params != nil {
		t.Error("params should be nil for exact match")
	}
}

func TestMatcherCacheLookupMiss(t *testing.T) {
	cache := newMatchCache(10)
	before := cache.Stats().Misses

	cache.Lookup("GET:/not/found")

	after := cache.Stats().Misses
	if after != before+1 {
		t.Errorf("misses should increment: before=%d after=%d", before, after)
	}
}

func TestMatcherCacheEviction(t *testing.T) {
	capacity := 5
	cache := newMatchCache(capacity)

	for i := 0; i < capacity+3; i++ {
		cache.Set(fmt.Sprintf("key-%d", i), &matchResult{})
	}

	stats := cache.Stats()
	if stats.ExactEntries > capacity {
		t.Errorf("ExactEntries = %d, should be <= %d after eviction", stats.ExactEntries, capacity)
	}
}
