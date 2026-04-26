// Package cache provides a small in-memory cache primitive with TTL-aware and
// memory-bounded eviction.
//
// This package implements an in-process cache primitive with features including:
//   - TTL (time-to-live) expiration for entries
//   - memory-bounded eviction
//   - thread-safe operations
//
// Example usage:
//
//	import (
//		"context"
//		"time"
//
//		"github.com/spcent/plumego/store/cache"
//	)
//
//	func example() error {
//		ctx := context.Background()
//		userData := []byte("profile")
//
//		c, err := cache.NewMemoryCacheWithConfig(cache.Config{
//			MaxMemoryUsage: 10 << 20, // 10 MiB
//			DefaultTTL:     5 * time.Minute,
//		})
//		if err != nil {
//			return err
//		}
//		defer c.Close()
//
//		if err := c.Set(ctx, "user:123", userData, 0); err != nil {
//			return err
//		}
//		if val, err := c.Get(ctx, "user:123"); err == nil {
//			_ = val
//		}
//		return c.Delete(ctx, "user:123")
//	}
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

	// ErrCacheMiss is retained as a compatibility name for ErrNotFound.
	ErrCacheMiss = ErrNotFound

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

	// DefaultTTL is the default time-to-live for items without explicit TTL.
	DefaultTTL time.Duration
}

// DefaultConfig returns a default configuration.
func DefaultConfig() Config {
	return Config{
		MaxKeyLength:    DefaultMaxKeyLength,
		MaxMemoryUsage:  DefaultMaxMemoryUsage,
		CleanupInterval: DefaultCleanupInterval,
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
	if c.DefaultTTL < 0 {
		return fmt.Errorf("%w: DefaultTTL cannot be negative", ErrInvalidConfig)
	}
	return nil
}

// MemoryCache is an in-memory cache implementation using sync.Map.
type MemoryCache struct {
	store     sync.Map
	config    Config
	writeMu   sync.Mutex
	stateMu   sync.RWMutex
	size      int
	memory    uint64
	stopChan  chan struct{}
	closeOnce sync.Once
	wg        sync.WaitGroup
}

type cacheItem struct {
	value      []byte
	expiration time.Time
}

// NewMemoryCache creates an empty MemoryCache instance.
func NewMemoryCache() *MemoryCache {
	cache, err := NewMemoryCacheWithConfig(DefaultConfig())
	if err != nil {
		return nil
	}
	return cache
}

// NewMemoryCacheWithConfig creates a MemoryCache with custom configuration.
func NewMemoryCacheWithConfig(config Config) (*MemoryCache, error) {
	if err := config.Validate(); err != nil {
		return nil, err
	}

	cache := &MemoryCache{
		config:   config,
		stopChan: make(chan struct{}),
	}

	if config.CleanupInterval > 0 {
		cache.startCleanup()
	}

	return cache, nil
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

// removeExpiredItem removes an expired item and updates internal state.
// Returns true if the item was removed.
func (mc *MemoryCache) removeExpiredItem(key any, item cacheItem) bool {
	if !expired(item.expiration) {
		return false
	}

	mc.writeMu.Lock()
	defer mc.writeMu.Unlock()
	return mc.removeExpiredItemLocked(key, item)
}

func (mc *MemoryCache) removeExpiredItemLocked(key any, item cacheItem) bool {
	current, ok := mc.store.Load(key)
	if !ok {
		return false
	}
	currentItem := current.(cacheItem)
	if currentItem.expiration != item.expiration || !bytes.Equal(currentItem.value, item.value) {
		return false
	}

	mc.store.Delete(key)
	mc.adjustStoredValue(-1, -int64(len(item.value)))
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

// validateKey checks if a key is valid for stable in-process cache storage.
func (mc *MemoryCache) validateKey(key string) error {
	if key == "" {
		return fmt.Errorf("%w: key cannot be empty", ErrInvalidConfig)
	}
	if mc.config.MaxKeyLength > 0 && len(key) > mc.config.MaxKeyLength {
		return fmt.Errorf("%w: key length %d exceeds maximum %d", ErrKeyTooLong, len(key), mc.config.MaxKeyLength)
	}

	// Reject control characters so keys remain safe for logs and serializers.
	for i := 0; i < len(key); i++ {
		c := key[i]
		if c < 0x20 || c == 0x7F {
			return fmt.Errorf("%w: key contains invalid control character at position %d", ErrInvalidConfig, i)
		}
	}

	return nil
}

// checkMemoryLimit checks if adding the value would exceed memory limit.
func (mc *MemoryCache) checkMemoryLimit(valueSize, existingSize uint64) error {
	if mc.config.MaxMemoryUsage == 0 {
		return nil
	}

	currentMemory := mc.currentMemoryUsage()
	if currentMemory >= existingSize {
		currentMemory -= existingSize
	}

	if currentMemory+valueSize > mc.config.MaxMemoryUsage {
		return fmt.Errorf("%w: memory limit exceeded (current: %d, adding: %d, limit: %d)",
			ErrCacheFull, currentMemory, valueSize, mc.config.MaxMemoryUsage)
	}

	return nil
}

// Get returns the cached value for the provided key if it exists and has not expired.
func (mc *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	if err := contextErr(ctx); err != nil {
		return nil, err
	}
	if err := mc.validateKey(key); err != nil {
		return nil, err
	}

	val, ok := mc.store.Load(key)
	if !ok {
		return nil, ErrNotFound
	}

	item := val.(cacheItem)
	if mc.removeExpiredItem(key, item) {
		return nil, ErrNotFound
	}

	return cloneBytes(item.value), nil
}

// Set stores a value with the specified TTL. A non-positive TTL uses DefaultTTL
// when configured; otherwise the value is stored without an expiration.
func (mc *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	if err := mc.validateKey(key); err != nil {
		return err
	}

	mc.writeMu.Lock()
	defer mc.writeMu.Unlock()
	return mc.setLocked(key, value, ttl)
}

func (mc *MemoryCache) setLocked(key string, value []byte, ttl time.Duration) error {
	return mc.setLockedWithExpiration(key, value, mc.expirationForTTL(ttl, time.Now()))
}

func (mc *MemoryCache) setLockedWithExpiration(key string, value []byte, exp time.Time) error {
	existingSize := uint64(0)
	existingFound := false
	if existing, ok := mc.store.Load(key); ok {
		existingItem := existing.(cacheItem)
		if expired(existingItem.expiration) {
			mc.removeExpiredItemLocked(key, existingItem)
		} else {
			existingSize = uint64(len(existingItem.value))
			existingFound = true
		}
	}

	valueSize := uint64(len(value))
	if err := mc.checkMemoryLimit(valueSize, existingSize); err != nil {
		return err
	}

	mc.store.Store(key, cacheItem{
		value:      cloneBytes(value),
		expiration: exp,
	})

	deltaSize := 0
	if !existingFound {
		deltaSize = 1
	}
	mc.adjustStoredValue(deltaSize, int64(valueSize)-int64(existingSize))

	return nil
}

func (mc *MemoryCache) expirationForTTL(ttl time.Duration, now time.Time) time.Time {
	if ttl > 0 {
		return now.Add(ttl)
	}
	if mc.config.DefaultTTL > 0 {
		return now.Add(mc.config.DefaultTTL)
	}
	return time.Time{}
}

// Delete removes the key from the cache.
func (mc *MemoryCache) Delete(ctx context.Context, key string) error {
	if err := contextErr(ctx); err != nil {
		return err
	}
	if err := mc.validateKey(key); err != nil {
		return err
	}

	mc.writeMu.Lock()
	defer mc.writeMu.Unlock()
	if existing, ok := mc.store.Load(key); ok {
		item := existing.(cacheItem)
		mc.store.Delete(key)
		mc.adjustStoredValue(-1, -int64(len(item.value)))
	}

	return nil
}

// Exists reports whether a key exists and has not expired.
func (mc *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	if err := contextErr(ctx); err != nil {
		return false, err
	}
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
	if err := contextErr(ctx); err != nil {
		return err
	}
	mc.writeMu.Lock()
	defer mc.writeMu.Unlock()
	mc.store.Range(func(key, value any) bool {
		item := value.(cacheItem)
		mc.store.Delete(key)
		mc.adjustStoredValue(-1, -int64(len(item.value)))
		return true
	})

	return nil
}

// Incr increments the integer value of a key by delta.
func (mc *MemoryCache) Incr(ctx context.Context, key string, delta int64) (int64, error) {
	if err := contextErr(ctx); err != nil {
		return 0, err
	}
	if err := mc.validateKey(key); err != nil {
		return 0, err
	}

	mc.writeMu.Lock()
	defer mc.writeMu.Unlock()

	// Get current value
	var currentVal int64
	exp := mc.expirationForTTL(0, time.Now())
	if val, ok := mc.store.Load(key); ok {
		item := val.(cacheItem)
		if expired(item.expiration) {
			mc.removeExpiredItemLocked(key, item)
		} else {
			exp = item.expiration
			// Try to parse as int64
			if len(item.value) > 0 {
				num, err := decodeInt64(item.value)
				if err != nil {
					return 0, ErrNotInteger
				}
				currentVal = num
			}
		}
	}

	// Calculate new value
	newVal := currentVal + delta

	encoded, err := encodeInt64(newVal)
	if err != nil {
		return 0, err
	}

	// Store new value
	if err := mc.setLockedWithExpiration(key, encoded, exp); err != nil {
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
	if err := contextErr(ctx); err != nil {
		return err
	}
	if err := mc.validateKey(key); err != nil {
		return err
	}

	mc.writeMu.Lock()
	defer mc.writeMu.Unlock()

	// Get existing value
	var existingData []byte
	exp := mc.expirationForTTL(0, time.Now())
	if val, ok := mc.store.Load(key); ok {
		item := val.(cacheItem)
		if expired(item.expiration) {
			mc.removeExpiredItemLocked(key, item)
		} else {
			existingData = item.value
			exp = item.expiration
		}
	}

	// Append new data
	newData := append(cloneBytes(existingData), data...)

	// Store new value
	return mc.setLockedWithExpiration(key, newData, exp)
}

// Close stops the background cleanup goroutine.
func (mc *MemoryCache) Close() error {
	mc.closeOnce.Do(func() {
		close(mc.stopChan)
		mc.wg.Wait()
	})
	return nil
}

func (mc *MemoryCache) currentMemoryUsage() uint64 {
	mc.stateMu.RLock()
	defer mc.stateMu.RUnlock()
	return mc.memory
}

func (mc *MemoryCache) adjustStoredValue(deltaSize int, deltaMemory int64) {
	mc.stateMu.Lock()
	defer mc.stateMu.Unlock()

	mc.size += deltaSize
	if mc.size < 0 {
		mc.size = 0
	}

	if deltaMemory < 0 {
		reduction := uint64(-deltaMemory)
		if mc.memory >= reduction {
			mc.memory -= reduction
		} else {
			mc.memory = 0
		}
		return
	}

	mc.memory += uint64(deltaMemory)
}

func expired(exp time.Time) bool {
	return expiredAt(exp, time.Now())
}

func expiredAt(exp, now time.Time) bool {
	return !exp.IsZero() && !exp.After(now)
}

func cloneBytes(in []byte) []byte {
	if len(in) == 0 {
		return nil
	}
	out := make([]byte, len(in))
	copy(out, in)
	return out
}

func contextErr(ctx context.Context) error {
	if ctx == nil {
		return nil
	}
	return ctx.Err()
}

func decodeInt64(data []byte) (int64, error) {
	var num int64
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&num); err != nil {
		return 0, err
	}
	return num, nil
}

func encodeInt64(num int64) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(num); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
