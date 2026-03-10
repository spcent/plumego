# Middleware Module

> **Package**: `github.com/spcent/plumego/middleware`

Middleware in Plumego is explicit, transport-layer only, and standard-library compatible.

---

## Canonical Type

```go
type Middleware func(http.Handler) http.Handler
```

---

## Registering Middleware with `core.App`

```go
app := core.New(core.WithAddr(":8080"))

if err := app.Use(
    observability.RequestID(),
    observability.Logging(app.Logger(), nil, nil),
    recovery.RecoveryMiddleware,
); err != nil {
    log.Fatalf("register middleware: %v", err)
}
```

Register before boot.

---

## Scope Options

### Global (`app.Use`)
Applies to all routes.

### Router / Group (`router.Use`)

```go
r := app.Router()
api := r.Group("/api")
api.Use(authMiddleware)
```

Applies only to that router/group subtree.

### Single route
Wrap one handler manually.

---

## Built-in Middleware (Common)

- request ID: `middleware/observability.RequestID(...)`
- structured logging: `middleware/observability.Logging(...)`
- panic recovery: `middleware/recovery.RecoveryMiddleware`
- CORS: `middleware/cors.CORS` / `middleware/cors.CORSWithOptions(...)`
- security headers: `middleware/security.SecurityHeaders(...)`
- abuse guard: `middleware/ratelimit.AbuseGuard(...)`
- token bucket limiter: `middleware/ratelimit.TokenBucket(...)`
- timeout: `middleware/timeout.Timeout(...)`
- gzip: `middleware/compression.Gzip(...)`

Use only the middleware needed by your transport boundary.

---

## Manual Composition

```go
h := middleware.Apply(
    http.HandlerFunc(finalHandler),
    observability.RequestID(),
    recovery.RecoveryMiddleware,
)

http.ListenAndServe(":8080", h)
```

---

## Ordering Guidance

Recommended baseline order:

1. request ID
2. logging/tracing
3. recovery
4. auth/rate/security headers/cors (by endpoint needs)

Keep ordering explicit in code and tests.

---

## `Registry` for Component Wiring

When components contribute middleware, use `middleware.Registry`:

```go
reg := middleware.NewRegistry()
reg.Use(observability.RequestID())
reg.Use(recovery.RecoveryMiddleware)

h := middleware.Apply(http.HandlerFunc(finalHandler), reg.Middlewares()...)
```

---

## Testing Middleware

Test middleware with `httptest` and assert:

- order of execution
- response/status/body/header behavior
- panic and error paths
- concurrency behavior (`go test -race`)
