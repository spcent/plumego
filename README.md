# Plumego — Standard Library Web Toolkit

[![Go Version](https://img.shields.io/badge/Go-1.26%2B-00ADD8?style=flat&logo=go)](https://go.dev/)
[![Status](https://img.shields.io/badge/status-v1.1.0-blue)](https://github.com/spcent/plumego/releases/tag/v1.1.0)
[![License](https://img.shields.io/badge/license-MIT-green)](LICENSE)

Plumego is a small Go HTTP toolkit built on the standard library. It keeps
`net/http` compatibility at the center: handlers are ordinary
`func(http.ResponseWriter, *http.Request)`, middleware wraps `http.Handler`,
and application wiring stays explicit in your own `main` package.

The stable surface is intentionally narrow. Start with `core`, `router`,
`contract`, and `middleware`; add `security`, `store`, `health`, `log`, and
`metrics` only when those responsibilities are needed.

## Quick Start

Create `main.go`:

```go
package main

import (
	"log"
	"net/http"

	"github.com/spcent/plumego/contract"
	"github.com/spcent/plumego/core"
)

func main() {
	cfg := core.DefaultConfig()
	cfg.Addr = ":8080"
	app := core.New(cfg, core.AppDependencies{})
	if err := app.Get("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = contract.WriteResponse(w, r, http.StatusOK, map[string]string{"message": "pong"}, nil)
	})); err != nil {
		log.Fatal(err)
	}
	if err := app.Prepare(); err != nil {
		log.Fatal(err)
	}
	srv, err := app.Server()
	if err != nil {
		log.Fatal(err)
	}
	log.Fatal(srv.ListenAndServe())
}
```

Run it:

```bash
go mod init example.com/hello
go get github.com/spcent/plumego@latest
go run main.go
```

Open `http://localhost:8080/ping`.

For the production-style application layout, read
[`reference/standard-service`](./reference/standard-service) after this
example.

## Choose Your Starting Point

Pick the scenario that matches your project:

| I want to build... | Start here |
| --- | --- |
| A plain JSON API | `reference/standard-service` → stable roots only |
| REST resources with CRUD conventions | `reference/with-rest` → `x/rest` |
| A multi-tenant SaaS API | `reference/with-tenant` → `x/tenant` |
| An API gateway or reverse proxy | `reference/with-gateway` → `x/gateway` |
| Real-time WebSocket features | `reference/with-websocket` → `x/websocket` |
| An AI-backed service | `reference/with-ai` → `x/ai/provider` |
| A service with rich messaging/webhooks | `reference/with-messaging` → `x/messaging` |
| A gRPC + HTTP service | `reference/with-rpc` → `x/rpc` |
| Observability (Prometheus / OpenTelemetry) | `reference/with-observability` → `x/observability` |
| A tenant administration console | `reference/with-tenant-admin` → `x/tenant` |

All paths keep `reference/standard-service` as the base layout. Extensions are
explicit additions, not alternate bootstraps.

## Why plumego

Plumego is for Go services that need more structure than raw `http.ServeMux`
without taking on a large framework model.

| Principle | How plumego applies it |
| --- | --- |
| Standard library first | Uses ordinary handlers, middleware, requests, response writers, and `*http.Server` values. |
| Explicit wiring | Routes, middleware, dependencies, and lifecycle are visible at construction sites. |
| Small stable surface | Stable roots have narrow ownership and avoid feature catalogs. |
| Agent-friendly maintenance | `specs/`, `tasks/`, and per-module `module.yaml` files make scope and validation discoverable. |
| Optional capabilities | Product features and protocol adapters live outside the stable learning path. |

## stdlib comparison

| Feature | `http.ServeMux` | plumego |
| --- | --- | --- |
| Basic routing | Method handling is caller-owned. | Method helpers such as `Get`, `Post`, and `AddRoute` register one method, one path, one handler. |
| `{param}` path extraction | Caller parses path segments manually. | Route params are matched by the router and read from the request context. |
| Route groups | Caller repeats prefixes manually. | Groups apply a prefix to related route registrations. |
| Per-group middleware | Caller composes handlers manually per subtree. | Groups can carry shared middleware while preserving `http.Handler` shape. |
| Named routes + reverse URL | Caller builds URLs manually. | Named routes support reverse URL generation through the app/router API. |
| Route freeze after preparation | Routes can be changed whenever caller mutates wiring. | `Prepare` freezes route mutation before serving. |
| Structured error responses | Caller defines every response shape. | `contract.WriteError` provides the canonical structured error path. |
| Request ID context carriage | Caller chooses and propagates a convention. | Request ID helpers use explicit context accessors and middleware support. |
| Graceful lifecycle | Caller builds server setup and shutdown policy. | `Prepare`, `Server`, and `Shutdown` keep lifecycle explicit and reusable. |

## Package overview

| Package | Role |
| --- | --- |
| `core` | App construction, route registration entry points, middleware attachment, and server lifecycle. |
| `router` | Route matching, path params, groups, route metadata, and reverse URL generation. |
| `contract` | Canonical response writers, structured error builders, request metadata, and transport binding helpers. |
| `middleware` | Transport-only middleware composition and first-party middleware packages. |
| `security` | Authentication, JWT, password, security header, input-safety, and abuse guard primitives. |
| `store` | Stable storage contracts and in-memory primitives for cache, KV, file, DB, and idempotency use. |
| `health` | Health and readiness models for application and dependency status. |
| `log` | Minimal logging interfaces and default logger implementation. |
| `metrics` | Minimal metrics contracts for counters, gauges, timings, and collectors. |

Optional capability families live under `x/*`. Treat them as additions to the
stable root path, not as alternate application layouts.

## Agent-First Design

Plumego is maintained with an agent-first control plane: `docs/` explains the
architecture, `specs/` records machine-checkable boundaries, `tasks/` defines
reviewable execution units, and `reference/` shows canonical wiring. The checks
under `internal/checks/` enforce the most important boundaries locally and in
CI, so automated changes produce reviewable evidence instead of relying on
implicit convention. See [`docs/AGENT_FIRST.md`](./docs/AGENT_FIRST.md) for the
external model and adoption path, and
[`docs/agent-first-operating-reference.md`](./docs/agent-first-operating-reference.md)
for the detailed internal operating reference.

## Getting Help

- Start with [`docs/getting-started.md`](./docs/getting-started.md) for the
  smallest runnable walkthrough.
- Use [`reference/standard-service`](./reference/standard-service) as the
  canonical app layout.
- Read [`docs/CANONICAL_STYLE_GUIDE.md`](./docs/CANONICAL_STYLE_GUIDE.md) for
  handler, middleware, routing, and dependency-injection conventions.
- Browse [`docs/modules`](./docs/modules) for package-specific primers.
- Check [`docs/ADOPTION_PATH.md`](./docs/ADOPTION_PATH.md) for the 5-minute,
  30-minute, and 1-day adoption path.
- See [`docs/troubleshooting.md`](./docs/troubleshooting.md) for common
  problems: route freeze, middleware order, JWT errors, and lifecycle issues.
- See [`docs/benchmarks/README.md`](./docs/benchmarks/README.md) for
  performance comparisons with Chi, Gin, and Echo.
