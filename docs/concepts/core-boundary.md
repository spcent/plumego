# Core Boundary

This document defines the engineering boundary for Plumego's stable root
packages. It is the authoritative reference for deciding whether new capability
belongs in a stable root or in an `x/*` extension family.

Read this before making any change to `core`, `router`, `contract`, `middleware`,
`security`, `store`, `health`, `log`, or `metrics`.

For machine-readable boundary rules, see `specs/dependency-rules.yaml`.
For the full repository shape, see `docs/concepts/agent-first-repo-blueprint.md`.

---

## Why the Boundary Exists

Stable roots carry a long-term compatibility promise. Every exported symbol in a
stable root is a commitment: changing it requires a deprecation period, a
migration path, and explicit release notes.

Keeping the boundary narrow means that commitment remains affordable. A stable
root that absorbs every fast-moving concern eventually becomes too coupled to
change safely, too wide for new engineers to understand, and too risky for AI
agents to modify without breaking hidden dependencies.

The boundary is not about technical purity. It is about protecting the surface
where the compatibility promise is strongest.

---

## Package Responsibilities

### `core`

**Owns:** App construction, dependency wiring entry point, middleware attachment,
route group setup, graceful shutdown, server lifecycle (prepare, serve, shutdown).

**Does not own:** Configuration file parsing, service discovery, ORM, connection
pooling, task scheduling, DI containers, plugin registration, global state.

The kernel is the wiring point, not a feature catalog. If a capability can start
outside `core` and be passed in through `core.AppDependencies`, it does not
belong in `core`.

---

### `router`

**Owns:** Route matching, path parameter extraction, route groups, static path
mounting, reverse routing, route tree management, route freeze.

**Does not own:** Controller scanning, annotation-based routing, response
formatting, request binding, JSON encoding, repository injection, middleware
policy decisions.

A route is a mapping from a method and path to a handler. Everything else is the
handler's responsibility.

---

### `contract`

**Owns:** Transport-level response helpers (`WriteResponse`, `WriteError`),
structured error types, request metadata extraction, context accessors
(`With{Type}` / `{Type}FromContext`), request binding helpers.

**Does not own:** Business domain types, service-layer error hierarchies, ORM
entities, business validation rules, service injection, session data.

`contract` defines how the transport layer communicates results. It does not
define what the results mean in the business domain.

---

### `middleware`

**Owns:** Transport-level cross-cutting concerns: request ID propagation,
structured access logging, panic recovery, response timeout, gzip, CORS,
authentication header extraction, rate limiting at the transport layer, request
body size limits, security headers.

**Does not own:** Business authorization decisions, tenant resolution, domain
policy, ORM lookups in request handling, response body transformation based on
business rules, service-layer calls.

Middleware runs before the handler. It must not know what the handler does with
the request. If a middleware needs to call a service, it is not transport-level
middleware — it is a handler component.

---

### `security`

**Owns:** JWT signing and verification, password hashing and comparison
(`bcrypt`-backed), security header policy helpers, input safety validators (XSS
prevention, path traversal checks), abuse-guard rate-limiting primitives,
timing-safe comparison utilities.

**Does not own:** Full account management systems, OAuth provider clients,
session storage backends, multi-factor authentication flows, identity provider
integration, role and permission models.

`security` provides the building blocks for secure handlers. It does not provide
a complete identity or authorization platform.

---

### `store`

**Owns:** Storage interface contracts, idempotency record types and repository
contracts, file storage contracts.

**Does not own:** ORM query builders, database migration runners, connection pool
management, Redis client wrappers, provider-specific storage implementations,
tenant-scoped storage routing.

`store` defines what persistent storage looks like from the application's
perspective. Concrete implementations and advanced topology live in `x/data`.

---

### `health`

**Owns:** Health check registration contracts, readiness check models, checker
interface, check result types, HTTP health handler that the caller mounts
explicitly.

**Does not own:** HTTP handler ownership at a fixed path, external orchestration
integration, service-mesh sidecar lifecycle, Kubernetes readiness/liveness probe
policy.

The caller mounts the health handler at a path they choose. `health` does not
decide where or how the endpoint is exposed.

---

### `log`

**Owns:** Structured logging contracts (`Logger` interface), default logger
construction, log level types, context-aware log entry helpers.

**Does not own:** Log aggregation backends, cloud-provider logging SDKs, log
shipping configuration, Loki/Datadog/CloudWatch adapters.

Logger adapters that integrate with external systems belong in `x/observability`.

---

### `metrics`

**Owns:** Metrics contracts (`Counter`, `Gauge`, `Histogram` interfaces), default
no-op implementations, basic in-process collectors.

**Does not own:** Prometheus exposition format, OpenTelemetry SDK, metrics
export configuration, dashboard definitions, alert rule templates.

Metric exporters and adapters belong in `x/observability`.

---

## Decision Checklist

Use this checklist before adding anything to a stable root:

```
[ ] Does this capability have a clear, narrow role in the HTTP transport layer?
[ ] Can we carry a three-year compatibility promise on every exported symbol?
[ ] Does it avoid third-party imports not already present in the package?
[ ] Does it remain useful without any x/* extension being present?
[ ] Does it work correctly without knowledge of the caller's business domain?
```

If any box is unchecked, start in `x/*` or `reference/` instead.

---

## Wiring Patterns

### Constructor injection — required

```go
// Correct: caller constructs, caller owns
logger := plog.NewLogger(cfg.Log)
db := sql.Open("postgres", cfg.DB.DSN)
app := core.New(cfg.Core, core.AppDependencies{
    Logger: logger,
})
```

```go
// Incorrect: hidden global, caller loses control
core.SetGlobalLogger(logger)
app := core.Default()
```

### Caller-owned lifecycle — required

```go
// Correct: shutdown is explicit, context is caller-controlled
ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
defer cancel()
if err := app.Shutdown(ctx); err != nil {
    log.Printf("shutdown: %v", err)
}
```

```go
// Incorrect: framework owns the process
framework.Run(":8080") // hides signal handling, TLS config, shutdown
```

### Explicit extension mounting — required

```go
// Correct: caller decides what is exposed and under which path
debugHandler := devtools.NewHandler(cfg.Debug)
app.Mount("/_debug", authMiddleware(debugHandler))
```

```go
// Incorrect: extension self-registers
devtools.Enable() // unknown path, unknown auth, hidden behavior
```

### App-local configuration — required

```go
// Correct: caller controls all parameters
cfg := core.DefaultConfig()
cfg.Addr = ":8080"
cfg.ReadTimeout = 5 * time.Second
cfg.WriteTimeout = 10 * time.Second
```

```go
// Incorrect: framework reads environment without caller involvement
core.AutoConfig() // opaque, untestable, hard to override
```

---

## What Belongs in `x/*` Instead

When a capability does not meet the checklist above, it belongs in an `x/*`
extension family. Common examples:

| Capability | Correct location | Reason |
|---|---|---|
| Redis cache | `x/data/cache` or `x/data` | External dependency |
| Tenant resolution and policy | `x/tenant` | Business topology |
| WebSocket hub | `x/websocket` | Not universal transport need |
| Admin and ops routes | `x/observability/ops` | Requires explicit auth boundary |
| OpenTelemetry exporters | `x/observability` | Ecosystem-specific |
| REST resource controllers | `x/rest` | Convention layer, not HTTP kernel |
| Circuit breaker | `x/resilience` | Optional reliability primitive |
| Message queue integration | `x/messaging` / `x/messaging/mq` | External system dependency |

The test: if removing the capability from `x/*` leaves the stable root intact
and all service code still compiles against the stable root, the separation is
correct.

---

## Enforcement

Boundary violations are caught automatically:

```bash
go run ./internal/checks/dependency-rules
go run ./internal/checks/module-manifests
```

These run as part of `make gates` and in CI. A stable root importing `x/*` is a
hard violation. A stable root acquiring a new third-party dependency without
explicit approval is a hard violation.
