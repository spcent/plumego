# Readiness Probes

> **Package**: `github.com/spcent/plumego/health` | **Purpose**: Detect whether service can receive traffic

Readiness controls traffic admission. If readiness fails, orchestrators remove the instance from load balancing without restarting the process.

## Canonical Wiring

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

    // Readiness state endpoint (MarkReady/MarkNotReady based)
    app.Get("/health/ready", health.ReadinessHandler(manager).ServeHTTP)

    // Optional: derive readiness from component checks
    app.Get("/health/ready-checks", health.ReadinessHandlerWithManager(manager).ServeHTTP)

    if err := app.Boot(); err != nil {
        log.Fatal(err)
    }
}
```

## Readiness Semantics

`ReadinessHandler(manager)`:

- `200` when `manager.Readiness().Ready == true`
- `503` when `Ready == false`

`ReadinessHandlerWithManager(manager)`:

- Runs `CheckAllComponents(...)`
- Returns `200` for aggregate `healthy/degraded`
- Returns `503` for aggregate `unhealthy`

## Component-Based Readiness

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

func register(manager health.HealthManager, db *sql.DB) error {
    return manager.RegisterComponent(dbChecker{db: db})
}
```

With this checker registered, `ReadinessHandlerWithManager` reflects database availability.

## Manual State Control (Optional)

```go
manager.MarkNotReady("warming cache")
// warmup...
manager.MarkReady()
```

This is useful for custom startup sequencing outside core lifecycle integration.

## Kubernetes Example

```yaml
readinessProbe:
  httpGet:
    path: /health/ready
    port: 8080
  initialDelaySeconds: 5
  periodSeconds: 5
  timeoutSeconds: 2
  failureThreshold: 2
```

## Notes

- Use readiness for dependency and initialization gating.
- Keep liveness independent from dependency checks.
- In core apps, prefer mounting manager with `core.WithHealthManager(...)` so readiness follows boot/shutdown lifecycle.
