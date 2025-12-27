# Middleware module

The **middleware** package provides cross-cutting handlers for recovery, logging, CORS, compression, timeouts, rate/concurrency limits, body size enforcement, auth helpers, and metrics/tracing adapters.

## Responsibilities
- Wrap standard `http.Handler` values; each middleware implements `func(http.Handler) http.Handler`.
- Register global middleware with `app.Use(...)` before boot, or attach to groups/routes through the router API.
- Provide pluggable collectors (`middleware.MetricsCollector`) and tracers (`middleware.Tracer`) for observability backends.

## Built-in options
- **Recovery & logging**: enable via `app.EnableRecovery()` and `app.EnableLogging()`; logging injects request metadata and trace IDs when a tracer is configured.
- **CORS**: `app.EnableCORS()` with sensible defaults; customize origins/headers in options or by wrapping your own handler.
- **Guards**: concurrency limiter, queueing, timeouts, gzip, body size limit, and simple token auth (see `middleware/simple_auth.go`).
- **Metrics**: attach Prometheus collector or OpenTelemetry tracer through `core.WithMetricsCollector` / `core.WithTracer`.

## Composition rules
- Order matters: recovery/logging should wrap outer layers; tighter controls (timeouts, auth) can sit closer to handlers.
- Group middleware is inherited by child routes, so prefer composing at boundary groups (e.g., `/api/admin`).
- Avoid global mutable state; pass configuration via closures or constructor functions.

## Custom middleware template
```go
func NewMaskingLogger() middleware.Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // redact authorization header
            r.Header.Del("Authorization")
            next.ServeHTTP(w, r)
        })
    }
}
```
Register it with `app.Use(NewMaskingLogger())` or on a router group.

## Troubleshooting
- **Duplicate writes**: ensure your middleware respects `http.ResponseWriter` semantics; wrap once and avoid writing after `next.ServeHTTP` returns when status is final.
- **Timeouts not firing**: verify the timeout middleware is placed before long-running handlers and that handlers propagate `context.Context` properly.
- **Metrics missing**: confirm a collector/tracer is supplied during `core.New`; logging middleware emits spans only when a tracer is present.

## Where to look in the repo
- `middleware` directory for concrete implementations (recovery, logging, cors, gzip, timeout, limiters, auth).
- `examples/reference/main.go` for the default middleware stack used by the demo app.
