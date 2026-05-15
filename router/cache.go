package router

import (
	"container/list"
	"sync"
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
}

// newMatchCache creates a new route cache with the given capacity.
func newMatchCache(capacity int) *matchCache {
	if capacity <= 0 {
		capacity = defaultCacheCapacity
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

	return value, true
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
}
