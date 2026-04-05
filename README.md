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
- `specs/task-routing.yaml`
- `specs/extension-taxonomy.yaml`
- `specs/package-hotspots.yaml`
- `specs/dependency-rules.yaml`
- `specs/checks.yaml`
- `specs/change-recipes/*`
- `<module>/module.yaml`

For current priorities and remaining extension work, see `docs/ROADMAP.md`.

Machine-enforced repo guardrails live under `internal/checks/*` and are enforced directly in CI.

For new application work, use a single canonical path:

- read `reference/standard-service` first for structure and wiring
- `reference/standard-service` intentionally depends only on stable root packages; treat `x/*` examples as non-canonical

## v1 Support Matrix

Plumego v1 release scope covers every checked-in module in this repository, but the compatibility promise differs by layer.

| Area | v1 status | Compatibility promise | Modules |
| --- | --- | --- | --- |
| Stable library roots | GA | Public package surface is the long-term stable API for v1 users | `core`, `router`, `contract`, `middleware`, `security`, `store`, `health`, `log`, `metrics` |
| Canonical reference app | Supported reference | Kept aligned with the canonical bootstrap and stable-root usage, but not treated as a reusable extension catalog | `reference/standard-service` |
| CLI | Included in v1 release scope | Supported as a command-line tool, not as a Go import surface; command behavior and generated output must stay aligned with canonical docs | `cmd/plumego` |
| App-facing extension families | Experimental | Included in repo quality gates and release scope, but API/config compatibility is not frozen | `x/ai`, `x/data`, `x/devtools`, `x/discovery`, `x/frontend`, `x/gateway`, `x/messaging`, `x/observability`, `x/ops`, `x/rest`, `x/tenant`, `x/websocket` |
| Subordinate extension primitives | Experimental | Maintained and tested, but discovery should start from the owning family entrypoint and compatibility is not frozen | `x/ipc`, `x/mq`, `x/pubsub`, `x/scheduler`, `x/webhook` |

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

For the canonical quick-start path, read [`docs/getting-started.md`](./docs/getting-started.md) first, then open [`reference/standard-service`](./reference/standard-service).

Smallest runnable example:

```go
package main

import (
    "context"
    "errors"
    "log"
    "net/http"

    "github.com/spcent/plumego/contract"
    "github.com/spcent/plumego/core"
    plog "github.com/spcent/plumego/log"
    "github.com/spcent/plumego/middleware/requestid"
    "github.com/spcent/plumego/middleware/recovery"
)

func main() {
    ctx := context.Background()
    cfg := core.DefaultConfig()
    cfg.Addr = ":8080"
    app := core.New(cfg, core.AppDependencies{Logger: plog.NewGLogger()})

    if err := app.Use(
        requestid.Middleware(),
        recovery.Recovery(app.Logger()),
    ); err != nil {
        log.Fatalf("register middleware: %v", err)
    }

    if err := app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
        if err := contract.WriteResponse(w, r, http.StatusOK, map[string]string{
            "message": "pong",
        }, nil); err != nil {
            http.Error(w, "write response", http.StatusInternalServerError)
        }
    }); err != nil {
        log.Fatalf("register route: %v", err)
    }

    if err := app.Prepare(); err != nil {
        log.Fatalf("prepare server: %v", err)
    }
    srv, err := app.Server()
    if err != nil {
        log.Fatalf("get server: %v", err)
    }
    defer app.Shutdown(ctx)

    log.Println("server started at :8080")
    serveErr := srv.ListenAndServe()
    if srv.TLSConfig != nil {
        serveErr = srv.ListenAndServeTLS("", "")
    }
    if serveErr != nil && !errors.Is(serveErr, http.ErrServerClosed) {
        log.Fatalf("server stopped: %v", serveErr)
    }
}
```

## Configuration Basics
- Environment variables should be loaded explicitly in your `main` package. Keep `.env` path ownership in app-local config, for example `cfg.App.EnvFile` in the reference layout, when tooling such as devtools reload needs to know which file is active.
- `core` construction is config-first: start from `core.DefaultConfig()`, adjust the typed `core.AppConfig`, then pass it to `core.New(cfg, ...)`.
- `core.New(cfg, ...)` defaults to a `NoOpLogger`. If you expect request/runtime logs, inject a real logger with `core.AppDependencies{Logger: ...}`.
- Logger lifecycle ownership stays with the caller. `Prepare()` and `Shutdown(ctx)` do not initialize, flush, or close injected logger implementations for you.
- Common variables: `AUTH_TOKEN` (used by ops component defaults), `WS_SECRET` (WebSocket JWT signing key, at least 32 bytes), `WEBHOOK_TRIGGER_TOKEN`, `GITHUB_WEBHOOK_SECRET`, and `STRIPE_WEBHOOK_SECRET` (see `env.example`).
- `core.AppConfig` owns server address, TLS, and HTTP server timeout/hardening settings. Request body limits and concurrency limits belong to explicit middleware wiring, not to `core` itself.
- TLS stays on the same explicit serve path: `Prepare()` loads cert/key material into the prepared `*http.Server`, then callers choose `ListenAndServe()` or `ListenAndServeTLS("", "")` on the server returned by `Server()`.
- Security baseline should be composed explicitly via `app.Use(...)`, for example `middleware/security.SecurityHeaders(...)` and `middleware/ratelimit.AbuseGuard(...)`.
- Debug mode and devtools are separate. Keep debug flags in app-local config, for example `cfg.App.Debug` in the reference layout; if you need devtools, wire its routes explicitly in an app-local package instead of treating it as part of the canonical kernel path.
- Devtools endpoints under `/_debug` (routes, middleware, config, metrics, pprof, reload) are provided by `x/devtools`, not by `core` itself. These endpoints are intended for local development or protected environments; disable or gate them in production.
- When `x/devtools` is wired, `/_debug/config` exposes the stable runtime snapshot used by first-party tooling: address, env file, server timeouts, drain settings, TLS config, and the kernel `preparation_state`.

## Agent-First Workflow
- Canonical app bootstrap starts from `reference/standard-service`.
- Machine-readable discovery rules live in `specs/task-routing.yaml`.
- Module ownership, risk, and default validation live in each `<module>/module.yaml`.
- Standard change recipes live in `specs/change-recipes/*`.
- Module primers live in `docs/modules/*` and should match each manifest's `doc_paths`.
- Secondary task-family defaults are also explicit: frontend asset work starts in `x/frontend`, local debug work starts in `x/devtools`, service discovery starts in `x/discovery`, and protected admin surfaces start in `x/ops`.
- These secondary extension roots are capability entrypoints, not application bootstrap surfaces.

## Capability Guides

Use the root README as an entry page. Detailed capability guidance lives in `docs/modules/*`.

Stable roots:

- [core](./docs/modules/core/README.md) — app kernel, lifecycle, shared runtime wiring
- [router](./docs/modules/router/README.md) — matching, params, groups, reverse routing
- [middleware](./docs/modules/middleware/README.md) — transport-only middleware
- [contract](./docs/modules/contract/README.md) — response and error contracts
- [security](./docs/modules/security/README.md) — auth, headers, input-safety, abuse guard
- [store](./docs/modules/store/README.md) — persistence primitives
- [health](./docs/modules/health/README.md) — readiness state and health models
- [log](./docs/modules/log/README.md) and [metrics](./docs/modules/metrics/README.md) — base logging and metrics contracts

App-facing extension families:

- [x/tenant](./docs/modules/x-tenant/README.md) — multi-tenancy, quota, policy, tenant-aware data paths
- [x/rest](./docs/modules/x-rest/README.md) — reusable resource APIs and CRUD standardization
- [x/websocket](./docs/modules/x-websocket/README.md) — WebSocket transport
- [x/messaging](./docs/modules/x-messaging/README.md) — messaging entrypoint
- [x/fileapi](./docs/modules/x-fileapi/README.md) — file upload/download transport
- [x/gateway](./docs/modules/x-gateway/README.md) and [x/discovery](./docs/modules/x-discovery/README.md) — edge transport and service discovery
- [x/frontend](./docs/modules/x-frontend/README.md) — frontend asset serving
- [x/observability](./docs/modules/x-observability/README.md), [x/ops](./docs/modules/x-ops/README.md), and [x/devtools](./docs/modules/x-devtools/README.md) — observability, protected admin surfaces, and local debug tools
- [x/data](./docs/modules/x-data/README.md), [x/cache](./docs/modules/x-cache/README.md), and [x/ai](./docs/modules/x-ai/README.md) — topology-heavy data features, cache adapters, and AI capabilities

## Reference App
`reference/standard-service` is the canonical reference application. It depends only on stable root packages and demonstrates:

- default application layout
- explicit bootstrap flow in `main.go`
- explicit route registration in `internal/app/routes.go`
- app-local configuration under `internal/config`
- minimal stable-root-only wiring

Run it with:

```bash
go run ./reference/standard-service
```

## Further Reading

- [`docs/getting-started.md`](./docs/getting-started.md) — smallest runnable example
- [`docs/README.md`](./docs/README.md) — docs entrypoint
- [`env.example`](./env.example) — environment variable reference
- [`cmd/plumego/DEV_SERVER.md`](./cmd/plumego/DEV_SERVER.md) — dev server and dashboard details

## Development and Testing
- Install Go 1.24+ (matching `go.mod`).
- Run tests: `go test ./...`
- Use Go toolchain for formatting and static checks (`go fmt`, `go vet`).

## Development Server with Dashboard

The `plumego` CLI includes a powerful development server built with the plumego framework itself. It provides hot reload, real-time monitoring, and a web-based dashboard for enhanced development experience.

The dashboard is **enabled by default** - simply run `plumego dev` to get started.

**Positioning & Production Guidance**
- `cfg.App.Debug = true` in the reference layout exposes application devtools under `/_debug`. These are app endpoints and should be disabled or protected in production.
- `plumego dev` dashboard is a local developer tool that runs a separate dashboard server; it is not intended to be exposed publicly in production environments.
- The dashboard may query the app’s `/_debug` endpoints for routes/config/metrics/pprof when available, so keep debug endpoints gated outside local/dev usage.

### Start the Dev Server

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
