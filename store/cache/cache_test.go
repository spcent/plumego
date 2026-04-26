package cache

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"
)

func mustNewMemoryCacheWithConfig(t *testing.T, config Config) *MemoryCache {
	t.Helper()
	cache, err := NewMemoryCacheWithConfig(config)
	if err != nil {
		t.Fatalf("NewMemoryCacheWithConfig: %v", err)
	}
	return cache
}

func TestMemoryCacheSetAndGet(t *testing.T) {
	cache := NewMemoryCache()

	err := cache.Set(t.Context(), "foo", []byte("bar"), time.Minute)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	val, err := cache.Get(t.Context(), "foo")
	if err != nil {
		t.Fatalf("expected value, got error %v", err)
	}

	if string(val) != "bar" {
		t.Fatalf("expected 'bar', got %q", val)
	}

	val[0] = 'B'
	val2, err := cache.Get(t.Context(), "foo")
	if err != nil {
		t.Fatalf("expected value, got error %v", err)
	}
	if string(val2) != "bar" {
		t.Fatalf("expected original value to remain unchanged, got %q", val2)
	}
}

func TestMemoryCacheExpiration(t *testing.T) {
	cache := NewMemoryCache()

	if err := cache.Set(t.Context(), "temp", []byte("value"), 10*time.Millisecond); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	time.Sleep(20 * time.Millisecond)

	if _, err := cache.Get(t.Context(), "temp"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	exists, err := cache.Exists(t.Context(), "temp")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if exists {
		t.Fatalf("expected key to be expired")
	}
}

func TestMemoryCacheClear(t *testing.T) {
	cache := NewMemoryCache()

	cache.Set(t.Context(), "a", []byte("1"), 0)
	cache.Set(t.Context(), "b", []byte("2"), 0)

	if err := cache.Clear(t.Context()); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if _, err := cache.Get(t.Context(), "a"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}

	exists, _ := cache.Exists(t.Context(), "b")
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
			name:    "negative default ttl",
			config:  Config{DefaultTTL: -1 * time.Second},
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
	if config.DefaultTTL != 10*time.Minute {
		t.Fatalf("expected DefaultTTL 10m, got %v", config.DefaultTTL)
	}
}

func TestMemoryCacheWithConfig(t *testing.T) {
	config := Config{
		MaxKeyLength:    100,
		MaxMemoryUsage:  1024,
		CleanupInterval: 1 * time.Second,
		DefaultTTL:      5 * time.Minute,
	}

	cache := mustNewMemoryCacheWithConfig(t, config)
	defer cache.Close()

	// Test Set with default TTL
	err := cache.Set(t.Context(), "test", []byte("value"), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Test Get
	val, err := cache.Get(t.Context(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "value" {
		t.Fatalf("expected 'value', got %q", val)
	}
}

func TestMemoryCacheDeleteAndMiss(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	// Set some values
	cache.Set(t.Context(), "key1", []byte("value1"), time.Minute)
	cache.Set(t.Context(), "key2", []byte("value2"), time.Minute)

	// Get existing key (hit)
	_, _ = cache.Get(t.Context(), "key1")

	// Get non-existing key (miss)
	_, _ = cache.Get(t.Context(), "nonexistent")

	// Delete a key
	_ = cache.Delete(t.Context(), "key2")

	if _, err := cache.Get(t.Context(), "key2"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected deleted key to be missing, got %v", err)
	}
}

func TestErrCacheMissCompatibility(t *testing.T) {
	if !errors.Is(ErrCacheMiss, ErrNotFound) {
		t.Fatal("ErrCacheMiss should match ErrNotFound")
	}
	if !errors.Is(ErrNotFound, ErrCacheMiss) {
		t.Fatal("ErrNotFound should match ErrCacheMiss")
	}
}

func TestMemoryCacheMemoryLimit(t *testing.T) {
	config := Config{
		MaxKeyLength:    100,
		MaxMemoryUsage:  10, // Very small limit
		CleanupInterval: 0,
		DefaultTTL:      1 * time.Minute,
	}

	cache := mustNewMemoryCacheWithConfig(t, config)
	defer cache.Close()

	// First set should succeed
	err := cache.Set(t.Context(), "key1", []byte("small"), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Second set should fail due to memory limit
	err = cache.Set(t.Context(), "key2", []byte("larger value that exceeds limit"), 0)
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
		DefaultTTL:      1 * time.Minute,
	}

	cache := mustNewMemoryCacheWithConfig(t, config)
	defer cache.Close()

	// Test empty key
	err := cache.Set(t.Context(), "", []byte("value"), 0)
	if err == nil {
		t.Fatal("expected error for empty key")
	}
	if !errors.Is(err, ErrInvalidConfig) {
		t.Fatalf("expected ErrInvalidConfig, got %v", err)
	}

	// Test key too long
	longKey := "this_is_a_very_long_key_that_exceeds_the_limit"
	err = cache.Set(t.Context(), longKey, []byte("value"), 0)
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

	cache := mustNewMemoryCacheWithConfig(t, config)
	defer cache.Close()

	// Set with zero TTL should use default TTL
	err := cache.Set(t.Context(), "test", []byte("value"), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should be available immediately
	val, err := cache.Get(t.Context(), "test")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "value" {
		t.Fatalf("expected 'value', got %q", val)
	}

	// Wait for default TTL to expire
	time.Sleep(100 * time.Millisecond)

	// Should be expired now
	_, err = cache.Get(t.Context(), "test")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound after TTL expiration, got %v", err)
	}
}

func TestMemoryCacheCleanup(t *testing.T) {
	config := Config{
		MaxKeyLength:    100,
		MaxMemoryUsage:  0,
		CleanupInterval: 50 * time.Millisecond,
		DefaultTTL:      1 * time.Minute,
	}

	cache := mustNewMemoryCacheWithConfig(t, config)
	defer cache.Close()

	// Set some items with short TTL
	cache.Set(t.Context(), "exp1", []byte("value1"), 10*time.Millisecond)
	cache.Set(t.Context(), "exp2", []byte("value2"), 10*time.Millisecond)
	cache.Set(t.Context(), "keep", []byte("value3"), 1*time.Minute)

	// Wait for items to expire
	time.Sleep(20 * time.Millisecond)

	// Wait for cleanup to run
	time.Sleep(100 * time.Millisecond)

	if _, err := cache.Get(t.Context(), "exp1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected exp1 to be cleaned up, got %v", err)
	}
	if _, err := cache.Get(t.Context(), "exp2"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected exp2 to be cleaned up, got %v", err)
	}
	if value, err := cache.Get(t.Context(), "keep"); err != nil || string(value) != "value3" {
		t.Fatalf("expected keep to remain available, got value=%q err=%v", value, err)
	}
}

func TestMemoryCacheClose(t *testing.T) {
	config := Config{
		MaxKeyLength:    100,
		MaxMemoryUsage:  0,
		CleanupInterval: 50 * time.Millisecond,
		DefaultTTL:      1 * time.Minute,
	}

	cache := mustNewMemoryCacheWithConfig(t, config)

	// Set some data
	cache.Set(t.Context(), "test", []byte("value"), 1*time.Minute)

	// Close the cache
	err := cache.Close()
	if err != nil {
		t.Fatalf("unexpected error closing cache: %v", err)
	}

	// Try to use cache after close (should still work for basic operations)
	// Note: In a real implementation, you might want to prevent usage after close
}

func TestMemoryCacheCloseIdempotent(t *testing.T) {
	cache := NewMemoryCache()

	if err := cache.Close(); err != nil {
		t.Fatalf("Close: %v", err)
	}
	if err := cache.Close(); err != nil {
		t.Fatalf("expected repeated Close to be nil, got %v", err)
	}
}

func TestMemoryCacheZeroTTL(t *testing.T) {
	config := DefaultConfig()
	config.DefaultTTL = 0
	cache := mustNewMemoryCacheWithConfig(t, config)
	defer cache.Close()

	// Set with zero TTL and no DefaultTTL stores indefinitely.
	err := cache.Set(t.Context(), "permanent", []byte("value"), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Wait a bit
	time.Sleep(50 * time.Millisecond)

	// Should still be available
	val, err := cache.Get(t.Context(), "permanent")
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
	err := cache.Set(t.Context(), "key", []byte("old"), 1*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Update with new value
	err = cache.Set(t.Context(), "key", []byte("new"), 1*time.Minute)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Get updated value
	val, err := cache.Get(t.Context(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "new" {
		t.Fatalf("expected 'new', got %q", val)
	}

	if exists, err := cache.Exists(t.Context(), "key"); err != nil || !exists {
		t.Fatalf("expected updated key to exist, exists=%v err=%v", exists, err)
	}
}

func TestMemoryCacheUpdateExistingEmptyValueKeepsSize(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	if err := cache.Set(t.Context(), "key", nil, 1*time.Minute); err != nil {
		t.Fatalf("Set empty value: %v", err)
	}
	if err := cache.Set(t.Context(), "key", []byte("value"), 1*time.Minute); err != nil {
		t.Fatalf("Set replacement value: %v", err)
	}

	cache.stateMu.RLock()
	got := cache.size
	cache.stateMu.RUnlock()
	if got != 1 {
		t.Fatalf("expected one tracked entry after replacement, got %d", got)
	}
}

func TestMemoryCacheConcurrentAccess(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	// Test concurrent Set operations
	var writers sync.WaitGroup
	for i := 0; i < 100; i++ {
		writers.Add(1)
		go func(i int) {
			defer writers.Done()
			key := fmt.Sprintf("key%d", i)
			value := []byte(fmt.Sprintf("value%d", i))
			cache.Set(t.Context(), key, value, 1*time.Minute)
		}(i)
	}
	writers.Wait()

	// Test concurrent Get operations
	var readers sync.WaitGroup
	for i := 0; i < 100; i++ {
		readers.Add(1)
		go func(i int) {
			defer readers.Done()
			key := fmt.Sprintf("key%d", i)
			cache.Get(t.Context(), key)
		}(i)
	}
	readers.Wait()

	for i := 0; i < 100; i++ {
		key := fmt.Sprintf("key%d", i)
		if _, err := cache.Get(t.Context(), key); err != nil {
			t.Fatalf("expected key %q to be readable after concurrent access, got %v", key, err)
		}
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
			err := cache.Set(t.Context(), tc.key, []byte("value"), 0)
			if (err != nil) != tc.wantErr {
				t.Fatalf("Set() error = %v, wantErr %v", err, tc.wantErr)
			}

			// Also test Get, Delete, Exists, Incr, Decr, Append
			if tc.wantErr {
				_, err := cache.Get(t.Context(), tc.key)
				if err == nil {
					t.Fatal("Get() should reject invalid key")
				}

				err = cache.Delete(t.Context(), tc.key)
				if err == nil {
					t.Fatal("Delete() should reject invalid key")
				}

				_, err = cache.Exists(t.Context(), tc.key)
				if err == nil {
					t.Fatal("Exists() should reject invalid key")
				}

				_, err = cache.Incr(t.Context(), tc.key, 1)
				if err == nil {
					t.Fatal("Incr() should reject invalid key")
				}

				_, err = cache.Decr(t.Context(), tc.key, 1)
				if err == nil {
					t.Fatal("Decr() should reject invalid key")
				}

				err = cache.Append(t.Context(), tc.key, []byte("data"))
				if err == nil {
					t.Fatal("Append() should reject invalid key")
				}
			}
		})
	}
}

func TestMemoryCacheCanceledContext(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	ctx, cancel := context.WithCancel(t.Context())
	cancel()

	if err := cache.Set(ctx, "key", []byte("value"), 0); !errors.Is(err, context.Canceled) {
		t.Fatalf("Set should return context.Canceled, got %v", err)
	}
	if _, err := cache.Get(ctx, "key"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Get should return context.Canceled, got %v", err)
	}
	if err := cache.Delete(ctx, "key"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Delete should return context.Canceled, got %v", err)
	}
	if _, err := cache.Exists(ctx, "key"); !errors.Is(err, context.Canceled) {
		t.Fatalf("Exists should return context.Canceled, got %v", err)
	}
	if err := cache.Clear(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("Clear should return context.Canceled, got %v", err)
	}
	if _, err := cache.Incr(ctx, "key", 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("Incr should return context.Canceled, got %v", err)
	}
	if _, err := cache.Decr(ctx, "key", 1); !errors.Is(err, context.Canceled) {
		t.Fatalf("Decr should return context.Canceled, got %v", err)
	}
	if err := cache.Append(ctx, "key", []byte("value")); !errors.Is(err, context.Canceled) {
		t.Fatalf("Append should return context.Canceled, got %v", err)
	}

	if exists, err := cache.Exists(t.Context(), "key"); err != nil || exists {
		t.Fatalf("canceled Set should not store value, exists=%v err=%v", exists, err)
	}
}

func TestExpiredAtBoundary(t *testing.T) {
	now := time.Now()
	if !expiredAt(now, now) {
		t.Fatal("expected value expiring at current time to be expired")
	}
	if expiredAt(now.Add(time.Nanosecond), now) {
		t.Fatal("expected future expiration to remain active")
	}
	if expiredAt(time.Time{}, now) {
		t.Fatal("expected zero expiration to remain active")
	}
}

func TestMemoryCacheIncr(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	// Test increment on non-existent key
	val1, err := cache.Incr(t.Context(), "counter", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val1 != 5 {
		t.Fatalf("expected 5, got %d", val1)
	}

	// Test increment on existing key
	val2, err := cache.Incr(t.Context(), "counter", 3)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val2 != 8 {
		t.Fatalf("expected 8, got %d", val2)
	}

	// Test increment by negative number
	val3, err := cache.Incr(t.Context(), "counter", -2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val3 != 6 {
		t.Fatalf("expected 6, got %d", val3)
	}

	raw, err := cache.Get(t.Context(), "counter")
	if err != nil {
		t.Fatalf("Get counter: %v", err)
	}
	decoded, err := decodeInt64(raw)
	if err != nil {
		t.Fatalf("decode counter: %v", err)
	}
	if decoded != 6 {
		t.Fatalf("expected encoded counter 6, got %d", decoded)
	}
}

func TestMemoryCacheDecr(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	// Test decrement on non-existent key
	val1, err := cache.Decr(t.Context(), "counter", 5)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if val1 != -5 {
		t.Fatalf("expected -5, got %d", val1)
	}

	// Test decrement on existing key
	val2, err := cache.Decr(t.Context(), "counter", 3)
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
	err := cache.Set(t.Context(), "key", []byte("not an integer"), 0)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Try to increment
	_, err = cache.Incr(t.Context(), "key", 1)
	if !errors.Is(err, ErrNotInteger) {
		t.Fatalf("expected ErrNotInteger, got %v", err)
	}
}

func TestMemoryCacheAppend(t *testing.T) {
	cache := NewMemoryCache()
	defer cache.Close()

	// Test append on non-existent key
	err := cache.Append(t.Context(), "key", []byte("hello"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err := cache.Get(t.Context(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "hello" {
		t.Fatalf("expected 'hello', got %q", val)
	}

	// Test append on existing key
	err = cache.Append(t.Context(), "key", []byte(" world"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	val, err = cache.Get(t.Context(), "key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if string(val) != "hello world" {
		t.Fatalf("expected 'hello world', got %q", val)
	}
}
