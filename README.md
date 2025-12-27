# Plumego — standard-library web toolkit

Plumego is a small Go HTTP toolkit that keeps everything inside the standard library while still covering routing, middleware, graceful shutdown, WebSocket helpers, webhook plumbing, and static frontend hosting. It is designed to be embedded in your own `main` package rather than acting as a framework binary.

## Highlights
- **Router with groups and params**: Trie-based matcher that supports `/:param` segments, route freezing, and middleware stacks per route/group.
- **Middleware chain**: Logging, recovery, gzip, CORS, timeout, rate limiting, concurrency limits, body size limits, and auth helpers that wrap standard `http.Handler` values.
- **Structured logging hooks**: Plug in a custom logger and collect metrics/traces via middleware hooks.
- **Graceful lifecycle**: Environment loading, connection draining, readiness flags, and optional TLS/HTTP2 configuration with sane defaults.
- **Optional services**: Built-in WebSocket hub with auth, in-process pub/sub with debug snapshots, inbound/outbound webhook routers, and static frontend mounting from disk or embedded assets.

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

## Development and testing
- Install Go 1.24+ (matching `go.mod`).
- Run tests: `go test ./...`
- Format and lint using the Go toolchain as needed (`go fmt`, `go vet`).

## Improvement ideas
- Provide a reference `main` package (or example folder) that demonstrates WebSocket setup, webhook routes, and static frontend mounting end-to-end to reduce onboarding friction.
- Add minimal HTTP examples for readiness/build info exposure so operators can plug Plumego into health check endpoints without re-implementing them.
- Offer metrics/tracing adapters (Prometheus/OpenTelemetry) alongside the `middleware.MetricsCollector` and `Tracer` hooks to make observability plug-and-play.
- Expand configuration docs to map every `AppConfig` field to environment variables or command-line flags for predictable deployments.
