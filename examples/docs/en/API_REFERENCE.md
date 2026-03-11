# API Reference (v1 Canonical)

This reference summarizes the stable, production-facing APIs used across Plumego v1 docs and examples.

## Core (`core`)

### App construction
```go
app := core.New(options...)
```

Common options:
- `core.WithAddr(string)`
- `core.WithEnvPath(string)`
- `core.WithShutdownTimeout(time.Duration)`
- `core.WithServerTimeouts(read, readHeader, write, idle time.Duration)`
- `core.WithMaxHeaderBytes(int)`
- `core.WithHTTP2(bool)`
- `core.WithTLS(certFile, keyFile)` / `core.WithTLSConfig(core.TLSConfig)`
- `core.WithDebug()`
- `core.WithDevTools()`
- `core.WithLogger(log.StructuredLogger)`
- `core.WithMethodNotAllowed(bool)`
- `core.WithComponent(component)` / `core.WithComponents(...)`
- `core.WithRunner(runner)` / `core.WithRunners(...)`
- `core.WithShutdownHook(hook)` / `core.WithShutdownHooks(...)`
- `core.WithMetricsCollector(collector)`
- `core.WithTracer(tracer)`
- `core.WithHealthManager(manager)`

### Route registration
```go
app.Get(path, handlerFunc)
app.Post(path, handlerFunc)
app.Put(path, handlerFunc)
app.Delete(path, handlerFunc)
app.Patch(path, handlerFunc)
app.Any(path, handlerFunc)
```

### Middleware registration
```go
err := app.Use(mw1, mw2, mw3)
```

### Lifecycle
```go
err := app.Prepare()
err := app.Start(ctx)
srv, err := app.Server()
err := app.Shutdown(ctx)
err := app.Register(runner)
err := app.OnShutdown(hook)
```

### Advanced wrappers
```go
r := app.Router()
_, err := app.ConfigureWebSocket()
_, err := app.ConfigureWebSocketWithOptions(core.DefaultWebSocketConfig())
err := app.ConfigureObservability(core.DefaultObservabilityConfig())
```

## Router (`router`)

### Group and middleware
```go
r := app.Router()
api := r.Group("/api")
api.Use(mw)
```

### Route metadata and reverse routing
```go
err := r.AddRouteWithOptions(router.GET, "/users/:id", http.HandlerFunc(show),
    router.WithRouteName("users.show"),
)

u := r.URL("users.show", "id", "42")
```

### Path params
```go
id, ok := contract.Param(req, "id")
```

## Middleware (`middleware/*`)

Canonical middleware type:
```go
type Middleware func(http.Handler) http.Handler
```

Common middleware:
- `observability.RequestID()`
- `observability.Logging(logger, metricsCollector, tracer)`
- `recovery.RecoveryMiddleware`
- `cors.CORS` / `cors.CORSWithOptions(...)`
- `compression.Gzip()`
- `timeout.Timeout(duration)` / `timeout.TimeoutWithConfig(...)`
- `limits.BodyLimit(maxBytes, logger)`
- `limits.ConcurrencyLimit(maxConcurrent, queueDepth, queueTimeout, logger)`
- `security.SecurityHeaders(policy)`
- `ratelimit.AbuseGuard(config)` / `ratelimit.TokenBucket(config)`
- `auth.SimpleAuth(token)`
- `auth.Authenticate(authenticator, ...)`
- `auth.Authorize(authorizer, action, resource, ...)`
- `auth.SessionCheck(sessionStore, sessionValidator, ...)`

## Contract (`contract`)

### Ctx adapter
```go
app.Post("/users", contract.AdaptCtxHandler(func(ctx *contract.Ctx) {
    // ...
}, app.Logger()).ServeHTTP)
```

### Binding and response helpers
```go
err := ctx.BindJSON(&payload)
err := ctx.BindAndValidateJSONWithOptions(&payload, contract.BindOptions{DisallowUnknownFields: true})
contract.WriteBindError(ctx.W, ctx.R, err)
contract.WriteError(w, r, apiErr)
err = ctx.Response(status, data, meta)
```

## Webhook (`core/components/webhook`, `net/webhookin`, `net/webhookout`)

### Mount components
```go
core.WithComponent(webhook.NewWebhookInComponent(webhook.WebhookInConfig{...}, bus, nil))
core.WithComponent(webhook.NewWebhookOutComponent(webhook.WebhookOutConfig{...}))
```

### Generic signature verify
```go
result, err := webhookin.VerifyHMAC(r, webhookin.HMACConfig{...})
```

## Pub/Sub (`pubsub`)

```go
bus := pubsub.New()
sub, err := bus.SubscribePattern("orders.*", pubsub.DefaultSubOptions())
err = bus.Publish("orders.created", pubsub.Message{...})
```

Read events from `sub.C()` and `defer sub.Cancel()`.

## Health / Metrics (`health`, `metrics`)

```go
prom := metrics.NewPrometheusCollector("plumego")
tracer := metrics.NewOpenTelemetryTracer("svc")
healthManager, err := health.NewHealthManager(health.HealthCheckConfig{})
if err != nil {
    log.Fatal(err)
}

app.Get("/metrics", prom.Handler().ServeHTTP)
app.Get("/health/ready", health.ReadinessHandler(healthManager).ServeHTTP)
app.Get("/health/build", health.BuildInfoHandler().ServeHTTP)
```

## Deprecated patterns to avoid in v1 docs
- `core.WithRecovery`, `core.WithLogging`, `core.WithCORS`
- `core.WithWebhookIn`, `core.WithWebhookOut`, `core.WithPubSub`
- `app.GetCtx` / `app.PostCtx`
- top-level `middleware` convenience APIs that do not exist in current package layout
