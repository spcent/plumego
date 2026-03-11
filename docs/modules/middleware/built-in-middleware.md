# Built-in Middleware

> **Package root**: `github.com/spcent/plumego/middleware`

This page lists canonical middleware in Plumego v1 and how to register them.

## Registration Pattern

```go
app := core.New(core.WithAddr(":8080"))

if err := app.Use(
    observability.RequestID(),
    observability.Logging(app.Logger(), nil, nil),
    recovery.Recovery(app.Logger()),
); err != nil {
    log.Fatal(err)
}
```

`app.Use(...)` order is execution order.

## Observability

- `observability.RequestID(...)`
- `observability.Logging(logger, metricsCollector, tracer)`

## Safety / Reliability

- `recovery.Recovery(logger)`
- `timeout.Timeout(duration)`
- `timeout.TimeoutWithConfig(timeout.TimeoutConfig{...})`
- `limits.BodyLimit(maxBytes, logger)`
- `limits.ConcurrencyLimit(maxConcurrent, queueDepth, queueTimeout, logger)`

## Transport / Performance

- `cors.CORS`
- `cors.CORSWithOptions(opts, next)` (handler wrapper form)
- `compression.Gzip()`
- `compression.GzipWithConfig(compression.GzipConfig{...})`

## Security

- `security.SecurityHeaders(policy)`
- `auth.SimpleAuth(token)`
- `auth.Authenticate(authenticator, ...)`
- `auth.Authorize(authorizer, action, resource, ...)`
- `auth.SessionCheck(store, validator, ...)`
- `ratelimit.AbuseGuard(config)`
- `ratelimit.TokenBucket(config)`
- `ratelimit.RateLimitMiddleware(config)`

## Protocol / Version / Tenant

- `protocol.Middleware(registry)` / `protocol.MiddlewareWithConfig(config)`
- `versioning.Middleware(versioning.Config{...})`
- `tenant.TenantResolver(options)`
- `tenant.TenantRateLimit(options)`

## Group-Level Middleware

```go
api := app.Router().Group("/api")
api.Use(auth.SimpleAuth(os.Getenv("AUTH_TOKEN")))
api.Use(timeout.Timeout(2 * time.Second))
```

## Testing Guidance

For middleware changes, add tests for:
- ordering
- panic/error path
- status/header/body behavior
- race safety (`go test -race`)
