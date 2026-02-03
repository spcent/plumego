package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/spcent/plumego/tenant"
)

// TenantCache wraps a Cache implementation and automatically scopes keys by tenant ID.
// This prevents cross-tenant cache pollution and ensures data isolation.
type TenantCache struct {
	cache     Cache
	keyPrefix string
	separator string
}

// NewTenantCache creates a new tenant-scoped cache wrapper.
// Keys will be automatically prefixed with the tenant ID from context.
func NewTenantCache(cache Cache, options ...TenantCacheOption) *TenantCache {
	tc := &TenantCache{
		cache:     cache,
		keyPrefix: "tenant",
		separator: ":",
	}

	for _, opt := range options {
		opt(tc)
	}

	return tc
}

// TenantCacheOption configures a TenantCache.
type TenantCacheOption func(*TenantCache)

// WithKeyPrefix sets a custom prefix for all tenant cache keys.
// Default is "tenant".
func WithKeyPrefix(prefix string) TenantCacheOption {
	return func(tc *TenantCache) {
		tc.keyPrefix = prefix
	}
}

// WithSeparator sets a custom separator between prefix, tenant ID, and key.
// Default is ":".
func WithSeparator(sep string) TenantCacheOption {
	return func(tc *TenantCache) {
		tc.separator = sep
	}
}

// Get retrieves a value from the cache, automatically scoping by tenant ID from context.
func (tc *TenantCache) Get(ctx context.Context, key string) ([]byte, error) {
	scopedKey, err := tc.scopeKey(ctx, key)
	if err != nil {
		return nil, err
	}
	return tc.cache.Get(ctx, scopedKey)
}

// Set stores a value in the cache, automatically scoping by tenant ID from context.
func (tc *TenantCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	scopedKey, err := tc.scopeKey(ctx, key)
	if err != nil {
		return err
	}
	return tc.cache.Set(ctx, scopedKey, value, ttl)
}

// Delete removes a value from the cache, automatically scoping by tenant ID from context.
func (tc *TenantCache) Delete(ctx context.Context, key string) error {
	scopedKey, err := tc.scopeKey(ctx, key)
	if err != nil {
		return err
	}
	return tc.cache.Delete(ctx, scopedKey)
}

// Exists checks if a key exists in the cache, automatically scoping by tenant ID from context.
func (tc *TenantCache) Exists(ctx context.Context, key string) (bool, error) {
	scopedKey, err := tc.scopeKey(ctx, key)
	if err != nil {
		return false, err
	}
	return tc.cache.Exists(ctx, scopedKey)
}

// Clear removes all cache entries (delegates to underlying cache).
// Note: This clears the ENTIRE cache, not just the current tenant.
// For tenant-specific clearing, use DeletePattern with tenant prefix.
func (tc *TenantCache) Clear(ctx context.Context) error {
	return tc.cache.Clear(ctx)
}

// Incr atomically increments a counter, scoped by tenant ID.
func (tc *TenantCache) Incr(ctx context.Context, key string, delta int64) (int64, error) {
	scopedKey, err := tc.scopeKey(ctx, key)
	if err != nil {
		return 0, err
	}
	return tc.cache.Incr(ctx, scopedKey, delta)
}

// Decr atomically decrements a counter, scoped by tenant ID.
func (tc *TenantCache) Decr(ctx context.Context, key string, delta int64) (int64, error) {
	scopedKey, err := tc.scopeKey(ctx, key)
	if err != nil {
		return 0, err
	}
	return tc.cache.Decr(ctx, scopedKey, delta)
}

// Append appends data to an existing value, scoped by tenant ID.
func (tc *TenantCache) Append(ctx context.Context, key string, data []byte) error {
	scopedKey, err := tc.scopeKey(ctx, key)
	if err != nil {
		return err
	}
	return tc.cache.Append(ctx, scopedKey, data)
}

// RawCache returns the underlying cache for operations that need direct access.
// Use with caution as this bypasses tenant isolation.
func (tc *TenantCache) RawCache() Cache {
	return tc.cache
}

// scopeKey builds a tenant-scoped cache key from context.
func (tc *TenantCache) scopeKey(ctx context.Context, key string) (string, error) {
	tenantID := tenant.TenantIDFromContext(ctx)
	if tenantID == "" {
		return "", tenant.ErrTenantNotFound
	}
	return tc.buildKey(tenantID, key), nil
}

// buildKey constructs a cache key with tenant prefix.
func (tc *TenantCache) buildKey(tenantID, key string) string {
	if tc.keyPrefix == "" {
		return fmt.Sprintf("%s%s%s", tenantID, tc.separator, key)
	}
	return fmt.Sprintf("%s%s%s%s%s", tc.keyPrefix, tc.separator, tenantID, tc.separator, key)
}
