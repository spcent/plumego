# Middleware module

Plumego’s middleware wraps standard `http.Handler` values. Chain them globally with `app.Use(...)`, per group via `Group("/x", m1, m2)`, or per route using handler wrappers.

## Built-in middleware
- **Recovery**: `core.WithRecovery()` catches panics, logs stack traces, returns JSON errors.
- **Logging**: `core.WithLogging()` records request/response metadata and plugs into metrics/tracing collectors provided through `core.WithMetricsCollector` and `core.WithTracer`.
- **Request ID**: `middleware.RequestID()` injects `X-Request-ID` and stores it in context for correlation.
- **CORS**: `core.WithCORS()` sets permissive defaults; customize with `core.WithCORSOptions(...)` or `middleware.CORSWithOptions(...)` (use `CORSWithOptionsFunc` for `http.HandlerFunc` return).
- **Gzip**: `middleware.Gzip()` compresses responses when clients send `Accept-Encoding: gzip`.
- **Timeout**: `middleware.Timeout(duration)` enforces per-request deadlines.
- **Body limit**: `middleware.BodyLimit(maxBytes, logger)` caps request size with a structured 413 response.
- **Concurrency limit**: `middleware.ConcurrencyLimit(maxConcurrent, queueDepth, queueTimeout, logger)` bounds parallel work and queueing.
- **Rate limit**: `middleware.RateLimit(ratePerSecond, burst, cleanupInterval, maxIdle)` applies an IP-based token bucket.
- **Auth helpers**: `middleware.SimpleAuth("token")` checks a bearer token header; `middleware.APIKey("X-API-Key", "secret")` validates custom headers.

Core wires protective defaults (body/concurrency limits) during setup. Call `app.Use(...)` for additional per-app chains; group-level middleware stacks are appended automatically.

## Composition examples
### Global guardrail chain
```go
app := core.New(core.WithRecovery(), core.WithLogging())
app.Use(
    middleware.Gzip(),
    middleware.Timeout(3*time.Second),
    middleware.ConcurrencyLimit(200, 400, 250*time.Millisecond, logger),
)
```

### Route-specific controls
```go
secured := app.Router().Group("/admin", middleware.SimpleAuth(os.Getenv("AUTH_TOKEN")))
secured.Get("/stats", middleware.TimeoutFunc(1*time.Second)(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("ok"))
}))
```

### Using `contract.Ctx` helpers
```go
app.GetCtx("/echo/:msg", middleware.WrapCtx(middleware.Timeout(2*time.Second), func(ctx *contract.Ctx) {
    _ = ctx.Response(http.StatusOK, map[string]any{"echo": ctx.Param("msg")}, nil)
}))
```

## Operational notes
- Chain order matters: loggers should typically wrap recoveries so panics include request metadata.
- Avoid global state in custom middleware; share dependencies through closures or application-scoped singletons.
- When enabling CORS or auth per group, keep public assets in a dedicated group to avoid unintended headers or auth checks.

## Where to look in the repo
- `middleware/` directory for implementations (recovery, logging, gzip, CORS, timeout, body limit, rate limit, auth, concurrency control).
- `examples/reference/main.go` for a realistic chain combining recovery, logging, CORS, and default safety limits.
