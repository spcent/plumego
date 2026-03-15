package redis

import (
	"bytes"
	"context"
	"encoding/gob"
	"errors"
	"fmt"
	"time"

	"github.com/spcent/plumego/store/cache"
)

var (
	ErrClearUnsupported = errors.New("redis cache: clear unsupported")
	ErrNilClient        = errors.New("redis cache: client is nil")
)

const (
	// DefaultMaxKeyLength is the default maximum key length.
	DefaultMaxKeyLength = 256
)

// Client captures the minimal Redis operations required by the adapter.
type Client interface {
	Get(ctx context.Context, key string) ([]byte, error)
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) (int64, error)
	Exists(ctx context.Context, keys ...string) (int64, error)
}

// Flusher allows the adapter to clear all keys.
type Flusher interface {
	FlushDB(ctx context.Context) error
}

// Adapter implements cache.Cache using a Redis client.
type Adapter struct {
	Client       Client
	IsNotFound   func(error) bool
	MaxKeyLength int
}

// NewAdapter wraps a Redis client in a cache.Cache adapter.
func NewAdapter(client Client, isNotFound func(error) bool) *Adapter {
	return &Adapter{
		Client:       client,
		IsNotFound:   isNotFound,
		MaxKeyLength: DefaultMaxKeyLength,
	}
}

// validateKey checks if a key is valid and safe.
// This prevents cache key pollution, tenant isolation bypass, and injection attacks.
func (a *Adapter) validateKey(key string) error {
	if key == "" {
		return fmt.Errorf("redis cache: key cannot be empty")
	}

	if a.MaxKeyLength > 0 && len(key) > a.MaxKeyLength {
		return fmt.Errorf("%w: key length %d exceeds maximum %d",
			cache.ErrKeyTooLong, len(key), a.MaxKeyLength)
	}

	// Prevent cache key pollution by rejecting keys with control characters
	// These characters could be used to bypass tenant isolation or manipulate logging
	for i := 0; i < len(key); i++ {
		c := key[i]
		// Reject ASCII control characters (0x00-0x1F, 0x7F)
		// and newlines which could pollute logs or break key formatting
		if c < 0x20 || c == 0x7F {
			return fmt.Errorf("redis cache: key contains invalid control character at position %d", i)
		}
	}

	return nil
}

// Get returns the cached value for the provided key.
func (a *Adapter) Get(ctx context.Context, key string) ([]byte, error) {
	if a == nil || a.Client == nil {
		return nil, ErrNilClient
	}

	if err := a.validateKey(key); err != nil {
		return nil, err
	}

	value, err := a.Client.Get(ctx, key)
	if err != nil {
		if a.IsNotFound != nil && a.IsNotFound(err) {
			return nil, cache.ErrNotFound
		}
		return nil, err
	}

	return value, nil
}

// Set stores a value with the specified TTL.
func (a *Adapter) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if a == nil || a.Client == nil {
		return ErrNilClient
	}

	if err := a.validateKey(key); err != nil {
		return err
	}

	return a.Client.Set(ctx, key, value, ttl)
}

// Delete removes the key from Redis.
func (a *Adapter) Delete(ctx context.Context, key string) error {
	if a == nil || a.Client == nil {
		return ErrNilClient
	}

	if err := a.validateKey(key); err != nil {
		return err
	}

	_, err := a.Client.Del(ctx, key)
	return err
}

// Exists reports whether a key exists.
func (a *Adapter) Exists(ctx context.Context, key string) (bool, error) {
	if a == nil || a.Client == nil {
		return false, ErrNilClient
	}

	if err := a.validateKey(key); err != nil {
		return false, err
	}

	count, err := a.Client.Exists(ctx, key)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

// Clear removes all keys if the client supports FlushDB.
func (a *Adapter) Clear(ctx context.Context) error {
	if a == nil || a.Client == nil {
		return ErrNilClient
	}
	flusher, ok := a.Client.(Flusher)
	if !ok {
		return ErrClearUnsupported
	}
	return flusher.FlushDB(ctx)
}

// Incr atomically increments the integer value of a key by delta.
// Returns the new value after increment.
// If the key doesn't exist, it's created with delta as the initial value.
// Returns cache.ErrNotInteger if the value is not an integer.
func (a *Adapter) Incr(ctx context.Context, key string, delta int64) (int64, error) {
	if a == nil || a.Client == nil {
		return 0, ErrNilClient
	}

	if err := a.validateKey(key); err != nil {
		return 0, err
	}

	// Get current value
	var currentVal int64
	if data, err := a.Client.Get(ctx, key); err == nil && len(data) > 0 {
		// Try to parse as int64
		buf := bytes.NewReader(data)
		if err := gob.NewDecoder(buf).Decode(&currentVal); err != nil {
			return 0, cache.ErrNotInteger
		}
	} else if err != nil && (a.IsNotFound == nil || !a.IsNotFound(err)) {
		return 0, err
	}

	// Calculate new value
	newVal := currentVal + delta

	// Encode new value
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(newVal); err != nil {
		return 0, err
	}

	// Store new value (use zero TTL to keep existing TTL)
	if err := a.Client.Set(ctx, key, buf.Bytes(), 0); err != nil {
		return 0, err
	}

	return newVal, nil
}

// Decr atomically decrements the integer value of a key by delta.
// Returns the new value after decrement.
// If the key doesn't exist, it's created with -delta as the initial value.
// Returns cache.ErrNotInteger if the value is not an integer.
func (a *Adapter) Decr(ctx context.Context, key string, delta int64) (int64, error) {
	return a.Incr(ctx, key, -delta)
}

// Append appends data to the end of an existing value.
// If the key doesn't exist, it's created with the data as the value.
func (a *Adapter) Append(ctx context.Context, key string, data []byte) error {
	if a == nil || a.Client == nil {
		return ErrNilClient
	}

	if err := a.validateKey(key); err != nil {
		return err
	}

	// Get existing value
	var existingData []byte
	if val, err := a.Client.Get(ctx, key); err == nil {
		existingData = val
	} else if err != nil && (a.IsNotFound == nil || !a.IsNotFound(err)) {
		return err
	}

	// Append new data
	newData := append(existingData, data...)

	// Store new value (use zero TTL to keep existing TTL)
	return a.Client.Set(ctx, key, newData, 0)
}
