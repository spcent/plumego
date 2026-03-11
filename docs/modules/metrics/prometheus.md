# Prometheus Collector

> **Package**: `github.com/spcent/plumego/metrics` | **Integration**: Prometheus-compatible text exposition

## Overview

`metrics.NewPrometheusCollector(namespace)` provides an in-memory collector with a built-in `http.Handler` for Prometheus scraping.

It does not register itself with `core`. Mount the handler explicitly.

## Canonical Setup

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
    collector := metrics.NewPrometheusCollector("plumego").WithMaxMemory(10000)

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

## What It Emits

Current built-in HTTP output includes:

- `<namespace>_http_requests_total`
- `<namespace>_http_request_duration_seconds_sum`
- `<namespace>_http_request_duration_seconds_count`
- `<namespace>_http_request_duration_seconds_min`
- `<namespace>_http_request_duration_seconds_max`
- `<namespace>_uptime_seconds`
- `<namespace>_http_requests_total_all`

Additional SMS gateway metrics are emitted when corresponding records are observed.

## Recording Data

```go
collector.ObserveHTTP(ctx, "GET", "/api/users", 200, 1024, 35*time.Millisecond)
collector.ObserveDB(ctx, "query", "postgres", "select_users", 10, 8*time.Millisecond, nil)
collector.ObserveKV(ctx, "get", "user:42", 2*time.Millisecond, nil, true)
```

## Operational Notes

- `WithMaxMemory(n)` bounds the number of active metric series.
- `GetStats()` returns `TotalRecords`, `ErrorRecords`, `ActiveSeries`, `StartTime`, and optional type breakdowns.
- `Clear()` resets all in-memory state.

## Prometheus scrape config

```yaml
scrape_configs:
  - job_name: 'plumego'
    static_configs:
      - targets: ['localhost:8080']
```
