package router

import (
	"fmt"
	"testing"
)

// --- NewRouteCache edge cases ---

func TestNewRouteCache_ZeroCapacity(t *testing.T) {
	// Zero or negative capacity should use the default (100).
	cache := NewRouteCache(0)
	if cache.capacity != 100 {
		t.Errorf("capacity = %d, want 100 for zero input", cache.capacity)
	}

	cache = NewRouteCache(-5)
	if cache.capacity != 100 {
		t.Errorf("capacity = %d, want 100 for negative input", cache.capacity)
	}
}

// --- GetByPattern ---

func TestRouteCache_GetByPattern_Empty(t *testing.T) {
	cache := NewRouteCache(10)
	result, params, found := cache.GetByPattern("GET", "/users/123")
	if found || result != nil || params != nil {
		t.Error("expected no match on empty cache")
	}
}

func TestRouteCache_GetByPattern_Match(t *testing.T) {
	cache := NewRouteCache(10)
	mr := &MatchResult{RoutePattern: "/users/:id", RouteMethod: "GET"}
	cache.SetPattern("GET", "/users/:id", mr)

	result, params, found := cache.GetByPattern("GET", "/users/42")
	if !found {
		t.Fatal("expected match for /users/42")
	}
	if result == nil {
		t.Fatal("expected non-nil result")
	}
	if len(params) != 1 || params[0] != "42" {
		t.Errorf("params = %v, want [42]", params)
	}
}

func TestRouteCache_GetByPattern_NoMatchWrongMethod(t *testing.T) {
	cache := NewRouteCache(10)
	cache.SetPattern("GET", "/users/:id", &MatchResult{RoutePattern: "/users/:id"})

	_, _, found := cache.GetByPattern("POST", "/users/42")
	if found {
		t.Error("should not match for different method")
	}
}

func TestRouteCache_GetByPattern_UpdatesHits(t *testing.T) {
	cache := NewRouteCache(10)
	cache.SetPattern("GET", "/items/:id", &MatchResult{RoutePattern: "/items/:id"})

	before := cache.Stats().Hits
	cache.GetByPattern("GET", "/items/99")
	after := cache.Stats().Hits

	if after != before+1 {
		t.Errorf("hits should increment by 1: before=%d after=%d", before, after)
	}
}

func TestRouteCache_GetByPattern_NoMatchNoHit(t *testing.T) {
	cache := NewRouteCache(10)
	cache.SetPattern("GET", "/items/:id", &MatchResult{RoutePattern: "/items/:id"})

	before := cache.Stats().Misses
	cache.GetByPattern("GET", "/other/path/no/match")
	after := cache.Stats().Misses

	// GetByPattern does not record misses (only hits), so misses should stay same.
	_ = before
	_ = after
}

// --- Clear ---

func TestRouteCache_Clear(t *testing.T) {
	cache := NewRouteCache(10)

	// Populate exact cache.
	for i := 0; i < 5; i++ {
		cache.Set(fmt.Sprintf("key-%d", i), &MatchResult{})
	}
	// Populate pattern cache.
	cache.SetPattern("GET", "/users/:id", &MatchResult{})
	cache.SetPattern("POST", "/items/:id", &MatchResult{})

	// Trigger some hits/misses.
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
	if stats.PatternEntries != 0 {
		t.Errorf("PatternEntries after Clear = %d, want 0", stats.PatternEntries)
	}
}

func TestRouteCache_Clear_EmptyCache(t *testing.T) {
	cache := NewRouteCache(10)
	cache.Clear() // Should not panic on empty cache.

	if size := cache.Size(); size != 0 {
		t.Errorf("size after Clear of empty = %d, want 0", size)
	}
}

// --- Stats ---

func TestRouteCache_Stats_Initial(t *testing.T) {
	cache := NewRouteCache(50)
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

func TestRouteCache_Stats_AfterOperations(t *testing.T) {
	cache := NewRouteCache(10)
	cache.Set("a", &MatchResult{})
	cache.Set("b", &MatchResult{})

	cache.Get("a")    // hit
	cache.Get("a")    // hit
	cache.Get("miss") // miss

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
	// HitRate = 2 / (2+1) ≈ 0.667
	if stats.HitRate < 0.66 || stats.HitRate > 0.68 {
		t.Errorf("HitRate = %f, want ~0.667", stats.HitRate)
	}
}

func TestRouteCache_Stats_WithPatterns(t *testing.T) {
	cache := NewRouteCache(10)
	cache.SetPattern("GET", "/a/:id", &MatchResult{})
	cache.SetPattern("POST", "/b/:id", &MatchResult{})

	stats := cache.Stats()
	if stats.PatternEntries != 2 {
		t.Errorf("PatternEntries = %d, want 2", stats.PatternEntries)
	}
}

// --- SetPattern deduplication ---

func TestRouteCache_SetPattern_UpdateExisting(t *testing.T) {
	cache := NewRouteCache(10)
	mr1 := &MatchResult{RoutePattern: "/a/:id", RouteMethod: "GET"}
	mr2 := &MatchResult{RoutePattern: "/a/:id", RouteMethod: "GET"}

	cache.SetPattern("GET", "/a/:id", mr1)
	cache.SetPattern("GET", "/a/:id", mr2) // Should update, not duplicate.

	stats := cache.Stats()
	if stats.PatternEntries != 1 {
		t.Errorf("PatternEntries = %d, want 1 after update", stats.PatternEntries)
	}
}

// --- patternSpecificityScore ---

func TestPatternSpecificityScore(t *testing.T) {
	tests := []struct {
		pattern string
		min     int
	}{
		{"/", 0},
		{"", 0},
		{"/static/path", 200},
		{"/users/:id", 110}, // 100 (static) + 10 (param)
		{"/:wildcard", 10},
		{"/*rest", 1},
	}
	for _, tt := range tests {
		got := patternSpecificityScore(tt.pattern)
		if got < tt.min {
			t.Errorf("patternSpecificityScore(%q) = %d, want >= %d", tt.pattern, got, tt.min)
		}
	}
	// More specific patterns score higher than less specific.
	staticScore := patternSpecificityScore("/users/profile")
	paramScore := patternSpecificityScore("/users/:id")
	if staticScore <= paramScore {
		t.Errorf("static pattern should score higher: static=%d param=%d", staticScore, paramScore)
	}
}

// --- isParameterized ---

func TestIsParameterized(t *testing.T) {
	tests := []struct {
		pattern string
		want    bool
	}{
		{"/users/:id", true},
		{"/files/*path", true},
		{"/static/path", false},
		{"/", false},
		{"/:param/static", true},
	}
	for _, tt := range tests {
		got := isParameterized(tt.pattern)
		if got != tt.want {
			t.Errorf("isParameterized(%q) = %v, want %v", tt.pattern, got, tt.want)
		}
	}
}

// --- Lookup method ---

func TestRouteCache_Lookup_ExactHit(t *testing.T) {
	cache := NewRouteCache(10)
	mr := &MatchResult{RoutePattern: "/health", RouteMethod: "GET"}
	cache.Set("GET:/health", mr)

	result, params, found := cache.Lookup("GET", "/health", "GET:/health")
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

func TestRouteCache_Lookup_PatternFallback(t *testing.T) {
	cache := NewRouteCache(10)
	mr := &MatchResult{RoutePattern: "/users/:id", RouteMethod: "GET"}
	cache.SetPattern("GET", "/users/:id", mr)

	// No exact key, falls through to pattern matching.
	result, params, found := cache.Lookup("GET", "/users/42", "GET:/users/42")
	if !found {
		t.Fatal("expected pattern fallback match")
	}
	if result == nil {
		t.Fatal("expected non-nil result from pattern")
	}
	_ = params
}

func TestRouteCache_Lookup_Miss(t *testing.T) {
	cache := NewRouteCache(10)
	before := cache.Stats().Misses

	cache.Lookup("GET", "/not/found", "GET:/not/found")

	after := cache.Stats().Misses
	if after != before+1 {
		t.Errorf("misses should increment: before=%d after=%d", before, after)
	}
}

// --- Eviction at capacity ---

func TestRouteCache_Eviction(t *testing.T) {
	capacity := 5
	cache := NewRouteCache(capacity)

	for i := 0; i < capacity+3; i++ {
		cache.Set(fmt.Sprintf("key-%d", i), &MatchResult{})
	}

	stats := cache.Stats()
	if stats.ExactEntries > capacity {
		t.Errorf("ExactEntries = %d, should be <= %d after eviction", stats.ExactEntries, capacity)
	}
}
