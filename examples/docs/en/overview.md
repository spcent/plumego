# Plumego Overview

**Plumego is a standard-library-first Go web framework built for real-world services.**

It provides a **production-grade runtime** that covers the parts most teams eventually have to build themselves:
security, lifecycle, observability, integrations, and safe extensibility —  
without locking you into a platform or hiding behavior behind magic.

Plumego is not optimized for demos.  
It is optimized for services that need to **run reliably for years**.

---

## Why Plumego exists

Most Go web projects start simple:

- a router
- a few middlewares
- some handlers

But as the service grows, teams inevitably need to add:

- graceful shutdown
- concurrency limits
- request timeouts and body limits
- authentication and session management
- webhooks and third-party integrations
- observability and health checks
- background tasks and lifecycle hooks

In many projects, these concerns are added **incrementally, inconsistently, and often incorrectly**.

Plumego exists to solve this exact problem.

> **Not by becoming a giant platform,  
but by providing a stable, explicit runtime foundation  
for real business services.**

---

## What Plumego focuses on (and what it does not)

### Plumego focuses on
- **Service runtime correctness**
- **Security and authentication boundaries**
- **Lifecycle and operational safety**
- **Integrations (webhooks, events, realtime)**
- **Observability without vendor lock-in**
- **Explicit, reviewable composition**

### Plumego deliberately does not include
- An ORM
- A dependency injection container
- An admin panel
- A plugin marketplace
- A “one true” application architecture

Those choices are left to your application.

Plumego defines the **runtime**, not the **business**.

---

## Core capabilities (business-driven)

### 1. Production-grade HTTP runtime

Plumego ships with runtime guarantees that most real services eventually need:

- Graceful startup and shutdown
- Concurrency limits and backpressure
- Request timeouts
- Request body size limits
- Panic recovery
- Readiness and liveness endpoints
- Optional TLS / HTTP2 support

These are not examples or suggestions.  
They are **first-class runtime features**.

---

### 2. Explicit routing and middleware

Routing and middleware are built directly on `net/http` semantics.

- Deterministic route matching
- Grouped routes with scoped middleware
- Explicit middleware ordering (code order = execution order)
- No hidden hooks or implicit lifecycle stages

If you understand `http.Handler`, you understand Plumego.

---

### 3. Security and authentication as architecture, not addons

Plumego treats security as a **core boundary**, not a plugin.

It defines clear contracts for:

- Authentication (who you are)
- Authorization (what you can do)
- Session and token lifecycle (whether you are still valid)

Instead of forcing a single solution, Plumego provides:
- stable interfaces
- recommended patterns
- reference implementations

Including support for:
- access + refresh tokens
- token rotation
- session revocation
- multi-device login
- role-based authorization

Security is structured so it can be **audited, tested, and evolved safely**.

---

### 4. Webhooks, events, and integrations

Modern services are rarely isolated.

Plumego includes built-in support for:

- Inbound webhooks (with signature verification and replay protection)
- Outbound webhooks (delivery, retry, replay)
- In-process Pub/Sub for event fan-out
- Debug and snapshot tools for integration flows

This makes Plumego well-suited for:
- SaaS backends
- platform services
- integration-heavy systems

---

### 5. Realtime capabilities (when you need them)

Plumego includes optional WebSocket utilities:

- Connection lifecycle management
- JWT-protected endpoints
- Broadcast and fan-out helpers
- Backpressure-aware design

It is not an IM framework.  
It is designed for **business-level realtime needs** such as notifications, status updates, and dashboards.

---

### 6. Observability without lock-in

Plumego provides observability as **interfaces and adapters**, not hard dependencies.

- Metrics collection (Prometheus adapter included)
- Distributed tracing (OpenTelemetry adapter included)
- Request ID and trace propagation
- Structured logging compatibility

You can start simple and grow — without rewriting your core runtime.

---

### 7. Configuration and operations consistency

Plumego provides a consistent model for:

- Environment-based configuration
- `.env` and flags
- Runtime validation
- Safe reloads
- Predictable startup order

This reduces deployment-time surprises and environment-specific bugs.

---

## Architecture model

At the center of Plumego is an explicit application runtime:

- `core.App` implements `http.Handler`
- Capabilities are added via **components**
- Components can:
  - register routes
  - register middleware
  - participate in startup and shutdown
  - expose health status

This makes the system:
- modular
- testable
- auditable

No global magic. No hidden registries.

---

## Quick start

### Minimal server

```go
app := plumego.New()

app.GetCtx("/health", func(ctx *plumego.Context) {
    ctx.JSON(200, map[string]string{"status": "ok"})
})

log.Fatal(http.ListenAndServe(":8080", app))
```

### Production-oriented server

```go
app := core.New(
    core.WithAddr(":8080"),
    core.WithRecovery(),
    core.WithLogging(),
    core.WithCORS(),
    core.WithGracefulShutdown(),
)

app.Get("/ping", func(w http.ResponseWriter, _ *http.Request) {
    w.Write([]byte("pong"))
})

if err := app.Boot(); err != nil {
    log.Fatal(err)
}
```

---

## Who Plumego is for

Plumego is a strong fit if you are:

* Building services expected to live for years
* Working in a team that values code review and clarity
* Operating internal platforms, SaaS backends, or infrastructure services
* Tired of re-implementing the same runtime and security concerns

It may not be the best choice if you want:

* maximum scaffolding
* rapid demos with minimal decisions
* a full application platform out of the box

---

## How Plumego compares

Plumego sits between minimal routers and full frameworks:

* More complete than “router + DIY runtime”
* Less opinionated than full application platforms
* More explicit and predictable than convenience-first frameworks

> Plumego optimizes for **long-term correctness**, not short-term velocity.

---

## Stability promise

Plumego follows semantic versioning.

Starting with v1.0:

* Core APIs are stable
* Behavior changes are documented
* Breaking changes require a major version
* Examples and contracts are treated as part of the API

Upgrading should never feel surprising.

---

## Final note

Plumego is intentionally calm.

It does not try to impress you with features.
It tries to earn your trust by staying understandable under pressure.

If you are building real services —
services that must be secure, observable, and maintainable —

**Plumego is built for that reality.**
