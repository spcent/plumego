# Database Module

> **Package**: `github.com/spcent/plumego/store/db` | **Backend**: database/sql

## Overview

The `db` package wraps the standard `database/sql` package with additional features: query metrics, slow query logging, tenant isolation, database sharding, and read/write splitting.

## Quick Start

```go
import "github.com/spcent/plumego/store/db"

// Open with monitoring
database, err := db.Open("postgres", os.Getenv("DATABASE_URL"),
    db.WithMaxOpenConns(25),
    db.WithMaxIdleConns(5),
    db.WithConnMaxLifetime(30*time.Minute),
    db.WithSlowQueryLog(100*time.Millisecond),
    db.WithMetrics(metricsRegistry),
)
if err != nil {
    log.Fatal(err)
}
defer database.Close()

// Standard usage
rows, err := database.QueryContext(ctx, "SELECT id, name FROM users WHERE active = $1", true)
```

## Connection Configuration

```go
database, err := db.Open("postgres", dsn,
    // Connection pool
    db.WithMaxOpenConns(25),          // Max simultaneous connections
    db.WithMaxIdleConns(5),           // Keep 5 idle connections
    db.WithConnMaxLifetime(30*time.Minute), // Recycle connections
    db.WithConnMaxIdleTime(10*time.Minute), // Close idle connections

    // Monitoring
    db.WithSlowQueryLog(100*time.Millisecond), // Log queries > 100ms
    db.WithMetrics(metricsRegistry),           // Prometheus metrics
    db.WithTracing(tracer),                    // Distributed tracing

    // Safety
    db.WithMaxQueryTime(30*time.Second), // Kill long-running queries
)
```

## Supported Drivers

```go
// PostgreSQL
db.Open("postgres", "postgres://user:pass@localhost/mydb?sslmode=require")

// MySQL
db.Open("mysql", "user:pass@tcp(localhost:3306)/mydb?parseTime=true")

// SQLite
db.Open("sqlite3", "./data/app.db")
```

## Metrics

When metrics are enabled, the following are tracked automatically:

| Metric | Type | Description |
|--------|------|-------------|
| `db_queries_total` | Counter | Total queries by operation and status |
| `db_query_duration_seconds` | Histogram | Query execution time |
| `db_connections_open` | Gauge | Open connection count |
| `db_connections_idle` | Gauge | Idle connection count |
| `db_slow_queries_total` | Counter | Queries exceeding slow query threshold |

## Slow Query Logging

```go
database, err := db.Open("postgres", dsn,
    db.WithSlowQueryLog(100*time.Millisecond), // Threshold
    db.WithSlowQueryLogger(func(query string, args []any, duration time.Duration) {
        log.Warnf("slow query [%s]: %s (args: %v)", duration, query, args)
    }),
)
```

**Example slow query log**:
```
WARN  slow query [523ms]: SELECT * FROM orders WHERE user_id = $1 AND status = $2 ORDER BY created_at DESC LIMIT 100 (args: [user-123 pending])
```

## Migrations

```go
// Run migrations on startup
migrator := db.NewMigrator(database,
    db.WithMigrationsDir("./migrations"),
    db.WithMigrationsTable("schema_migrations"),
)

if err := migrator.Up(ctx); err != nil {
    log.Fatalf("migration failed: %v", err)
}
```

**Migration file** `migrations/001_create_users.sql`:
```sql
-- +migrate Up
CREATE TABLE users (
    id          VARCHAR(36) PRIMARY KEY,
    tenant_id   VARCHAR(36) NOT NULL,
    email       VARCHAR(255) UNIQUE NOT NULL,
    name        VARCHAR(255) NOT NULL,
    created_at  TIMESTAMP DEFAULT NOW()
);
CREATE INDEX idx_users_tenant ON users(tenant_id);

-- +migrate Down
DROP TABLE users;
```

## Read/Write Splitting

Route read queries to replicas to distribute load:

```go
primary := db.Open("postgres", os.Getenv("DB_PRIMARY"))
replica1 := db.Open("postgres", os.Getenv("DB_REPLICA_1"))
replica2 := db.Open("postgres", os.Getenv("DB_REPLICA_2"))

rwDB := db.NewReadWriteSplit(primary,
    db.WithReplicas(replica1, replica2),
    db.WithReplicaSelector(db.RoundRobin),
)

// Writes go to primary
rwDB.ExecContext(ctx, "INSERT INTO users ...")

// Reads go to replicas
rwDB.QueryContext(ctx, "SELECT * FROM users ...")
```

## Sharding

Distribute data across multiple database instances:

```go
shardedDB := db.NewSharded(
    db.ShardConfig{
        Shards: []db.ShardInstance{
            {ID: 0, DB: shard0DB, Range: db.HashRange(0, 255)},
            {ID: 1, DB: shard1DB, Range: db.HashRange(256, 511)},
            {ID: 2, DB: shard2DB, Range: db.HashRange(512, 767)},
            {ID: 3, DB: shard3DB, Range: db.HashRange(768, 1023)},
        },
        ShardKey: func(ctx context.Context) string {
            return tenant.TenantIDFromContext(ctx)
        },
    },
)

// Automatically routes to correct shard
shardedDB.QueryContext(ctx, "SELECT * FROM orders")
```

## Query Builder

```go
import "github.com/spcent/plumego/store/db"

// Build safe parameterized queries
q := db.Query("SELECT * FROM users").
    Where("active = ?", true).
    Where("created_at > ?", since).
    OrderBy("name ASC").
    Limit(20).
    Offset(40)

rows, err := database.QueryContext(ctx, q.SQL(), q.Args()...)
```

## Transaction Helper

```go
err := db.WithTransaction(ctx, database, func(tx *sql.Tx) error {
    // All operations in a transaction
    _, err := tx.ExecContext(ctx, "INSERT INTO orders (id, user_id) VALUES (?, ?)", orderID, userID)
    if err != nil {
        return err // Auto-rollback on error
    }

    _, err = tx.ExecContext(ctx, "UPDATE inventory SET quantity = quantity - 1 WHERE product_id = ?", productID)
    if err != nil {
        return err
    }

    return nil // Auto-commit on success
})
```

## Health Check

```go
func (db *DB) Ping(ctx context.Context) error {
    return database.PingContext(ctx)
}

// Register as health check
checker.AddReadiness("database", func(ctx context.Context) error {
    ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
    defer cancel()
    return database.PingContext(ctx)
})
```

## Best Practices

### 1. Always Use Context

```go
// ✅ Good: context-aware
rows, err := db.QueryContext(ctx, query, args...)

// ❌ Bad: no context
rows, err := db.Query(query, args...)
```

### 2. Close Rows

```go
rows, err := db.QueryContext(ctx, "SELECT * FROM users")
if err != nil {
    return err
}
defer rows.Close() // Always defer Close
```

### 3. Use Prepared Statements for Repeated Queries

```go
stmt, err := db.PrepareContext(ctx, "SELECT * FROM users WHERE id = $1")
if err != nil {
    return err
}
defer stmt.Close()

// Reuse prepared statement
for _, id := range userIDs {
    row := stmt.QueryRowContext(ctx, id)
    // ...
}
```

## Related Documentation

- [Tenant Isolation](tenant-isolation.md) — Multi-tenant query filtering
- [Cache Module](../cache/README.md) — Caching database results
- [Tenant Module](../../tenant/database-isolation.md) — Tenant database patterns
