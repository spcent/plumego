# Compatibility Guide

This document describes how Plumego integrates with the Go standard library, existing middleware, and common ecosystem patterns.

## Standard Library Compatibility

Plumego is built **on** the standard library, not **around** it. Every component respects `net/http` conventions and can interoperate with stdlib-only code.

### Handlers

Plumego handlers are ordinary `http.Handler` or `http.HandlerFunc`. There is no special handler type.

```go
// This is a valid Plumego handler — it is standard `net/http`
app.Get("/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    id := router.Param(r.Context(), "id")
    w.WriteHeader(http.StatusOK)
    fmt.Fprintf(w, "User: %s", id)
}))
```

### Middleware

All Plumego middleware have the signature `func(http.Handler) http.Handler`. There is no custom middleware type.

```go
// This middleware works in Plumego and any other stdlib-compatible system
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("%s %s", r.Method, r.RequestURI)
        next.ServeHTTP(w, r)
    })
}

// Use it like any other Plumego middleware
app.Use(LoggingMiddleware)
```

### Response Writers and Requests

Plumego operates on standard `http.ResponseWriter` and `*http.Request`. Response helpers are pure functions that write to the response writer.

```go
// contract.WriteResponse is a pure function: takes an http.ResponseWriter
contract.WriteResponse(w, r, http.StatusOK, data, nil)

// You can write to the response writer directly
w.WriteHeader(http.StatusCreated)
fmt.Fprintf(w, "Created")
```

### Server Lifecycle

The Plumego app is a standard `http.Handler`. You can run it with any stdlib server or custom server setup.

```go
// Plumego app is an http.Handler
app := core.New(config, logger)
app.Get("/ping", pingHandler)

// Run with stdlib http.ListenAndServe
http.ListenAndServe(":8080", app)

// Or with *http.Server for custom lifecycle
server := &http.Server{
    Addr:    ":8080",
    Handler: app,
}
server.ListenAndServe()
```

---

## Middleware Compatibility

Plumego middleware works with any third-party middleware that respects the `func(http.Handler) http.Handler` signature.

### Stdlib-First Middleware

Plumego bundles first-party middleware in the `middleware` package. All are compatible with any `http.Handler` stack:

```go
import (
    "github.com/spcent/plumego/middleware"
)

app := core.New(config, logger)

// Mix Plumego and third-party middleware freely
app.Use(middleware.RequestID())
app.Use(myCustomMiddleware())
app.Use(middleware.Logging(logger))
```

### Third-Party Middleware

Any middleware written for `net/http` works with Plumego:

```go
// Third-party middleware examples
app.Use(someThirdPartyMiddleware)
app.Use(cors.Default())  // github.com/rs/cors
app.Use(csrf.Protect([]byte("secret")))  // gorilla/csrf
```

### Middleware Ordering

Plumego `Use()` attaches middleware in order; the first middleware attached is the outermost (first to see the request):

```go
app.Use(middleware.RequestID())      // 1. outermost
app.Use(myCustomMiddleware())        // 2. middle
app.Use(middleware.Logging(logger))  // 3. innermost

// Request flow: RequestID → custom → Logging → handler
```

This is the standard stdlib convention used by all `func(http.Handler) http.Handler` stacks.

---

## Logging Compatibility

Plumego is **logger-neutral**. The `log` package defines a minimal interface; you bring your own implementation.

### Logger Interface

```go
// log.Logger is a minimal, stdlib-friendly interface
type Logger interface {
    Info(ctx context.Context, msg string, fields ...field.Field)
    Error(ctx context.Context, msg string, fields ...field.Field)
}
```

### Bring Your Own Logger

Plumego ships with a basic implementation but **does not** require it. Inject any logger that matches the interface:

```go
import (
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/log"
)

cfg := core.DefaultConfig()

// Use Plumego's basic logger
logger := log.NewLogger(os.Stderr)

// OR use your own (as long as it matches the log.Logger interface)
logger := myZerologAdapter    // or slog, logrus, zap, etc.

app := core.New(cfg, logger)
```

### Structured Logging

Plumego uses structured logging with field helpers:

```go
import "github.com/spcent/plumego/log/field"

logger.Info(r.Context(), "user_login", 
    field.String("user_id", userID),
    field.Int("status_code", 200),
)
```

This approach is compatible with any structured logger you wrap.

---

## Metrics Compatibility

Plumego is **metrics-neutral**. The `metrics` package defines minimal interfaces; you bring your own backend.

### Metrics Interface

```go
// metrics.Collector is the minimal interface
type Collector interface {
    Counter(name string) metrics.Counter
    Gauge(name string) metrics.Gauge
    Timer(name string) metrics.Timer
}
```

### Bring Your Own Metrics

Plumego ships with an in-memory collector for testing but **does not** require it. Implement the interface for your backend:

```go
import "github.com/spcent/plumego/metrics"

// Inject your implementation
var collector metrics.Collector = myPrometheusAdapter

// OR use Plumego's test collector
collector := metrics.NewMemoryCollector()
```

### Exemplar Patterns

Integrate with Prometheus, Datadog, or your observability stack:

```go
// x/observability exports adapters for common stacks
import "github.com/spcent/plumego/x/observability"

prometheus := observability.PrometheusAdapter()
app.RegisterMetricsCollector(prometheus)
```

---

## Config Neutrality

Plumego **does not** prescribe a config format or library. Use whatever works for your project.

### Core Config

The core package defines a minimal `Config` struct for server setup:

```go
cfg := core.DefaultConfig()
cfg.Addr = ":9090"
cfg.ReadTimeout = 10 * time.Second
app := core.New(cfg, logger)
```

### Application Config

Your application config is completely separate; manage it as you see fit:

```go
// Your config, your choice of library
type AppConfig struct {
    DatabaseURL string
    JWTSecret   string
    CacheTTL    time.Duration
}

cfg := AppConfig{}
// Load from env, YAML, TOML, or anywhere else
// Then use it in your handlers
```

---

## Migration and Exit Strategy

### Migrating To Plumego

If you are coming from another framework:

1. Start with `reference/standard-service` as your app shape.
2. Follow the migration guide for your current framework:
   - [`docs/guides/migration/from-chi.md`](./docs/guides/migration/from-chi.md)
   - [`docs/guides/migration/from-gin.md`](./docs/guides/migration/from-gin.md)
   - [`docs/guides/migration/from-echo.md`](./docs/guides/migration/from-echo.md)
   - [`docs/guides/migration/middleware-compat.md`](./docs/guides/migration/middleware-compat.md)

3. Update your handlers to use standard `http.HandlerFunc`.
4. Replace framework-specific routing with Plumego `router`.
5. Rewrite middleware to match `func(http.Handler) http.Handler`.

All Plumego components are stdlib-compatible, so the migration is straightforward.

### Exiting Plumego (If Needed)

Because Plumego is built on stdlib, exiting to raw `net/http` or another framework is straightforward:

1. **Handlers:** Plumego handlers are already `http.Handler` or `http.HandlerFunc`; no changes needed.
2. **Middleware:** All Plumego middleware follow `func(http.Handler) http.Handler`; works with any stdlib stack.
3. **Response helpers:** Replace `contract.WriteResponse` with direct `w.WriteHeader()` and `json.NewEncoder(w).Encode()` calls.
4. **Routing:** Rewrite route registration using your target framework's routing API.
5. **Lifecycle:** Use your target framework's server setup instead of `core.New` and `app.Shutdown`.

**Key point:** No vendor lock-in. Plumego handlers and middleware are pure stdlib, so you are not trapped.

---

## Storage Layer Compatibility

Plumego's `store` package defines contracts for common storage patterns. You provide the implementations.

### Storage Contracts

```go
// Plumego defines interfaces for common storage needs
type Cache interface { /* ... */ }
type KVStore interface { /* ... */ }
type FileStore interface { /* ... */ }
```

### Bring Your Own Implementation

Plumego ships with in-memory reference implementations; for production, implement the contract for your backend:

```go
// Implement the store interface for your database
type PostgresStore struct {
    db *sql.DB
}

func (s *PostgresStore) Get(ctx context.Context, key string) ([]byte, error) {
    // Your implementation
}

// Use it in your handlers
app.RegisterStore(myPostgresStore)
```

### x/data Adapters

The `x/data` extension family provides adapters for common storage backends:

```go
import "github.com/spcent/plumego/x/data/pgx"

pgStore := pgx.NewStore(pgConn)
app.RegisterStore(pgStore)
```

---

## Transport Compatibility

Plumego apps are standard `http.Handler` instances, compatible with any transport that wraps HTTP.

### Reverse Proxy

Use Plumego behind a reverse proxy without modification:

```
Nginx/Caddy/HAProxy → Plumego app (http.Handler)
```

### Service Mesh (Istio, etc.)

Plumego apps work seamlessly with service meshes:

```
Service Mesh Sidecar → Plumego app (http.Handler)
```

### Embedded in Larger Services

Plumego apps can be composed with other stdlib servers:

```go
mux := http.NewServeMux()
plumego_app := core.New(cfg, logger)

mux.Handle("/api/", plumego_app)      // Plumego API
mux.HandleFunc("/health", healthCheck) // Custom health check
mux.Handle("/metrics", metricsHandler)  // Metrics endpoint

http.ListenAndServe(":8080", mux)
```

---

## Testing Compatibility

Plumego apps work with the standard library testing tools (`net/http/httptest`):

```go
import "net/http/httptest"

app := core.New(config, logger)
app.Get("/ping", pongHandler)

// Use stdlib httptest; Plumego is an http.Handler
rec := httptest.NewRecorder()
req := httptest.NewRequest("GET", "/ping", nil)
app.ServeHTTP(rec, req)

if rec.Code != http.StatusOK {
    t.Fatalf("Expected 200, got %d", rec.Code)
}
```

---

## Observability Compatibility

### Tracing

Plumego apps carry `context.Context` through every handler and middleware. Use stdlib `context` for trace propagation:

```go
import "context"

ctx := r.Context()
ctx = context.WithValue(ctx, "trace_id", traceID)
// Propagate for tracing
```

### Metrics

Inject any metrics collector via the `metrics` interface. No framework coupling.

### Logging

Use the minimal `log.Logger` interface with your choice of underlying logger.

---

## Quick Compat Checklist

- ✅ **Handlers:** Ordinary `http.Handler` or `http.HandlerFunc`
- ✅ **Middleware:** Standard `func(http.Handler) http.Handler` signature
- ✅ **Response writers:** Direct `http.ResponseWriter` operations
- ✅ **Logger:** Bring your own; Plumego uses an interface, not a concrete type
- ✅ **Metrics:** Bring your own; Plumego uses an interface, not a concrete type
- ✅ **Config:** Completely your choice; Plumego does not prescribe
- ✅ **Storage:** Implement Plumego's storage interface for your backend
- ✅ **Exit:** All core components are stdlib-compatible; easy migration away

---

## Support and Questions

For compatibility questions:

1. Check if your component matches the `http.Handler` or middleware pattern.
2. Verify logger/metrics implement the minimal Plumego interfaces.
3. Consult `docs/guides/migration/` for examples from similar frameworks.
4. Open an issue on [GitHub](https://github.com/spcent/plumego).
