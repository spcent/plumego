# Plumego — Standard Library Web Toolkit

[![Go Version](https://img.shields.io/badge/Go-1.24%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Version](https://img.shields.io/badge/version-v1.0.0--rc.1-blue)](https://github.com/spcent/plumego/releases)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

Plumego is a lightweight Go HTTP toolkit built entirely on the standard library. It covers routing, middleware, graceful shutdown, security helpers, transport adapters, and optional `x/*` capability packs. It is designed to be embedded into your own `main` package rather than acting as a standalone framework binary.

## Repository Direction

The target repository layout is now:

- stable root packages: `core`, `router`, `contract`, `middleware`, `security`, `store`, `health`, `log`, `metrics`
- extension packages: `x/*`
- canonical architecture docs: `docs/architecture/*`
- machine-readable repo rules: `specs/*`
- repo-native execution cards: `tasks/*`

Repository control-plane split:

- `docs/`: human-readable explanation, architecture, primers, and roadmap
- `specs/`: machine-readable rules, ownership, dependency policy, and change recipes
- `tasks/`: executable work cards and agent-facing task queue

Do not move `specs/` into `docs/`. In Plumego, `specs/` is a first-class repository control surface rather than supporting prose.

For architecture planning and future refactors, prefer the rules in:

- `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
- `docs/CANONICAL_STYLE_GUIDE.md`
- `specs/repo.yaml`
- `specs/agent-entrypoints.yaml`
- `specs/ownership.yaml`
- `specs/dependency-rules.yaml`
- `specs/change-recipes/*`
- `<module>/module.yaml`

For the staged future plan, see `docs/ROADMAP.md`.

Machine-enforced repo guardrails live under `internal/checks/*`. Historical exceptions are tracked explicitly in `specs/check-baseline/*` so CI can block new drift while existing migration debt is burned down.

For new application work, use a single canonical path:

- read `reference/standard-service` first for structure and wiring
- `reference/standard-service` intentionally depends only on stable root packages; treat `x/*` examples as non-canonical

## Highlights
- **Router with Groups and Parameters**: Trie-based matcher supporting `/:param` segments, route freezing, and per-route/group middleware stacks.
- **Middleware Chain**: Logging, recovery, gzip, CORS, timeout (buffers up to 10 MiB by default), rate limiting, concurrency limits, body size limits, security headers, and authentication helpers, all wrapping standard `http.Handler`.
- **Security Helpers**: JWT + password utilities, security header policies, input-safety helpers, and abuse guard primitives for baseline hardening.
- **Integration Helpers**: Lightweight adapters for `database/sql`, Redis-backed caches, and extension-backed discovery and messaging. Start from `x/discovery` and `x/messaging`; use lower-level roots like `x/mq` only when you need queue primitives directly.
- **Idempotency Utilities**: Simple KV/SQL helpers for request deduplication via `store/idempotency`.
- **Structured Logging Hooks**: Hook into custom loggers and collect metrics/tracing through middleware hooks.
- **Graceful Lifecycle**: Environment variable loading, connection draining, ready flags, and optional TLS/HTTP2 configuration with sensible defaults.
- **Optional Services**: WebSocket, webhook, frontend, gateway, messaging, and other capability packs live under `x/*` and are intentionally excluded from the canonical app path.
- **Task Scheduling**: In-process cron, delayed jobs, and retryable tasks via the `scheduler` package.

Wire routes, middleware, and background tasks explicitly in your application package. Plumego no longer carries a compatibility component layer in `core`.

## Quick Start
Create a small `main.go`, wire routes and middleware explicitly, then start the server:

```go
package main

import (
    "log"
    "net/http"

    "github.com/spcent/plumego/core"
    plog "github.com/spcent/plumego/log"
    "github.com/spcent/plumego/middleware/requestid"
    "github.com/spcent/plumego/middleware/recovery"
)

func main() {
    app := core.New(
        core.WithAddr(":8080"),
        core.WithLogger(plog.NewGLogger()),
    )

    app.Use(
        requestid.Middleware(),
        recovery.Recovery(app.Logger()),
    )

    app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("pong"))
    })

    log.Println("server started at :8080")
    log.Fatal(http.ListenAndServe(":8080", app))
}
```

`core.App` also implements `http.Handler`, so it can be mounted directly into a standard library server:

```go
package main

import (
    "log"
    "net/http"

    "github.com/spcent/plumego/contract"
    "github.com/spcent/plumego/core"
)

func main() {
    app := core.New(core.WithAddr(":8080"))

    app.Get("/health", func(w http.ResponseWriter, r *http.Request) {
        _ = contract.WriteResponse(w, r, http.StatusOK, map[string]string{
            "status": "ok",
        }, nil)
    })

    log.Println("server started at :8080")
    log.Fatal(http.ListenAndServe(":8080", app))
}
```

## Configuration Basics
- Environment variables should be loaded explicitly in your `main` package. `core.WithEnvPath` only records the path for components that need it, such as devtools reload support.
- `core.New(...)` defaults to a `NoOpLogger`. If you expect request/runtime logs, inject a real logger with `core.WithLogger(...)`.
- Common variables: `AUTH_TOKEN` (used by ops component defaults), `WS_SECRET` (WebSocket JWT signing key, at least 32 bytes), `WEBHOOK_TRIGGER_TOKEN`, `GITHUB_WEBHOOK_SECRET`, and `STRIPE_WEBHOOK_SECRET` (see `env.example`).
- The app defaults to a 10485760 byte (10 MiB) request body limit, 256 concurrent requests (with queue), HTTP read/write timeouts, and a 5000ms (5s) graceful shutdown window. Override via `core.With...` options.
- Security baseline should be composed explicitly via `app.Use(...)`, for example `middleware/security.SecurityHeaders(...)` and `middleware/ratelimit.AbuseGuard(...)`.
- Debug mode and devtools are separate. Use `core.WithDebug()` for debug behavior; if you need devtools, wire its routes explicitly in an app-local package instead of treating it as part of the canonical kernel path.
- Devtools endpoints under `/_debug` (routes, middleware, config, metrics, pprof, reload) are provided by `x/devtools`, not by `core` itself. These endpoints are intended for local development or protected environments; disable or gate them in production.

## Agent-First Workflow
- Canonical app bootstrap starts from `reference/standard-service`.
- Machine-readable discovery rules live in `specs/agent-entrypoints.yaml`.
- Module ownership and default validation live in `specs/ownership.yaml`.
- Standard change recipes live in `specs/change-recipes/*`.
- Module primers live in `docs/modules/*` and should match each manifest's `doc_paths`.
- Secondary task-family defaults are also explicit: frontend asset work starts in `x/frontend`, local debug work starts in `x/devtools`, service discovery starts in `x/discovery`, and protected admin surfaces start in `x/ops`.
- These secondary extension roots are capability entrypoints, not application bootstrap surfaces.

## Key Components
- **Router**: Register handlers with `Get`, `Post`, and other standard-library style methods that accept `func(w http.ResponseWriter, r *http.Request)`. Groups allow attaching shared middleware, and static frontends can be mounted via `frontend.RegisterFromDir` with cache/fallback options (`frontend.WithCacheControl`, `frontend.WithIndexCacheControl`, `frontend.WithFallback`, `frontend.WithHeaders`).
- **Middleware**: Chain middleware before boot with `app.Use(...)`. Keep middleware transport-only and explicit. Canonical observability order is `middleware/requestid.Middleware`, `middleware/tracing.Middleware`, `middleware/httpmetrics.Middleware`, `middleware/accesslog.Middleware`, then `middleware/recovery.Recovery(logger)`.
- **Multi-Tenancy (experimental)**: Tenant isolation with quota enforcement, policy controls, and database filtering. The API is experimental and may change. See [Multi-Tenancy](#multi-tenancy) for details.
- **Ops/Admin Endpoints**: Optional protected operations API for queue stats/replay, receipt lookup, channel health, and tenant quota inspection. Mount via `x/ops` and secure with a token or custom middleware. If auth is missing and `AllowInsecure` is false (default), requests are denied.
- **Contract Helpers**: Use `contract.WriteError` for error payloads and `contract.WriteResponse` / `Ctx.Response` for consistent JSON responses with trace IDs.
- **WebSocket Transport**: `x/websocket` provides an app-facing server with explicit route registration, a JWT-protected `/ws` endpoint, and an optional broadcast endpoint protected by a shared secret.
- **Messaging + Webhook**: Start new app-facing messaging work in `x/messaging`. Use `x/webhook` directly only for narrow inbound or outbound webhook mechanics such as signature verification, delivery replay, and trigger-token-protected webhook operations.
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
- **Database Isolation**: Automatic tenant filtering for all SQL queries via the `x/tenant/store/db` `TenantDB` wrapper
- **Middleware Stack**: Tenant resolution → Rate limiting → Quota checking → Policy enforcement
- **Audit Hooks**: Optional callbacks for monitoring quota/policy violations

### Quick Setup

```go
import (
    "context"
    "database/sql"
    "log"
    "net/http"
    "time"

    "github.com/spcent/plumego/contract"
    "github.com/spcent/plumego/core"
    tenantconfig "github.com/spcent/plumego/x/tenant/config"
    tenant "github.com/spcent/plumego/x/tenant/core"
    tenantpolicy "github.com/spcent/plumego/x/tenant/policy"
    tenantquota "github.com/spcent/plumego/x/tenant/quota"
    tenantratelimit "github.com/spcent/plumego/x/tenant/ratelimit"
    tenantresolve "github.com/spcent/plumego/x/tenant/resolve"
    tenantdb "github.com/spcent/plumego/x/tenant/store/db"
)

func setupTenantApp(database *sql.DB) *core.App {
    // Create tenant config manager with caching.
    tenantMgr := tenantconfig.NewDBTenantConfigManager(
        database,
        tenantconfig.WithTenantCache(1000, 5*time.Minute),
    )

    // Create managers used by tenant middleware.
    quotaMgr := tenant.NewWindowQuotaManager(tenantMgr, tenant.NewInMemoryQuotaStore())
    policyEval := tenant.NewConfigPolicyEvaluator(tenantMgr)
    rateLimiter := tenant.NewTokenBucketRateLimiter(
        &tenant.RateLimitConfigProviderFromConfig{Manager: tenantMgr},
    )

    // Tenant-aware DB wrapper for query isolation.
    tenantDB := tenantdb.NewTenantDB(database)

    app := core.New(core.WithAddr(":8080"))
    api := app.Router().Group("/api")

    // Canonical explicit middleware chain.
    api.Use(tenantresolve.Middleware(tenantresolve.Options{
        HeaderName: "X-Tenant-ID",
    }))
    api.Use(tenantratelimit.Middleware(tenantratelimit.Options{
        Limiter: rateLimiter,
    }))
    api.Use(tenantquota.Middleware(tenantquota.Options{
        Manager: quotaMgr,
        Hooks: tenant.Hooks{
            OnQuota: func(ctx context.Context, decision tenant.QuotaDecision) {
                if !decision.Allowed {
                    log.Printf("quota exceeded for %s", decision.TenantID)
                }
            },
        },
    }))
    api.Use(tenantpolicy.Middleware(tenantpolicy.Options{
        Evaluator: policyEval,
    }))

    api.Get("/users", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        rows, err := tenantDB.QueryFromContext(
            r.Context(),
            "SELECT id, email FROM users WHERE active = ?",
            true,
        )
        if err != nil {
            contract.WriteError(w, r, contract.APIError{
                Status:   http.StatusInternalServerError,
                Code:     "db_query_failed",
                Message:  "query failed",
                Category: contract.CategoryServer,
            })
            return
        }
        defer rows.Close()
        _ = contract.WriteResponse(w, r, http.StatusOK, map[string]string{"status": "ok"}, nil)
    }))

    return app
}
```

If you manage tenant config in code (or provide your own config provider), you can set multi-window quotas like:

```go
tenantMgr := tenant.NewInMemoryConfigManager()
tenantMgr.SetTenantConfig(tenant.Config{
    TenantID: "tenant-id",
    Quota: tenant.QuotaConfig{
        Limits: []tenant.QuotaLimit{
            {Window: tenant.QuotaWindowDay, Requests: 200000},
            {Window: tenant.QuotaWindowMonth, Tokens: 10_000_000},
        },
    },
    Policy: tenant.PolicyConfig{
        AllowedModels: []string{"gpt-4o-mini"},
        AllowedTools:  []string{"search"},
    },
    RateLimit: tenant.RateLimitConfig{
        RequestsPerSecond: 50,
        Burst:             100,
    },
})

// Route policy cache (optional)
routePolicyStore := tenant.NewInMemoryRoutePolicyStore()
_ = routePolicyStore.SetRoutePolicy(context.Background(), tenant.RoutePolicy{
    TenantID: "tenant-id",
    Strategy: "weighted",
    Payload:  []byte(`{"rules":[{"provider":"a","weight":70},{"provider":"b","weight":30}]}`),
})
routePolicyCache := tenant.NewInMemoryRoutePolicyCache(1000, 5*time.Minute)
routePolicyProvider := tenant.NewCachedRoutePolicyProvider(routePolicyStore, routePolicyCache)

// Optional: tenant-aware token bucket from the same config manager.
rateLimiter := tenant.NewTokenBucketRateLimiter(
    &tenant.RateLimitConfigProviderFromConfig{Manager: tenantMgr},
)
_ = routePolicyProvider
_ = rateLimiter
```

### Automatic Query Filtering

The `x/tenant/store/db` `TenantDB` wrapper automatically filters all queries by tenant ID:

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
app.Use(auth.Authenticate(jwtManager.Authenticator(jwt.TokenTypeAccess)))
app.Use(auth.SessionCheck(sessionStore, sessionValidator))
app.Use(auth.Authorize(jwt.PolicyAuthorizer{Policy: jwt.AuthZPolicy{AnyRole: []string{"admin"}}}, "", ""))

protected := middleware.Apply(
	http.HandlerFunc(adminHandler),
	auth.Authenticate(jwtManager.Authenticator(jwt.TokenTypeAccess)),
	auth.SessionCheck(sessionStore, sessionValidator),
	auth.Authorize(jwt.PolicyAuthorizer{Policy: jwt.AuthZPolicy{AnyRole: []string{"admin"}}}, "", ""),
)
```

The `security/jwt` package provides adapters (`jwtManager.Authenticator`, `jwt.PolicyAuthorizer`, `jwt.PermissionAuthorizer`) that implement these contracts while keeping your own storage and policy engines decoupled.

## Reference App
`reference/standard-service` is the canonical minimal `main` package. It depends only on stable root packages and demonstrates explicit wiring instead of extension assembly:

- Configured WebSocket hub with JWT keys and broadcast endpoint
- Inbound GitHub/Stripe Webhooks, publishing to in-process Pub/Sub
- In-memory store for outbound Webhook management
- Static frontend served from embedded resources
- Prometheus metrics, OpenTelemetry tracing, and health endpoints mounted to the router

Run it with:

```bash
go run ./reference/standard-service
```

## Health Endpoints
HTTP probe and diagnostics handlers now live in `x/ops/healthhttp`, while `health` stays focused on managers, state, and check primitives:

```go
healthManager, err := health.NewHealthManager(health.HealthCheckConfig{})
if err != nil {
    log.Fatal(err)
}

app := core.New(core.WithHealthManager(healthManager))
app.Get("/health/ready", opshealth.ReadinessHandler(healthManager).ServeHTTP)
app.Get("/health", opshealth.SummaryHandler(healthManager).ServeHTTP)
app.Get("/health/build", opshealth.BuildInfoHandler().ServeHTTP)
```

```go
import (
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/health"
    opshealth "github.com/spcent/plumego/x/ops/healthhttp"
)

app := core.New(core.WithHealthManager(healthManager))
app.Get("/health/ready", opshealth.ReadinessHandler(healthManager).ServeHTTP)
app.Get("/health", opshealth.SummaryHandler(healthManager).ServeHTTP)
app.Get("/health/build", opshealth.BuildInfoHandler().ServeHTTP)
```

`opshealth.ReadinessHandler` returns readiness from the provided `HealthManager` (200 when ready, otherwise 503). When the manager is attached via `core.WithHealthManager`, the core lifecycle updates readiness automatically.

## Observability Adapters
No need to write your own adapters to hook logging middleware into metrics/tracing backends:

- `metrics.NewPrometheusCollector(namespace)` implements `httpmetrics.Observer`; pair it with `metrics.NewPrometheusExporter(collector)` for `/metrics`.
- `metrics.NewOpenTelemetryTracer(name)` implements `tracing.Tracer`, emitting spans with HTTP metadata.

For observability-heavy applications, keep the concrete collector and tracer in your application wiring, pass the collector to `core.WithHTTPMetrics(...)`, and mount request metrics explicitly with `httpmetrics.Middleware(app.HTTPMetrics())`.
For narrower DI at module boundaries, prefer `metrics.HTTPObserver`, `metrics.MQObserver`, `metrics.DBObserver`, or `metrics.Recorder` instead of the full `metrics.AggregateCollector` when a call site only needs one capability.

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

Keep configuration loading in your `main` package. Parse env vars, flags, or files into an `AppConfig`, then pass concrete values into `core.New(...)`. Canonical app scaffolds keep any shared helpers under app-local `internal/config` rather than a public root package.

## Development and Testing
- Install Go 1.24+ (matching `go.mod`).
- Run tests: `go test ./...`
- Use Go toolchain for formatting and static checks (`go fmt`, `go vet`).

## Development Server with Dashboard

The `plumego` CLI includes a powerful development server built with the plumego framework itself. It provides hot reload, real-time monitoring, and a web-based dashboard for enhanced development experience.

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
Canonical docs entrypoint and priority order: `docs/README.md`.
