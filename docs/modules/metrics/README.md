# Metrics Module

> **Package Path**: `github.com/spcent/plumego/metrics` | **Stability**: High | **Priority**: P1

## Overview

`metrics/` provides dependency-free collectors and helpers for request, storage, queue, IPC, and database metrics.

Current package surface is centered on explicit collectors:

- `metrics.AggregateCollector` is the full collector interface.
- `metrics.NewPrometheusCollector(namespace)` provides the in-memory collector.
- `metrics.NewPrometheusExporter(collector)` provides the Prometheus HTTP exporter.
- `metrics.NewOpenTelemetryTracer(name)` provides tracing hooks compatible with the middleware observability layer.
- `metrics.NewNoopCollector()` is the zero-cost default when you want the call sites but no export.

There is no `core.WithPrometheusMetrics(...)` or `core.WithMetricsExporter(...)` option in the current API. Metrics are attached explicitly.

## Quick Start

### Expose a Prometheus endpoint

```go
package main

import (
    "context"
    "log"
    "net/http"
    "os/signal"
    "syscall"
    "time"

    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/metrics"
    "github.com/spcent/plumego/middleware/httpmetrics"
)

func main() {
    collector := metrics.NewPrometheusCollector("plumego")
    exporter := metrics.NewPrometheusExporter(collector)

    app := core.New(
        core.WithAddr(":8080"),
        core.WithPrometheusCollector(collector),
    )

    if err := app.Use(httpmetrics.Middleware(app.HTTPMetrics())); err != nil {
        log.Fatal(err)
    }

    if err := app.Router().AddRoute(http.MethodGet, "/metrics", exporter.Handler()); err != nil {
        log.Fatal(err)
    }

    if err := app.Prepare(); err != nil {
        log.Fatal(err)
    }

    ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
    defer stop()

    if err := app.Start(ctx); err != nil {
        log.Fatal(err)
    }

    srv, err := app.Server()
    if err != nil {
        log.Fatal(err)
    }

    go func() {
        <-ctx.Done()
        shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        _ = app.Shutdown(shutdownCtx)
    }()

    if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
        log.Fatal(err)
    }
}
```

### Record request metrics explicitly

```go
collector := metrics.NewPrometheusCollector("plumego")
collector.ObserveHTTP(context.Background(), "GET", "/users", 200, 512, 42*time.Millisecond)
```

### Time arbitrary work

```go
err := metrics.MeasureFunc(ctx, collector, "db_query", func() error {
    return repo.ListUsers(ctx)
})
```

For domain-specific measurement helpers, prefer the explicit variant that matches the dependency:

```go
err := metrics.MeasureKVFunc(ctx, collector, "get", "users:42", true, func() error {
    return repo.Get(ctx, "users:42")
})
```

## Collector Model

### `AggregateCollector`

Full collectors implement:

- `Record(ctx, record)`
- `ObserveHTTP(...)`
- `ObservePubSub(...)`
- `ObserveMQ(...)`
- `ObserveKV(...)`
- `ObserveIPC(...)`
- `ObserveDB(...)`
- `GetStats()`
- `Clear()`

This keeps instrumentation sites transport-agnostic and lets you swap exporters without changing call sites.

### `HTTPObserver`

Use `metrics.HTTPObserver` in transport middleware when you only need request metrics:

- `ObserveHTTP(...)`

This keeps HTTP-facing code from depending on MQ/KV/DB observer methods it never calls.

### Narrow observer interfaces

Use the smallest contract that matches the call site:

- `metrics.Recorder` for generic `Record(...)` instrumentation
- `metrics.PubSubObserver` for pub/sub metrics
- `metrics.MQObserver` for queue and worker metrics
- `metrics.KVObserver` for key-value metrics
- `metrics.IPCObserver` for IPC metrics
- `metrics.DBObserver` for database metrics

This keeps module boundaries explicit and stops storage, queue, and AI instrumentation from depending on unrelated metrics methods.

### Prometheus collector

Use when you want:

- a collector that feeds a scrapeable `/metrics` endpoint
- no third-party dependency
- bounded in-memory series retention

```go
collector := metrics.NewPrometheusCollector("plumego").WithMaxMemory(20000)
exporter := metrics.NewPrometheusExporter(collector)
```

### No-op collector

Use when you want instrumentation calls to remain in place but skip storage/export:

```go
collector := metrics.NewNoopCollector()
```

## Helpers

### Timer

```go
timer := metrics.NewTimer()
// ... work ...
duration := timer.Elapsed()
collector.ObserveHTTP(ctx, "POST", "/jobs", 202, 0, duration)
```

### Success / error helpers

```go
metrics.RecordSuccess(ctx, collector, "cache_warm", 150*time.Millisecond)
metrics.RecordError(ctx, collector, "cache_warm", 150*time.Millisecond, err)
```

## Integration Points

- `core.WithPrometheusCollector(...)` stores the Prometheus collector on the app.
- `core.WithHTTPMetrics(...)` stores a non-Prometheus HTTP observer on the app.
- `app.HTTPMetrics()` returns the current app-managed HTTP observer so middleware wiring stays explicit.
- `core/components/observability` can wire `/metrics` and tracing when you opt into that component-level configuration.
- `health.AttachMetricsTracker(...)` bridges health checks into the health metrics tracker.

## Related Documentation

- [Prometheus](prometheus.md) - scrape endpoint wiring
- [OpenTelemetry](opentelemetry.md) - tracing support
- [Core Module](../core/README.md) - app lifecycle and options
