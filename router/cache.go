package router

import (
	"container/list"
	"sync"
	"sync/atomic"
)

// cacheEntry represents a cached route match result
type cacheEntry struct {
	key   string
	value *matchResult
}

// matchCache implements a simple LRU cache for route matching results
// for concrete request paths. Route selection always comes from the trie before
// a path is cached so warm-cache dispatch preserves trie precedence.
type matchCache struct {
	capacity int
	cache    map[string]*list.Element
	list     *list.List
	mu       sync.RWMutex

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
		capacity: capacity,
		cache:    make(map[string]*list.Element),
		list:     list.New(),
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

// Lookup performs a single cache lookup by concrete route cache key.
// It records one hit or one miss for the entire lookup.
func (rc *matchCache) Lookup(key string) (*matchResult, []string, bool) {
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

	atomic.AddUint64(&rc.misses, 1)
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

// Clear removes all entries from the cache
func (rc *matchCache) Clear() {
	rc.mu.Lock()
	rc.cache = make(map[string]*list.Element)
	rc.list = list.New()
	rc.mu.Unlock()

	atomic.StoreUint64(&rc.hits, 0)
	atomic.StoreUint64(&rc.misses, 0)
}

// Size returns the current number of cached entries
func (rc *matchCache) Size() int {
	rc.mu.RLock()
	size := len(rc.cache)
	rc.mu.RUnlock()
	return size
}

// Stats returns cache statistics
func (rc *matchCache) Stats() matchStats {
	rc.mu.RLock()
	exactSize := len(rc.cache)
	rc.mu.RUnlock()

	hits := atomic.LoadUint64(&rc.hits)
	misses := atomic.LoadUint64(&rc.misses)

	var hitRate float64
	total := hits + misses
	if total > 0 {
		hitRate = float64(hits) / float64(total)
	}

	return matchStats{
		ExactEntries: exactSize,
		Capacity:     rc.capacity,
		Hits:         hits,
		Misses:       misses,
		HitRate:      hitRate,
	}
}

// matchStats holds cache statistics for internal matcher-cache tests.
type matchStats struct {
	ExactEntries int     // Number of exact path cache entries
	Capacity     int     // Maximum capacity for exact cache
	Hits         uint64  // Number of cache hits
	Misses       uint64  // Number of cache misses
	HitRate      float64 // Hit rate (hits / total)
}
