# Metrics Package

A comprehensive, zero-dependency metrics collection package for the Plumego framework, providing unified interfaces for Prometheus, OpenTelemetry, and custom metric collectors.

## Features

- **Zero External Dependencies**: Built entirely on Go standard library
- **Unified Interface**: Single API for all metric types and collectors
- **Multiple Collectors**: Prometheus, OpenTelemetry, Base, and No-op implementations
- **Type Safety**: Strong typing for all metric types
- **Thread Safe**: All collectors are safe for concurrent use
- **Memory Efficient**: Configurable memory limits and automatic eviction
- **Rich Helpers**: Timer, MeasureFunc, MultiCollector, and convenience functions
- **Comprehensive**: Support for HTTP, PubSub, MQ, KV, and IPC metrics

## Quick Start

```go
import "github.com/spcent/plumego/metrics"

// Create a Prometheus collector
collector := metrics.NewPrometheusCollector("myapp")

// Record HTTP metrics
ctx := context.Background()
collector.ObserveHTTP(ctx, "GET", "/api/users", 200, 1024, 50*time.Millisecond)

// Expose metrics endpoint
http.Handle("/metrics", collector.Handler())
```

## Installation

The metrics package is part of the Plumego framework:

```bash
go get github.com/spcent/plumego
```

## Collector Implementations

### PrometheusCollector

Prometheus-compatible metrics exposition in text format.

```go
collector := metrics.NewPrometheusCollector("myapp").
    WithMaxMemory(10000)

// Record metrics
collector.ObserveHTTP(ctx, "GET", "/api/users", 200, 1024, 50*time.Millisecond)

// Expose endpoint
http.Handle("/metrics", collector.Handler())
```

**Metrics Exposed:**
- `{namespace}_http_requests_total` - Total HTTP requests by method, path, status
- `{namespace}_http_request_duration_seconds_sum` - Sum of request durations
- `{namespace}_http_request_duration_seconds_count` - Count of requests
- `{namespace}_http_request_duration_seconds_min` - Minimum request duration
- `{namespace}_http_request_duration_seconds_max` - Maximum request duration
- `{namespace}_uptime_seconds` - Application uptime

### OpenTelemetryTracer

OpenTelemetry-compatible distributed tracing.

```go
tracer := metrics.NewOpenTelemetryTracer("my-service")

// In HTTP handler
ctx, span := tracer.Start(context.Background(), r)
defer span.End(middleware.RequestMetrics{
    Status:  w.Status(),
    Bytes:   w.BytesWritten(),
    TraceID: traceID,
})

// Get spans for export
spans := tracer.Spans()
```

### BaseMetricsCollector

In-memory collector for testing and development.

```go
collector := metrics.NewBaseMetricsCollector().
    WithMaxRecords(5000)

collector.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)

// Inspect records
records := collector.GetRecords()
for _, record := range records {
    fmt.Printf("%s: %v\n", record.Name, record.Value)
}
```

### NoopCollector

Zero-overhead collector that discards all metrics.

```go
// For production when metrics are disabled
collector := metrics.NewNoopCollector()

// All operations are no-ops
collector.ObserveHTTP(ctx, "GET", "/test", 200, 100, 50*time.Millisecond)
```

## Metric Types

The package supports metrics for various subsystems:

### HTTP Metrics

```go
collector.ObserveHTTP(ctx, "POST", "/api/data", 201, 512, 100*time.Millisecond)
```

### Pub/Sub Metrics

```go
collector.ObservePubSub(ctx, "publish", "user.created", 5*time.Millisecond, nil)
collector.ObservePubSub(ctx, "subscribe", "events", 10*time.Millisecond, nil)
```

### Message Queue Metrics

```go
collector.ObserveMQ(ctx, "publish", "jobs", 8*time.Millisecond, nil, false)
collector.ObserveMQ(ctx, "subscribe", "orders", 10*time.Millisecond, err, false)
```

### Key-Value Store Metrics

```go
collector.ObserveKV(ctx, "get", "user:123", 2*time.Millisecond, nil, true)  // cache hit
collector.ObserveKV(ctx, "get", "user:456", 5*time.Millisecond, nil, false) // cache miss
collector.ObserveKV(ctx, "set", "user:789", 3*time.Millisecond, nil, false)
```

### IPC Metrics

```go
collector.ObserveIPC(ctx, "read", "/tmp/app.sock", "unix", 256, 1*time.Millisecond, nil)
collector.ObserveIPC(ctx, "write", "/tmp/app.sock", "unix", 512, 2*time.Millisecond, nil)
```

## Helper Functions

### Timer

Convenient duration measurement:

```go
timer := metrics.NewTimer()

// Perform operation
doWork()

// Record duration
duration := timer.Elapsed()
collector.ObserveHTTP(ctx, "GET", "/api", 200, 100, duration)
```

### MeasureFunc

Automatic metric recording with function wrapping:

```go
err := metrics.MeasureFunc(ctx, collector, "database_query", "users", func() error {
    return db.Query("SELECT * FROM users")
}, metrics.MeasureWithKV(true))
```

**Measurement Options:**
- `MeasureWithKV(hit bool)` - For KV operations
- `MeasureWithPubSub()` - For PubSub operations
- `MeasureWithMQ(panicked bool)` - For MQ operations
- `MeasureWithIPC(transport, bytes)` - For IPC operations

### Convenience Functions

```go
// Record success
metrics.RecordSuccess(ctx, collector, "api_call", 50*time.Millisecond)

// Record error
if err := doOperation(); err != nil {
    metrics.RecordError(ctx, collector, "operation", 100*time.Millisecond, err)
}

// Record with custom labels
metrics.RecordWithLabels(ctx, collector, "custom_metric", 75*time.Millisecond,
    metrics.MetricLabels{
        "endpoint": "/api/users",
        "region":   "us-east",
    })
```

### MultiCollector

Combine multiple collectors:

```go
prom := metrics.NewPrometheusCollector("myapp")
otel := metrics.NewOpenTelemetryTracer("myapp")
base := metrics.NewBaseMetricsCollector()

multi := metrics.NewMultiCollector(prom, otel, base)

// Metrics go to all collectors
multi.ObserveHTTP(ctx, "GET", "/api/users", 200, 100, 50*time.Millisecond)

// Get combined statistics
stats := multi.GetStats()
```

## Memory Management

### Prometheus Collector

Limit unique metric series to prevent unbounded memory growth:

```go
collector := metrics.NewPrometheusCollector("myapp").
    WithMaxMemory(10000) // Max 10,000 unique series

// Oldest/least-used series are evicted when limit is reached
```

### Base Collector

Limit retained records:

```go
collector := metrics.NewBaseMetricsCollector().
    WithMaxRecords(5000) // Keep only last 5,000 records

// Set to 0 to disable limit
collector.WithMaxRecords(0)
```

## Statistics

All collectors provide statistics:

```go
stats := collector.GetStats()

fmt.Printf("Total records: %d\n", stats.TotalRecords)
fmt.Printf("Error records: %d\n", stats.ErrorRecords)
fmt.Printf("Active series: %d\n", stats.ActiveSeries)
fmt.Printf("Uptime: %v\n", time.Since(stats.StartTime))

// Type breakdown
for metricType, count := range stats.TypeBreakdown {
    fmt.Printf("%s: %d\n", metricType, count)
}
```

For OpenTelemetry tracer:

```go
stats := tracer.GetStats()

fmt.Printf("Total spans: %d\n", stats.TotalSpans)
fmt.Printf("Error spans: %d\n", stats.ErrorSpans)
fmt.Printf("Average duration: %v\n", stats.AverageDuration)
```

## Integration with Plumego

### Basic Setup

```go
import (
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/metrics"
    "github.com/spcent/plumego/middleware"
)

app := core.New(
    core.WithAddr(":8080"),
)

// Create metrics collector
collector := metrics.NewPrometheusCollector("myapp")

// Expose metrics endpoint
app.Get("/metrics", collector.Handler().ServeHTTP)

// Use in middleware
app.Use(middleware.Metrics(collector))
```

### Custom Middleware

```go
func MetricsMiddleware(collector metrics.MetricsCollector) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            timer := metrics.NewTimer()

            // Wrap response writer to capture status and bytes
            rw := &responseWriter{ResponseWriter: w, status: 200}

            next.ServeHTTP(rw, r)

            collector.ObserveHTTP(
                r.Context(),
                r.Method,
                r.URL.Path,
                rw.status,
                rw.bytes,
                timer.Elapsed(),
            )
        })
    }
}
```

## Best Practices

1. **Choose the Right Collector**
   - Production: `PrometheusCollector` or `OpenTelemetryTracer`
   - Testing: `BaseMetricsCollector` or `NoopCollector`
   - Disabled metrics: `NoopCollector` (zero overhead)

2. **Set Memory Limits**
   ```go
   collector := metrics.NewPrometheusCollector("myapp").
       WithMaxMemory(10000) // Prevent unbounded growth
   ```

3. **Use Appropriate Labels**
   - Keep cardinality low (avoid user IDs, session IDs)
   - Use consistent label names across metrics
   - Avoid high-cardinality values

4. **Clear Periodically**
   ```go
   // In long-running applications
   ticker := time.NewTicker(24 * time.Hour)
   go func() {
       for range ticker.C {
           collector.Clear()
       }
   }()
   ```

5. **Use MultiCollector for Multiple Backends**
   ```go
   multi := metrics.NewMultiCollector(
       metrics.NewPrometheusCollector("myapp"),
       metrics.NewOpenTelemetryTracer("myapp"),
   )
   ```

6. **Measure Critical Paths**
   ```go
   err := metrics.MeasureFunc(ctx, collector, "db_query", "users", func() error {
       return db.Query(...)
   })
   ```

## Performance

### Overhead

- **NoopCollector**: ~0-5 ns/op (effectively zero)
- **BaseMetricsCollector**: ~500-1000 ns/op
- **PrometheusCollector**: ~1000-2000 ns/op
- **OpenTelemetryTracer**: ~2000-3000 ns/op

### Benchmarks

```bash
go test -bench=. -benchmem ./metrics/...
```

## Testing

### Unit Tests

```bash
go test ./metrics/...
```

### Race Detection

```bash
go test -race ./metrics/...
```

### Coverage

```bash
go test -cover ./metrics/...
```

## Examples

See [example_test.go](example_test.go) for comprehensive runnable examples:

- Basic Prometheus usage
- OpenTelemetry tracing
- Timer usage
- MeasureFunc with different metric types
- MultiCollector setup
- Memory management
- Statistics collection

## API Reference

See [doc.go](doc.go) for complete package documentation.

## Thread Safety

All collectors are safe for concurrent use from multiple goroutines. Internal synchronization is handled automatically with minimal lock contention.

## License

Part of the Plumego framework. See the main repository for license details.

## Contributing

Contributions are welcome! Please see the main Plumego repository for contribution guidelines.
