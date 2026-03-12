# Plumego — Web Toolkit Based on Go Standard Library

Plumego is a lightweight Go HTTP toolkit built around `net/http`.
It focuses on explicit routing, middleware composition, lifecycle safety, and optional components (webhook, websocket, observability, tenant, etc.).

## Highlights
- Router with groups, params, and route metadata/reverse routing support.
- Middleware chain with explicit order (`func(http.Handler) http.Handler`).
- Lifecycle hooks for components/runners and graceful shutdown.
- Optional observability adapters (Prometheus/OpenTelemetry).
- Standard-library compatibility end-to-end.

## Components
`core.App` orchestrates pluggable components:

```go
type Component interface {
    RegisterRoutes(r *router.Router)
    RegisterMiddleware(m *middleware.Registry)
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Health() (name string, status health.HealthStatus)
}
```

Mount components via `core.WithComponent(...)` / `core.WithComponents(...)`.

## Quick Start
```go
package main

import (
    "context"
    "log"
    "net/http"

    "github.com/spcent/plumego/core"
    plumelog "github.com/spcent/plumego/log"
    "github.com/spcent/plumego/middleware/cors"
    "github.com/spcent/plumego/middleware/observability"
    "github.com/spcent/plumego/middleware/recovery"
)

func main() {
    ctx := context.Background()
    app := core.New(
        core.WithAddr(":8080"),
        core.WithDebug(),
        core.WithDevTools(),
        core.WithLogger(plumelog.NewGLogger()),
    )

    if err := app.Use(
        observability.RequestID(),
        observability.Tracing(nil),
        observability.HTTPMetrics(nil),
        observability.AccessLog(app.Logger()),
        recovery.Recovery(app.Logger()),
        cors.CORS,
    ); err != nil {
        log.Fatalf("register middleware: %v", err)
    }

    if err := app.AddRoute(http.MethodGet, "/ping", http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
        w.Write([]byte("pong"))
    })); err != nil {
        log.Fatalf("register route: %v", err)
    }

    if err := app.Prepare(); err != nil {
        log.Fatalf("prepare app: %v", err)
    }
    if err := app.Start(ctx); err != nil {
        log.Fatalf("start runtime: %v", err)
    }
    srv, err := app.Server()
    if err != nil {
        log.Fatalf("build server: %v", err)
    }
    defer app.Shutdown(ctx)

    log.Fatal(srv.ListenAndServe())
}
```

## Configuration Basics
- `.env` loading should happen explicitly in `main`; `core.WithEnvPath(...)` only records the path for components that need it.
- Address / server timeouts / graceful shutdown / header size / TLS / HTTP2 via `core.With...` options.
- Metrics/tracing hooks via `core.WithMetricsCollector(...)` and `core.WithTracer(...)`.

## Key Runtime Pieces
- Router: `app.Get/Post/Put/Delete/Patch/Any`; advanced operations from `app.Router()`.
- Middleware: register explicitly with `app.Use(...)`, group scope with `group.Use(...)`.
- Contract helpers: `contract.WriteError`, `contract.WriteResponse`, `contract.AdaptCtxHandler`.
- WebSocket: `app.ConfigureWebSocket()` / `app.ConfigureWebSocketWithOptions(...)`.
- Health: `health.ReadinessHandler(manager)` and `health.BuildInfoHandler()`.

## Reference App
`examples/reference` wires a production-like setup:
- Request ID + logging + recovery + CORS middleware chain.
- Optional metrics/tracing collector/tracer injection.
- WebSocket setup with explicit config.
- Webhook components mounted through `core.WithComponent(...)`.

Run:

```bash
go run ./examples/reference
```

## Development
- `go test -timeout 20s ./...`
- `go vet ./...`
- `gofmt -w .`
