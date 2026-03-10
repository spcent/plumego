# Health Module

> **Package Path**: `github.com/spcent/plumego/health` | **Stability**: High | **Priority**: P1

The `health` package provides standard HTTP handlers and a `HealthManager` for liveness, readiness, component health checks, and build/runtime diagnostics.

## Canonical Quick Start

```go
package main

import (
    "log"

    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/health"
)

func main() {
    manager, err := health.NewHealthManager(health.HealthCheckConfig{})
    if err != nil {
        log.Fatal(err)
    }

    app := core.New(
        core.WithAddr(":8080"),
        core.WithHealthManager(manager),
    )

    app.Get("/health/live", health.LiveHandler().ServeHTTP)
    app.Get("/health/ready", health.ReadinessHandler(manager).ServeHTTP)
    app.Get("/health", health.HealthHandler(manager, false).ServeHTTP)
    app.Get("/health/build", health.BuildInfoHandler().ServeHTTP)

    if err := app.Boot(); err != nil {
        log.Fatal(err)
    }
}
```

## Handler Matrix

- `health.LiveHandler()`
  - Always returns `200` with `alive` body.
  - Use for liveness probes.
- `health.ReadinessHandler(manager)`
  - Returns manager readiness (`200` when ready, `503` when not ready).
  - Works with explicit `MarkReady/MarkNotReady` state.
- `health.ReadinessHandlerWithManager(manager)`
  - Runs component checks and derives readiness from aggregate health.
- `health.HealthHandler(manager, debug)`
  - Full component health summary (`healthy/degraded/unhealthy`) plus build info.
  - With `debug=true`, includes runtime diagnostics.
- `health.BuildInfoHandler()`
  - Returns version/commit/build metadata.

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

When `HealthManager` is mounted via `core.WithHealthManager(...)`, core lifecycle updates readiness automatically:

- startup complete -> `MarkReady()`
- shutdown/drain path -> `MarkNotReady(...)`

This keeps readiness endpoints aligned with actual serving state.

## Related APIs

- Per-component: `health.ComponentHealthHandler(manager, "component")`
- All components: `health.AllComponentsHealthHandler(manager)`
- Component list: `health.ComponentsListHandler(manager)`
- History/debug: `health.HealthHistoryHandler(...)`, `health.DebugHealthHandler(...)`
- Metrics bridge: `health.AttachMetrics(...)`, `health.MetricsHandler(...)`
