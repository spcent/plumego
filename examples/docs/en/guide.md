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
    "log"
    "net/http"

    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/middleware/cors"
    "github.com/spcent/plumego/middleware/observability"
    "github.com/spcent/plumego/middleware/recovery"
)

func main() {
    app := core.New(
        core.WithAddr(":8080"),
        core.WithDebug(),
    )

    if err := app.Use(
        observability.RequestID(),
        observability.Logging(app.Logger(), nil, nil),
        recovery.RecoveryMiddleware,
        cors.CORS,
    ); err != nil {
        log.Fatalf("register middleware: %v", err)
    }

    app.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
        w.Write([]byte("pong"))
    })

    if err := app.Boot(); err != nil {
        log.Fatalf("server stopped: %v", err)
    }
}
```

## Configuration Basics
- `.env` loading: `core.WithEnvPath(...)`.
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
