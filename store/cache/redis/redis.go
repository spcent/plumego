package redis

import (
	"context"
	"errors"
	"time"

	"github.com/spcent/plumego/store/cache"
)

var (
	ErrClearUnsupported = errors.New("redis cache: clear unsupported")
	ErrNilClient        = errors.New("redis cache: client is nil")
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
	Client     Client
	IsNotFound func(error) bool
}

// NewAdapter wraps a Redis client in a cache.Cache adapter.
func NewAdapter(client Client, isNotFound func(error) bool) *Adapter {
	return &Adapter{Client: client, IsNotFound: isNotFound}
}

// Get returns the cached value for the provided key.
func (a *Adapter) Get(ctx context.Context, key string) ([]byte, error) {
	if a == nil || a.Client == nil {
		return nil, ErrNilClient
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
	return a.Client.Set(ctx, key, value, ttl)
}

// Delete removes the key from Redis.
func (a *Adapter) Delete(ctx context.Context, key string) error {
	if a == nil || a.Client == nil {
		return ErrNilClient
	}
	_, err := a.Client.Del(ctx, key)
	return err
}

// Exists reports whether a key exists.
func (a *Adapter) Exists(ctx context.Context, key string) (bool, error) {
	if a == nil || a.Client == nil {
		return false, ErrNilClient
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
