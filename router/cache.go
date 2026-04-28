package router

import (
	"container/list"
	"strings"
	"sync"
	"sync/atomic"
)

// Cache configuration constants (unexported; exposed only via RouterOption).
const (
	defaultPatternCacheSize = 50
	minCacheCapacity        = 10
)

// cacheEntry represents a cached route match result
type cacheEntry struct {
	key   string
	value *matchResult
}

// patternCacheEntry represents a cached route pattern for parameterized routes
type patternCacheEntry struct {
	pattern     string             // Route pattern like /users/:id
	result      *matchResult       // Cached match result (without param values)
	precompiled precompiledPattern // Pre-split pattern for fast matching
}

// matchCache implements a simple LRU cache for route matching results
// It supports two caching strategies:
// 1. Exact path caching for static routes (e.g., /users, /api/health)
// 2. Pattern-based caching for parameterized routes (e.g., /users/:id)
type matchCache struct {
	capacity int
	cache    map[string]*list.Element
	list     *list.List
	mu       sync.RWMutex

	// Pattern cache for parameterized routes
	// Key: method, Value: list of pattern entries sorted by specificity
	patternCache map[string][]patternCacheEntry
	patternMu    sync.RWMutex

	// Metrics
	hits   uint64
	misses uint64
}

// newMatchCache creates a new route cache with the given capacity.
func newMatchCache(capacity int) *matchCache {
	if capacity <= 0 {
		capacity = 100 // Default capacity
	}
	return &matchCache{
		capacity:     capacity,
		cache:        make(map[string]*list.Element),
		list:         list.New(),
		patternCache: make(map[string][]patternCacheEntry),
	}
}

// Get retrieves a cached route match result
func (rc *matchCache) Get(key string) (*matchResult, bool) {
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
	value := element.Value.(*cacheEntry).value
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

// Lookup performs a single cache lookup across exact and parameterized routes.
// It records one hit or one miss for the entire lookup.
func (rc *matchCache) Lookup(method, path, key string) (*matchResult, []string, bool) {
	rc.mu.RLock()
	element, exists := rc.cache[key]
	if exists {
		isAtFront := element == rc.list.Front()
		value := element.Value.(*cacheEntry).value
		rc.mu.RUnlock()

		if !isAtFront {
			rc.mu.Lock()
			if elem, ok := rc.cache[key]; ok && elem == element {
				rc.list.MoveToFront(elem)
			}
			rc.mu.Unlock()
		}

		atomic.AddUint64(&rc.hits, 1)
		return value, nil, true
	}
	rc.mu.RUnlock()

	result, paramValues, found := rc.matchPatternCache(method, path)
	if found {
		atomic.AddUint64(&rc.hits, 1)
	} else {
		atomic.AddUint64(&rc.misses, 1)
	}
	return result, paramValues, found
}

// GetByPattern tries to match a path against cached patterns for parameterized routes.
// Returns the match result and extracted parameter values if found.
func (rc *matchCache) GetByPattern(method, path string) (*matchResult, []string, bool) {
	result, paramValues, found := rc.matchPatternCache(method, path)
	if found {
		atomic.AddUint64(&rc.hits, 1)
	}
	return result, paramValues, found
}

// matchPatternCache is the shared implementation for pattern-based cache lookup.
// It does NOT update hit/miss counters; callers are responsible for that.
func (rc *matchCache) matchPatternCache(method, path string) (*matchResult, []string, bool) {
	rc.patternMu.RLock()
	patterns, exists := rc.patternCache[method]
	if !exists || len(patterns) == 0 {
		rc.patternMu.RUnlock()
		return nil, nil, false
	}

	bufPtr := pathBufPool.Get().(*[]string)
	buf := *bufPtr
	pathParts := fastSplitPath(path, buf)

	paramBufPtr := matchParamBufPool.Get().(*[]string)
	paramBuf := *paramBufPtr

	for i := range patterns {
		entry := patterns[i]
		matchedParams, ok := matchPrecompiled(&entry.precompiled, pathParts, paramBuf)
		if ok {
			paramValues := make([]string, len(matchedParams))
			copy(paramValues, matchedParams)
			rc.patternMu.RUnlock()

			*bufPtr = pathParts[:0]
			pathBufPool.Put(bufPtr)
			*paramBufPtr = matchedParams[:0]
			matchParamBufPool.Put(paramBufPtr)

			return entry.result, paramValues, true
		}
		paramBuf = matchedParams[:0]
	}
	rc.patternMu.RUnlock()

	*bufPtr = pathParts[:0]
	pathBufPool.Put(bufPtr)
	*paramBufPtr = paramBuf[:0]
	matchParamBufPool.Put(paramBufPtr)

	return nil, nil, false
}

// Set adds a route match result to the cache
func (rc *matchCache) Set(key string, value *matchResult) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	// Check if key already exists
	if element, exists := rc.cache[key]; exists {
		rc.list.MoveToFront(element)
		element.Value.(*cacheEntry).value = value
		return
	}

	// Check if cache is full
	if len(rc.cache) >= rc.capacity {
		// Remove least recently used
		oldest := rc.list.Back()
		if oldest != nil {
			rc.list.Remove(oldest)
			delete(rc.cache, oldest.Value.(*cacheEntry).key)
		}
	}

	// Add new entry
	entry := &cacheEntry{key: key, value: value}
	element := rc.list.PushFront(entry)
	rc.cache[key] = element
}

// SetPattern adds a parameterized route pattern to the cache
// This is used for routes like /users/:id where we cache the pattern
// instead of each individual path
func (rc *matchCache) SetPattern(method, pattern string, result *matchResult) {
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

	// Add new pattern entry with precompiled parts for fast matching.
	// Sort by specificity so cache lookup mirrors trie matching:
	// more static segments first, then params, wildcard last.
	entry := patternCacheEntry{
		pattern:     pattern,
		result:      result,
		precompiled: newPrecompiledPattern(pattern),
	}
	newPatternScore := patternSpecificityScore(pattern)
	insertAt := len(patterns) // default: append at end
	for i, existing := range patterns {
		if patternSpecificityScore(existing.pattern) < newPatternScore {
			insertAt = i
			break
		}
	}

	// Insert without creating a temporary slice.
	patterns = append(patterns, patternCacheEntry{}) // grow by one
	copy(patterns[insertAt+1:], patterns[insertAt:]) // shift right
	patterns[insertAt] = entry

	rc.patternCache[method] = patterns
}

func patternSpecificityScore(pattern string) int {
	if pattern == "/" || pattern == "" {
		return 0
	}

	score := 0
	for _, part := range strings.Split(strings.Trim(pattern, "/"), "/") {
		switch {
		case strings.HasPrefix(part, "*"):
			score += 1
		case strings.HasPrefix(part, ":"):
			score += 10
		default:
			score += 100
		}
	}
	return score
}

// isParameterized checks if a route pattern contains parameters or wildcards
func isParameterized(pattern string) bool {
	return strings.Contains(pattern, ":") || strings.Contains(pattern, "*")
}

// Clear removes all entries from the cache
func (rc *matchCache) Clear() {
	rc.mu.Lock()
	rc.cache = make(map[string]*list.Element)
	rc.list = list.New()
	rc.mu.Unlock()

	rc.patternMu.Lock()
	rc.patternCache = make(map[string][]patternCacheEntry)
	rc.patternMu.Unlock()

	atomic.StoreUint64(&rc.hits, 0)
	atomic.StoreUint64(&rc.misses, 0)
}

// Size returns the current number of cached entries
func (rc *matchCache) Size() int {
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
func (rc *matchCache) Stats() matchStats {
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

	return matchStats{
		ExactEntries:   exactSize,
		PatternEntries: patternSize,
		Capacity:       rc.capacity,
		Hits:           hits,
		Misses:         misses,
		HitRate:        hitRate,
	}
}

// matchStats holds cache statistics for internal matcher-cache tests.
type matchStats struct {
	ExactEntries   int     // Number of exact path cache entries
	PatternEntries int     // Number of pattern cache entries
	Capacity       int     // Maximum capacity for exact cache
	Hits           uint64  // Number of cache hits
	Misses         uint64  // Number of cache misses
	HitRate        float64 // Hit rate (hits / total)
}

// withCacheCapacity overrides route cache capacity for internal tests and benchmarks.
func withCacheCapacity(capacity int) RouterOption {
	return func(r *Router) {
		r.state.matchCache = newMatchCache(capacity)
	}
}

// newRouterWithMatchCapacity creates a new router with an explicit matcher-cache capacity.
func newRouterWithMatchCapacity(capacity int, opts ...RouterOption) *Router {
	allOpts := append([]RouterOption{withCacheCapacity(capacity)}, opts...)
	return NewRouter(allOpts...)
}
