# Plumego — Standard Library Web Toolkit

[![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Version](https://img.shields.io/badge/version-v1.0.0--rc.1-blue)](https://github.com/spcent/plumego/releases)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

Plumego is a lightweight Go HTTP toolkit built entirely on the standard library. It covers routing, middleware, graceful shutdown, WebSocket utilities, webhook pipelines, and static frontend hosting. It is designed to be embedded into your own `main` package rather than acting as a standalone framework binary.

The `core` package is the stable, primary entrypoint. The top-level `plumego` package provides convenience re-exports for common types and options.

## Highlights
- **Router with Groups and Parameters**: Trie-based matcher supporting `/:param` segments, route freezing, and per-route/group middleware stacks.
- **Middleware Chain**: Logging, recovery, gzip, CORS, timeout (buffers up to 10 MiB by default), rate limiting, concurrency limits, body size limits, security headers, and authentication helpers, all wrapping standard `http.Handler`.
- **Security Helpers**: JWT + password utilities, security header policies, input-safety helpers, and abuse guard primitives for baseline hardening.
- **Integration Helpers**: Lightweight adapters for `database/sql`, Redis-backed caches, and message queues (the `net/mq` module includes a durable task queue mode; still experimental).
- **Idempotency Utilities**: Simple KV/SQL helpers for request deduplication via `store/idempotency`.
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
- The app defaults to a 10485760 byte (10 MiB) request body limit, 256 concurrent requests (with queue), HTTP read/write timeouts, and a 5000ms (5s) graceful shutdown window. Override via `core.With...` options.
- Security guardrails (security headers + abuse guard) are enabled by default. Abuse guard defaults to 100 req/s with a burst of 200 per client and tracks up to 100k active keys. Disable or tune via `core.WithSecurityHeadersEnabled`, `core.WithSecurityHeadersPolicy`, `core.WithAbuseGuardEnabled`, and `core.WithAbuseGuardConfig`.
- Debug mode (`core.WithDebug`) enables devtools endpoints under `/_debug` (routes, middleware, config, metrics, pprof, reload), friendly JSON error output, and `.env` hot reload. These endpoints are intended for local development or protected environments; disable or gate them in production.

## Key Components
- **Router**: Register handlers with `Get`, `Post`, etc., or the context-aware variants (`GetCtx`) that expose a unified request context wrapper. Groups allow attaching shared middleware, and static frontends can be mounted via `frontend.RegisterFromDir` with cache/fallback options (`frontend.WithCacheControl`, `frontend.WithIndexCacheControl`, `frontend.WithFallback`, `frontend.WithHeaders`).
- **Middleware**: Chain middleware before boot with `app.Use(...)`; guards (security headers, abuse guard, body size limits, concurrency limits) are auto-injected during setup. Recovery/logging/CORS helpers can be enabled via `core.WithRecovery`, `core.WithLogging`, and `core.WithCORS`. For a production-safe baseline, `core.WithRecommendedMiddleware()` enables RequestID + Logging + Recovery in the recommended order.
- **Multi-Tenancy (experimental)**: Tenant isolation with quota enforcement, policy controls, and database filtering. The API is experimental and may change. See [Multi-Tenancy](#multi-tenancy) for details.
- **Ops/Admin Endpoints**: Optional protected operations API for queue stats/replay, receipt lookup, channel health, and tenant quota inspection. Mount via `core/components/ops` and secure with a token or custom middleware. If auth is missing and `AllowInsecure` is false (default), requests are denied.
- **Contract Helpers**: Use `contract.WriteError` for error payloads and `contract.WriteResponse` / `Ctx.Response` for consistent JSON responses with trace IDs.
- **WebSocket Hub**: `ConfigureWebSocket()` mounts a JWT-protected `/ws` endpoint, plus an optional broadcast endpoint (protected by a shared secret). Customize worker count and queue size via `WebSocketConfig`.
- **Pub/Sub + Webhook**: Provides `pubsub.PubSub` to enable webhook fan-out. Outbound Webhook management includes target CRUD, delivery replay, and trigger tokens; inbound receivers handle GitHub/Stripe signatures plus generic HMAC verification with replay protection and IP allowlists.
- **Migrations**: Optional SQL schemas for modules/examples live in `docs/migrations/` (see notes for sms-gateway `sent_at` backfill).
- **Health + Readiness**: Lifecycle hooks mark readiness during startup/shutdown; build metadata (`Version`, `Commit`, `BuildTime`) can be injected via ldflags.

## Multi-Tenancy

Plumego provides experimental multi-tenancy primitives for SaaS applications with tenant isolation, quota enforcement, and policy controls.

### Features

- **Tenant Configuration**: Flexible storage backends (in-memory, database with LRU caching)
- **Rate Limiting**: Per-tenant token bucket (requests/second with burst control)
- **Quota Management**: Per-tenant usage limits across minute/hour/day/month windows with fixed-window enforcement
- **Policy Controls**: Per-tenant allow-lists for models and tools
- **Routing Policy Cache**: Tenant-specific routing policy with cache wrapper
- **Database Isolation**: Automatic tenant filtering for all SQL queries via `TenantDB` wrapper
- **Middleware Stack**: Tenant resolution → Rate limiting → Quota checking → Policy enforcement
- **Audit Hooks**: Optional callbacks for monitoring quota/policy violations

### Quick Setup

```go
import (
    "github.com/spcent/plumego"
    "github.com/spcent/plumego/store/db"
)

// Create tenant config manager with caching
tenantMgr := plumego.NewDBTenantConfigManager(
    database,
    plumego.WithTenantCache(1000, 5*time.Minute),
)

// Create rate limit provider and limiter
rateLimitProvider := plumego.NewInMemoryRateLimitProvider()
rateLimitProvider.SetRateLimit("tenant-id", plumego.TenantRateLimitConfig{
    RequestsPerSecond: 50,
    Burst:             100,
})
rateLimiter := plumego.NewTokenBucketRateLimiter(rateLimitProvider)

// Create quota and policy managers
quotaStore := plumego.NewInMemoryQuotaStore()
quotaMgr := plumego.NewWindowQuotaManager(tenantMgr, quotaStore)
policyEval := plumego.NewConfigPolicyEvaluator(tenantMgr)

// Create tenant-aware database wrapper
tenantDB := plumego.NewTenantDB(database)

// Configure application
app := plumego.New(
    plumego.WithTenantConfigManager(tenantMgr),
    plumego.WithTenantMiddleware(plumego.TenantMiddlewareOptions{
        HeaderName:      "X-Tenant-ID",
        AllowMissing:    false,
        RateLimiter:     rateLimiter,
        QuotaManager:    quotaMgr,
        PolicyEvaluator: policyEval,
        Hooks: plumego.TenantHooks{
            OnQuota: func(ctx context.Context, decision plumego.TenantQuotaDecision) {
                if !decision.Allowed {
                    log.Printf("Quota exceeded for %s", decision.TenantID)
                }
            },
        },
    }),
)
```

If you manage tenant config in code (or provide your own config provider), you can set multi-window quotas like:

```go
tenantMgr := plumego.NewInMemoryTenantConfigManager()
tenantMgr.SetTenantConfig(plumego.TenantConfig{
    TenantID: "tenant-id",
    Quota: plumego.TenantQuotaConfig{
        Limits: []plumego.TenantQuotaLimit{
            {Window: plumego.TenantQuotaWindowDay, Requests: 200000},
            {Window: plumego.TenantQuotaWindowMonth, Tokens: 10_000_000},
        },
    },
})

// Route policy cache (optional)
routePolicyStore := plumego.NewInMemoryRoutePolicyStore()
_ = routePolicyStore.SetRoutePolicy(context.Background(), plumego.TenantRoutePolicy{
    TenantID: "tenant-id",
    Strategy: "weighted",
    Payload:  []byte(`{"rules":[{"provider":"a","weight":70},{"provider":"b","weight":30}]}`),
})
routePolicyCache := plumego.NewInMemoryRoutePolicyCache(1000, 5*time.Minute)
routePolicyProvider := plumego.NewCachedRoutePolicyProvider(routePolicyStore, routePolicyCache)
```

### Automatic Query Filtering

The `TenantDB` wrapper automatically filters all queries by tenant ID:

```go
// Your query
rows, err := tenantDB.QueryFromContext(ctx,
    "SELECT * FROM users WHERE active = ?", true)

// Automatically becomes
"SELECT * FROM users WHERE tenant_id = ? AND active = ?"
// with tenant_id from context
```

This prevents cross-tenant data leaks and simplifies business logic by removing manual tenant filtering.

### Example Application

See `examples/multi-tenant-saas/` for a complete working example with:
- Admin API for tenant CRUD operations
- Tenant-scoped business API
- Quota enforcement with retry-after headers
- Policy validation for models/tools
- Request analytics per tenant

Run it:
```bash
cd examples/multi-tenant-saas
go run .
```

### Production Considerations

- **Performance**: Use database-backed config manager with LRU caching (1000+ tenants)
- **Security**: Replace header-based tenant ID with signed JWT tokens
- **Monitoring**: Enable quota/policy hooks for metrics collection
- **Scaling**: Run multiple instances behind load balancer with shared database

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
	Use(auth.Authenticate(jwtManager.Authenticator(jwt.TokenTypeAccess))).
	Use(auth.SessionCheck(sessionStore, sessionValidator)).
	Use(auth.Authorize(jwt.PolicyAuthorizer{Policy: jwt.AuthZPolicy{AnyRole: []string{"admin"}}}, "", ""))
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

- `metrics.NewPrometheusCollector(namespace)` implements `observability.MetricsCollector`, and exposes a `/metrics` handler via `collector.Handler()`.
- `metrics.NewOpenTelemetryTracer(name)` implements `observability.Tracer`, emitting spans with HTTP metadata.

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
Use `config.LoadEnv` to load environment variables, or bind command-line flags; `config.ConfigManager` also provides `LoadBestEffort` to skip optional source failures and `ReloadWithValidation` for transactional hot reloads. Config keys are normalized to lower_snake_case for lookups, so CamelCase and UPPER_SNAKE resolve to the same value. Durations in environment variables use milliseconds (the `_MS` suffix). Use the table below for predictable deployments.

| AppConfig Field          | Default        | Environment Variable           | Flag Example                     |
|--------------------------|----------------|--------------------------------|----------------------------------|
| Addr                     | :8080          | APP_ADDR                      | --addr :8080                    |
| EnvFile                  | .env           | APP_ENV_FILE                  | --env-file .env                 |
| Debug                    | false          | APP_DEBUG                     | --debug                         |
| ShutdownTimeout          | 5000ms         | APP_SHUTDOWN_TIMEOUT_MS       | --shutdown-timeout 5000ms       |
| ReadTimeout              | 30000ms        | APP_READ_TIMEOUT_MS           | --read-timeout 30000ms          |
| ReadHeaderTimeout        | 5000ms         | APP_READ_HEADER_TIMEOUT_MS    | --read-header-timeout 5000ms    |
| WriteTimeout             | 30000ms        | APP_WRITE_TIMEOUT_MS          | --write-timeout 30000ms         |
| IdleTimeout              | 60000ms        | APP_IDLE_TIMEOUT_MS           | --idle-timeout 60000ms          |
| MaxHeaderBytes           | 1048576        | APP_MAX_HEADER_BYTES          | --max-header-bytes 1048576      |
| EnableHTTP2              | true           | APP_ENABLE_HTTP2              | --http2=false                   |
| DrainInterval            | 500ms          | APP_DRAIN_INTERVAL_MS         | --drain-interval 500ms          |
| MaxBodyBytes             | 10485760       | APP_MAX_BODY_BYTES            | --max-body-bytes 10485760       |
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
| WebhookIn.MaxBodyBytes   | 1048576        | WEBHOOK_MAX_BODY_BYTES        | --webhook-max-body 1048576      |
| WebhookIn.StripeTolerance| 300000ms       | WEBHOOK_STRIPE_TOLERANCE_MS   | --stripe-tolerance 300000ms     |
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

## Development Server with Dashboard

The `plumego` CLI includes a powerful development server built with the plumego framework itself (dogfooding). It provides hot reload, real-time monitoring, and a web-based dashboard for enhanced development experience.

The dashboard is **enabled by default** - simply run `plumego dev` to get started.

**Positioning & Production Guidance**
- `core.WithDebug` exposes application devtools under `/_debug`. These are app endpoints and should be disabled or protected in production.
- `plumego dev` dashboard is a local developer tool that runs a separate dashboard server; it is not intended to be exposed publicly in production environments.
- The dashboard may query the app’s `/_debug` endpoints for routes/config/metrics/pprof when available, so keep debug endpoints gated outside local/dev usage.

### Quick Start

```bash
plumego dev
# Dashboard: http://localhost:9999
# Your app:  http://localhost:8080
```

### Dashboard Features

Every `plumego dev` session includes:

- **Real-time Logs**: Stream application stdout/stderr with filtering
- **Route Browser**: Auto-discover and display all HTTP routes from your app
- **Metrics Dashboard**: Monitor uptime, PID, health status, and performance
- **Build Management**: View build output and manually trigger rebuilds
- **App Control**: Start, stop, and restart your application from the UI
- **Hot Reload**: Automatic rebuild and restart on file changes (< 5 seconds)

### Customization

```bash
# Custom application port
plumego dev --addr :3000

# Custom dashboard port
plumego dev --dashboard-addr :8888

# Custom watch patterns
plumego dev --watch "**/*.go,**/*.yaml"

# Adjust hot reload sensitivity
plumego dev --debounce 1s
```

For complete documentation, see `cmd/plumego/DEV_SERVER.md`.

## Documentation
For detailed documentation, see the `examples/docs` directory:
- `examples/docs/en/` - English documentation
- `examples/docs/zh/` - Chinese documentation
