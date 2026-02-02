# Plumego — Standard Library Web Toolkit

Plumego is a lightweight Go HTTP toolkit built entirely on the standard library. It covers routing, middleware, graceful shutdown, WebSocket utilities, webhook pipelines, and static frontend hosting. It is designed to be embedded into your own `main` package rather than acting as a standalone framework binary.

## Highlights
- **Router with Groups and Parameters**: Trie-based matcher supporting `/:param` segments, route freezing, and per-route/group middleware stacks.
- **Middleware Chain**: Logging, recovery, gzip, CORS, timeout (buffers up to 10 MiB by default), rate limiting, concurrency limits, body size limits, security headers, and authentication helpers, all wrapping standard `http.Handler`.
- **Security Helpers**: JWT + password utilities, security header policies, input-safety helpers, and abuse guard primitives for baseline hardening.
- **Integration Helpers**: Lightweight adapters for `database/sql`, Redis-backed caches, and message queues (the `net/mq` module is experimental).
- **Structured Logging Hooks**: Hook into custom loggers and collect metrics/tracing through middleware hooks.
- **Graceful Lifecycle**: Environment variable loading, connection draining, ready flags, and optional TLS/HTTP2 configuration with sensible defaults.
- **Optional Services**: Built-in authenticated WebSocket hub, in-process Pub/Sub (with debug snapshots), inbound/outbound webhook routers, and static frontend serving from disk or embedded resources.
- **Task Scheduling**: In-process cron, delayed jobs, and retryable tasks via the `scheduler` package.

## Components
`core.App` orchestrates through pluggable components instead of hardcoded functionality. Components can register routes, middleware, and lifecycle hooks:

```go
type Component interface {
    RegisterRoutes(r *router.Router)
    RegisterMiddleware(m *middleware.Registry)
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Health() (name string, status health.HealthStatus)
}
```

`HealthStatus` uses constrained values (`healthy`, `degraded`, `unhealthy`) to ensure components report health in a structured, type-safe way.

Use `core.WithComponent` (or `WithComponents`) when constructing the app to add functionality. Built-in features (Webhook management, inbound Webhook receiver, PubSub debug, WebSocket utilities, frontend serving) can all be mounted as components, so examples can mix only the parts they need.

## Quick Start
Create a small `main.go`, wire routes and middleware, then start the server:

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
        core.WithRecovery(),
        core.WithLogging(),
        core.WithCORS(),
    )

    app.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
        w.Write([]byte("pong"))
    })

    if err := app.Boot(); err != nil {
        log.Fatalf("server stopped: %v", err)
    }
}
```

`plumego.App` also implements `http.Handler`, so it can be mounted directly into a standard library server. Contextual handlers can use the unified `plumego.Context`:

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

## Configuration Basics
- Environment variables can be loaded from a `.env` file (default path `.env`; override via `core.WithEnvPath`).
- Common variables: `AUTH_TOKEN` (SimpleAuth middleware), `WS_SECRET` (WebSocket JWT signing key, at least 32 bytes), `WEBHOOK_TRIGGER_TOKEN`, `GITHUB_WEBHOOK_SECRET`, and `STRIPE_WEBHOOK_SECRET` (see `env.example`).
- The app defaults to a 10 MiB request body limit, 256 concurrent requests (with queue), HTTP read/write timeouts, and a 5-second graceful shutdown window. Override via `core.With...` options.
- Security guardrails (security headers + abuse guard) are enabled by default. Abuse guard defaults to 100 req/s with a burst of 200 per client and tracks up to 100k active keys. Disable or tune via `core.WithSecurityHeadersEnabled`, `core.WithSecurityHeadersPolicy`, `core.WithAbuseGuardEnabled`, and `core.WithAbuseGuardConfig`.
- Debug mode (`core.WithDebug`) enables devtools endpoints under `/_debug` (routes, middleware, config, reload), friendly JSON error output, and `.env` hot reload.

## Key Components
- **Router**: Register handlers with `Get`, `Post`, etc., or the context-aware variants (`GetCtx`) that expose a unified request context wrapper. Groups allow attaching shared middleware, and static frontends can be mounted via `frontend.RegisterFromDir` with cache/fallback options (`frontend.WithCacheControl`, `frontend.WithIndexCacheControl`, `frontend.WithFallback`, `frontend.WithHeaders`).
- **Middleware**: Chain middleware before boot with `app.Use(...)`; guards (security headers, abuse guard, body size limits, concurrency limits) are auto-injected during setup. Recovery/logging/CORS helpers can be enabled via `core.WithRecovery`, `core.WithLogging`, and `core.WithCORS`. For a production-safe baseline, `core.WithRecommendedMiddleware()` enables RequestID + Logging + Recovery in the recommended order.
- **Tenant toolkit**: `tenant` provides tenant config, policy, and quota primitives plus middleware helpers (`middleware.TenantResolver`, `middleware.TenantPolicy`, `middleware.TenantQuota`) and a `core.TenantConfigComponent` for wiring tenant configuration into the core lifecycle.
- **Contract Helpers**: Use `contract.WriteError` for error payloads and `contract.WriteResponse` / `Ctx.Response` for consistent JSON responses with trace IDs.
- **WebSocket Hub**: `ConfigureWebSocket()` mounts a JWT-protected `/ws` endpoint, plus an optional broadcast endpoint (protected by a shared secret). Customize worker count and queue size via `WebSocketConfig`.
- **Pub/Sub + Webhook**: Provides `pubsub.PubSub` to enable webhook fan-out. Outbound Webhook management includes target CRUD, delivery replay, and trigger tokens; inbound receivers handle GitHub/Stripe signatures with deduplication and size limits.
- **Health + Readiness**: Lifecycle hooks mark readiness during startup/shutdown; build metadata (`Version`, `Commit`, `BuildTime`) can be injected via ldflags.

## Background Runners
Register background tasks with a minimal lifecycle interface:

```go
type Runner interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
}

app.Register(myRunner)
```

Runners start before the HTTP server and stop during graceful shutdown.

## Task Scheduling
Use the `scheduler` package for in-process cron and delayed jobs:

```go
sch := scheduler.New(scheduler.WithWorkers(2))
sch.Start()
defer sch.Stop(context.Background())

sch.AddCron("cleanup", "0 * * * *", func(ctx context.Context) error {
    // hourly task
    return nil
})

sch.Delay("one-off", 5*time.Second, func(ctx context.Context) error {
    return nil
})
```

Optional helpers include a minimal admin handler (`scheduler.NewAdminHandler`) and pluggable persistence (`scheduler.WithStore` with the in-memory or KV store).
You can also register a panic handler and metrics sink via `scheduler.WithPanicHandler` and `scheduler.WithMetricsSink`.
Job status snapshots now include a unified state machine (`queued`, `scheduled`, `running`, `retrying`, `failed`, `canceled`, `completed`) with a `StateUpdated` timestamp. The `JobQuery` filter supports state-based filtering via the `States` field.
The admin handler accepts a `state` query parameter on `/scheduler/jobs` (repeatable) to filter by job state.

## Auth Contracts
Plumego keeps authentication, authorization, and session validation separate through interfaces in `contract`. Compose them with middleware rather than relying on framework magic:

```go
chain := middleware.NewChain().
	Use(middleware.Authenticate(jwtManager.Authenticator(jwt.TokenTypeAccess))).
	Use(middleware.SessionCheck(sessionStore, sessionValidator)).
	Use(middleware.Authorize(jwt.PolicyAuthorizer{Policy: jwt.AuthZPolicy{AnyRole: []string{"admin"}}}, "", ""))
```

The `security/jwt` package provides adapters (`jwtManager.Authenticator`, `jwt.PolicyAuthorizer`, `jwt.PermissionAuthorizer`) that implement these contracts while keeping your own storage and policy engines decoupled.

## Reference App
`examples/reference` is an out-of-the-box `main` package that integrates common components:

- Configured WebSocket hub with JWT keys and broadcast endpoint
- Inbound GitHub/Stripe Webhooks, publishing to in-process Pub/Sub
- In-memory store for outbound Webhook management
- Static frontend served from embedded resources
- Prometheus metrics, OpenTelemetry tracing, and health endpoints mounted to the router

Run it with:

```bash
go run ./examples/reference
```

## Health Endpoints
The `health` package now exposes HTTP handlers, eliminating the need to implement ready/build info checks yourself:

```go
app.Get("/health/ready", health.ReadinessHandler().ServeHTTP)
app.Get("/health/build", health.BuildInfoHandler().ServeHTTP)
```

`ReadinessHandler` returns 200 after `health.SetReady()` is called (the startup lifecycle calls this automatically), otherwise 503. `BuildInfoHandler` returns the current `health.BuildInfo` struct as JSON.

## Observability Adapters
No need to write your own adapters to hook logging middleware into metrics/tracing backends:

- `metrics.NewPrometheusCollector(namespace)` implements `middleware.MetricsCollector`, and exposes a `/metrics` handler via `collector.Handler()`.
- `metrics.NewOpenTelemetryTracer(name)` implements `middleware.Tracer`, emitting spans with HTTP metadata.

As shown in `examples/reference`, wire them into `core.New` using `core.WithMetricsCollector(...)` and `core.WithTracer(...)`.

To enable a built-in Prometheus endpoint and OpenTelemetry-style tracer in one call:

```go
obs := core.DefaultObservabilityConfig()
obs.Metrics.Enabled = true
obs.Tracing.Enabled = true

if err := app.ConfigureObservability(obs); err != nil {
    log.Fatal(err)
}
```

When tracing is enabled, logs include `trace_id` and `span_id`, and responses include `X-Span-ID` for correlation.

## Configuration Reference
Use `config.LoadEnv` to load environment variables, or bind command-line flags; `config.ConfigManager` also provides `LoadBestEffort` to skip optional source failures and `ReloadWithValidation` for transactional hot reloads. Config keys are normalized to lower_snake_case for lookups, so CamelCase and UPPER_SNAKE resolve to the same value. Use the table below for predictable deployments.

| AppConfig Field          | Default        | Environment Variable           | Flag Example                     |
|--------------------------|----------------|--------------------------------|----------------------------------|
| Addr                     | :8080          | APP_ADDR                      | --addr :8080                    |
| EnvFile                  | .env           | APP_ENV_FILE                  | --env-file .env                 |
| Debug                    | false          | APP_DEBUG                     | --debug                         |
| ShutdownTimeout          | 5s             | APP_SHUTDOWN_TIMEOUT_MS       | --shutdown-timeout 5s           |
| ReadTimeout              | 30s            | APP_READ_TIMEOUT_MS           | --read-timeout 30s              |
| ReadHeaderTimeout        | 5s             | APP_READ_HEADER_TIMEOUT_MS    | --read-header-timeout 5s        |
| WriteTimeout             | 30s            | APP_WRITE_TIMEOUT_MS          | --write-timeout 30s             |
| IdleTimeout              | 60s            | APP_IDLE_TIMEOUT_MS           | --idle-timeout 60s              |
| MaxHeaderBytes           | 1 MiB          | APP_MAX_HEADER_BYTES          | --max-header-bytes 1048576      |
| EnableHTTP2              | true           | APP_ENABLE_HTTP2              | --http2=false                   |
| DrainInterval            | 500ms          | APP_DRAIN_INTERVAL_MS         | --drain-interval 500ms          |
| MaxBodyBytes             | 10 MiB         | APP_MAX_BODY_BYTES            | --max-body-bytes 10485760       |
| MaxConcurrency           | 256            | APP_MAX_CONCURRENCY           | --max-concurrency 256           |
| QueueDepth               | 512            | APP_QUEUE_DEPTH               | --queue-depth 512               |
| QueueTimeout             | 250ms          | APP_QUEUE_TIMEOUT_MS          | --queue-timeout 250ms           |
| TLS.Enabled              | false          | TLS_ENABLED                   | --tls                           |
| TLS.CertFile             | (empty)        | TLS_CERT_FILE                 | --tls-cert /path/cert.pem       |
| TLS.KeyFile              | (empty)        | TLS_KEY_FILE                  | --tls-key /path/key.pem         |
| PubSub.Enabled           | false          | PUBSUB_DEBUG_ENABLED          | --pubsub-debug                  |
| PubSub.Path              | /_debug/pubsub | PUBSUB_DEBUG_PATH             | --pubsub-path /_debug/pubsub    |
| WebhookOut.TriggerToken  | (empty)        | WEBHOOK_TRIGGER_TOKEN         | --webhook-trigger-token TOKEN   |
| WebhookOut.BasePath      | /webhooks      | (inherit)                     | --webhook-base /webhooks        |
| WebhookOut.IncludeStats  | false          | WEBHOOK_INCLUDE_STATS         | --webhook-include-stats         |
| WebhookOut.DefaultPageLimit| 0 (no default) | WEBHOOK_DEFAULT_PAGE_LIMIT    | --webhook-page-limit 50         |
| WebhookIn.GitHubSecret   | env or config  | GITHUB_WEBHOOK_SECRET         | --github-secret value           |
| WebhookIn.StripeSecret   | env or config  | STRIPE_WEBHOOK_SECRET         | --stripe-secret value           |
| WebhookIn.MaxBodyBytes   | 1 MiB          | WEBHOOK_MAX_BODY_BYTES        | --webhook-max-body 1048576      |
| WebhookIn.StripeTolerance| 5m             | WEBHOOK_STRIPE_TOLERANCE_MS   | --stripe-tolerance 5m           |
| WebhookIn.TopicPrefixGitHub| in.github.     | WEBHOOK_TOPIC_PREFIX_GITHUB   | --github-topic-prefix in.github.|
| WebhookIn.TopicPrefixStripe | in.stripe.     | WEBHOOK_TOPIC_PREFIX_STRIPE   | --stripe-topic-prefix in.stripe.|
| WebSocket.Secret         | env or config  | WS_SECRET                     | --ws-secret value               |
| WebSocket.WSRoutePath    | /ws            | WS_ROUTE_PATH                 | --ws-route /ws                  |
| WebSocket.BroadcastPath  | /_admin/broadcast | WS_BROADCAST_PATH          | --ws-broadcast /_admin/broadcast|
| WebSocket.BroadcastEnabled| true           | WS_BROADCAST_ENABLED          | --ws-broadcast-enabled=false    |
| WebSocket.MaxConnections | 0 (unlimited)  | (config only)                 | -                               |
| WebSocket.MaxRoomConnections | 0 (unlimited) | (config only)               | -                               |

Use `config.Get*` helpers (see `config/env.go`) or Go's `flag` package to transform these sources into an `AppConfig`, then call `core.New(...)`.

## Development and Testing
- Install Go 1.24+ (matching `go.mod`).
- Run tests: `go test ./...`
- Use Go toolchain for formatting and static checks (`go fmt`, `go vet`).

## Documentation
For detailed documentation, see the `examples/docs` directory:
- `examples/docs/en/` - English documentation
- `examples/docs/zh/` - Chinese documentation
