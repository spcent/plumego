package cache

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"time"
)

var (
	// ErrNotFound is returned when a cache entry is missing or expired.
	ErrNotFound = errors.New("cache: key not found")
)

// Cache defines the minimal contract for cache backends.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Delete(ctx context.Context, key string) error
	Exists(ctx context.Context, key string) (bool, error)
	Clear(ctx context.Context) error
}

// MemoryCache is an in-memory cache implementation using sync.Map.
type MemoryCache struct {
	store sync.Map
}

type cacheItem struct {
	value      []byte
	expiration time.Time
}

// NewMemoryCache creates an empty MemoryCache instance.
func NewMemoryCache() *MemoryCache {
	return &MemoryCache{}
}

// Get returns the cached value for the provided key if it exists and has not expired.
func (mc *MemoryCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, ok := mc.store.Load(key)
	if !ok {
		return nil, ErrNotFound
	}

	item := val.(cacheItem)
	if expired(item.expiration) {
		mc.store.Delete(key)
		return nil, ErrNotFound
	}

	return cloneBytes(item.value), nil
}

// Set stores a value with the specified TTL. A zero or negative TTL stores the value indefinitely.
func (mc *MemoryCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	var exp time.Time
	if ttl > 0 {
		exp = time.Now().Add(ttl)
	}

	mc.store.Store(key, cacheItem{value: cloneBytes(value), expiration: exp})
	return nil
}

// Delete removes the key from the cache.
func (mc *MemoryCache) Delete(ctx context.Context, key string) error {
	mc.store.Delete(key)
	return nil
}

// Exists reports whether a key exists and has not expired.
func (mc *MemoryCache) Exists(ctx context.Context, key string) (bool, error) {
	val, ok := mc.store.Load(key)
	if !ok {
		return false, nil
	}

	item := val.(cacheItem)
	if expired(item.expiration) {
		mc.store.Delete(key)
		return false, nil
	}

	return true, nil
}

// Clear removes all keys from the cache.
func (mc *MemoryCache) Clear(ctx context.Context) error {
	mc.store.Range(func(key, value any) bool {
		mc.store.Delete(key)
		return true
	})
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
