# Database Metrics Tracking

This package provides comprehensive database metrics tracking through the `InstrumentedDB` wrapper and integration with the plumego metrics system.

## Features

- **Automatic Operation Tracking**: All database operations (query, exec, transaction, ping, connect, close) are automatically recorded
- **Connection Pool Monitoring**: Track connection pool health metrics like open connections, idle connections, wait times
- **Low Overhead**: Minimal performance impact with optional metrics collection
- **Flexible Integration**: Works with any `sql.DB` and metrics collector
- **Production Ready**: Supports nil collectors for zero-overhead when metrics are disabled

## Quick Start

### Basic Usage

```go
import (
    "database/sql"
    "github.com/spcent/plumego/metrics"
    "github.com/spcent/plumego/store/db"
    _ "github.com/lib/pq" // your database driver
)

// Create a metrics collector
collector := metrics.NewPrometheusCollector("myapp")

// Open database connection
rawDB, err := sql.Open("postgres", "postgres://user:pass@localhost/db")
if err != nil {
    log.Fatal(err)
}

// Wrap with instrumentation
instrumentedDB := db.NewInstrumentedDB(rawDB, collector, "postgres")

// Use as normal - all operations are automatically tracked
result, err := instrumentedDB.ExecContext(ctx,
    "INSERT INTO users (name, email) VALUES ($1, $2)",
    "Alice", "alice@example.com")
```

## Tracked Metrics

### Operation Metrics

All database operations are tracked with the following labels:

| Operation | Description | Labels |
|-----------|-------------|--------|
| `query` | SELECT queries | operation, driver, query, rows |
| `exec` | INSERT/UPDATE/DELETE | operation, driver, query, rows |
| `transaction` | BEGIN/COMMIT/ROLLBACK | operation, driver, query |
| `ping` | Database health checks | operation, driver |
| `connect` | Connection establishment | operation, driver |
| `close` | Connection closure | operation, driver |

**Recorded Data:**
- Duration (milliseconds)
- Error status
- Affected rows (for exec operations)
- Query text (truncated to 100 chars)

### Connection Pool Metrics

Use `RecordPoolStats()` to periodically track connection pool health:

```go
// Start periodic pool stats collection
ticker := time.NewTicker(30 * time.Second)
go func() {
    for range ticker.C {
        instrumentedDB.RecordPoolStats(context.Background())
    }
}()
```

**Pool Metrics:**
- `pool_open_connections`: Total open connections
- `pool_in_use`: Connections currently in use
- `pool_idle`: Idle connections waiting for reuse
- `pool_wait_count`: Total number of times waited for a connection
- `pool_wait_duration`: Total time spent waiting for connections
- `pool_max_idle_closed`: Connections closed due to SetMaxIdleConns
- `pool_max_lifetime_closed`: Connections closed due to SetConnMaxLifetime

## Advanced Usage

### With Prometheus

```go
import (
    "net/http"
    "github.com/spcent/plumego/metrics"
    "github.com/spcent/plumego/store/db"
)

// Create Prometheus collector
collector := metrics.NewPrometheusCollector("myapp")

// Create instrumented database
instrumentedDB := db.NewInstrumentedDB(rawDB, collector, "postgres")

// Configure connection pool
instrumentedDB.SetMaxOpenConns(25)
instrumentedDB.SetMaxIdleConns(5)
instrumentedDB.SetConnMaxLifetime(5 * time.Minute)

// Start periodic pool monitoring
go func() {
    ticker := time.NewTicker(30 * time.Second)
    for range ticker.C {
        instrumentedDB.RecordPoolStats(context.Background())
    }
}()

// Expose metrics endpoint
http.Handle("/metrics", collector.Handler())
http.ListenAndServe(":9090", nil)
```

### With OpenTelemetry

```go
import (
    "github.com/spcent/plumego/metrics"
    "github.com/spcent/plumego/store/db"
)

// Create OpenTelemetry tracer
tracer := metrics.NewOpenTelemetryTracer("myapp")

// Create instrumented database
instrumentedDB := db.NewInstrumentedDB(rawDB, tracer, "postgres")

// Operations are automatically traced
rows, err := instrumentedDB.QueryContext(ctx, "SELECT * FROM users WHERE id = $1", userID)
```

### With Multiple Collectors

```go
import (
    "github.com/spcent/plumego/metrics"
    "github.com/spcent/plumego/store/db"
)

// Combine multiple collectors
prom := metrics.NewPrometheusCollector("myapp")
otel := metrics.NewOpenTelemetryTracer("myapp")
multi := metrics.NewMultiCollector(prom, otel)

// All metrics are sent to both collectors
instrumentedDB := db.NewInstrumentedDB(rawDB, multi, "postgres")
```

### Disable Metrics

```go
// Pass nil collector for zero overhead
instrumentedDB := db.NewInstrumentedDB(rawDB, nil, "postgres")

// Or use NoopCollector
noop := metrics.NewNoopCollector()
instrumentedDB := db.NewInstrumentedDB(rawDB, noop, "postgres")
```

## Integration with Database Cluster

The instrumented DB works seamlessly with the plumego database cluster package:

```go
import (
    "github.com/spcent/plumego/metrics"
    "github.com/spcent/plumego/store/db"
    "github.com/spcent/plumego/store/db/rw"
)

// Create instrumented databases
collector := metrics.NewPrometheusCollector("myapp")
primary := db.NewInstrumentedDB(primaryDB, collector, "postgres")
replica1 := db.NewInstrumentedDB(replica1DB, collector, "postgres")
replica2 := db.NewInstrumentedDB(replica2DB, collector, "postgres")

// Create cluster with instrumented databases
cluster := rw.NewCluster(primary, []db.DB{replica1, replica2})

// All queries through the cluster are tracked with metrics
rows, err := cluster.QueryContext(ctx, "SELECT * FROM users")
```

## Prometheus Query Examples

Once metrics are exposed, you can query them in Prometheus:

```promql
# Average query duration by operation
rate(db_query_duration_milliseconds_sum[5m]) / rate(db_query_duration_milliseconds_count[5m])

# Queries per second by operation
rate(db_query_total[5m])

# Error rate
rate(db_query_errors_total[5m]) / rate(db_query_total[5m])

# Connection pool utilization
db_pool_in_use / db_pool_open_connections

# Connection pool wait time
rate(db_pool_wait_duration_sum[5m])
```

## Performance Considerations

### Overhead

The `InstrumentedDB` adds minimal overhead:

- **Nil collector**: Zero overhead (checked at runtime)
- **With collector**: ~100-200ns per operation
- **Query truncation**: Prevents memory issues with large queries
- **Async recording**: Metrics are recorded synchronously but collectors may buffer

### Best Practices

1. **Query Truncation**: Queries are automatically truncated to 100 characters to prevent excessive label cardinality
2. **Pool Monitoring**: Record pool stats every 30-60 seconds, not more frequently
3. **Production Use**: Always configure max connections and connection lifetime
4. **Sensitive Data**: Query parameters are NOT logged, only the query structure

## Monitoring Recommendations

### Key Metrics to Alert On

1. **High Error Rate**: `db_query_errors_total` / `db_query_total` > 1%
2. **Slow Queries**: `db_query_duration_p95 > 1000ms`
3. **Pool Exhaustion**: `db_pool_in_use` / `db_pool_open_connections` > 90%
4. **Connection Waits**: `rate(db_pool_wait_count[5m])` > 0
5. **Connection Churn**: `rate(db_pool_max_lifetime_closed[5m])` high

### Dashboard Setup

Create dashboards with:
- Query rate by operation type
- Query duration percentiles (p50, p95, p99)
- Error rate over time
- Connection pool utilization
- Connection pool wait times
- Slow query identification

## Example Application

See `examples/db-metrics/` for a complete example application demonstrating:
- Database setup with metrics
- Connection pool monitoring
- Prometheus integration
- Grafana dashboard configuration

## API Reference

### InstrumentedDB

```go
// NewInstrumentedDB creates a new instrumented database connection
func NewInstrumentedDB(db *sql.DB, collector MetricsCollector, driver string) *InstrumentedDB

// RecordPoolStats records current connection pool statistics
func (idb *InstrumentedDB) RecordPoolStats(ctx context.Context)

// Unwrap returns the underlying sql.DB instance
func (idb *InstrumentedDB) Unwrap() *sql.DB

// All standard sql.DB methods (ExecContext, QueryContext, etc.)
```

### MetricsCollector Interface

```go
type MetricsCollector interface {
    ObserveDB(ctx context.Context, operation, driver, query string,
              rows int, duration time.Duration, err error)
}
```

## Troubleshooting

### Metrics Not Appearing

1. Check collector is not nil: `if collector == nil`
2. Verify metrics endpoint is exposed: `curl http://localhost:9090/metrics`
3. Check operation type is recognized

### High Cardinality

If you see too many unique metric series:

1. Query truncation is automatic (100 chars)
2. Use query normalization in your application
3. Consider parameterized queries

### Performance Impact

If metrics add too much overhead:

1. Use `NoopCollector` in non-production environments
2. Sample metrics (e.g., 10% of requests)
3. Use asynchronous collectors

## License

See the main plumego LICENSE file.
