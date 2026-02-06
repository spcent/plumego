# AGENTS.md — plumego

This document provides **operational guidance for AI coding agents**
(Copilot, Codex, Cursor, Claude Code, etc.) working in the `spcent/plumego`
repository.

Its purpose is to ensure that any automated or semi-automated changes are:

- Consistent with the existing codebase
- Aligned with the project's architectural philosophy
- Safe, reviewable, and reversible
- Compatible with Go's standard library (`net/http`) design model

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
- Zero external dependencies (standard library only)
- Composable router and middleware
- Built-in support for:
  - Routing with path parameters and reverse routing
  - Middleware chains (19 subpackages)
  - Graceful startup / shutdown with connection draining
  - WebSocket hub with JWT authentication
  - Webhook handling (GitHub, Stripe) with signature verification
  - In-process Pub/Sub for event distribution
  - Task scheduling (cron, delayed jobs, retries)
  - Static frontend serving (disk or embedded assets)
  - Embedded KV storage with WAL and LRU eviction
  - AI agent gateway (SSE streaming, provider abstraction, session management)
  - Multi-tenancy with quota enforcement and policy evaluation
  - Service discovery (static, Consul)
  - HTTP reverse proxy with circuit breaker

**Non-goals**:

- Replacing large Go frameworks (Gin, Echo, Fiber, etc.)
- Introducing heavy dependency trees
- Hiding or abstracting away `net/http`
- Providing opinionated ORM, RPC, or persistence layers

Agents must respect these boundaries.

---

## 2. Supported Go Environment

- **Go version**: 1.24+ (see `go.mod`)
- **Toolchain**: go1.24.4
- **Dependencies**: Zero external dependencies (standard library only)
- The codebase assumes modern Go features (modules, context, embed, generics, etc.)

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

* Application construction (`core.New(...)`) with functional options
* Global configuration wiring
* Enabling built-in middleware (`WithRecovery`, `WithLogging`, `WithCORS`, etc.)
* Router registration
* Server startup and graceful shutdown
* `Boot()` execution flow
* Component and Runner lifecycle management
* Shutdown hooks

Subpackages:

* `core/di/` — Dependency injection container (Singleton, Transient, Scoped lifecycles)
* `core/components/` — Built-in components (devtools, observability, ops, tenant, webhook, websocket)
* `core/internal/` — Internal utilities

Public APIs in `core` are **high-stability**.
Breaking changes here require strong justification and documentation.

---

#### `router/`

HTTP routing and request dispatch.

Responsibilities:

* Trie-based path matching
* Route groups and prefixes
* Path parameters (e.g. `/:id`)
* Route-level and group-level middleware composition
* Reverse routing (`router.URL(name, params...)`)
* Named routes and route metadata
* Route validation and caching
* `ResourceController` interface for REST resources
* Registration vs. frozen routing state

Key types:

* `Handler = http.Handler` (standard library alias)
* `HandlerFunc = http.HandlerFunc` (standard library alias)
* `RouteRegistrar` interface for modular route registration

Do **not** move routing logic into `core`.

---

#### `contract/`

Request context, error types, and response helpers.

Responsibilities:

* `Ctx` struct — unified request context with `W`, `R`, `Params`, `Query`, `Headers`, `ClientIP`, `Logger`, `TraceID`
* `CtxHandlerFunc` — context-aware handler signature
* `APIError` struct — structured errors with status, code, message, category, trace ID
* Error categories: `Client`, `Server`, `Business`, `Timeout`, `Validation`, `Authentication`, `RateLimit`
* Error helpers: `NewValidationError()`, `NewNotFoundError()`, `NewUnauthorizedError()`, `WriteError()`
* `ErrorBuilder` — fluent builder pattern for constructing errors
* `RequestContext` — route pattern and parameter extraction from context

Subpackages:

* `contract/protocol/` — Protocol adapters (gRPC, GraphQL, HTTP)

---

#### `middleware/`

HTTP middleware implementations (19 subpackages).

Subpackages:

* `auth/` — Authentication and authorization (`Authenticator`, `Authorizer` interfaces)
* `bind/` — Request binding
* `cache/` — HTTP caching (`KeyStrategy`, `Store` interfaces)
* `circuitbreaker/` — Circuit breaker pattern
* `coalesce/` — Request coalescing
* `compression/` — Response compression (gzip)
* `cors/` — CORS handling (`CORSOptions`)
* `debug/` — Debug utilities
* `limits/` — Request limits (body size, concurrency)
* `observability/` — Tracing and metrics (`Tracer` interface)
* `protocol/` — Protocol adapters
* `proxy/` — HTTP reverse proxy
* `ratelimit/` — Rate limiting (`RateLimiter` interface)
* `recovery/` — Panic recovery
* `security/` — Security headers
* `tenant/` — Tenant routing and policies
* `timeout/` — Request timeouts
* `transform/` — Response transformation
* `versioning/` — API versioning

Middleware must:

* Follow `func(http.Handler) http.Handler` signature
* Be composable via `Chain.Use()` and `Chain.Apply()`
* Avoid global mutable state unless explicitly documented

---

#### `ai/`

AI agent gateway capabilities (21 subpackages).

Responsibilities:

* Unified LLM provider interface (Claude, OpenAI, etc.)
* SSE (Server-Sent Events) for real-time streaming
* Conversation session management with context window control
* Token counting and quota management
* Function calling / tool framework
* Semantic caching with embeddings
* Circuit breaker and resilience for LLM calls
* AI workflow orchestration

Subpackages:

* `circuitbreaker/` — Circuit breaker for LLM calls
* `distributed/` — Distributed AI features
* `filter/` — Request/response filtering
* `instrumentation/` — AI metrics and observability
* `llmcache/` — LLM response caching
* `logging/` — AI logging
* `marketplace/` — Model marketplace
* `metrics/` — AI metrics collection
* `multimodal/` — Multi-modal AI support
* `orchestration/` — AI workflow orchestration
* `prompt/` — Prompt management and engineering
* `provider/` — Unified LLM provider interface
* `ratelimit/` — AI endpoint rate limiting
* `resilience/` — Error handling and retries
* `semanticcache/` — Semantic caching with embeddings
* `session/` — Conversation session management
* `sse/` — Server-Sent Events streaming
* `streaming/` — Streaming response support
* `tokenizer/` — Token counting and management
* `tool/` — Function calling framework

Integration via `core.WithAIProvider()` and `core.WithSessionManager()`.

---

#### `tenant/`

Multi-tenancy primitives.

Responsibilities:

* Tenant configuration (`Config`, `ConfigManager` interface)
* Quota enforcement (`QuotaManager` interface)
* Policy evaluation (`PolicyEvaluator` interface)
* Rate limiting (`RateLimiter` interface)
* Route policies (`RoutePolicyStore` interface)
* Context helpers: `TenantIDFromContext()`, `ContextWithTenantID()`

Implementations:

* `InMemoryConfigManager` — for testing
* Database-backed config manager with LRU cache — for production (in `store/db/`)
* Sliding window and window-based quota managers

Rules:

* API may still evolve (treat as **high-stability** but watch for changes)
* Tenant isolation must be enforced at the database layer via `store/db/TenantDB`

---

#### `security/`

Security-critical logic.

Subpackages:

* `jwt/` — JWT token management with key rotation (`Manager`, `Verify()`, `TokenTypeAccess`)
* `headers/` — Security header policies (CSP, HSTS, etc.)
* `password/` — Bcrypt hashing, strength validation (`Hash()`, `Verify()`)
* `input/` — Input validation (email, URL, phone: `ValidateEmail()`, etc.)
* `abuse/` — Rate limiting and anti-abuse guard (`Guard`, `NewGuard()`, `Allow()`)

Rules:

* No secrets in logs
* Fail closed on verification errors
* Use timing-safe comparisons for secrets
* Prefer explicit verification APIs

---

#### `store/`

Data persistence abstractions.

Subpackages:

* `cache/` — Caching interface and implementations (distributed cache, Redis integration)
* `db/` — `database/sql` wrapper with tenant isolation (`TenantDB`), read/write splitting, sharding
* `file/` — File storage backend with migrations
* `idempotency/` — Idempotent request handling
* `kv/` — Embedded key-value store with WAL and LRU eviction

Rules:

* `store/kv/` is for lightweight internal persistence, not a general database layer
* `store/db/TenantDB` auto-injects `tenant_id` into queries for isolation
* Use `RawDB()` only when explicit bypass is needed (e.g. audit logs)

---

#### `net/`

Network utilities.

Subpackages:

* `discovery/` — Service discovery (static config, Consul integration)
* `http/` — HTTP client helpers
* `ipc/` — Inter-process communication (Unix/Windows)
* `mq/` — In-memory message queue
* `webhookin/` — Inbound webhook receivers
* `webhookout/` — Outbound webhook delivery with retry
* `websocket/` — WebSocket hub and connections

---

#### `pubsub/`

In-process publish/subscribe system.

Responsibilities:

* Internal event distribution
* Webhook fan-out support
* Lightweight event channels
* MQTT pattern matching
* Message scheduling and TTL management
* Debug or inspection hooks (if present)

This is **not** a distributed message queue.

---

#### `scheduler/`

Task scheduling.

Responsibilities:

* Cron job scheduling (`AddCron()`)
* Delayed task execution (`Delay()`)
* Retry policies (exponential backoff, etc.)
* Backpressure configuration
* Job status tracking and querying
* Injectable clock for testing

---

#### `health/`

Health and readiness signaling.

Responsibilities:

* Liveness / readiness indicators (`HealthState`: healthy, degraded, unhealthy)
* `ComponentChecker` interface (`Name()`, `Check()`)
* `HealthStatus` and `ComponentHealth` structs
* Build info tracking (`Version`, `Commit`, `BuildTime`)
* Startup and shutdown state tracking

---

#### `log/`

Structured logging abstraction.

Key types:

* `Fields` — `map[string]any` for structured fields
* `StructuredLogger` interface — `WithFields()`, `Debug/Info/Warn/Error()`, `DebugCtx/InfoCtx/WarnCtx/ErrorCtx()`
* `Lifecycle` interface — optional `Start()` / `Stop()` for logger lifecycle

---

#### `metrics/`

Metrics collection adapters.

Responsibilities:

* Prometheus metrics exporting
* OpenTelemetry integration

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

#### `frontend/`

Static frontend serving.

Responsibilities:

* Serving static files from disk or embedded assets
* SPA / static site mounting
* Safe defaults for caching and paths

---

#### `validator/`

Request validation.

Responsibilities:

* Request body validation
* Field-level validation helpers

---

#### `utils/`

Small shared helpers.

Subpackages:

* `httpx/` — HTTP utilities
* `jsonx/` — JSON utilities
* `pool/` — Object pooling
* `semver/` — Semantic versioning
* `stringsx/` — String manipulation

Root-level files:

* `html.go` — HTML helpers
* `http_response.go` — HTTP response helpers

Rules:

* No business logic
* No cross-layer coupling
* No hidden dependencies

---

#### `cmd/`

CLI tooling (separate Go module).

* Located at `cmd/plumego/` with its own `go.mod`
* Core plumego has zero dependencies; CLI adds only `gopkg.in/yaml.v3`
* Keeps CLI-only dependencies isolated from the library
* See `cmd/plumego/MODULE.md` for details

---

#### `examples/`

Reference implementations (19 examples).

Key examples:

* `reference/` — Full-featured example application
* `ai-agent-gateway/` — AI agent gateway
* `api-gateway/` — API gateway with proxy
* `multi-tenant-saas/` — Multi-tenant SaaS
* `resilient-gateway/` — Resilient gateway with circuit breaker
* `docs/` — API documentation (en/, zh/)
* Others: `agents/`, `bind-example/`, `cache-combined/`, `cache-distributed/`, `cache-leaderboard/`, `crud-demo/`, `db-metrics/`, `db-sharding/`, `file-storage-tests/`, `mq-task-queue/`, `rw/`, `scheduler/`, `sms-gateway/`

---

#### `docs/`

Internal design documents and migration guides (30+ markdown files).

---

## 4. Module Boundaries (Strict)

| Module | Responsibility | Stability |
|--------|----------------|-----------|
| `core/` | App lifecycle, DI, configuration, Boot() | **High** |
| `router/` | Path matching, route groups, parameters, reverse routing | **High** |
| `middleware/` | Request processing chain (19 subpackages) | **High** |
| `contract/` | Context, errors, response helpers, protocol adapters | **High** |
| `config/` | Environment loading, validation | Medium |
| `security/` | Cryptographic operations, signatures, abuse guard | **Critical** |
| `tenant/` | Multi-tenancy primitives, quota, policy, rate limit | **High** |
| `ai/` | AI agent gateway, LLM providers, sessions, streaming | Medium |
| `scheduler/` | Task scheduling, cron | Medium |
| `store/` | Persistence abstractions (cache, db, file, kv, idempotency) | Medium |
| `net/` | Network utilities, service discovery, webhooks | Medium |
| `pubsub/` | Event distribution | Medium |
| `health/` | Liveness/readiness probes | Medium |
| `log/` | Structured logging | Medium |
| `metrics/` | Prometheus and OpenTelemetry adapters | Medium |
| `frontend/` | Static file serving | Medium |
| `utils/` | Shared helpers only | Low |
| `cmd/` | CLI tooling (separate module) | Low |

**Rule**: Changes to `core/`, `router/`, `middleware/`, or `security/` require extra scrutiny and thorough testing.

---

## 5. Change Rules (Strict)

### Compatibility Rules

Agents **must**:

* Preserve `net/http` compatibility
* Avoid introducing incompatible abstractions
* Keep handlers as `http.Handler` or `http.HandlerFunc`
* Use `contract.CtxHandlerFunc` for context-aware handlers
* Avoid hidden global side effects

---

### API Stability Rules

* Public APIs in `core`, `router`, `middleware`, and `contract` are stable by default
* Breaking changes require:

  * Clear motivation
  * Migration notes
  * Preferably backward compatibility

---

### Dependency Rules

* New dependencies require strong justification
* Prefer standard library solutions
* The main module has **zero** external dependencies — preserve this
* `cmd/plumego/` is a separate module and may have limited dependencies
* No "utility" dependencies without necessity

---

## 6. Key Patterns

Agents must follow these established patterns:

### Functional Options

```go
type Option func(*App)

func WithAddr(addr string) Option {
    return func(a *App) {
        a.config.Addr = addr
    }
}
```

### Middleware Signature

```go
type Middleware func(http.Handler) http.Handler
```

### Middleware Chain

```go
chain := middleware.NewChain().
    Use(middleware.RequestID).
    Use(middleware.Logging).
    Use(middleware.Recovery)
handler := chain.Apply(baseHandler)
```

### Component Interface

```go
type Component interface {
    RegisterRoutes(r *router.Router)
    RegisterMiddleware(m *middleware.Registry)
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Health() (name string, status health.HealthStatus)
    Dependencies() []reflect.Type
}
```

### Runner Interface

```go
type Runner interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

### Error Handling

```go
// Use structured errors
err := contract.NewValidationError("email", "invalid format")
contract.WriteError(w, r, err)

// Or use the fluent builder
err := contract.NewError(
    contract.WithStatus(http.StatusBadRequest),
    contract.WithCode("INVALID_INPUT"),
    contract.WithMessage("Email is required"),
    contract.WithCategory(contract.CategoryValidation),
)
```

### Handler Registration

```go
// Standard library handler
app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("pong"))
})

// Context-aware handler
app.GetCtx("/health", func(ctx *plumego.Context) {
    ctx.JSON(http.StatusOK, map[string]string{"status": "ok"})
})
```

---

## 7. Testing & Quality Gates

Before submitting or generating patches, agents must ensure:

```bash
go test -timeout 20s ./...
go vet ./...
gofmt -w .
```

### Additional Verification by Change Type

| Change Type | Required Verification |
|-------------|----------------------|
| Routing | Route matching tests, parameter extraction, reverse routing |
| Middleware | Chain order, error path tests, handler semantics |
| Security | Negative tests (invalid signature, invalid token, timing attacks) |
| Config | Default and missing-value behavior tests |
| Scheduler | Overlap policy, retry behavior, injectable clock tests |
| AI | Provider abstraction, streaming, token counting tests |
| Tenant | Quota enforcement, policy evaluation, database isolation tests |
| Store | Persistence correctness, concurrent access, WAL integrity |

### Testing Patterns

* Standard `*_test.go` files, table-driven tests preferred
* Race detection: `go test -race ./...`
* Resettable state: `app.ResetForTesting()`
* Mocking time: `scheduler.WithClock(mockClock)`

---

## 8. Safe Refactor Zones

* **Zone A (Free)**: `docs/`, `examples/`, `utils/`, internal utilities
* **Zone B (Constrained)**: `router/`, `middleware/`, `contract/`, `ai/`, `tenant/`
* **Zone C (API Boundary)**: `core/`, public exports in `plumego.go`
* **Zone D (Do Not Refactor)**: Breaking changes require RFC

---

## 9. Common Agent Tasks

### A. Adding a New Middleware

1. Implement in `middleware/` (create subpackage if complex)
2. Follow `func(http.Handler) http.Handler` signature
3. Avoid global state
4. Add at least minimal tests
5. If auto-enabled via `core`, provide a `core.With*` option toggle

---

### B. Modifying Routing Behavior

1. Work exclusively in `router/`
2. Preserve existing semantics
3. Add tests for:

   * Static routes
   * Parameterized routes (`:id`)
   * Group prefixes
   * Middleware stacking order
   * Reverse routing (`URL()`)

---

### C. Adding Environment Configuration

1. Add to `config/`
2. Update `env.example`
3. Document default behavior
4. Fail fast on invalid critical config
5. Add tests for missing-value handling

---

### D. Webhook Enhancements

1. Signature and verification logic belongs in `security/`
2. Inbound webhook routing in `net/webhookin/`
3. Outbound delivery in `net/webhookout/`
4. Never log raw secrets or signatures
5. Prefer explicit verification APIs

---

### E. Adding an AI Provider

1. Implement the provider interface in `ai/provider/`
2. Add streaming support via `ai/sse/` or `ai/streaming/`
3. Add token counting in `ai/tokenizer/`
4. Add tests for provider responses and error handling
5. Integrate via `core.WithAIProvider()`

---

### F. Modifying Multi-Tenancy

1. Config and interfaces belong in `tenant/`
2. Database isolation logic in `store/db/` (`TenantDB`)
3. Middleware integration in `middleware/tenant/`
4. Test quota enforcement, policy evaluation, and isolation
5. See `examples/multi-tenant-saas/` for reference

---

### G. Adding a Store Backend

1. Implement in appropriate `store/` subpackage
2. Follow existing interface patterns
3. Test concurrent access and edge cases
4. If tenant-aware, integrate with `store/db/TenantDB`

---

## 10. Documentation Synchronization Rules

Agents **must update documentation** when changing:

* Public APIs
* Environment variables
* Default limits (timeouts, body size, concurrency)
* Security behavior
* Startup or shutdown semantics
* Module boundaries or new packages

Primary docs to keep in sync:

* `README.md`
* `README_CN.md`
* `CLAUDE.md`
* `AGENTS.md`
* `env.example`

---

## 11. Security & Disclosure

* Follow `SECURITY.md` for vulnerability reporting
* Never expose secrets in logs, issues, or PRs
* Assume logs may be public or centralized
* Use `security/abuse/` for rate limiting and anti-abuse
* Use `security/jwt/` for token management
* Use timing-safe comparisons for secret verification

---

## 12. Recommended Agent Workflow

1. Read `CLAUDE.md` and `AGENTS.md` first
2. Identify the correct module boundary
3. Read existing code before modifying
4. Add or update tests first (or alongside changes)
5. Make minimal, focused changes
6. Run full test suite: `go test -timeout 20s ./... && go vet ./... && gofmt -w .`
7. Ensure documentation parity
8. Keep commits small and reversible

---

**plumego values clarity, restraint, and correctness over feature volume.
Agents should optimize for maintainability, not cleverness.**
