package router

import (
	"container/list"
	"sync"
)

// CacheEntry represents a cached route match result
type CacheEntry struct {
	key   string
	value *MatchResult
}

// RouteCache implements a simple LRU cache for route matching results
type RouteCache struct {
	capacity int
	cache    map[string]*list.Element
	list     *list.List
	mu       sync.RWMutex
}

// NewRouteCache creates a new route cache with the given capacity
func NewRouteCache(capacity int) *RouteCache {
	if capacity <= 0 {
		capacity = 100 // Default capacity
	}
	return &RouteCache{
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		list:     list.New(),
	}
}

// Get retrieves a cached route match result
func (rc *RouteCache) Get(key string) (*MatchResult, bool) {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	element, exists := rc.cache[key]
	if !exists {
		return nil, false
	}

	// Move to front (most recently used)
	rc.list.MoveToFront(element)
	return element.Value.(*CacheEntry).value, true
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

// Clear removes all entries from the cache
func (rc *RouteCache) Clear() {
	rc.mu.Lock()
	defer rc.mu.Unlock()

	rc.cache = make(map[string]*list.Element)
	rc.list = list.New()
}

// Size returns the current number of cached entries
func (rc *RouteCache) Size() int {
	rc.mu.RLock()
	defer rc.mu.RUnlock()
	return len(rc.cache)
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
