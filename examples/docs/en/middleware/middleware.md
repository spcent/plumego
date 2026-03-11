# Middleware module

Plumego middleware is standard-library compatible (`func(http.Handler) http.Handler`).
Register globally with `app.Use(...)`, or scope to a router group via `group.Use(...)`.

## Common built-in middleware
- Request ID: `observability.RequestID()`
- Structured logging: `observability.Logging(app.Logger(), metricsCollector, tracer)`
- Panic recovery: `recovery.Recovery(logger)`
- CORS: `cors.CORS` / `cors.CORSWithOptions(...)`
- Gzip compression: `compression.Gzip()`
- Timeout: `timeout.Timeout(duration)`
- Body limit: `limits.BodyLimit(maxBytes, logger)`
- Concurrency limit: `limits.ConcurrencyLimit(maxConcurrent, queueDepth, queueTimeout, logger)`
- Auth helpers: `auth.SimpleAuth(token)`
- Security headers: `security.SecurityHeaders(nil)`
- Abuse guard / token bucket: `ratelimit.AbuseGuard(...)`, `ratelimit.TokenBucket(...)`

## Global chain example
```go
app := core.New(core.WithAddr(":8080"))

if err := app.Use(
    observability.RequestID(),
    observability.Logging(app.Logger(), nil, nil),
    recovery.Recovery(app.Logger()),
    cors.CORS,
    compression.Gzip(),
    timeout.Timeout(3*time.Second),
    limits.ConcurrencyLimit(200, 400, 250*time.Millisecond, app.Logger()),
); err != nil {
    log.Fatal(err)
}
```

## Group-scoped controls
```go
api := app.Router().Group("/api")
api.Use(auth.SimpleAuth(os.Getenv("AUTH_TOKEN")))
api.Use(timeout.Timeout(2 * time.Second))

api.Get("/stats", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("ok"))
}))
```

## Context binding/validation with canonical adapter
```go
app.Post("/v1/users", contract.AdaptCtxHandler(func(ctx *contract.Ctx) {
    var payload CreateUserRequest
    if err := ctx.BindAndValidateJSONWithOptions(&payload, contract.BindOptions{
        DisallowUnknownFields: true,
        Logger:                ctx.Logger,
    }); err != nil {
        contract.WriteBindError(ctx.W, ctx.R, err)
        return
    }

    _ = ctx.Response(http.StatusCreated, map[string]any{"ok": true}, nil)
}, app.Logger()).ServeHTTP)
```

## Operational notes
- Keep middleware order explicit and regression-tested.
- Prefer constructor/closure injection for dependencies; avoid hidden global state.
- Scope auth/security middleware per group when public and private routes differ.
