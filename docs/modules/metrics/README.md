# Metrics Module

> **Package Path**: `github.com/spcent/plumego/metrics` | **Stability**: High | **Priority**: P1

## Overview

`metrics/` provides dependency-free collectors and helpers for request, storage, queue, IPC, and database metrics.

Current package surface is centered on explicit collectors:

- `metrics.MetricsCollector` is the shared interface.
- `metrics.NewPrometheusCollector(namespace)` provides an in-memory Prometheus-compatible HTTP exporter.
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
)

func main() {
    collector := metrics.NewPrometheusCollector("plumego")

    app := core.New(
        core.WithAddr(":8080"),
        core.WithMetricsCollector(collector),
    )

    if err := app.Router().AddRoute(http.MethodGet, "/metrics", collector.Handler()); err != nil {
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
err := metrics.MeasureFunc(ctx, collector, "db_query", "users", func() error {
    return repo.ListUsers(ctx)
})
```

## Collector Model

### `MetricsCollector`

All collectors implement:

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

### Prometheus collector

Use when you want:

- a scrapeable `/metrics` endpoint
- no third-party dependency
- bounded in-memory series retention

```go
collector := metrics.NewPrometheusCollector("plumego").WithMaxMemory(20000)
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

- `core.WithMetricsCollector(...)` stores the collector on the app for components and middleware that need it.
- `core/components/observability` can wire `/metrics` and tracing when you opt into that component-level configuration.
- `health.AttachMetrics(...)` bridges health checks into the health metrics collector.

## Related Documentation

- [Prometheus](prometheus.md) - scrape endpoint wiring
- [OpenTelemetry](opentelemetry.md) - tracing support
- [Core Module](../core/README.md) - app lifecycle and options
