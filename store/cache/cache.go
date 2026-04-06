// Package cache provides an in-memory caching system with TTL and LRU eviction.
//
// This package implements an in-process cache primitive with features including:
//   - TTL (time-to-live) expiration for entries
//   - memory-bounded eviction
//   - metrics collection
//   - thread-safe operations
//
// Example usage:
//
//	import "github.com/spcent/plumego/store/cache"
//
//	// Create a cache with 1000 items max, 5 minute TTL
//	c := cache.New(cache.Config{
//		MaxSize: 1000,
//		TTL:     5 * time.Minute,
//	})
//
//	// Set a value
//	c.Set("user:123", userData)
//
//	// Get a value
//	if val, found := c.Get("user:123"); found {
//		// Use val
//	}
//
//	// Delete a value
//	c.Delete("user:123")
package cache

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"sync"
	"time"
)

var (
	// ErrNotFound is returned when a cache entry is missing or expired.
	ErrNotFound = errors.New("cache: key not found")

	// ErrInvalidConfig is returned when configuration is invalid.
	ErrInvalidConfig = errors.New("cache: invalid config")

	// ErrCacheMiss is returned when cache lookup fails.
	ErrCacheMiss = errors.New("cache: cache miss")

	// ErrCacheFull is returned when cache reaches capacity limit.
	ErrCacheFull = errors.New("cache: cache full")

	// ErrKeyTooLong is returned when cache key exceeds maximum length.
	ErrKeyTooLong = errors.New("cache: key too long")

	// ErrNotInteger is returned when attempting increment/decrement on non-integer value.
	ErrNotInteger = errors.New("cache: value is not an integer")
)

const (
	// DefaultMaxKeyLength is the default maximum key length.
	DefaultMaxKeyLength = 256

	// DefaultMaxMemoryUsage is the default maximum memory usage in bytes (0 = no limit).
	DefaultMaxMemoryUsage = 0

	// DefaultCleanupInterval is the default cleanup interval for expired items.
	DefaultCleanupInterval = 5 * time.Minute
)

// Cache defines the minimal contract for cache backends.
type Cache interface {
	// Basic operations
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Clear(ctx context.Context) error

	// Atomic operations
	// Incr increments the integer value of a key by delta.
	// Returns the new value after increment.
	// If the key doesn't exist, it's created with delta as the initial value.
	// Returns ErrNotInteger if the value is not an integer.
	Incr(ctx context.Context, key string, delta int64) (int64, error)

	// Decr decrements the integer value of a key by delta.
	// Returns the new value after decrement.
	// If the key doesn't exist, it's created with -delta as the initial value.
	// Returns ErrNotInteger if the value is not an integer.
	Decr(ctx context.Context, key string, delta int64) (int64, error)

	// Append appends data to the end of an existing value.
	// If the key doesn't exist, it's created with the data as the value.
	Append(ctx context.Context, key string, data []byte) error
}

// Config defines the configuration for cache backends.
type Config struct {
	// MaxKeyLength is the maximum allowed key length (0 = no limit).
	MaxKeyLength int

	// MaxMemoryUsage is the maximum memory usage in bytes (0 = no limit).
	MaxMemoryUsage uint64

	// CleanupInterval is the interval for cleaning up expired items.
	CleanupInterval time.Duration

	// EnableMetrics enables metrics collection.
	EnableMetrics bool

	// DefaultTTL is the default time-to-live for items without explicit TTL.
	DefaultTTL time.Duration
}

// DefaultConfig returns a default configuration.
func DefaultConfig() Config {
	return Config{
		MaxKeyLength:    DefaultMaxKeyLength,
		MaxMemoryUsage:  DefaultMaxMemoryUsage,
		CleanupInterval: DefaultCleanupInterval,
		EnableMetrics:   true,
		DefaultTTL:      10 * time.Minute,
	}
}

// Validate checks if the configuration is valid.
func (c Config) Validate() error {
	if c.MaxKeyLength < 0 {
		return fmt.Errorf("%w: MaxKeyLength cannot be negative", ErrInvalidConfig)
	}
	if c.CleanupInterval < 0 {
		return fmt.Errorf("%w: CleanupInterval cannot be negative", ErrInvalidConfig)
	}
	return nil
}

// MemoryCache is an in-memory cache implementation using sync.Map.
type MemoryCache struct {
	store    sync.Map
	config   Config
	metrics  *MetricsTracker
	stopChan chan struct{}
	wg       sync.WaitGroup
}

type cacheItem struct {
	value      []byte
	expiration time.Time
	key        string // Store key for cleanup
}

// MetricsTracker collects cache metrics.
type MetricsTracker struct {
	Hits          uint64
	Misses        uint64
	Sets          uint64
	Deletes       uint64
	Clears        uint64
	Expired       uint64
	CurrentSize   int
	CurrentMemory uint64
	mu            sync.RWMutex
}

// NewMetricsTracker creates a new cache metrics tracker.
func NewMetricsTracker() *MetricsTracker {
	return &MetricsTracker{}
}

// GetStats returns the current metrics snapshot.
func (mc *MetricsTracker) GetStats() MetricsSnapshot {
	mc.mu.RLock()
	defer mc.mu.RUnlock()

	return MetricsSnapshot{
		Hits:          mc.Hits,
		Misses:        mc.Misses,
		Sets:          mc.Sets,
		Deletes:       mc.Deletes,
		Clears:        mc.Clears,
		Expired:       mc.Expired,
		CurrentSize:   mc.CurrentSize,
		CurrentMemory: mc.CurrentMemory,
	}
}

// MetricsSnapshot represents a snapshot of cache metrics.
type MetricsSnapshot struct {
	Hits          uint64 `json:"hits"`
	Misses        uint64 `json:"misses"`
	Sets          uint64 `json:"sets"`
	Deletes       uint64 `json:"deletes"`
	Clears        uint64 `json:"clears"`
	Expired       uint64 `json:"expired"`
	CurrentSize   int    `json:"current_size"`
	CurrentMemory uint64 `json:"current_memory"`
}

// NewMemoryCache creates an empty MemoryCache instance.
func NewMemoryCache() *MemoryCache {
	return NewMemoryCacheWithConfig(DefaultConfig())
}

// NewMemoryCacheWithConfig creates a MemoryCache with custom configuration.
//
// Panics if the configuration is invalid. Call config.Validate() beforehand
// if you need to handle validation errors gracefully.
func NewMemoryCacheWithConfig(config Config) *MemoryCache {
	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("invalid cache config: %v", err))
	}

	cache := &MemoryCache{
		config:   config,
		metrics:  NewMetricsTracker(),
		stopChan: make(chan struct{}),
	}

	if config.CleanupInterval > 0 {
		cache.startCleanup()
	}

	return cache
}

// startCleanup starts the background cleanup goroutine.
func (mc *MemoryCache) startCleanup() {
	mc.wg.Add(1)
	go func() {
		defer mc.wg.Done()

		ticker := time.NewTicker(mc.config.CleanupInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				mc.cleanupExpired()
			case <-mc.stopChan:
				return
			}
		}
	}()
}

// removeExpiredItem removes an expired item and updates metrics.
// Returns true if the item was removed.
func (mc *MemoryCache) removeExpiredItem(key any, item cacheItem) bool {
	if !expired(item.expiration) {
		return false
	}

	mc.store.Delete(key)
	mc.updateMetrics(func(m *MetricsTracker) {
		m.Expired++
		m.CurrentSize--
		if m.CurrentMemory >= uint64(len(item.value)) {
			m.CurrentMemory -= uint64(len(item.value))
		}
	})
	return true
}

// cleanupExpired removes expired items from the cache.
// To avoid O(N) scan on every cleanup cycle, we limit the number of items checked.
func (mc *MemoryCache) cleanupExpired() {
	const maxItemsPerCleanup = 1000 // Check at most 1000 items per cleanup cycle
	checked := 0

	mc.store.Range(func(key, value any) bool {
		// Limit the number of items checked per cycle
		if checked >= maxItemsPerCleanup {
			return false // Stop iteration
		}
		checked++

		item := value.(cacheItem)
		mc.removeExpiredItem(key, item)
		return true
	})
}

// validateKey checks if a key is valid.
func (mc *MemoryCache) validateKey(key string) error {
	if key == "" {
		return fmt.Errorf("%w: key cannot be empty", ErrInvalidConfig)
	}
	if mc.config.MaxKeyLength > 0 && len(key) > mc.config.MaxKeyLength {
		return fmt.Errorf("%w: key length %d exceeds maximum %d", ErrKeyTooLong, len(key), mc.config.MaxKeyLength)
	}

	// Prevent cache key pollution by rejecting keys with control characters
	// These characters could be used to bypass tenant isolation or manipulate logging
	for i := 0; i < len(key); i++ {
		c := key[i]
		// Reject ASCII control characters (0x00-0x1F, 0x7F)
		// and newlines which could pollute logs or break key formatting
		if c < 0x20 || c == 0x7F {
			return fmt.Errorf("%w: key contains invalid control character at position %d", ErrInvalidConfig, i)
		}
	}

	return nil
}

// checkMemoryLimit checks if adding the value would exceed memory limit.
func (mc *MemoryCache) checkMemoryLimit(valueSize uint64) error {
	if mc.config.MaxMemoryUsage == 0 {
		return nil
	}

	mc.metrics.mu.RLock()
	currentMemory := mc.metrics.CurrentMemory
	mc.metrics.mu.RUnlock()

	if currentMemory+valueSize > mc.config.MaxMemoryUsage {
		return fmt.Errorf("%w: memory limit exceeded (current: %d, adding: %d, limit: %d)",
			ErrCacheFull, currentMemory, valueSize, mc.config.MaxMemoryUsage)
	}

	return nil
}

// updateMetrics updates metrics with a write lock.
func (mc *MemoryCache) updateMetrics(fn func(*MetricsTracker)) {
	if !mc.config.EnableMetrics {
		return
	}

	mc.metrics.mu.Lock()
	defer mc.metrics.mu.Unlock()
	fn(mc.metrics)
}

// Get returns the cached value for the provided key if it exists and has not expired.
func (mc *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	if err := mc.validateKey(key); err != nil {
		return nil, err
	}

	val, ok := mc.store.Load(key)
	if !ok {
		mc.updateMetrics(func(m *MetricsTracker) {
			m.Misses++
		})
		return nil, ErrNotFound
	}

	item := val.(cacheItem)
	if mc.removeExpiredItem(key, item) {
		return nil, ErrNotFound
	}

	mc.updateMetrics(func(m *MetricsTracker) {
		m.Hits++
	})

	return cloneBytes(item.value), nil
}

// Set stores a value with the specified TTL. A zero or negative TTL stores the value indefinitely.
func (mc *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := mc.validateKey(key); err != nil {
		return err
	}

	valueSize := uint64(len(value))
	if err := mc.checkMemoryLimit(valueSize); err != nil {
		return err
	}

	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	} else if mc.config.DefaultTTL > 0 {
		exp = time.Now().Add(mc.config.DefaultTTL)
	}

	// Check if key already exists and update memory usage
	existingSize := uint64(0)
	if existing, ok := mc.store.Load(key); ok {
		existingItem := existing.(cacheItem)
		existingSize = uint64(len(existingItem.value))
	}

	mc.store.Store(key, cacheItem{
		value:      cloneBytes(value),
		expiration: exp,
		key:        key,
	})

	mc.updateMetrics(func(m *MetricsTracker) {
		m.Sets++
		// Only increment size if this is a new key
		if existingSize == 0 {
			m.CurrentSize++
		}
		// Update memory: subtract old size, add new size
		if m.CurrentMemory >= existingSize {
			m.CurrentMemory -= existingSize
		}
		m.CurrentMemory += valueSize
	})

	return nil
}

// Delete removes the key from the cache.
func (mc *MemoryCache) Delete(ctx context.Context, key string) error {
	if err := mc.validateKey(key); err != nil {
		return err
	}

	if existing, ok := mc.store.Load(key); ok {
		item := existing.(cacheItem)
		mc.store.Delete(key)
		mc.updateMetrics(func(m *MetricsTracker) {
			m.Deletes++
			m.CurrentSize--
			if m.CurrentMemory >= uint64(len(item.value)) {
				m.CurrentMemory -= uint64(len(item.value))
			}
		})
	}

	return nil
}

// Exists reports whether a key exists and has not expired.
func (mc *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	if err := mc.validateKey(key); err != nil {
		return false, err
	}

	val, ok := mc.store.Load(key)
	if !ok {
		return false, nil
	}

	item := val.(cacheItem)
	if mc.removeExpiredItem(key, item) {
		return false, nil
	}

	return true, nil
}

// Clear removes all keys from the cache.
func (mc *MemoryCache) Clear(ctx context.Context) error {
	mc.store.Range(func(key, value any) bool {
		item := value.(cacheItem)
		mc.store.Delete(key)
		mc.updateMetrics(func(m *MetricsTracker) {
			m.CurrentSize--
			if m.CurrentMemory >= uint64(len(item.value)) {
				m.CurrentMemory -= uint64(len(item.value))
			}
		})
		return true
	})

	mc.updateMetrics(func(m *MetricsTracker) {
		m.Clears++
	})

	return nil
}

// Incr increments the integer value of a key by delta.
func (mc *MemoryCache) Incr(ctx context.Context, key string, delta int64) (int64, error) {
	if err := mc.validateKey(key); err != nil {
		return 0, err
	}

	// Get current value
	var currentVal int64
	if val, ok := mc.store.Load(key); ok {
		item := val.(cacheItem)
		if !expired(item.expiration) {
			// Try to parse as int64
			if len(item.value) > 0 {
				var num int64
				buf := bytes.NewReader(item.value)
				if err := gob.NewDecoder(buf).Decode(&num); err != nil {
					return 0, ErrNotInteger
				}
				currentVal = num
			}
		}
	}

	// Calculate new value
	newVal := currentVal + delta

	// Encode new value
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(newVal); err != nil {
		return 0, err
	}

	// Store new value
	if err := mc.Set(ctx, key, buf.Bytes(), 0); err != nil {
		return 0, err
	}

	return newVal, nil
}

// Decr decrements the integer value of a key by delta.
func (mc *MemoryCache) Decr(ctx context.Context, key string, delta int64) (int64, error) {
	return mc.Incr(ctx, key, -delta)
}

// Append appends data to the end of an existing value.
func (mc *MemoryCache) Append(ctx context.Context, key string, data []byte) error {
	if err := mc.validateKey(key); err != nil {
		return err
	}

	// Get existing value
	var existingData []byte
	if val, ok := mc.store.Load(key); ok {
		item := val.(cacheItem)
		if !expired(item.expiration) {
			existingData = item.value
		}
	}

	// Append new data
	newData := append(cloneBytes(existingData), data...)

	// Store new value
	return mc.Set(ctx, key, newData, 0)
}

// GetMetrics returns the current cache metrics.
func (mc *MemoryCache) GetMetrics() MetricsSnapshot {
	return mc.metrics.GetStats()
}

// Close stops the background cleanup goroutine.
func (mc *MemoryCache) Close() error {
	close(mc.stopChan)
	mc.wg.Wait()
	return nil
}

func expired(exp time.Time) bool {
	return !exp.IsZero() && time.Now().After(exp)
}

func cloneBytes(in []byte) []byte {
	if len(in) == 0 {
		return nil
	}
	out := make([]byte, len(in))
	copy(out, in)
	return out
}
