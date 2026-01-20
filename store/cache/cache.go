package cache

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
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
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Clear(ctx context.Context) error
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
	metrics  *MetricsCollector
	stopChan chan struct{}
	wg       sync.WaitGroup
}

type cacheItem struct {
	value      []byte
	expiration time.Time
	key        string // Store key for cleanup
}

// MetricsCollector collects cache metrics.
type MetricsCollector struct {
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

// NewMetricsCollector creates a new metrics collector.
func NewMetricsCollector() *MetricsCollector {
	return &MetricsCollector{}
}

// GetStats returns the current metrics snapshot.
func (mc *MetricsCollector) GetStats() MetricsSnapshot {
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
func NewMemoryCacheWithConfig(config Config) *MemoryCache {
	if err := config.Validate(); err != nil {
		panic(fmt.Sprintf("invalid cache config: %v", err))
	}

	cache := &MemoryCache{
		config:   config,
		metrics:  NewMetricsCollector(),
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

// cleanupExpired removes expired items from the cache.
func (mc *MemoryCache) cleanupExpired() {
	mc.store.Range(func(key, value any) bool {
		item := value.(cacheItem)
		if expired(item.expiration) {
			mc.store.Delete(key)
			mc.updateMetrics(func(m *MetricsCollector) {
				m.Expired++
				m.CurrentSize--
				if m.CurrentMemory >= uint64(len(item.value)) {
					m.CurrentMemory -= uint64(len(item.value))
				}
			})
		}
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
func (mc *MemoryCache) updateMetrics(fn func(*MetricsCollector)) {
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
		mc.updateMetrics(func(m *MetricsCollector) {
			m.Misses++
		})
		return nil, ErrNotFound
	}

	item := val.(cacheItem)
	if expired(item.expiration) {
		mc.store.Delete(key)
		mc.updateMetrics(func(m *MetricsCollector) {
			m.Expired++
			m.CurrentSize--
			if m.CurrentMemory >= uint64(len(item.value)) {
				m.CurrentMemory -= uint64(len(item.value))
			}
		})
		return nil, ErrNotFound
	}

	mc.updateMetrics(func(m *MetricsCollector) {
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

	mc.updateMetrics(func(m *MetricsCollector) {
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
		mc.updateMetrics(func(m *MetricsCollector) {
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
	if expired(item.expiration) {
		mc.store.Delete(key)
		mc.updateMetrics(func(m *MetricsCollector) {
			m.Expired++
			m.CurrentSize--
			if m.CurrentMemory >= uint64(len(item.value)) {
				m.CurrentMemory -= uint64(len(item.value))
			}
		})
		return false, nil
	}

	return true, nil
}

// Clear removes all keys from the cache.
func (mc *MemoryCache) Clear(ctx context.Context) error {
	mc.store.Range(func(key, value any) bool {
		item := value.(cacheItem)
		mc.store.Delete(key)
		mc.updateMetrics(func(m *MetricsCollector) {
			m.CurrentSize--
			if m.CurrentMemory >= uint64(len(item.value)) {
				m.CurrentMemory -= uint64(len(item.value))
			}
		})
		return true
	})

	mc.updateMetrics(func(m *MetricsCollector) {
		m.Clears++
	})

	return nil
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

// Cached decorates an http.Handler with cache read/write logic.
func Cached(c Cache, ttl time.Duration, keyFn func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFn(r)

			if data, err := c.Get(r.Context(), key); err == nil {
				resp, err := decodeCachedResponse(data)
				if err == nil {
					writeCachedResponse(w, resp)
					return
				}
				_ = c.Delete(r.Context(), key)
			}

			recorder := httptest.NewRecorder()
			next.ServeHTTP(recorder, r)

			copyHeaders(w.Header(), recorder.Header())
			w.Header().Set("X-Cache", "MISS")
			w.WriteHeader(recorder.Code)
			_, _ = w.Write(recorder.Body.Bytes())

			if recorder.Code == http.StatusOK {
				resp := cachedResponse{
					Status: recorder.Code,
					Header: cloneHeader(recorder.Header()),
					Body:   recorder.Body.Bytes(),
				}
				if encoded, err := encodeCachedResponse(resp); err == nil {
					_ = c.Set(r.Context(), key, encoded, ttl)
				}
			}
		})
	}
}

// CachedWithConfig decorates an http.Handler with configurable cache logic.
func CachedWithConfig(c Cache, config Config, keyFn func(*http.Request) string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := keyFn(r)

			// Try to get from cache
			if data, err := c.Get(r.Context(), key); err == nil {
				resp, err := decodeCachedResponse(data)
				if err == nil {
					writeCachedResponse(w, resp)
					return
				}
				// Cache is corrupted, delete it
				_ = c.Delete(r.Context(), key)
			}

			// Cache miss or corrupted, call the handler
			recorder := httptest.NewRecorder()
			next.ServeHTTP(recorder, r)

			copyHeaders(w.Header(), recorder.Header())
			w.Header().Set("X-Cache", "MISS")
			w.WriteHeader(recorder.Code)
			_, _ = w.Write(recorder.Body.Bytes())

			// Only cache successful responses
			if recorder.Code >= 200 && recorder.Code < 300 {
				resp := cachedResponse{
					Status: recorder.Code,
					Header: cloneHeader(recorder.Header()),
					Body:   recorder.Body.Bytes(),
				}
				if encoded, err := encodeCachedResponse(resp); err == nil {
					ttl := config.DefaultTTL
					if ttl == 0 {
						ttl = 10 * time.Minute
					}
					_ = c.Set(r.Context(), key, encoded, ttl)
				}
			}
		})
	}
}

// KeyFromRequest is a helper function to generate cache keys from requests.
func KeyFromRequest(r *http.Request) string {
	var builder strings.Builder
	builder.WriteString(r.Method)
	builder.WriteString(":")
	builder.WriteString(r.URL.Path)

	if query := r.URL.RawQuery; query != "" {
		builder.WriteString("?")
		builder.WriteString(query)
	}

	return builder.String()
}

// KeyFromRequestWithHeaders is a helper function to generate cache keys including specific headers.
func KeyFromRequestWithHeaders(r *http.Request, headers ...string) string {
	var builder strings.Builder
	builder.WriteString(r.Method)
	builder.WriteString(":")
	builder.WriteString(r.URL.Path)

	if query := r.URL.RawQuery; query != "" {
		builder.WriteString("?")
		builder.WriteString(query)
	}

	for _, header := range headers {
		if value := r.Header.Get(header); value != "" {
			builder.WriteString(":")
			builder.WriteString(header)
			builder.WriteString("=")
			builder.WriteString(value)
		}
	}

	return builder.String()
}

type cachedResponse struct {
	Status int
	Header http.Header
	Body   []byte
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

func copyHeaders(dst, src http.Header) {
	for k, v := range src {
		dst[k] = append([]string(nil), v...)
	}
}

func cloneHeader(h http.Header) http.Header {
	cloned := make(http.Header, len(h))
	copyHeaders(cloned, h)
	return cloned
}

func encodeCachedResponse(resp cachedResponse) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(resp); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func decodeCachedResponse(data []byte) (cachedResponse, error) {
	var resp cachedResponse
	if err := gob.NewDecoder(bytes.NewReader(data)).Decode(&resp); err != nil {
		return cachedResponse{}, err
	}
	return resp, nil
}

func writeCachedResponse(w http.ResponseWriter, resp cachedResponse) {
	copyHeaders(w.Header(), resp.Header)
	w.Header().Set("X-Cache", "HIT")
	status := resp.Status
	if status == 0 {
		status = http.StatusOK
	}
	w.WriteHeader(status)
	_, _ = w.Write(resp.Body)
}
