# Core module

The `core` package owns the HTTP server lifecycle. It wires routing, middleware registration, component startup/shutdown, and graceful termination around standard `net/http` handlers.

## Create and boot an app
```go
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
```

## Configuration and defaults
Most server knobs map to `AppConfig` / `core.With...` options:

- `WithAddr`, `WithEnvPath`
- `WithShutdownTimeout`
- `WithServerTimeouts` (read/read-header/write/idle)
- `WithMaxHeaderBytes`
- `WithHTTP2`, `WithTLS`, `WithTLSConfig`
- `WithDebug`, `WithLogger`
- `WithMethodNotAllowed`

## Components and lifecycle hooks
`core.App` can orchestrate components so background work stays aligned with server start/stop.

```go
type worker struct{ bus *pubsub.InProcPubSub }

func (w *worker) RegisterRoutes(r *router.Router)           {}
func (w *worker) RegisterMiddleware(m *middleware.Registry) {}
func (w *worker) Start(ctx context.Context) error           { return nil }
func (w *worker) Stop(ctx context.Context) error            { return nil }
func (w *worker) Health() (string, health.HealthStatus) {
    return "worker", health.HealthStatus{Status: health.StatusHealthy}
}

app := core.New(core.WithComponent(&worker{bus: pubsub.New()}))
```

You can also register runners and shutdown hooks via `WithRunner` / `app.Register(...)` and `WithShutdownHook` / `app.OnShutdown(...)`.

## Reference-style wiring
```go
prom := metrics.NewPrometheusCollector("plumego")
tracer := metrics.NewOpenTelemetryTracer("plumego")
healthManager, err := health.NewHealthManager(health.HealthCheckConfig{})
if err != nil {
    log.Fatal(err)
}

app := core.New(
    core.WithAddr(":8080"),
    core.WithMetricsCollector(prom),
    core.WithTracer(tracer),
    core.WithHealthManager(healthManager),
)

if err := app.Use(
    observability.RequestID(),
    observability.Logging(app.Logger(), prom, tracer),
    recovery.RecoveryMiddleware,
); err != nil {
    log.Fatal(err)
}

app.Get("/metrics", prom.Handler().ServeHTTP)
app.Get("/health/ready", health.ReadinessHandler(healthManager).ServeHTTP)
app.Get("/health/build", health.BuildInfoHandler().ServeHTTP)

if err := app.Boot(); err != nil {
    log.Fatal(err)
}
```

## Context-style handlers (canonical adapter)
`core.App` keeps canonical handlers as `func(http.ResponseWriter, *http.Request)`. For `contract.Ctx`, adapt explicitly:

```go
app.Post("/users", contract.AdaptCtxHandler(func(ctx *contract.Ctx) {
    var req CreateUserRequest
    if err := ctx.BindAndValidateJSONWithOptions(&req, contract.BindOptions{}); err != nil {
        contract.WriteBindError(ctx.W, ctx.R, err)
        return
    }
    _ = ctx.Response(http.StatusOK, map[string]any{"ok": true}, nil)
}, app.Logger()).ServeHTTP)
```

## Safety and troubleshooting
- Register routes/middleware before `Boot()`; mutating after boot is rejected.
- Keep handlers cancellation-aware so shutdown drains quickly.
- WebSocket and webhook features should be mounted via explicit components/config, not hidden globals.
