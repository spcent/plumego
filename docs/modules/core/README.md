# Core Module

> **Package**: `github.com/spcent/plumego/core`  
> **Stability**: High (v1 GA)

The `core` package is Plumego's application entrypoint. It manages app construction, route registration, middleware composition, component/runner lifecycle, and explicit server preparation/startup.

---

## What `core` Owns

- App construction via `core.New(...)`
- HTTP route registration (`Get`, `Post`, `Put`, `Delete`, `Patch`, `Any`)
- Middleware registration via `app.Use(...)`
- Component lifecycle (`WithComponent`, `WithComponents`)
- Runner lifecycle (`WithRunner`, `WithRunners`)
- Graceful shutdown hooks (`WithShutdownHook`, `WithShutdownHooks`)
- Explicit lifecycle (`app.Prepare()`, `app.Start(ctx)`, `app.Server()`, `app.Shutdown(ctx)`)

`core` stays transport-first and remains compatible with `net/http`.

---

## Quick Start (Canonical)

```go
package main

import (
    "context"
    "log"
    "net/http"

    "github.com/spcent/plumego/core"
    plumelog "github.com/spcent/plumego/log"
    "github.com/spcent/plumego/middleware/observability"
    "github.com/spcent/plumego/middleware/recovery"
    xdevtools "github.com/spcent/plumego/x/devtools"
)

func main() {
    app := core.New(
        core.WithAddr(":8080"),
        core.WithDebug(),
        core.WithLogger(plumelog.NewGLogger()),
    )

    if err := app.MountComponent(xdevtools.NewAppComponent(app)); err != nil {
        log.Fatalf("mount devtools: %v", err)
    }

    if err := app.Use(
        observability.RequestID(),
        observability.Tracing(nil),
        observability.HTTPMetrics(nil),
        observability.AccessLog(app.Logger()),
        recovery.Recovery(app.Logger()),
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
    if err := app.Start(context.Background()); err != nil {
        log.Fatalf("start runtime: %v", err)
    }
    srv, err := app.Server()
    if err != nil {
        log.Fatalf("build server: %v", err)
    }
    defer app.Shutdown(context.Background())

    log.Fatal(srv.ListenAndServe())
}
```

---

## Lifecycle Model

1. Construct app with `core.New(options...)`.
2. Register middleware and routes while app is mutable.
3. Prepare app with `app.Prepare()`.
4. Start runtime hooks with `app.Start(ctx)`.
5. Serve using the prepared `*http.Server` or `app.Run(ctx)`.
6. During shutdown, call `app.Shutdown(ctx)` to stop server, runners, components, and hooks.

After boot starts, mutating operations (adding routes/middleware/runners/hooks) are rejected.

`core.New(...)` defaults to `log.NewNoOpLogger()`. Inject a real logger with `core.WithLogger(...)` before wiring logging middleware around `app.Logger()`.

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

Use explicit order with `app.Use(...)` before prepare:

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
- `WithPrometheusCollector`
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
