# Embedded Key-Value Store

> **Package**: `github.com/spcent/plumego/store/kv` | **Feature**: Persistent embedded storage

## Overview

The `kv` package provides an embedded persistent key-value store with Write-Ahead Log (WAL) durability and LRU eviction. It is suitable for application state, configuration, and small datasets that require persistence without a separate database.

## Quick Start

```go
import "github.com/spcent/plumego/store/kv"

// Open persistent store
store, err := kv.Open("./data/app.kv",
    kv.WithMaxSizeBytes(100*1024*1024), // 100MB max
    kv.WithLRUEviction(true),           // Evict LRU when full
    kv.WithSyncWrites(true),            // fsync on each write (durable)
)
if err != nil {
    log.Fatal(err)
}
defer store.Close()

// Set
err = store.Set("user:123:theme", []byte("dark"))

// Get
value, err := store.Get("user:123:theme")
if err == kv.ErrNotFound {
    // Key doesn't exist
}

// Delete
err = store.Delete("user:123:theme")

// Exists
exists, err := store.Exists("user:123:theme")
```

## Configuration

```go
store, err := kv.Open("./data/app.kv",

    // Storage limits
    kv.WithMaxSizeBytes(100*1024*1024),  // 100MB storage cap
    kv.WithMaxKeys(1000000),             // Max 1M keys

    // Eviction
    kv.WithLRUEviction(true),           // Evict least-recently-used when full

    // Durability
    kv.WithSyncWrites(true),            // Flush WAL on each write (slower, safer)
    kv.WithSyncInterval(time.Second),   // Periodic fsync (balanced)

    // Serialization
    kv.WithBinarySerializer(),          // Compact binary format (default)
    kv.WithJSONSerializer(),            // Human-readable JSON format

    // Compression
    kv.WithCompression(kv.Snappy),      // Compress values
)
```

## Serialization

The KV store stores raw `[]byte`. Use the built-in serializers for typed values:

```go
// JSON serializer
store, _ := kv.Open("./data", kv.WithJSONSerializer())

type UserPref struct {
    Theme    string `json:"theme"`
    Language string `json:"language"`
}

// Marshal manually
pref := UserPref{Theme: "dark", Language: "en"}
data, _ := json.Marshal(pref)
store.Set("user:123:prefs", data)

// Unmarshal on get
raw, _ := store.Get("user:123:prefs")
var prefs UserPref
json.Unmarshal(raw, &prefs)
```

### Typed Store Wrapper

```go
// Type-safe wrapper
prefStore := kv.NewTyped[UserPref](store)

// Set
prefStore.Set("user:123", UserPref{Theme: "dark"})

// Get
prefs, err := prefStore.Get("user:123")
```

## WAL Durability

The Write-Ahead Log ensures data survives crashes:

```
Write Request
     │
     ▼
Write to WAL (sequential, fast)
     │
     ▼
Apply to in-memory hash map
     │
     ▼
Periodic compaction (WAL → snapshot)
```

### Durability Modes

```go
// Maximum durability: sync on every write (~1ms overhead)
kv.WithSyncWrites(true)

// Balanced: sync every second (good for most cases)
kv.WithSyncInterval(time.Second)

// Maximum performance: no sync (risk of data loss on crash)
kv.WithSyncInterval(0)
```

## Iteration

```go
// Iterate all keys with prefix
err := store.Scan("user:123:", func(key string, value []byte) error {
    fmt.Printf("%s = %s\n", key, value)
    return nil
})

// Iterate all keys
err = store.ScanAll(func(key string, value []byte) error {
    // Process each entry
    return nil
})

// Get all keys matching prefix
keys, err := store.Keys("config:")
```

## Batch Operations

```go
// Atomic batch write
batch := store.NewBatch()
batch.Set("config:theme", []byte("dark"))
batch.Set("config:lang", []byte("en"))
batch.Set("config:timezone", []byte("UTC"))
batch.Delete("config:old_setting")

err := store.WriteBatch(batch)
```

## Compaction

The WAL grows over time. Compaction merges all WAL entries into a compact snapshot:

```go
// Manual compaction
err := store.Compact()

// Automatic compaction (when WAL exceeds threshold)
store, _ := kv.Open("./data",
    kv.WithAutoCompact(true),
    kv.WithCompactThreshold(50*1024*1024), // Compact when WAL > 50MB
)
```

## Use Cases

### Application Configuration

```go
// Store runtime config
store.Set("feature:new_ui", []byte("true"))
store.Set("feature:beta_users", []byte(`["user-1","user-2"]`))

// Read in handlers
raw, _ := store.Get("feature:new_ui")
enabled := string(raw) == "true"
```

### Rate Limiting State

```go
// Track request counts
key := "ratelimit:" + clientIP + ":" + windowID
raw, err := store.Get(key)
if err == kv.ErrNotFound {
    store.Set(key, []byte("1"))
} else {
    count, _ := strconv.Atoi(string(raw))
    store.Set(key, []byte(strconv.Itoa(count+1)))
}
```

### Session Storage

```go
// Store session data
sessionData, _ := json.Marshal(session)
store.Set("session:"+sessionID, sessionData)

// Retrieve session
raw, err := store.Get("session:" + sessionID)
if err == kv.ErrNotFound {
    return nil, ErrSessionExpired
}
var session Session
json.Unmarshal(raw, &session)
```

## Statistics

```go
stats, err := store.Stats()
fmt.Printf("Total keys: %d\n", stats.KeyCount)
fmt.Printf("Data size: %d MB\n", stats.DataSizeBytes/1024/1024)
fmt.Printf("WAL size: %d MB\n", stats.WALSizeBytes/1024/1024)
fmt.Printf("Hit rate: %.1f%%\n", stats.HitRate*100)
```

## Related Documentation

- [Cache Module](../cache/README.md) — In-memory and Redis caching
- [Idempotency](../idempotency/README.md) — KV-backed idempotency store
