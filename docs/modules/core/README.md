# Core Module

> **Package**: `github.com/spcent/plumego/core`  
> **Stability**: High (v1 GA)

The `core` package is Plumego's application entrypoint. It manages app construction, route registration, middleware composition, component/runner lifecycle, and server boot.

---

## What `core` Owns

- App construction via `core.New(...)`
- HTTP route registration (`Get`, `Post`, `Put`, `Delete`, `Patch`, `Any`)
- Middleware registration via `app.Use(...)`
- Component lifecycle (`WithComponent`, `WithComponents`)
- Runner lifecycle (`WithRunner`, `WithRunners`)
- Graceful shutdown hooks (`WithShutdownHook`, `WithShutdownHooks`)
- Server boot (`app.Boot()`)

`core` stays transport-first and remains compatible with `net/http`.

---

## Quick Start (Canonical)

```go
package main

import (
    "log"
    "net/http"

    "github.com/spcent/plumego/core"
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

---

## Lifecycle Model

1. Construct app with `core.New(options...)`.
2. Register middleware and routes while app is mutable.
3. Boot app with `app.Boot()`.
4. During shutdown, hooks/components/runners are stopped in managed order.

After boot starts, mutating operations (adding routes/middleware/runners/hooks) are rejected.

---

## Core Interfaces

### `Component`

```go
type Component interface {
    RegisterRoutes(r *router.Router)
    RegisterMiddleware(m *middleware.Registry)
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Health() (name string, status health.HealthStatus)
}
```

### `Runner`

```go
type Runner interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

### `ShutdownHook`

```go
type ShutdownHook func(context.Context) error
```

---

## Route Registration

Use standard-library handler shape:

```go
app.Get("/users", listUsers)
app.Post("/users", createUser)
app.Delete("/users/:id", deleteUser)
```

For strict boot-time wiring (fail on duplicate/frozen registration), use explicit error-return APIs:

```go
if err := app.AddRoute(http.MethodGet, "/users/:id", http.HandlerFunc(getUser)); err != nil {
    log.Fatalf("register route: %v", err)
}
if err := app.AddRouteWithName(http.MethodGet, "/users/:id", "users.show", http.HandlerFunc(getUser)); err != nil {
    log.Fatalf("register named route: %v", err)
}
```

For advanced routing (groups, route metadata, reverse routing), use `app.Router()` and the `router` module directly.

---

## Middleware Registration

Use explicit order with `app.Use(...)` before boot:

```go
if err := app.Use(mw1, mw2, mw3); err != nil {
    // handle registration error
}
```

Recommended baseline order:
1. request ID
2. logging/tracing
3. recovery
4. security/limits/cors (as needed)

---

## Configuration Options

See full option signatures and examples in:

- [configuration-options.md](configuration-options.md)

Common options:
- `WithAddr`
- `WithEnvPath`
- `WithServerTimeouts`
- `WithMaxHeaderBytes`
- `WithShutdownTimeout`
- `WithHTTP2`
- `WithTLS` / `WithTLSConfig`
- `WithDebug`
- `WithLogger`
- `WithMethodNotAllowed`
- `WithComponent` / `WithComponents`
- `WithRunner` / `WithRunners`
- `WithShutdownHook` / `WithShutdownHooks`
- `WithMetricsCollector`
- `WithTracer`
- `WithHealthManager`
- `WithRouter`

---

## Compatibility Notes (v1)

The following historical option names are not part of the current core v1 API:

- `WithRecommendedMiddleware`
- `WithRequestID`
- `WithLogging`
- `WithRecovery`
- `WithCORS`
- `WithTenantConfigManager`
- `WithTenantMiddleware`
- `WithMiddlewareChain`
- `WithServer`

Use explicit `app.Use(...)` and router-group middleware composition instead.
