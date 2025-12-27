# Plumego documentation hub (English)

This folder mirrors the source tree so every feature has a dedicated, copy-pastable guide. Start here if you want a tour of what ships with Plumego and how to stitch the pieces together in your own `main` package.

## What you get
- Lightweight HTTP lifecycle built on the Go standard library (`net/http`, `context`).
- Trie-based router with groups, parameters, and middleware inheritance.
- Guardrail middleware (recovery, logging, gzip, CORS, timeout, body/concurrency limits, rate limit, auth helpers).
- Built-in WebSocket hub with JWT auth + broadcast endpoint.
- In-process Pub/Sub bus that can fan out webhook events or WebSocket payloads.
- Webhook receivers (GitHub/Stripe) and outbound webhook service with trigger tokens.
- Metrics (Prometheus), tracing (OpenTelemetry), readiness/build health endpoints.
- Static frontend mounting from disk or embedded assets.

## Quick start (reference app)
Run the reference application to see everything working end to end:

```bash
go run ./examples/reference
```

It starts an HTTP server on `:8080`, serves the docs in this folder, exposes `/metrics`, `/health/ready`, `/health/build`, mounts a WebSocket hub at `/ws`, and wires inbound/outbound webhooks through the Pub/Sub bus.

## Build your own app
Create a small `main.go` and register routes before calling `Boot()`:

```go
package main

import (
    "log"
    "net/http"

    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/middleware"
)

func main() {
    app := core.New(core.WithAddr(":8080"))
    app.EnableRecovery()
    app.EnableLogging()

    app.Use(middleware.Gzip())
    app.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
        w.Write([]byte("pong"))
    })

    if err := app.Boot(); err != nil {
        log.Fatalf("server stopped: %v", err)
    }
}
```

Want the same features as the reference app? Add `core.WithPubSub`, `core.ConfigureWebSocketWithOptions`, webhook configs, and the metrics/health handlers described in the module pages below.

## Module index
- [Core](core.md): lifecycle, configuration, components, graceful shutdown.
- [Router](router.md): groups, parameters, context-aware handlers, static frontends.
- [Middleware](middleware.md): built-in guardrails and composition patterns.
- [Pub/Sub & WebSocket](pubsub-websocket.md): real-time messaging, hub wiring, and fan-out examples.
- [Metrics & Health](metrics-health.md): Prometheus/OpenTelemetry hooks, readiness/build probes.
- [Security & Webhook](security-webhook.md): inbound signature verification, outbound delivery management, and secrets.

Each page includes runnable code snippets. Copy them into `examples/reference` or your own service to verify behavior.
