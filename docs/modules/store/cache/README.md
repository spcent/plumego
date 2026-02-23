# Cache Module

> **Package**: `github.com/spcent/plumego/store/cache` | **Backends**: Memory, Redis, Distributed

## Overview

The `cache` package provides a unified caching interface with multiple backends. All backends implement the same `Cache` interface, enabling easy switching between in-memory, Redis, or distributed caching without changing application code.

## Cache Interface

```go
type Cache interface {
    Get(key string) ([]byte, bool)
    Set(key string, value []byte, opts ...Option) error
    Delete(key string) error
    Clear() error
    Exists(key string) bool
    TTL(key string) time.Duration
}
```

## In-Memory Cache

```go
import "github.com/spcent/plumego/store/cache"

c := cache.NewMemory(
    cache.WithMaxSize(10000),          // Max 10k entries
    cache.WithTTL(5*time.Minute),      // Default TTL
    cache.WithEvictionPolicy(cache.LRU), // LRU eviction
)

// Store
c.Set("user:123", userJSON)
c.Set("session:abc", sessionData, cache.WithTTL(30*time.Minute)) // Override TTL

// Retrieve
data, ok := c.Get("user:123")
if !ok {
    // Cache miss — fetch from database
}

// Delete
c.Delete("user:123")

// Check existence
if c.Exists("session:abc") {
    // Session still valid
}
```

## Redis Cache

```go
// Connect
c := cache.NewRedis(os.Getenv("REDIS_URL"),
    cache.WithKeyPrefix("myapp:"),        // Namespace all keys
    cache.WithDefaultTTL(10*time.Minute),
    cache.WithMaxRetries(3),
    cache.WithTimeout(2*time.Second),
)

// Use same interface as in-memory
c.Set("user:123", userJSON)
data, ok := c.Get("user:123")

// With expiry
c.Set("token:abc", tokenData, cache.WithTTL(15*time.Minute))
```

### Redis with Connection Pool

```go
c := cache.NewRedisPool(cache.RedisPoolConfig{
    URL:         os.Getenv("REDIS_URL"),
    MaxActive:   20,
    MaxIdle:     5,
    IdleTimeout: 5 * time.Minute,
    KeyPrefix:   "myapp:",
})
```

## Typed Cache Helpers

```go
// Generic typed cache wrapper
userCache := cache.NewTyped[User](c, cache.WithTTL(5*time.Minute))

// Set
userCache.Set("user:123", user)

// Get
user, ok := userCache.Get("user:123")

// GetOrLoad (cache-aside pattern)
user, err := userCache.GetOrLoad("user:123", func(key string) (User, error) {
    return db.GetUser(ctx, "123")
})
```

## Cache-Aside Pattern

```go
func GetUser(ctx context.Context, userID string) (User, error) {
    cacheKey := "user:" + userID

    // 1. Check cache
    if data, ok := userCache.Get(cacheKey); ok {
        var user User
        json.Unmarshal(data, &user)
        return user, nil
    }

    // 2. Cache miss: load from DB
    user, err := db.QueryRow(ctx, "SELECT * FROM users WHERE id = ?", userID)
    if err != nil {
        return User{}, err
    }

    // 3. Store in cache
    data, _ := json.Marshal(user)
    userCache.Set(cacheKey, data, cache.WithTTL(5*time.Minute))

    return user, nil
}
```

## Write-Through Cache

```go
func UpdateUser(ctx context.Context, user User) error {
    // Update database first
    if err := db.UpdateUser(ctx, user); err != nil {
        return err
    }

    // Then update cache
    data, _ := json.Marshal(user)
    cacheKey := "user:" + user.ID
    userCache.Set(cacheKey, data, cache.WithTTL(5*time.Minute))

    return nil
}
```

## Cache Invalidation

```go
// Single key
func DeleteUser(ctx context.Context, userID string) error {
    if err := db.DeleteUser(ctx, userID); err != nil {
        return err
    }
    userCache.Delete("user:" + userID)
    return nil
}

// Pattern-based invalidation (Redis)
func InvalidateUserCache(userID string) {
    redisCache.DeletePattern("user:" + userID + ":*")
}

// Clear all
func ClearAllCache() {
    c.Clear()
}
```

## Distributed Cache

Multi-node aware cache with consistent hashing:

```go
distributedCache := cache.NewDistributed(
    cache.WithNodes([]string{
        "redis-1:6379",
        "redis-2:6379",
        "redis-3:6379",
    }),
    cache.WithReplication(2),     // Store in 2 nodes
    cache.WithConsistentHashing(), // Even distribution
)
```

## HTTP Response Cache

Cache entire HTTP responses:

```go
import "github.com/spcent/plumego/middleware/cache"

app.Get("/api/products", func(w http.ResponseWriter, r *http.Request) {
    products := loadProducts(r.Context())
    json.NewEncoder(w).Encode(products)
}, middleware.Cache(c, cache.WithTTL(5*time.Minute)))
```

## Cache Metrics

```go
c := cache.NewMemory(
    cache.WithMetrics(metricsRegistry),
)

// Emits:
// cache_hits_total{cache="memory"}
// cache_misses_total{cache="memory"}
// cache_evictions_total{cache="memory"}
// cache_entries_total{cache="memory"}
// cache_hit_ratio{cache="memory"}
```

## Multi-Level Cache

```go
// L1: fast in-memory, L2: Redis
multiLevel := cache.NewMultiLevel(
    cache.NewMemory(cache.WithMaxSize(1000)),  // L1
    cache.NewRedis(os.Getenv("REDIS_URL")),    // L2
)

// Get checks L1, then L2, then cache miss
// Set writes to both L1 and L2
```

## Best Practices

### 1. Namespace Keys

```go
// ✅ Good: namespaced keys
"user:123:profile"
"session:abc123"
"product:catalog:page:1"

// ❌ Bad: generic keys (collision risk)
"123"
"data"
```

### 2. Set Appropriate TTLs

```go
// User data: relatively stable
cache.WithTTL(5 * time.Minute)

// Session tokens: match auth timeout
cache.WithTTL(sessionDuration)

// Rate limit counters: sliding window
cache.WithTTL(time.Minute)
```

### 3. Handle Cache Errors Gracefully

```go
data, ok := c.Get(key)
if !ok {
    // Cache miss — not an error, just fetch from source
    data, err = db.Get(ctx, key)
    if err != nil {
        return err
    }
    c.Set(key, data) // Best effort — ignore cache set error
}
```

## Related Documentation

- [KV Store](../kv/README.md) — Persistent embedded storage
- [Database](../db/README.md) — Database access patterns
- [Middleware: Cache](../../middleware/cache.md) — HTTP response caching
