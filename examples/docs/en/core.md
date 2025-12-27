# Core module

The **core** package owns the HTTP server lifecycle. It wires routing, middleware, component startup/shutdown, WebSocket helpers, webhook services, observability, and graceful shutdown into a single `core.App` built on the Go standard library.

## Create and boot an app
```go
app := core.New(
    core.WithAddr(":8080"),
    core.WithDebug(),
)
app.EnableRecovery()
app.EnableLogging()
app.EnableCORS()

// Register routes _before_ Boot freezes the router.
app.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
    w.Write([]byte("pong"))
})

if err := app.Boot(); err != nil {
    log.Fatalf("server stopped: %v", err)
}
```

## Configuration and defaults
Most knobs map to `env.example` for deploy-time control.

- **Address & shutdown**: `core.WithAddr` overrides `:8080`. `ShutdownTimeout` drains inflight requests on SIGTERM (defaults to 5s).
- **Limits**: 10 MiB body cap, 256 concurrent requests with queueing, HTTP read/write timeouts. Override via `core.WithMaxBodyBytes`, `core.WithMaxConcurrency`, `core.WithHTTPTimeouts`, or env vars.
- **TLS / HTTP2**: enable with `core.WithTLSConfig` or `AppConfig.TLS`; disable HTTP/2 via `core.WithHTTP2(false)`.
- **Env loading**: `core.WithEnvPath` controls the `.env` file location.
- **Observability**: plug `metrics.NewPrometheusCollector` and `metrics.NewOpenTelemetryTracer` so the logging middleware emits metrics/traces automatically.
- **Debug aids**: `WithDebug()` logs registered routes and relaxes some failure paths; disable in production.

## Components and lifecycle hooks
`core.App` can orchestrate components so long-running tasks stay aligned with server start/stop.

```go
type worker struct{ bus *pubsub.PubSub }
func (w *worker) RegisterRoutes(r *router.Router) {}
func (w *worker) RegisterMiddleware(m *middleware.Registry) {}
func (w *worker) Start(ctx context.Context) error {
    // subscribe to events after the HTTP server is ready
    return w.bus.Subscribe("jobs.*", func(ctx context.Context, evt pubsub.Event) error {
        // ...handle work...
        return nil
    })
}
func (w *worker) Stop(ctx context.Context) error { return nil }
func (w *worker) Health() (string, health.HealthStatus) { return "worker", health.Healthy }

app := core.New(core.WithComponent(&worker{bus: pubsub.New()}))
```

Built-in components include inbound webhook receivers, outbound webhook service, WebSocket hub registration, Pub/Sub debug UI, and frontend mounting helpers. Add them with `core.WithComponent` or through dedicated options such as `core.WithWebhookIn`, `core.WithWebhookOut`, and `core.ConfigureWebSocketWithOptions`.

## Putting the pieces together (reference-style)
```go
bus := pubsub.New()
prom := metrics.NewPrometheusCollector("plumego")
tracer := metrics.NewOpenTelemetryTracer("plumego")

app := core.New(
    core.WithAddr(":8080"),
    core.WithPubSub(bus),
    core.WithMetricsCollector(prom),
    core.WithTracer(tracer),
    core.WithWebhookIn(core.WebhookInConfig{Enabled: true, Pub: bus}),
    core.WithWebhookOut(core.WebhookOutConfig{Enabled: true, TriggerToken: "secret"}),
)
app.EnableRecovery()
app.EnableLogging()

app.GetHandler("/metrics", prom.Handler())
app.GetHandler("/health/ready", health.ReadinessHandler())
app.GetHandler("/health/build", health.BuildInfoHandler())
_, _ = app.ConfigureWebSocketWithOptions(core.DefaultWebSocketConfig())

if err := app.Boot(); err != nil { log.Fatal(err) }
```

## Safety and troubleshooting
- Late route/middleware registration after `Boot()` panics by design—fix init order instead of suppressing it.
- Keep handlers cancellation-aware so shutdown drains quickly; adjust `ShutdownTimeout` if orchestration needs shorter/longer drains.
- Missing secrets for webhooks or WebSocket JWT signing are the most common startup errors; confirm `.env` was loaded.

## Where to look in the repo
- `core/app.go`: `App` structure, options, and lifecycle wiring.
- `core/options.go`: option helpers for address, debug mode, TLS, webhook, WebSocket, and Pub/Sub configuration.
- `examples/reference/main.go`: end-to-end wiring that exercises most core capabilities.
