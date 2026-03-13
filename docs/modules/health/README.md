# Health Module

> **Package Path**: `github.com/spcent/plumego/health` | **Stability**: High | **Priority**: P1

The `health` package provides the `HealthManager`, readiness state, component checks, history, and build metadata primitives.
HTTP probe and diagnostics handlers now live in `x/ops/healthhttp`.

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
    opshealth "github.com/spcent/plumego/x/ops/healthhttp"
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

    if err := app.Router().AddRoute(http.MethodGet, "/health/live", opshealth.LiveHandler()); err != nil {
        log.Fatal(err)
    }
    if err := app.Router().AddRoute(http.MethodGet, "/health/ready", opshealth.ReadinessHandler(manager)); err != nil {
        log.Fatal(err)
    }
    if err := app.Router().AddRoute(http.MethodGet, "/health", opshealth.SummaryHandler(manager)); err != nil {
        log.Fatal(err)
    }
    if err := app.Router().AddRoute(http.MethodGet, "/health/build", opshealth.BuildInfoHandler()); err != nil {
        log.Fatal(err)
    }
    if err := app.Router().AddRoute(http.MethodGet, "/health/runtime", opshealth.RuntimeInfoHandler()); err != nil {
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

## HTTP Handler Matrix

Handlers are extension-owned in `github.com/spcent/plumego/x/ops/healthhttp`.

- `opshealth.LiveHandler()`
  Returns `200` with `alive` body.
- `opshealth.ReadinessHandler(manager)`
  Returns `200` when `manager.Readiness().Ready` is true, otherwise `503`.
- `opshealth.ReadinessHandlerWithManager(manager)`
  Recomputes component health and derives readiness from aggregate status.
- `opshealth.SummaryHandler(manager)`
  Returns aggregate component health only.
- `opshealth.DetailedHandler(manager)`
  Returns aggregate health plus build metadata.
- `opshealth.HealthHandler(manager, debug)`
  Legacy convenience wrapper around `DetailedHandler` / runtime-inclusive diagnostics.
- `opshealth.BuildInfoHandler()`
  Returns version, commit, and build time metadata.
- `opshealth.RuntimeInfoHandler()`
  Returns Go runtime diagnostics only.
- `opshealth.ComponentHealthHandler(manager, name)`
  Returns the latest status for a single component.
- `opshealth.AllComponentsHealthHandler(manager)`
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
