# Readiness Probes

> **Package**: `github.com/spcent/plumego/health` | **Purpose**: traffic admission control

Readiness decides whether an instance should receive traffic. It should follow startup completion and shutdown/drain state, not just process aliveness.

## Canonical Wiring

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

    if err := app.Router().AddRoute(http.MethodGet, "/health/ready", opshealth.ReadinessHandler(manager)); err != nil {
        log.Fatal(err)
    }
    if err := app.Router().AddRoute(http.MethodGet, "/health/ready-checks", opshealth.ReadinessHandlerWithManager(manager)); err != nil {
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

## Semantics

`opshealth.ReadinessHandler(manager)`:

- returns `200` when `manager.Readiness().Ready == true`
- returns `503` when readiness is false

`opshealth.ReadinessHandlerWithManager(manager)`:

- runs `CheckAllComponents(...)`
- returns `200` for aggregate `healthy` or `degraded`
- returns `503` for aggregate `unhealthy`

## Manual state transitions

```go
manager.MarkNotReady("warming cache")
// warmup logic
manager.MarkReady()
```

Use this only when your startup sequence has steps outside normal core lifecycle management.

## Kubernetes example

```yaml
readinessProbe:
  httpGet:
    path: /health/ready
    port: 8080
  periodSeconds: 5
  timeoutSeconds: 2
  failureThreshold: 2
```
