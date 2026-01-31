# AGENTS.md — plumego

This document provides **operational guidance for AI coding agents**
(Copilot, Codex, Cursor, Claude Code, etc.) working in the `spcent/plumego`
repository.

Its purpose is to ensure that any automated or semi-automated changes are:

- Consistent with the existing codebase
- Aligned with the project’s architectural philosophy
- Safe, reviewable, and reversible
- Compatible with Go’s standard library (`net/http`) design model

This file is authoritative for agent behavior.

---

## 1. Project Overview (Read First)

**plumego** is a **lightweight Go web toolkit** built primarily on the Go
standard library.
It is designed to be **embedded into applications**, not used as a
monolithic framework.

Core characteristics:

- Standard library–first (`net/http`, `context`, `http.Handler`)
- Explicit lifecycle (`core.New(...)` → configuration → `Boot()`)
- Minimal dependencies
- Composable router and middleware
- Built-in support for:
  - Routing and middleware chains
  - Graceful startup / shutdown
  - WebSocket helpers
  - Webhook handling and verification
  - In-process Pub/Sub
  - Static frontend serving
  - Lightweight internal KV storage

**Non-goals**:

- Replacing large Go frameworks (Gin, Echo, Fiber, etc.)
- Introducing heavy dependency trees
- Hiding or abstracting away `net/http`
- Providing opinionated ORM, RPC, or persistence layers

Agents must respect these boundaries.

---

## 2. Supported Go Environment

- Go version: defined by `go.mod`
- The codebase assumes modern Go features (modules, context, embed, etc.)

Before making changes, agents should assume:

```bash
go test -timeout 20s ./...
go vet ./...
gofmt -w .
```

must pass without errors.

---

## 3. Repository Structure & Responsibilities

Agents **must not blur module boundaries**.

### Core Directories

#### `core/`

Application core and lifecycle management.

Responsibilities:

* Application construction (`core.New(...)`)
* Global configuration wiring
* Enabling built-in middleware (`WithRecovery`, `WithLogging`, `WithCORS`, etc.)
* Router registration
* Server startup and graceful shutdown
* `Boot()` execution flow

Public APIs in `core` are **high-stability**.
Breaking changes here require strong justification and documentation.

---

#### `router/`

HTTP routing and request dispatch.

Responsibilities:

* Path matching
* Route groups and prefixes
* Path parameters (e.g. `/:id`)
* Route-level and group-level middleware composition
* Registration vs. frozen routing state

Do **not** move routing logic into `core`.

---

#### `middleware/`

HTTP middleware implementations.

Responsibilities:

* Logging
* Recovery / panic protection
* CORS
* Gzip / compression
* Timeout
* Rate limiting / concurrency limiting
* Body size limits
* Authentication helpers

Middleware must:

* Be composable
* Follow `http.Handler` semantics
* Avoid global mutable state unless explicitly documented

---

#### `frontend/`

Static frontend serving.

Responsibilities:

* Serving static files from disk or embedded assets
* SPA / static site mounting
* Safe defaults for caching and paths

---

#### `pubsub/`

In-process publish/subscribe system.

Responsibilities:

* Internal event distribution
* Webhook fan-out support
* Lightweight event channels
* Debug or inspection hooks (if present)

This is **not** a distributed message queue.

---

#### `security/`

Security-critical logic.

Responsibilities:

* Webhook signature verification
* Token / secret validation
* Cryptographic helpers
* Authentication primitives

Rules:

* No secrets in logs
* Fail closed
* Prefer explicit verification APIs

---

#### `health/`

Health and readiness signaling.

Responsibilities:

* Liveness / readiness indicators
* Startup and shutdown state tracking
* Health endpoints or internal probes

---

#### `log/`

Logging abstraction and helpers.

Responsibilities:

* Structured logging adapters
* Log levels and formatting
* Integration points for middleware and core

---

#### `config/`

Configuration loading.

Responsibilities:

* Environment variable parsing
* `.env` file loading (if enabled)
* Defaults and overrides
* Explicit configuration errors

Any new config must update `env.example`.

---

#### `store/kv/`

Internal key-value storage.

Responsibilities:

* Lightweight persistence
* Deduplication helpers
* Webhook delivery state
* Internal coordination

Not a general database layer.

---

#### `net/`

Network utilities.

Responsibilities:

* HTTP/network helpers
* Timeout or connection wrappers
* Standard-library–compatible extensions

---

#### `ipc/`

Local inter-process or intra-process communication helpers.

Responsibilities:

* Internal signaling
* Local control channels (if present)

---

#### `utils/`

Small shared helpers.

Rules:

* No business logic
* No cross-layer coupling
* No hidden dependencies

---

## 4. Change Rules (Strict)

### Compatibility Rules

Agents **must**:

* Preserve `net/http` compatibility
* Avoid introducing incompatible abstractions
* Keep handlers as `http.Handler` or `http.HandlerFunc`
* Avoid hidden global side effects

---

### API Stability Rules

* Public APIs in `core`, `router`, and `middleware` are stable by default
* Breaking changes require:

  * Clear motivation
  * Migration notes
  * Preferably backward compatibility

---

### Dependency Rules

* New dependencies require strong justification
* Prefer standard library solutions
* No “utility” dependencies without necessity

---

## 5. Testing & Quality Gates

Before submitting or generating patches, agents must ensure:

```bash
go test -timeout 20s ./...
go vet ./...
gofmt -w .
```

Additional expectations:

* Routing changes → route matching tests
* Middleware changes → chain order and error path tests
* Security changes → negative tests (invalid signature, invalid token)
* Config changes → default and missing-value behavior tests

---

## 6. Common Agent Tasks

### A. Adding a New Middleware

1. Implement in `middleware/`
2. Follow existing middleware signatures
3. Avoid global state
4. Add at least minimal tests
5. If auto-enabled via `core`, provide a configuration toggle

---

### B. Modifying Routing Behavior

1. Work exclusively in `router/`
2. Preserve existing semantics
3. Add tests for:

   * Static routes
   * Parameterized routes
   * Group prefixes
   * Middleware stacking order

---

### C. Adding Environment Configuration

1. Add to `config/`
2. Update `env.example`
3. Document default behavior
4. Fail fast on invalid critical config

---

### D. Webhook Enhancements

1. Signature and verification logic belongs in `security/`
2. Routing and dispatch belong in `core` or webhook-specific wiring
3. Never log raw secrets or signatures
4. Prefer explicit verification APIs

---

## 7. Documentation Synchronization Rules

Agents **must update documentation** when changing:

* Public APIs
* Environment variables
* Default limits (timeouts, body size, concurrency)
* Security behavior
* Startup or shutdown semantics

Primary docs:

* `README.md`
* `README_CN.md`
* `env.example`

---

## 8. Security & Disclosure

* Follow `SECURITY.md` for vulnerability reporting
* Never expose secrets in logs, issues, or PRs
* Assume logs may be public or centralized

---

## 9. Recommended Agent Workflow

1. Identify the correct module boundary
2. Add or update tests first (or alongside changes)
3. Make minimal, focused changes
4. Run full test suite
5. Ensure documentation parity
6. Keep commits small and reversible

---

**plumego values clarity, restraint, and correctness over feature volume.
Agents should optimize for maintainability, not cleverness.**
