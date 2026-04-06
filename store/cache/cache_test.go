package cache

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"
)

func TestMemoryCacheSetAndGet(t *testing.T) {
	cache := NewMemoryCache()

	err := cache.Set(context.Background(), "foo", []byte("bar"), time.Minute)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	val, err := cache.Get(context.Background(), "foo")
	if err != nil {
		t.Fatalf("expected value, got error %v", err)
	}

	if string(val) != "bar" {
		t.Fatalf("expected 'bar', got %q", val)
	}

	val[0] = 'B'
	val2, err := cache.Get(context.Background(), "foo")
	if err != nil {
		t.Fatalf("expected value, got error %v", err)
	}
	if string(val2) != "bar" {
		t.Fatalf("expected original value to remain unchanged, got %q", val2)
	}
}

func TestMemoryCacheExpiration(t *testing.T) {
	cache := NewMemoryCache()

	if err := cache.Set(context.Background(), "temp", []byte("value"), 10*time.Millisecond); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	time.Sleep(20 * time.Millisecond)

	if _, err := cache.Get(context.Background(), "temp"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	exists, err := cache.Exists(context.Background(), "temp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Fatalf("expected key to be expired")
	}
}

func TestMemoryCacheClear(t *testing.T) {
	cache := NewMemoryCache()

	cache.Set(context.Background(), "a", []byte("1"), 0)
	cache.Set(context.Background(), "b", []byte("2"), 0)

	if err := cache.Clear(context.Background()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := cache.Get(context.Background(), "a"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	exists, _ := cache.Exists(context.Background(), "b")
	if exists {
		t.Fatalf("expected cache to be cleared")
	}
}

func TestConfigValidate(t *testing.T) {
	cases := []struct {
		name    string
		config  Config
		wantErr bool
	}{
		{
			name:    "negative max key length",
			config:  Config{MaxKeyLength: -1},
			wantErr: true,
		},
		{
			name:    "negative cleanup interval",
			config:  Config{CleanupInterval: -1 * time.Second},
			wantErr: true,
		},
		{
			name:    "valid config",
			config:  Config{MaxKeyLength: 100, CleanupInterval: 1 * time.Second},
			wantErr: false,
		},
	}

	for _, tc := range cases {
		if err := tc.config.Validate(); (err != nil) != tc.wantErr {
			t.Fatalf("%s: Validate() error = %v", tc.name, err)
		}
	}
}

func TestDefaultConfig(t *testing.T) {
	config := DefaultConfig()

	if config.MaxKeyLength != DefaultMaxKeyLength {
		t.Fatalf("expected MaxKeyLength %d, got %d", DefaultMaxKeyLength, config.MaxKeyLength)
	}
	if config.MaxMemoryUsage != DefaultMaxMemoryUsage {
		t.Fatalf("expected MaxMemoryUsage %d, got %d", DefaultMaxMemoryUsage, config.MaxMemoryUsage)
	}
	if config.CleanupInterval != DefaultCleanupInterval {
		t.Fatalf("expected CleanupInterval %v, got %v", DefaultCleanupInterval, config.CleanupInterval)
	}
	if !config.EnableMetrics {
		t.Fatal("expected EnableMetrics to be true")
	}
	if config.DefaultTTL != 10*time.Minute {
		t.Fatalf("expected DefaultTTL 10m, got %v", config.DefaultTTL)
	}
}

func TestMemoryCacheWithConfig(t *testing.T) {
	config := Config{
		MaxKeyLength:    100,
		MaxMemoryUsage:  1024,
		CleanupInterval: 1 * time.Second,
		EnableMetrics:   true,
		DefaultTTL:      5 * time.Minute,
	}

	cache := NewMemoryCacheWithConfig(config)
	defer cache.Close()

	// Test Set with default TTL
	err := cache.Set(context.Background(), "test", []byte("value"), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test Get
	val, err := cache.Get(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "value" {
		t.Fatalf("expected 'value', got %q", val)
	}
}

func TestMemoryCacheMetrics(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	// Set some values
	cache.Set(context.Background(), "key1", []byte("value1"), time.Minute)
	cache.Set(context.Background(), "key2", []byte("value2"), time.Minute)

	// Get existing key (hit)
	_, _ = cache.Get(context.Background(), "key1")

	// Get non-existing key (miss)
	_, _ = cache.Get(context.Background(), "nonexistent")

	// Delete a key
	_ = cache.Delete(context.Background(), "key2")

	// Get metrics
	metrics := cache.GetMetrics()

	if metrics.Sets != 2 {
		t.Fatalf("expected 2 sets, got %d", metrics.Sets)
	}
	if metrics.Hits != 1 {
		t.Fatalf("expected 1 hit, got %d", metrics.Hits)
	}
	if metrics.Misses != 1 {
		t.Fatalf("expected 1 miss, got %d", metrics.Misses)
	}
	if metrics.Deletes != 1 {
		t.Fatalf("expected 1 delete, got %d", metrics.Deletes)
	}
	if metrics.CurrentSize != 1 {
		t.Fatalf("expected 1 current size, got %d", metrics.CurrentSize)
	}
}

func TestMemoryCacheMemoryLimit(t *testing.T) {
	config := Config{
		MaxKeyLength:    100,
		MaxMemoryUsage:  10, // Very small limit
		CleanupInterval: 0,
		EnableMetrics:   false,
		DefaultTTL:      1 * time.Minute,
	}

	cache := NewMemoryCacheWithConfig(config)
	defer cache.Close()

	// First set should succeed
	err := cache.Set(context.Background(), "key1", []byte("small"), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second set should fail due to memory limit
	err = cache.Set(context.Background(), "key2", []byte("larger value that exceeds limit"), 0)
	if err == nil {
		t.Fatal("expected error when exceeding memory limit")
	}
	if !errors.Is(err, ErrCacheFull) {
		t.Fatalf("expected ErrCacheFull, got %v", err)
	}
}

func TestMemoryCacheKeyValidation(t *testing.T) {
	config := Config{
		MaxKeyLength:    10,
		MaxMemoryUsage:  0,
		CleanupInterval: 0,
		EnableMetrics:   false,
		DefaultTTL:      1 * time.Minute,
	}

	cache := NewMemoryCacheWithConfig(config)
	defer cache.Close()

	// Test empty key
	err := cache.Set(context.Background(), "", []byte("value"), 0)
	if err == nil {
		t.Fatal("expected error for empty key")
	}
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}

	// Test key too long
	longKey := "this_is_a_very_long_key_that_exceeds_the_limit"
	err = cache.Set(context.Background(), longKey, []byte("value"), 0)
	if err == nil {
		t.Fatal("expected error for key too long")
	}
	if !errors.Is(err, ErrKeyTooLong) {
		t.Fatalf("expected ErrKeyTooLong, got %v", err)
	}
}

func TestMemoryCacheDefaultTTL(t *testing.T) {
	config := DefaultConfig()
	config.DefaultTTL = 50 * time.Millisecond

	cache := NewMemoryCacheWithConfig(config)
	defer cache.Close()

	// Set with zero TTL should use default TTL
	err := cache.Set(context.Background(), "test", []byte("value"), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be available immediately
	val, err := cache.Get(context.Background(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "value" {
		t.Fatalf("expected 'value', got %q", val)
	}

	// Wait for default TTL to expire
	time.Sleep(100 * time.Millisecond)

	// Should be expired now
	_, err = cache.Get(context.Background(), "test")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after TTL expiration, got %v", err)
	}
}

func TestMemoryCacheCleanup(t *testing.T) {
	config := Config{
		MaxKeyLength:    100,
		MaxMemoryUsage:  0,
		CleanupInterval: 50 * time.Millisecond,
		EnableMetrics:   true,
		DefaultTTL:      1 * time.Minute,
	}

	cache := NewMemoryCacheWithConfig(config)
	defer cache.Close()

	// Set some items with short TTL
	cache.Set(context.Background(), "exp1", []byte("value1"), 10*time.Millisecond)
	cache.Set(context.Background(), "exp2", []byte("value2"), 10*time.Millisecond)
	cache.Set(context.Background(), "keep", []byte("value3"), 1*time.Minute)

	// Wait for items to expire
	time.Sleep(20 * time.Millisecond)

	// Wait for cleanup to run
	time.Sleep(100 * time.Millisecond)

	// Check metrics
	metrics := cache.GetMetrics()
	if metrics.Expired != 2 {
		t.Fatalf("expected 2 expired items, got %d", metrics.Expired)
	}
	if metrics.CurrentSize != 1 {
		t.Fatalf("expected 1 remaining item, got %d", metrics.CurrentSize)
	}
}

func TestMemoryCacheClose(t *testing.T) {
	config := Config{
		MaxKeyLength:    100,
		MaxMemoryUsage:  0,
		CleanupInterval: 50 * time.Millisecond,
		EnableMetrics:   false,
		DefaultTTL:      1 * time.Minute,
	}

	cache := NewMemoryCacheWithConfig(config)

	// Set some data
	cache.Set(context.Background(), "test", []byte("value"), 1*time.Minute)

	// Close the cache
	err := cache.Close()
	if err != nil {
		t.Fatalf("unexpected error closing cache: %v", err)
	}

	// Try to use cache after close (should still work for basic operations)
	// Note: In a real implementation, you might want to prevent usage after close
}

func TestMemoryCacheZeroTTL(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	// Set with zero TTL (should store indefinitely)
	err := cache.Set(context.Background(), "permanent", []byte("value"), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Should still be available
	val, err := cache.Get(context.Background(), "permanent")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "value" {
		t.Fatalf("expected 'value', got %q", val)
	}
}

func TestMemoryCacheUpdateExistingKey(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	// Set initial value
	err := cache.Set(context.Background(), "key", []byte("old"), 1*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Update with new value
	err = cache.Set(context.Background(), "key", []byte("new"), 1*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Get updated value
	val, err := cache.Get(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "new" {
		t.Fatalf("expected 'new', got %q", val)
	}

	// Check metrics
	metrics := cache.GetMetrics()
	if metrics.Sets != 2 {
		t.Fatalf("expected 2 sets, got %d", metrics.Sets)
	}
	if metrics.CurrentSize != 1 {
		t.Fatalf("expected 1 current size, got %d", metrics.CurrentSize)
	}
}

func TestMemoryCacheConcurrentAccess(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	// Test concurrent Set operations
	for i := 0; i < 100; i++ {
		go func(i int) {
			key := fmt.Sprintf("key%d", i)
			value := []byte(fmt.Sprintf("value%d", i))
			cache.Set(context.Background(), key, value, 1*time.Minute)
		}(i)
	}

	time.Sleep(100 * time.Millisecond)

	// Test concurrent Get operations
	for i := 0; i < 100; i++ {
		go func(i int) {
			key := fmt.Sprintf("key%d", i)
			cache.Get(context.Background(), key)
		}(i)
	}

	time.Sleep(100 * time.Millisecond)

	// Verify metrics
	metrics := cache.GetMetrics()
	if metrics.Sets != 100 {
		t.Fatalf("expected 100 sets, got %d", metrics.Sets)
	}
}

func TestMemoryCacheControlCharacterValidation(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	tests := []struct {
		name    string
		key     string
		wantErr bool
	}{
		{
			name:    "valid key",
			key:     "valid:key",
			wantErr: false,
		},
		{
			name:    "key with newline",
			key:     "key\nwith\nnewline",
			wantErr: true,
		},
		{
			name:    "key with tab",
			key:     "key\twith\ttab",
			wantErr: true,
		},
		{
			name:    "key with null byte",
			key:     "key\x00null",
			wantErr: true,
		},
		{
			name:    "key with DEL character",
			key:     "key\x7Fdel",
			wantErr: true,
		},
		{
			name:    "key with carriage return",
			key:     "key\rwith\rCR",
			wantErr: true,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := cache.Set(context.Background(), tc.key, []byte("value"), 0)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Set() error = %v, wantErr %v", err, tc.wantErr)
			}

			// Also test Get, Delete, Exists, Incr, Decr, Append
			if tc.wantErr {
				_, err := cache.Get(context.Background(), tc.key)
				if err == nil {
					t.Fatal("Get() should reject invalid key")
				}

				err = cache.Delete(context.Background(), tc.key)
				if err == nil {
					t.Fatal("Delete() should reject invalid key")
				}

				_, err = cache.Exists(context.Background(), tc.key)
				if err == nil {
					t.Fatal("Exists() should reject invalid key")
				}

				_, err = cache.Incr(context.Background(), tc.key, 1)
				if err == nil {
					t.Fatal("Incr() should reject invalid key")
				}

				_, err = cache.Decr(context.Background(), tc.key, 1)
				if err == nil {
					t.Fatal("Decr() should reject invalid key")
				}

				err = cache.Append(context.Background(), tc.key, []byte("data"))
				if err == nil {
					t.Fatal("Append() should reject invalid key")
				}
			}
		})
	}
}

func TestMemoryCacheIncr(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	// Test increment on non-existent key
	val1, err := cache.Incr(context.Background(), "counter", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val1 != 5 {
		t.Fatalf("expected 5, got %d", val1)
	}

	// Test increment on existing key
	val2, err := cache.Incr(context.Background(), "counter", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val2 != 8 {
		t.Fatalf("expected 8, got %d", val2)
	}

	// Test increment by negative number
	val3, err := cache.Incr(context.Background(), "counter", -2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val3 != 6 {
		t.Fatalf("expected 6, got %d", val3)
	}
}

func TestMemoryCacheDecr(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	// Test decrement on non-existent key
	val1, err := cache.Decr(context.Background(), "counter", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val1 != -5 {
		t.Fatalf("expected -5, got %d", val1)
	}

	// Test decrement on existing key
	val2, err := cache.Decr(context.Background(), "counter", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val2 != -8 {
		t.Fatalf("expected -8, got %d", val2)
	}
}

func TestMemoryCacheIncrNonInteger(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	// Set a non-integer value
	err := cache.Set(context.Background(), "key", []byte("not an integer"), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Try to increment
	_, err = cache.Incr(context.Background(), "key", 1)
	if !errors.Is(err, ErrNotInteger) {
		t.Fatalf("expected ErrNotInteger, got %v", err)
	}
}

func TestMemoryCacheAppend(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	// Test append on non-existent key
	err := cache.Append(context.Background(), "key", []byte("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := cache.Get(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "hello" {
		t.Fatalf("expected 'hello', got %q", val)
	}

	// Test append on existing key
	err = cache.Append(context.Background(), "key", []byte(" world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err = cache.Get(context.Background(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "hello world" {
		t.Fatalf("expected 'hello world', got %q", val)
	}
}
