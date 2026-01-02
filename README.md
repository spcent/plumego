# Plumego — standard-library web toolkit

Plumego is a small Go HTTP toolkit that keeps everything inside the standard library while still covering routing, middleware, graceful shutdown, WebSocket helpers, webhook plumbing, and static frontend hosting. It is designed to be embedded in your own `main` package rather than acting as a framework binary.

## Highlights
- **Router with groups and params**: Trie-based matcher that supports `/:param` segments, route freezing, and middleware stacks per route/group.
- **Middleware chain**: Logging, recovery, gzip, CORS, timeout, rate limiting, concurrency limits, body size limits, and auth helpers that wrap standard `http.Handler` values.
- **Structured logging hooks**: Plug in a custom logger and collect metrics/traces via middleware hooks.
- **Graceful lifecycle**: Environment loading, connection draining, readiness flags, and optional TLS/HTTP2 configuration with sane defaults.
- **Optional services**: Built-in WebSocket hub with auth, in-process pub/sub with debug snapshots, inbound/outbound webhook routers, and static frontend mounting from disk or embedded assets.

## Components
`core.App` orchestrates pluggable components instead of hard-wiring features. A component can register routes, middleware, and lifecycle hooks:

```
type Component interface {
    RegisterRoutes(r *router.Router)
    RegisterMiddleware(m *middleware.Registry)
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Health() (name string, status health.HealthStatus)
}
```

`HealthStatus` uses typed states (`healthy`, `degraded`, `unhealthy`) so components report health in a structured, type-safe man
ner.

Use `core.WithComponent` (or `WithComponents`) when constructing the app to add capabilities. Built-in features (webhook management, inbound webhook receivers, pubsub debugging, websocket helpers, and frontend serving) can all be mounted as components so examples can mix only what they need.

## Migration Notes
- `plumego.ComponentFunc` re-export has been removed. Implement `core.Component` directly (see the interface above), or keep a local adapter type if you prefer functional hooks.

## Quick start
Create a small `main.go` that wires routes and middleware, then boot the server:

```go
package main

import (
    "log"
    "net/http"

    "github.com/spcent/plumego/core"
)

func main() {
    app := core.New(
        core.WithAddr(":8080"),
        core.WithDebug(),
    )

    app.EnableRecovery()
    app.EnableLogging()
    app.EnableCORS()

    app.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
        w.Write([]byte("pong"))
    })

    if err := app.Boot(); err != nil {
        log.Fatalf("server stopped: %v", err)
    }
}
```

You can also embed **plumego** directly into a standard `net/http` server. `plumego.App`
implements `http.Handler`, and context-aware handlers can use the unified `plumego.Context`:

```go
package main

import (
    "log"
    "net/http"

    "github.com/spcent/plumego"
)

func main() {
    app := plumego.New()

    app.GetCtx("/health", func(ctx *plumego.Context) {
        ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
    })

    log.Println("server started at :8080")
    log.Fatal(http.ListenAndServe(":8080", app))
}
```

## Configuration basics
- Environment variables can be loaded from a `.env` file (default path `.env`; override with `core.WithEnvPath`).
- Useful variables: `AUTH_TOKEN` (SimpleAuth middleware), `WS_SECRET` (WebSocket JWT signing secret), `WEBHOOK_TRIGGER_TOKEN`, `GITHUB_WEBHOOK_SECRET`, and `STRIPE_WEBHOOK_SECRET` (see `env.example`).
- App defaults include 10 MiB body limit, 256 concurrent request limit with queueing, HTTP read/write timeouts, and a 5s graceful shutdown window. Override via the `core.With...` options.

## Key components
- **Router**: Register handlers with `Get`, `Post`, etc., or context-aware variants (`GetCtx`) that expose a unified request context wrapper. Grouping lets you attach shared middleware, and static frontends can be mounted via `frontend.RegisterFromDir`.
- **Middleware**: Chain middlewares with `app.Use(...)` before boot; guardrails (body limit, concurrency limit) are injected automatically during setup. Recovery and logging helpers are available via `EnableRecovery` and `EnableLogging`.
- **WebSocket hub**: `ConfigureWebSocket()` mounts a JWT-protected `/ws` endpoint plus an optional broadcast endpoint guarded by a shared secret. Customize worker counts and queue sizes through `WebSocketConfig`.
- **Pub/Sub + Webhooks**: Supply a `pubsub.PubSub` implementation to enable webhook fan-out. Outbound webhook management includes target CRUD, delivery replay, and trigger tokens; inbound receivers handle GitHub/Stripe signatures with deduplication and size limits.
- **Health + readiness**: Lifecycle hooks mark readiness during startup/shutdown, and build metadata (`Version`, `Commit`, `BuildTime`) can be injected via ldflags.

## Reference application
`examples/reference` is a ready-to-run `main` package that wires the common pieces together:

- WebSocket hub configured with JWT secrets and broadcast endpoint
- Inbound GitHub/Stripe webhooks that publish to the in-process pub/sub
- Outbound webhook management backed by the memory store
- Static frontend served from embedded assets
- Prometheus metrics, OpenTelemetry tracing, and health endpoints mounted on the router

Run it with:

```bash
go run ./examples/reference
```

## Health endpoints
The `health` package now exposes HTTP handlers so you don’t have to reimplement readiness/build checks:

```go
app.Get("/health/ready", health.ReadinessHandler().ServeHTTP)
app.Get("/health/build", health.BuildInfoHandler().ServeHTTP)
```

`ReadinessHandler` returns 200 when `health.SetReady()` has been called (the boot lifecycle does this automatically) and 503 otherwise. `BuildInfoHandler` returns the current `health.BuildInfo` struct as JSON.

## Observability adapters
Plug the logging middleware into metrics/tracing backends without writing adapters yourself:

- `metrics.NewPrometheusCollector(namespace)` implements `middleware.MetricsCollector` and exposes a `/metrics` handler via `collector.Handler()`.
- `metrics.NewOpenTelemetryTracer(name)` implements `middleware.Tracer` and emits spans with HTTP metadata.

Wire them into `core.New` using `core.WithMetricsCollector(...)` and `core.WithTracer(...)` as shown in `examples/reference`.

## Configuration reference
Load environment variables with `config.LoadEnv` and/or bind command-line flags; use the mapping below for predictable deploys.

| AppConfig field | Default | Environment variable | Flag example |
| --- | --- | --- | --- |
| Addr | :8080 | APP_ADDR | --addr :8080 |
| EnvFile | .env | APP_ENV_FILE | --env-file .env |
| Debug | false | APP_DEBUG | --debug |
| ShutdownTimeout | 5s | APP_SHUTDOWN_TIMEOUT_MS | --shutdown-timeout 5s |
| ReadTimeout | 30s | APP_READ_TIMEOUT_MS | --read-timeout 30s |
| ReadHeaderTimeout | 5s | APP_READ_HEADER_TIMEOUT_MS | --read-header-timeout 5s |
| WriteTimeout | 30s | APP_WRITE_TIMEOUT_MS | --write-timeout 30s |
| IdleTimeout | 60s | APP_IDLE_TIMEOUT_MS | --idle-timeout 60s |
| MaxHeaderBytes | 1 MiB | APP_MAX_HEADER_BYTES | --max-header-bytes 1048576 |
| EnableHTTP2 | true | APP_ENABLE_HTTP2 | --http2=false |
| DrainInterval | 500ms | APP_DRAIN_INTERVAL_MS | --drain-interval 500ms |
| MaxBodyBytes | 10 MiB | APP_MAX_BODY_BYTES | --max-body-bytes 10485760 |
| MaxConcurrency | 256 | APP_MAX_CONCURRENCY | --max-concurrency 256 |
| QueueDepth | 512 | APP_QUEUE_DEPTH | --queue-depth 512 |
| QueueTimeout | 250ms | APP_QUEUE_TIMEOUT_MS | --queue-timeout 250ms |
| TLS.Enabled | false | TLS_ENABLED | --tls |
| TLS.CertFile | (empty) | TLS_CERT_FILE | --tls-cert /path/cert.pem |
| TLS.KeyFile | (empty) | TLS_KEY_FILE | --tls-key /path/key.pem |
| PubSub.Enabled | false | PUBSUB_DEBUG_ENABLED | --pubsub-debug |
| PubSub.Path | /_debug/pubsub | PUBSUB_DEBUG_PATH | --pubsub-path /_debug/pubsub |
| WebhookOut.TriggerToken | (empty) | WEBHOOK_TRIGGER_TOKEN | --webhook-trigger-token TOKEN |
| WebhookOut.BasePath | /webhooks | (inherit) | --webhook-base /webhooks |
| WebhookOut.IncludeStats | false | WEBHOOK_INCLUDE_STATS | --webhook-include-stats |
| WebhookOut.DefaultPageLimit | 0 (no default) | WEBHOOK_DEFAULT_PAGE_LIMIT | --webhook-page-limit 50 |
| WebhookIn.GitHubSecret | env or config | GITHUB_WEBHOOK_SECRET | --github-secret value |
| WebhookIn.StripeSecret | env or config | STRIPE_WEBHOOK_SECRET | --stripe-secret value |
| WebhookIn.MaxBodyBytes | 1 MiB | WEBHOOK_MAX_BODY_BYTES | --webhook-max-body 1048576 |
| WebhookIn.StripeTolerance | 5m | WEBHOOK_STRIPE_TOLERANCE_MS | --stripe-tolerance 5m |
| WebhookIn.TopicPrefixGitHub | in.github. | WEBHOOK_TOPIC_PREFIX_GITHUB | --github-topic-prefix in.github. |
| WebhookIn.TopicPrefixStripe | in.stripe. | WEBHOOK_TOPIC_PREFIX_STRIPE | --stripe-topic-prefix in.stripe. |
| WebSocket.Secret | env or config | WS_SECRET | --ws-secret value |
| WebSocket.WSRoutePath | /ws | WS_ROUTE_PATH | --ws-route /ws |
| WebSocket.BroadcastPath | /_admin/broadcast | WS_BROADCAST_PATH | --ws-broadcast /_admin/broadcast |
| WebSocket.BroadcastEnabled | true | WS_BROADCAST_ENABLED | --ws-broadcast-enabled=false |

Use the `config.Get*` helpers (see `config/env.go`) or Go’s `flag` package to translate these sources into an `AppConfig` before calling `core.New(...)`.

## Development and testing
- Install Go 1.24+ (matching `go.mod`).
- Run tests: `go test ./...`
- Format and lint using the Go toolchain as needed (`go fmt`, `go vet`).
