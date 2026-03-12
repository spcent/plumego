# Health Module

> **Package Path**: `github.com/spcent/plumego/health` | **Stability**: High | **Priority**: P1

The `health` package provides explicit probe and diagnostics handlers plus a `HealthManager` for liveness, readiness, component checks, health history, and build/runtime diagnostics.

## Canonical Quick Start

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
    "github.com/spcent/plumego/health"
)

func main() {
    manager, err := health.NewHealthManager(health.HealthCheckConfig{})
    if err != nil {
        log.Fatal(err)
    }
    defer manager.Close()

    app := core.New(
        core.WithAddr(":8080"),
        core.WithHealthManager(manager),
    )

    if err := app.Router().AddRoute(http.MethodGet, "/health/live", health.LiveHandler()); err != nil {
        log.Fatal(err)
    }
    if err := app.Router().AddRoute(http.MethodGet, "/health/ready", health.ReadinessHandler(manager)); err != nil {
        log.Fatal(err)
    }
    if err := app.Router().AddRoute(http.MethodGet, "/health", health.SummaryHandler(manager)); err != nil {
        log.Fatal(err)
    }
    if err := app.Router().AddRoute(http.MethodGet, "/health/build", health.BuildInfoHandler()); err != nil {
        log.Fatal(err)
    }
    if err := app.Router().AddRoute(http.MethodGet, "/health/runtime", health.RuntimeInfoHandler()); err != nil {
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

## Handler Matrix

- `health.LiveHandler()`
  Returns `200` with `alive` body.
- `health.ReadinessHandler(manager)`
  Returns `200` when `manager.Readiness().Ready` is true, otherwise `503`.
- `health.ReadinessHandlerWithManager(manager)`
  Recomputes component health and derives readiness from aggregate status.
- `health.SummaryHandler(manager)`
  Returns aggregate component health only.
- `health.DetailedHandler(manager)`
  Returns aggregate health plus build metadata.
- `health.HealthHandler(manager, debug)`
  Legacy convenience wrapper around `DetailedHandler` / runtime-inclusive diagnostics.
- `health.BuildInfoHandler()`
  Returns version, commit, and build time metadata.
- `health.RuntimeInfoHandler()`
  Returns Go runtime diagnostics only.
- `health.ComponentHealthHandler(manager, name)`
  Returns the latest status for a single component.
- `health.AllComponentsHealthHandler(manager)`
  Returns the current health map for all registered components.

## Registering Component Checks

Implement `health.ComponentChecker` and register it with the manager.

```go
package main

import (
    "context"
    "database/sql"

    "github.com/spcent/plumego/health"
)

type dbChecker struct {
    db *sql.DB
}

func (c dbChecker) Name() string { return "database" }

func (c dbChecker) Check(ctx context.Context) error {
    return c.db.PingContext(ctx)
}

func wire(manager health.HealthManager, db *sql.DB) error {
    return manager.RegisterComponent(dbChecker{db: db})
}
```

## Core Lifecycle Integration

When a `HealthManager` is attached with `core.WithHealthManager(...)`, core keeps readiness aligned with lifecycle:

- after successful startup: `MarkReady()`
- during shutdown/drain: `MarkNotReady(...)`

That makes readiness probes reflect actual serving state instead of a separate flag path.
