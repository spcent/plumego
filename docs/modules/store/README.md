# Store Module

> **Package Path**: `github.com/spcent/plumego/store` | **Stability**: Medium | **Priority**: P2

## Overview

The `store/` package provides data persistence abstractions for Plumego applications. It covers caching, database access, embedded key-value storage, file storage, and idempotent request handling.

High-level data topology such as read-write splitting and sharding now lives under `x/data/*`, not under the stable `store` root.

**Key Features**:
- **Cache**: Interface-based caching with in-memory and Redis backends
- **Database**: `database/sql` wrapper with metrics and query monitoring
- **KV Store**: Embedded key-value store with WAL persistence and LRU eviction
- **File Storage**: Local and S3-compatible file storage with metadata
- **Idempotency**: Duplicate request detection and replay

## Subpackages

| Package | Description |
|---------|-------------|
| `store/cache` | Caching interface (in-memory, Redis, distributed) |
| `store/db` | SQL database wrapper with metrics and monitoring |
| `store/kv` | Embedded key-value store with WAL |
| `store/file` | File storage (local filesystem, S3) |
| `store/idempotency` | Idempotent request handling |
| `x/data/rw` | Read-write splitting orchestration |
| `x/data/sharding` | Sharding and topology routing |

## Quick Start

### Cache

```go
import "github.com/spcent/plumego/store/cache"

// In-memory cache
c := cache.NewMemory(cache.WithMaxSize(1000), cache.WithTTL(5*time.Minute))
c.Set("key", "value")
val, ok := c.Get("key")

// Redis cache
c := cache.NewRedis(os.Getenv("REDIS_URL"))
c.Set("key", "value", cache.WithTTL(10*time.Minute))
```

### Database

```go
import (
    "github.com/spcent/plumego/store/db"
    tenantdb "github.com/spcent/plumego/x/tenant/store/db"
)

// Open database with monitoring
database, err := db.Open("postgres", os.Getenv("DATABASE_URL"),
    db.WithSlowQueryLog(100*time.Millisecond),
    db.WithMetrics(metricsRegistry),
)

// Tenant-aware query filtering lives in the x/tenant extension.
tenantDB := tenantdb.NewTenantDB(database)
rows, err := tenantDB.QueryFromContext(ctx, "SELECT * FROM users")
```

### Key-Value Store

```go
import "github.com/spcent/plumego/store/kv"

// Embedded persistent KV store
store, err := kv.Open("./data/app.kv",
    kv.WithMaxSize(100*1024*1024), // 100MB
    kv.WithLRUEviction(true),
)
store.Set("config:theme", []byte("dark"))
value, err := store.Get("config:theme")
```

### File Storage

```go
import "github.com/spcent/plumego/store/file"

// Local storage
fs := file.NewLocal("./uploads")

// S3-compatible storage
fs := file.NewS3(file.S3Config{
    Endpoint:  os.Getenv("S3_ENDPOINT"),
    Bucket:    os.Getenv("S3_BUCKET"),
    AccessKey: os.Getenv("S3_ACCESS_KEY"),
    SecretKey: os.Getenv("S3_SECRET_KEY"),
})

// Upload
err := fs.Store(ctx, "uploads/avatar.png", imageData, file.WithContentType("image/png"))

// Download
reader, err := fs.Get(ctx, "uploads/avatar.png")
```

### Idempotency

```go
import "github.com/spcent/plumego/store/idempotency"

// Prevent duplicate payment processing
store := idempotency.NewKVStore(kvStore)
handler := idempotency.Middleware(store)(paymentHandler)
```

## Module Documentation

- **[Cache](cache/README.md)** — Caching strategies and backends
- **[Database](db/README.md)** — SQL wrapper with monitoring
- **[Database Isolation](db/tenant-isolation.md)** — Tenant data isolation
- **[KV Store](kv/README.md)** — Embedded key-value storage
- **[File Storage](file/README.md)** — File upload and retrieval
- **[Idempotency](idempotency/README.md)** — Duplicate request handling

## Related Documentation

- [Tenant Module](../tenant/database-isolation.md) — Multi-tenant DB access
- [Security Module](../security/) — Secure file handling
- [x/data](../x-data/README.md) — Data topology extensions beyond the stable store layer
