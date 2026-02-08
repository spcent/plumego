package router

import (
	"container/list"
	"strings"
	"sync"
	"sync/atomic"
)

// Cache configuration constants
const (
	// DefaultPatternCacheSize is the default size for pattern cache per method
	DefaultPatternCacheSize = 50

	// MinCacheCapacity is the minimum allowed cache capacity
	MinCacheCapacity = 10
)

// CacheEntry represents a cached route match result
type CacheEntry struct {
	key   string
	value *MatchResult
}

// PatternCacheEntry represents a cached route pattern for parameterized routes
type PatternCacheEntry struct {
	pattern     string             // Route pattern like /users/:id
	result      *MatchResult       // Cached match result (without param values)
	precompiled precompiledPattern // Pre-split pattern for fast matching
}

// RouteCache implements a simple LRU cache for route matching results
// It supports two caching strategies:
// 1. Exact path caching for static routes (e.g., /users, /api/health)
// 2. Pattern-based caching for parameterized routes (e.g., /users/:id)
type RouteCache struct {
	capacity int
	cache    map[string]*list.Element
	list     *list.List
	mu       sync.RWMutex

	// Pattern cache for parameterized routes
	// Key: method, Value: list of pattern entries sorted by specificity
	patternCache map[string][]PatternCacheEntry
	patternMu    sync.RWMutex

	// Metrics
	hits   uint64
	misses uint64
}

// NewRouteCache creates a new route cache with the given capacity
func NewRouteCache(capacity int) *RouteCache {
	if capacity <= 0 {
		capacity = 100 // Default capacity
	}
	return &RouteCache{
		capacity:     capacity,
		cache:        make(map[string]*list.Element),
		list:         list.New(),
		patternCache: make(map[string][]PatternCacheEntry),
	}
}

// Get retrieves a cached route match result
func (rc *RouteCache) Get(key string) (*MatchResult, bool) {
	// First try with read lock
	rc.mu.RLock()
	element, exists := rc.cache[key]
	if !exists {
		rc.mu.RUnlock()
		atomic.AddUint64(&rc.misses, 1)
		return nil, false
	}

	// Check if already at front - if so, no need to move
	isAtFront := element == rc.list.Front()
	value := element.Value.(*CacheEntry).value
	rc.mu.RUnlock()

	// Only acquire write lock if we need to move the element
	if !isAtFront {
		rc.mu.Lock()
		// Double-check element still exists and reacquire it
		if elem, ok := rc.cache[key]; ok && elem == element {
			rc.list.MoveToFront(elem)
		}
		rc.mu.Unlock()
	}

	atomic.AddUint64(&rc.hits, 1)
	return value, true
}

// GetByPattern tries to match a path against cached patterns for parameterized routes
// Returns the match result and extracted parameter values if found
func (rc *RouteCache) GetByPattern(method, path string) (*MatchResult, []string, bool) {
	rc.patternMu.RLock()
	patterns, exists := rc.patternCache[method]
	rc.patternMu.RUnlock()

	if !exists || len(patterns) == 0 {
		return nil, nil, false
	}

	// Use pooled buffer for path splitting to avoid allocation
	bufPtr := pathBufPool.Get().(*[]string)
	buf := *bufPtr
	numParts := fastSplitPath(path, buf)
	pathParts := buf[:numParts]

	// Use pooled buffer for extracted parameter values
	paramBufPtr := matchParamBufPool.Get().(*[]string)
	paramBuf := *paramBufPtr

	for i := range patterns {
		entry := &patterns[i]
		if paramCount, ok := matchPrecompiled(&entry.precompiled, pathParts, paramBuf); ok {
			// Copy params out of the pooled buffer before returning it
			paramValues := make([]string, paramCount)
			copy(paramValues, paramBuf[:paramCount])

			*bufPtr = buf
			pathBufPool.Put(bufPtr)
			*paramBufPtr = paramBuf
			matchParamBufPool.Put(paramBufPtr)

			atomic.AddUint64(&rc.hits, 1)
			return entry.result, paramValues, true
		}
	}

	*bufPtr = buf
	pathBufPool.Put(bufPtr)
	*paramBufPtr = paramBuf
	matchParamBufPool.Put(paramBufPtr)

	return nil, nil, false
}

// matchPattern checks if path parts match a pattern and extracts parameter values
func matchPattern(pattern string, pathParts []string) ([]string, bool) {
	patternParts := strings.Split(strings.Trim(pattern, "/"), "/")

	// Check for wildcard
	hasWildcard := false
	wildcardIdx := -1
	for i, part := range patternParts {
		if strings.HasPrefix(part, "*") {
			hasWildcard = true
			wildcardIdx = i
			break
		}
	}

	if !hasWildcard {
		if len(pathParts) != len(patternParts) {
			return nil, false
		}
	} else {
		if len(pathParts) < wildcardIdx {
			return nil, false
		}
	}

	var paramValues []string

	for i, patternPart := range patternParts {
		if strings.HasPrefix(patternPart, "*") {
			// Wildcard matches rest of path
			paramValues = append(paramValues, strings.Join(pathParts[i:], "/"))
			return paramValues, true
		}

		if i >= len(pathParts) {
			return nil, false
		}

		if strings.HasPrefix(patternPart, ":") {
			// Parameter - capture value
			paramValues = append(paramValues, pathParts[i])
		} else if patternPart != pathParts[i] {
			// Static part must match exactly
			return nil, false
		}
	}

	return paramValues, true
}

// Set adds a route match result to the cache
func (rc *RouteCache) Set(key string, value *MatchResult) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// Check if key already exists
	if element, exists := rc.cache[key]; exists {
		rc.list.MoveToFront(element)
		element.Value.(*CacheEntry).value = value
		return
	}

	// Check if cache is full
	if len(rc.cache) >= rc.capacity {
		// Remove least recently used
		oldest := rc.list.Back()
		if oldest != nil {
			rc.list.Remove(oldest)
			delete(rc.cache, oldest.Value.(*CacheEntry).key)
		}
	}

	// Add new entry
	entry := &CacheEntry{key: key, value: value}
	element := rc.list.PushFront(entry)
	rc.cache[key] = element
}

// SetPattern adds a parameterized route pattern to the cache
// This is used for routes like /users/:id where we cache the pattern
// instead of each individual path
func (rc *RouteCache) SetPattern(method, pattern string, result *MatchResult) {
	rc.patternMu.Lock()
	defer rc.patternMu.Unlock()

	// Check if pattern already exists
	patterns := rc.patternCache[method]
	for i, entry := range patterns {
		if entry.pattern == pattern {
			// Update existing entry
			patterns[i].result = result
			return
		}
	}

	// Add new pattern entry with precompiled parts for fast matching
	// Sort by specificity (more segments = higher priority)
	entry := PatternCacheEntry{
		pattern:     pattern,
		result:      result,
		precompiled: newPrecompiledPattern(pattern),
	}
	inserted := false

	newPatternSegments := countSegments(pattern)
	for i, existing := range patterns {
		if countSegments(existing.pattern) < newPatternSegments {
			// Insert before this entry
			patterns = append(patterns[:i], append([]PatternCacheEntry{entry}, patterns[i:]...)...)
			inserted = true
			break
		}
	}

	if !inserted {
		patterns = append(patterns, entry)
	}

	rc.patternCache[method] = patterns
}

// countSegments counts the number of segments in a path pattern
func countSegments(pattern string) int {
	if pattern == "/" || pattern == "" {
		return 0
	}
	return len(strings.Split(strings.Trim(pattern, "/"), "/"))
}

// IsParameterized checks if a route pattern contains parameters or wildcards
func IsParameterized(pattern string) bool {
	return strings.Contains(pattern, ":") || strings.Contains(pattern, "*")
}

// Clear removes all entries from the cache
func (rc *RouteCache) Clear() {
	rc.mu.Lock()
	rc.cache = make(map[string]*list.Element)
	rc.list = list.New()
	rc.mu.Unlock()

	rc.patternMu.Lock()
	rc.patternCache = make(map[string][]PatternCacheEntry)
	rc.patternMu.Unlock()

	atomic.StoreUint64(&rc.hits, 0)
	atomic.StoreUint64(&rc.misses, 0)
}

// Size returns the current number of cached entries
func (rc *RouteCache) Size() int {
	rc.mu.RLock()
	exactSize := len(rc.cache)
	rc.mu.RUnlock()

	rc.patternMu.RLock()
	patternSize := 0
	for _, patterns := range rc.patternCache {
		patternSize += len(patterns)
	}
	rc.patternMu.RUnlock()

	return exactSize + patternSize
}

// Stats returns cache statistics
func (rc *RouteCache) Stats() CacheStats {
	rc.mu.RLock()
	exactSize := len(rc.cache)
	rc.mu.RUnlock()

	rc.patternMu.RLock()
	patternSize := 0
	for _, patterns := range rc.patternCache {
		patternSize += len(patterns)
	}
	rc.patternMu.RUnlock()

	hits := atomic.LoadUint64(&rc.hits)
	misses := atomic.LoadUint64(&rc.misses)

	var hitRate float64
	total := hits + misses
	if total > 0 {
		hitRate = float64(hits) / float64(total)
	}

	return CacheStats{
		ExactEntries:   exactSize,
		PatternEntries: patternSize,
		Capacity:       rc.capacity,
		Hits:           hits,
		Misses:         misses,
		HitRate:        hitRate,
	}
}

// CacheStats holds cache statistics
type CacheStats struct {
	ExactEntries   int     // Number of exact path cache entries
	PatternEntries int     // Number of pattern cache entries
	Capacity       int     // Maximum capacity for exact cache
	Hits           uint64  // Number of cache hits
	Misses         uint64  // Number of cache misses
	HitRate        float64 // Hit rate (hits / total)
}

// WithCache creates a router option that enables route caching
func WithCache(capacity int) RouterOption {
	return func(r *Router) {
		r.routeCache = NewRouteCache(capacity)
	}
}

// NewRouterWithCache creates a new router with route caching enabled
func NewRouterWithCache(capacity int, opts ...RouterOption) *Router {
	allOpts := append([]RouterOption{WithCache(capacity)}, opts...)
	return NewRouter(allOpts...)
}
