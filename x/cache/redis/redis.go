package redis

import (
	"context"
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

	minInt64 = -1 << 63
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

// Incrementer allows the adapter to expose atomic counter operations.
type Incrementer interface {
	IncrBy(ctx context.Context, key string, delta int64) (int64, error)
}

// Appender allows the adapter to expose atomic append operations.
type Appender interface {
	Append(ctx context.Context, key string, data []byte) (int64, error)
}

// Adapter implements cache.Cache using a Redis client.
type Adapter struct {
	Client       Client
	IsNotFound   func(error) bool
	MaxKeyLength int
}

// CounterAdapter exposes cache.CounterCache only for clients with Incrementer support.
type CounterAdapter struct {
	*Adapter
	incrementer Incrementer
}

// AppenderAdapter exposes cache.AppenderCache only for clients with Appender support.
type AppenderAdapter struct {
	*Adapter
	appender Appender
}

// AtomicAdapter exposes both cache.CounterCache and cache.AppenderCache for clients with both capabilities.
type AtomicAdapter struct {
	*Adapter
	incrementer Incrementer
	appender    Appender
}

// NewAdapter wraps a Redis client in a cache.Cache adapter.
func NewAdapter(client Client, isNotFound func(error) bool) *Adapter {
	return &Adapter{
		Client:       client,
		IsNotFound:   isNotFound,
		MaxKeyLength: DefaultMaxKeyLength,
	}
}

// NewCounterAdapter wraps a Redis client with atomic counter support.
func NewCounterAdapter(client Client, isNotFound func(error) bool) (*CounterAdapter, error) {
	incrementer, ok := client.(Incrementer)
	if !ok {
		return nil, cache.ErrCapabilityUnsupported
	}
	return &CounterAdapter{
		Adapter:     NewAdapter(client, isNotFound),
		incrementer: incrementer,
	}, nil
}

// NewAppenderAdapter wraps a Redis client with atomic append support.
func NewAppenderAdapter(client Client, isNotFound func(error) bool) (*AppenderAdapter, error) {
	appender, ok := client.(Appender)
	if !ok {
		return nil, cache.ErrCapabilityUnsupported
	}
	return &AppenderAdapter{
		Adapter:  NewAdapter(client, isNotFound),
		appender: appender,
	}, nil
}

// NewAtomicAdapter wraps a Redis client with counter and append support.
func NewAtomicAdapter(client Client, isNotFound func(error) bool) (*AtomicAdapter, error) {
	incrementer, ok := client.(Incrementer)
	if !ok {
		return nil, cache.ErrCapabilityUnsupported
	}
	appender, ok := client.(Appender)
	if !ok {
		return nil, cache.ErrCapabilityUnsupported
	}
	return &AtomicAdapter{
		Adapter:     NewAdapter(client, isNotFound),
		incrementer: incrementer,
		appender:    appender,
	}, nil
}

// validateKey checks if a key is valid and safe.
// This prevents cache key pollution, tenant isolation bypass, and injection attacks.
func (a *Adapter) validateKey(key string) error {
	if key == "" {
		return fmt.Errorf("%w: redis key cannot be empty", cache.ErrInvalidKey)
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
			return fmt.Errorf("%w: redis key contains invalid control character at position %d", cache.ErrInvalidKey, i)
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

	return cloneBytes(value), nil
}

// Set stores a value with the specified TTL.
func (a *Adapter) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if a == nil || a.Client == nil {
		return ErrNilClient
	}

	if err := a.validateKey(key); err != nil {
		return err
	}

	return a.Client.Set(ctx, key, cloneBytes(value), ttl)
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

func incr(ctx context.Context, a *Adapter, incrementer Incrementer, key string, delta int64) (int64, error) {
	if a == nil || a.Client == nil {
		return 0, ErrNilClient
	}
	if incrementer == nil {
		return 0, cache.ErrCapabilityUnsupported
	}

	if err := a.validateKey(key); err != nil {
		return 0, err
	}

	return incrementer.IncrBy(ctx, key, delta)
}

func decr(ctx context.Context, a *Adapter, incrementer Incrementer, key string, delta int64) (int64, error) {
	if delta == minInt64 {
		return 0, fmt.Errorf("%w: integer overflow", cache.ErrNotInteger)
	}
	return incr(ctx, a, incrementer, key, -delta)
}

func appendValue(ctx context.Context, a *Adapter, appender Appender, key string, data []byte) error {
	if a == nil || a.Client == nil {
		return ErrNilClient
	}
	if appender == nil {
		return cache.ErrCapabilityUnsupported
	}

	if err := a.validateKey(key); err != nil {
		return err
	}

	_, err := appender.Append(ctx, key, cloneBytes(data))
	return err
}

// Incr atomically increments the integer value of a key by delta.
// Returns the new value after increment.
// If the key doesn't exist, it's created with delta as the initial value.
// Returns cache.ErrNotInteger if the value is not an integer.
func (a *CounterAdapter) Incr(ctx context.Context, key string, delta int64) (int64, error) {
	if a == nil {
		return 0, ErrNilClient
	}
	return incr(ctx, a.Adapter, a.incrementer, key, delta)
}

// Decr atomically decrements the integer value of a key by delta.
// Returns the new value after decrement.
// If the key doesn't exist, it's created with -delta as the initial value.
// Returns cache.ErrNotInteger if the value is not an integer.
func (a *CounterAdapter) Decr(ctx context.Context, key string, delta int64) (int64, error) {
	if a == nil {
		return 0, ErrNilClient
	}
	return decr(ctx, a.Adapter, a.incrementer, key, delta)
}

// Append appends data to the end of an existing value.
// If the key doesn't exist, it's created with the data as the value.
func (a *AppenderAdapter) Append(ctx context.Context, key string, data []byte) error {
	if a == nil {
		return ErrNilClient
	}
	return appendValue(ctx, a.Adapter, a.appender, key, data)
}

// Incr atomically increments the integer value of a key by delta.
func (a *AtomicAdapter) Incr(ctx context.Context, key string, delta int64) (int64, error) {
	if a == nil {
		return 0, ErrNilClient
	}
	return incr(ctx, a.Adapter, a.incrementer, key, delta)
}

// Decr atomically decrements the integer value of a key by delta.
func (a *AtomicAdapter) Decr(ctx context.Context, key string, delta int64) (int64, error) {
	if a == nil {
		return 0, ErrNilClient
	}
	return decr(ctx, a.Adapter, a.incrementer, key, delta)
}

// Append appends data to the end of an existing value.
func (a *AtomicAdapter) Append(ctx context.Context, key string, data []byte) error {
	if a == nil {
		return ErrNilClient
	}
	return appendValue(ctx, a.Adapter, a.appender, key, data)
}

func cloneBytes(value []byte) []byte {
	if value == nil {
		return nil
	}
	return append([]byte(nil), value...)
}
