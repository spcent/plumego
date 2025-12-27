# Core module

The **core** package builds the HTTP application lifecycle. It wires routing, middleware, component startup/shutdown, WebSocket helpers, webhook services, and observability into a single `core.App`.

## Responsibilities
- Construct an app with `core.New` using functional options (address, debug toggle, metrics collector, tracer, Pub/Sub bus, WebSocket config, webhook settings).
- Register routes and middleware before freezing the router on `Boot()`.
- Drive component lifecycle via the `Component` interface (`RegisterRoutes`, `RegisterMiddleware`, `Start`, `Stop`, `Health`).
- Manage environment loading (`core.WithEnvPath`), graceful shutdown, TLS/HTTP2 toggles, and default limits (body size, concurrency, queue depth, HTTP timeouts).

## Minimal setup
```go
app := core.New(
    core.WithAddr(":8080"),
    core.WithDebug(),
)
app.EnableRecovery()
app.EnableLogging()
app.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
    w.Write([]byte("pong"))
})
if err := app.Boot(); err != nil {
    log.Fatalf("server stopped: %v", err)
}
```

## Configuration checklist
- **Address & shutdown**: override `:8080` with `core.WithAddr`; graceful shutdown drains inflight requests on SIGTERM and honors `ShutdownTimeout`.
- **Limits**: defaults include 10 MiB body limit, 256 concurrency with queueing, HTTP read/write timeouts. Override with `core.WithMaxBodyBytes`, `core.WithMaxConcurrency`, `core.WithHTTPTimeouts`, or environment variables described in `env.example`.
- **TLS/HTTP2**: enable TLS via `core.WithTLSConfig` or `AppConfig.TLS`; HTTP/2 can be disabled with `core.WithHTTP2(false)`.
- **Components**: attach reusable capabilities with `core.WithComponent(...)`; built-ins include webhook services, Pub/Sub debug UI, WebSocket hub, and frontend serving.
- **Observability**: plug `metrics.NewPrometheusCollector(namespace)` and `metrics.NewOpenTelemetryTracer(name)` through options so logging middleware emits metrics/traces automatically.

## Startup and safety notes
- Call `Boot()` only after all routes and middleware are registered; late registrations panic because the router is frozen.
- Copy long-running goroutines (webhook schedulers, pub/sub dispatchers) into components and align their `Start/Stop` with the app lifecycle.
- In production, drop `WithDebug()` and control verbosity from your logger backend.

## Troubleshooting
- **Port conflict**: double-check `WithAddr` and inspect listeners with `netstat -tnlp` inside the container.
- **Missing secrets**: webhook signature errors typically mean `GITHUB_WEBHOOK_SECRET`, `STRIPE_WEBHOOK_SECRET`, or `WS_SECRET` are absent; verify `.env` loading.
- **Slow shutdown**: ensure handlers respect context cancellation; consider lowering `ShutdownTimeout` if pods hang during rolling updates.

## Where to look in the repo
- `core/app.go`: `App` structure, options, and lifecycle wiring.
- `core/options.go`: option helpers for address, debug mode, TLS, webhook, WebSocket, and Pub/Sub configuration.
- `examples/reference/main.go`: end-to-end wiring that exercises most core capabilities.
