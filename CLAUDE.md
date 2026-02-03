# CLAUDE.md — plumego

> **Version**: v1.0.0-rc.1 | **Status**: Release Candidate | **Go**: 1.24+

This document is the **authoritative reference for AI assistants** (Claude, Copilot, Cursor, etc.) working in the `spcent/plumego` repository. It explains the codebase structure, development workflows, and key conventions to follow.

---

## Quick Reference

```bash
# Build and test
go test ./...                    # Run all tests
go test -race ./...              # Run tests with race detector
go test -timeout 20s ./...       # Run with timeout (recommended)
go vet ./...                     # Static analysis
gofmt -w .                       # Format code

# Run the reference application
go run ./examples/reference
```

**Go Version**: 1.24+ (see `go.mod`)

---

## Project Overview

**Plumego** is a lightweight Go HTTP toolkit built entirely on the standard library (`net/http`). It is designed to be **embedded into applications** rather than acting as a standalone framework.

### Core Characteristics

- **Standard library–first**: Uses `net/http`, `context`, and `http.Handler`
- **Explicit lifecycle**: `core.New(...)` → configuration → `Boot()`
- **Minimal dependencies**: Only Go standard library
- **Composable architecture**: Pluggable router, middleware, and components

### What Plumego Provides

- Trie-based HTTP router with path parameters (`:id`)
- Middleware chain (logging, recovery, CORS, rate limiting, etc.)
- Graceful startup/shutdown with connection draining
- WebSocket hub with JWT authentication
- Webhook handling (GitHub, Stripe) with signature verification
- In-process Pub/Sub for event distribution
- Task scheduling (cron, delayed jobs, retries)
- Embedded KV storage with WAL and LRU eviction
- Static frontend serving from disk or embedded assets

### Non-Goals

- Replacing large frameworks (Gin, Echo, Fiber)
- Heavy dependency trees
- Hiding `net/http` abstractions
- Opinionated ORM or persistence layers

---

## Repository Structure

```
plumego/
├── core/           # Application lifecycle, DI container, configuration
├── router/         # HTTP routing and request dispatch
├── middleware/     # Request processing chain (logging, auth, CORS, etc.)
├── contract/       # Request context, error types, response helpers
├── config/         # Environment variable loading, .env parsing
├── health/         # Liveness/readiness probes, health endpoints
├── log/            # Structured logging abstraction
├── metrics/        # Prometheus and OpenTelemetry adapters
├── pubsub/         # In-process publish/subscribe
├── scheduler/      # Cron jobs, delayed tasks, retry policies
├── security/       # JWT, password hashing, input validation, headers
│   ├── jwt/        # Token management with key rotation
│   ├── headers/    # Security header policies (CSP, HSTS, etc.)
│   ├── password/   # Bcrypt hashing, strength validation
│   └── input/      # Email, URL, phone validation
├── store/          # Data persistence abstractions
│   ├── cache/      # Caching interface
│   ├── db/         # database/sql wrapper
│   └── kv/         # Embedded key-value store with WAL
├── net/            # Network utilities
│   ├── http/       # HTTP client helpers
│   ├── ipc/        # Inter-process communication (Unix/Windows)
│   ├── mq/         # In-memory message queue
│   ├── webhookin/  # Inbound webhook receivers
│   ├── webhookout/ # Outbound webhook delivery
│   └── websocket/  # WebSocket hub and connections
├── frontend/       # Static file serving
├── validator/      # Request validation
├── utils/          # Small shared helpers
├── examples/       # Reference implementations
│   ├── reference/  # Full-featured example application
│   └── docs/       # Documentation (en/, zh/)
├── plumego.go      # Main package re-exports
├── go.mod          # Module definition
├── env.example     # Environment variable template
├── README.md       # Project documentation
├── AGENTS.md       # Detailed agent guidelines
└── SECURITY.md     # Security policy
```

---

## Module Boundaries (Strict)

Agents **must respect module boundaries**. Do not blur these separations:

| Module | Responsibility | Stability |
|--------|----------------|-----------|
| `core/` | App lifecycle, DI, configuration, Boot() | **High** |
| `router/` | Path matching, route groups, parameters | **High** |
| `middleware/` | Request processing chain | **High** |
| `contract/` | Context, errors, response helpers | **High** |
| `config/` | Environment loading, validation | Medium |
| `security/` | Cryptographic operations, signatures | **Critical** |
| `scheduler/` | Task scheduling, cron | Medium |
| `store/` | Persistence abstractions | Medium |
| `net/` | Network utilities | Medium |
| `pubsub/` | Event distribution | Medium |
| `frontend/` | Static file serving | Medium |
| `utils/` | Shared helpers only | Low |

**Rule**: Changes to `core/`, `router/`, `middleware/`, or `security/` require extra scrutiny and thorough testing.

---

## Key Types and Interfaces

### Application Core (`core/`)

```go
// App is the main application instance
type App struct { ... }

// Create with functional options
app := core.New(
    core.WithAddr(":8080"),
    core.WithDebug(),
    core.WithRecovery(),
    core.WithLogging(),
)

// Boot starts the server
if err := app.Boot(); err != nil {
    log.Fatal(err)
}
```

### Component Interface

```go
type Component interface {
    RegisterRoutes(r *router.Router)
    RegisterMiddleware(m *middleware.Registry)
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Health() (name string, status health.HealthStatus)
    Dependencies() []reflect.Type  // For topological sorting
}
```

### Handler Signatures

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

### Middleware Signature

```go
type Middleware func(http.Handler) http.Handler
```

### Error Handling

```go
// Structured errors with categories
err := contract.NewValidationError("email", "invalid format")
contract.WriteError(w, r, err)

// Error categories: Client, Server, Business, Timeout, Validation, Authentication, RateLimit
```

---

## Configuration

### Environment Variables

Key variables (see `env.example` for full list):

| Variable | Default | Description |
|----------|---------|-------------|
| `APP_ADDR` | `:8080` | Server listen address |
| `APP_DEBUG` | `false` | Enable debug mode |
| `APP_SHUTDOWN_TIMEOUT_MS` | `5000` | Graceful shutdown timeout |
| `APP_MAX_BODY_BYTES` | `10485760` | Request body limit (10 MiB) |
| `APP_MAX_CONCURRENCY` | `256` | Max concurrent requests |
| `WS_SECRET` | - | WebSocket JWT secret (32+ bytes) |
| `GITHUB_WEBHOOK_SECRET` | - | GitHub webhook HMAC secret |
| `STRIPE_WEBHOOK_SECRET` | - | Stripe webhook secret |

### Functional Options Pattern

```go
app := core.New(
    core.WithAddr(":8080"),
    core.WithServerTimeouts(30*time.Second, 5*time.Second, 30*time.Second, 60*time.Second),
    core.WithMaxBodyBytes(10 << 20),
    core.WithSecurityHeadersEnabled(true),
    core.WithAbuseGuardEnabled(true),
    core.WithRecommendedMiddleware(), // RequestID + Logging + Recovery
    core.WithComponent(myComponent),
)
```

---

## Development Workflow

### Before Making Changes

1. Read existing code before modifying
2. Understand module boundaries
3. Plan changes with minimal scope

### Verification Checklist

```bash
# All changes must pass these:
go test -timeout 20s ./...
go vet ./...
gofmt -w .
```

### Additional Checks by Change Type

| Change Type | Required Verification |
|-------------|----------------------|
| Routing | Route matching tests, parameter extraction |
| Middleware | Chain order, error path tests |
| Security | Negative tests (invalid signature/token) |
| Config | Default and missing-value tests |
| Scheduler | Overlap policy, retry behavior tests |

### Safe Refactor Zones

- **Zone A (Free)**: `docs/`, `examples/`, internal utilities
- **Zone B (Constrained)**: `router/`, `middleware/`, `contract/`
- **Zone C (API Boundary)**: `core/`, public exports
- **Zone D (Do Not Refactor)**: Breaking changes require RFC

---

## Common Patterns

### 1. Functional Options

```go
func WithTimeout(d time.Duration) Option {
    return func(c *Config) {
        c.Timeout = d
    }
}
```

### 2. Middleware Chain

```go
chain := middleware.NewChain().
    Use(middleware.RequestID).
    Use(middleware.Logging).
    Use(middleware.Recovery)
handler := chain.Apply(baseHandler)
```

### 3. Component Registration

```go
app := core.New(
    core.WithComponent(&MyComponent{}),
)
```

### 4. Error Handling

```go
// Create structured error
err := contract.NewError(
    contract.WithStatus(http.StatusBadRequest),
    contract.WithCode("INVALID_INPUT"),
    contract.WithMessage("Email is required"),
    contract.WithCategory(contract.CategoryValidation),
)

// Write error response
contract.WriteError(w, r, err)
```

### 5. Scheduler Jobs

```go
sch := scheduler.New(scheduler.WithWorkers(4))
sch.Start()

sch.AddCron("cleanup", "0 * * * *", func(ctx context.Context) error {
    return nil
}, scheduler.WithTimeout(5*time.Minute))

sch.Delay("task", 10*time.Second, myFunc,
    scheduler.WithRetry(scheduler.RetryExponential(time.Second, 5)),
)
```

---

## Testing Patterns

### Unit Tests

- Standard `*_test.go` files
- Table-driven tests preferred
- Use `testing.T` helpers

### Race Condition Tests

```bash
go test -race ./...
```

### Resettable State for Testing

```go
// In test files
app.ResetForTesting()
```

### Mocking Time

```go
// Scheduler supports injectable clock
sch := scheduler.New(
    scheduler.WithClock(mockClock),
)
```

---

## Security Guidelines

### Critical Rules

1. **Never log secrets** (tokens, keys, passwords)
2. **Fail closed** on verification errors
3. **Validate all input** at system boundaries
4. **Use timing-safe comparisons** for secrets

### Security Package Usage

```go
// JWT verification
manager := jwt.NewManager(secret)
claims, err := manager.Verify(token, jwt.TokenTypeAccess)

// Password hashing
hash, err := password.Hash(plaintext)
ok := password.Verify(plaintext, hash)

// Input validation
if !input.ValidateEmail(email) {
    return errors.New("invalid email")
}
```

### Webhook Verification

- GitHub: HMAC-SHA256 signature in `X-Hub-Signature-256`
- Stripe: Signature with timestamp tolerance

---

## PR Guidelines

### Required Information

1. **Summary**: What problem does this solve?
2. **Scope**: What directories/packages were touched?
3. **Type**: Bugfix, Feature, Refactor, Breaking, Docs
4. **Zone**: A/B/C/D (see Safe Refactor Zones)
5. **Verification**: Test output or CI links

### Commit Messages

- Use imperative mood: "Add feature" not "Added feature"
- Reference issue numbers when applicable
- Keep first line under 72 characters

---

## Agent Best Practices

### Do

- Read existing code before modifying
- Keep changes minimal and focused
- Add tests alongside code changes
- Update documentation when changing APIs
- Use `contract.WriteError` for error responses
- Follow the functional options pattern

### Don't

- Introduce new dependencies without strong justification
- Blur module boundaries
- Add global mutable state
- Skip the verification checklist
- Make breaking changes without migration notes
- Log secrets or sensitive data

### Common Tasks

**Adding a new middleware:**
1. Implement in `middleware/`
2. Follow `func(http.Handler) http.Handler` signature
3. Add tests
4. Optionally add `core.With*` option if auto-enabled

**Adding environment configuration:**
1. Add to `config/`
2. Update `env.example`
3. Document default behavior
4. Add tests for missing-value handling

**Modifying routing:**
1. Work in `router/` only
2. Preserve existing route semantics
3. Test static routes, parameters, groups, middleware order

---

## Additional Resources

- `README.md` — Project documentation
- `README_CN.md` — Chinese documentation
- `AGENTS.md` — Detailed agent guidelines
- `SECURITY.md` — Security policy and disclosure
- `env.example` — Environment variable reference
- `examples/reference/` — Full-featured example application
- `examples/docs/` — API documentation

---

## Quick Command Reference

```bash
# Development
go run ./examples/reference          # Run reference app
go test ./...                        # Run tests
go test -race ./...                  # Race detection
go test -v -run TestName ./pkg/...   # Run specific test

# Quality
go vet ./...                         # Static analysis
gofmt -w .                           # Format code
go mod tidy                          # Clean dependencies

# Build
go build ./...                       # Build all packages
go build -o app ./examples/reference # Build reference app
```
