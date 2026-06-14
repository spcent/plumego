# mini-saas-api

A realistic multi-tenant SaaS API use-case built on Plumego, demonstrating explicit
wiring, stable API patterns, and maintainable service structure. This is **not** a
complete SaaS product — it's a bounded showcase of how Plumego supports authentication,
multi-tenancy, RBAC, audit logging, idempotency, and observability in a single
service with zero framework magic.

## Purpose

`mini-saas-api` teaches how to build a maintainable, explicit SaaS API using stable
Plumego roots and carefully chosen beta extensions. It is copyable and suitable as a
starting point for real projects that need multi-tenant isolation, per-tenant rate
limiting, audit trails, and standardized error handling. All wiring is explicit in
source code — no auto-discovery, no controller scanning, no reflection-based routing.

## What it demonstrates

| Concern | Implementation | Plumego module |
|---|---|---|
| App lifecycle, graceful shutdown | Canonical layout matching `reference/standard-service` | `core` |
| Route registration and middleware order | All routes explicit in `internal/app/routes.go`; middleware chain in `internal/app/app.go` | `core`, `router`, `middleware/*` |
| User authentication | Signup with password hashing; login with bcrypt comparison | `security/password`, `security/authn` |
| Token-based access control | JWT HS256 per-route guards (not global auth middleware) | `security/jwt` |
| Token lifecycle management | Access + refresh token pair; single-use rotation; family revocation on reuse | App domain over `store/kv` |
| Brute-force protection | Abuse guard (configurable requests/minute) on public auth endpoints | `middleware/abuseguard` |
| Multi-tenancy isolation | Tenant context extracted from JWT claim; tenant ID threaded through all domain operations | `x/tenant` (beta) |
| Per-tenant rate limiting | Token-bucket rate limiter per tenant over a uniform limiter interface | `x/tenant/ratelimit` (beta) |
| Per-tenant quota enforcement | Fixed-window request quota enforcement (per-minute) | `x/tenant/quota` (beta) |
| RBAC enforcement | Role lattice (owner > admin > member) with last-owner invariant; checked per route | App `access` domain |
| REST resource patterns | CRUD controller for projects with `x/rest` conventions | `x/rest` (beta) |
| Idempotency guarantee | Idempotency-Key request deduplication; replay detection and same-payload verification | `store/idempotency` (stable) |
| Audit trail | Append-only log of every mutation; queryable by tenant | App `audit` domain |
| Observability | Prometheus metrics on `/metrics` endpoint; HTTP latency instrumentation | `x/observability` (beta) |
| Health checks | Liveness probe (always ready); readiness probes for app state | `health` |
| Error handling | Canonical envelope with typed error codes, messages, and field-level details | `contract` |

## What it intentionally excludes

This use-case is **not** a full SaaS product. It excludes:

| Capability | Why | Where to find it |
|---|---|---|
| Real database (SQL, NoSQL, etc.) | In-memory stores teach the pattern; swapping storage means implementing `Repository` interfaces | Replace `internal/domain/{user,tenantspace,project}/store.go` with a DB driver; test compatibility with the interface |
| Billing / payment processing | Orthogonal to the core SaaS patterns; would hide the main teaching | Build this as an independent service or standalone package |
| Frontend / web UI | API-only focus keeps wiring clear | Pair with `reference/with-frontend` if you need embedded assets |
| WebSocket / real-time features | Separate concern; demonstrated in `reference/with-websocket` | Combine their wiring patterns with this service's app shape |
| Email / SMS notifications | Messaging belongs in `x/messaging`; not shown here | See `reference/with-messaging` for async pubsub/webhook wiring |
| GraphQL API | REST patterns are explicit; GraphQL requires different handler and routing patterns | Build with this app as the domain layer; expose over a separate GraphQL endpoint |
| gRPC server | This is HTTP-only; see `reference/with-rpc` for hybrid HTTP/gRPC | Combine this service's domain logic with the gRPC server wiring from that reference |
| Distributed tracing (OpenTelemetry) | Basic Prometheus metrics shown; full OTEL setup in `reference/with-observability` | Add `x/observability` tracing per that reference's patterns |
| API gateway / reverse proxy | Single service focus; see `reference/with-gateway` for multi-service routing | This service could be a backend behind that gateway |

## Directory layout

```
use-cases/mini-saas-api/
├── main.go                    Process entrypoint (signal handling, config load, app startup)
├── go.mod                     Go module; replace directive points to repo root
├── env.example                Environment variables and defaults (copy to .env for local dev)
├── ARCHITECTURE.md            Ownership, dependency direction, security boundaries, request flow
├── PRODUCTION_CHECKLIST.md    Hardening steps before production deployment
├── api/
│   ├── curl.sh               End-to-end workflow walkthrough (bash)
│   ├── examples.http         HTTP examples for REST Client / JetBrains IDEs
│   └── postman_collection.json   Postman import for manual testing
├── internal/
│   ├── config/
│   │   ├── config.go         Config struct, Defaults(), Load(), Validate()
│   │   └── config_test.go    Config precedence and validation tests
│   ├── app/
│   │   ├── app.go            App struct, New() (stable middleware), service wiring
│   │   ├── routes.go         RegisterRoutes() — all HTTP routes declared explicitly (no scanning)
│   │   ├── (extension files) Tenant limits, JWT key management, KV setup, etc.
│   │   ├── auth_flow_test.go Full signup/login/refresh/token rotation flow
│   │   ├── tenant_flow_test.go  Workspace, membership, RBAC enforcement
│   │   ├── project_flow_test.go  CRUD operations, idempotency, audit trail
│   │   ├── observability_test.go Metrics emission
│   │   └── testhelper_test.go   Test fixtures and HTTP client builders
│   ├── handler/
│   │   ├── auth.go           SignUp, Login, Refresh (with token rotation)
│   │   ├── me.go             GET /me (current principal)
│   │   ├── tenant.go         Workspace Get/Update
│   │   ├── members.go        Membership List/Add/ChangeRole/Remove
│   │   ├── projects.go       x/rest controller for projects CRUD
│   │   ├── audit.go          Audit log query
│   │   ├── health.go         /healthz, /readyz handlers
│   │   ├── guard.go          Per-route middleware: RequireAuth, RequireRole, RequireMetricsToken
│   │   ├── idempotency.go    Idempotency-Key enforcement
│   │   ├── metricsguard.go   /metrics token guard
│   │   ├── write.go          logWriteErr helper
│   │   ├── errors.go         writeDomainError — domain sentinel → contract error mapping
│   │   └── health_test.go    Health probe tests
│   ├── domain/
│   │   ├── user/
│   │   │   ├── user.go       User model (id, email, password hash, created_at)
│   │   │   ├── store.go      Repository interface + MemoryStore (thread-safe, in-memory)
│   │   │   ├── service.go    Service interface + UserService (registration, auth, lookup)
│   │   │   └── service_test.go  Hashing, strength validation, duplicate email detection
│   │   ├── tenantspace/
│   │   │   ├── tenantspace.go  Tenant model (id, slug, name, plan) + Membership model
│   │   │   ├── store.go        Repository interface + MemoryStore
│   │   │   ├── service.go      Service interface + TenantService (CRUD + membership)
│   │   │   └── service_test.go  Last-owner invariant, slug uniqueness, role enforcement
│   │   ├── project/
│   │   │   ├── project.go       Project model (id, tenant_id, name, status, created_at)
│   │   │   ├── store.go         Repository interface + MemoryStore (tenant-scoped)
│   │   │   ├── service.go       Service interface + ProjectService (CRUD + quota checking)
│   │   │   └── service_test.go  Project limit enforcement, status transitions
│   │   ├── access/
│   │   │   ├── access.go        RBAC role lattice (Owner > Admin > Member); Role(string) validation
│   │   │   └── access_test.go   Permission checks (Can, Implies)
│   │   ├── session/
│   │   │   └── session.go       Token pair issuance, single-use refresh rotation, family revocation
│   │   ├── audit/
│   │   │   ├── audit.go         AuditEntry (event, actor, resource, timestamp); Recorder
│   │   │   └── audit_test.go    Append-only append, time-ordered queries
│   │   └── ident/
│   │       └── ident.go         Random 128-bit hex identifiers
│   └── platform/
│       └── idemstore/           In-memory implementation of store/idempotency contract
│           └── idemstore.go
└── internal/app/ extensions (tenant limits, JWT key storage, KV initialization)
```

## Run from source

```bash
# Prerequisites: Go 1.26+

# Clone the repo
git clone https://github.com/spcent/plumego && cd plumego

# Copy env.example to .env and edit APP_JWT_SECRET (optional for local dev)
cd use-cases/mini-saas-api
cp env.example .env

# Start the server (listens on :8090 by default)
go run .
```

The server logs on startup:
- App version and config
- Deprecation warnings for demo-only defaults (e.g., CORS=*, JWT_SECRET too short)

## Try it

### Using curl (end-to-end workflow)

Run this from another terminal with the server running:

```bash
# All routes in api/curl.sh
bash api/curl.sh
```

Or step by step:

```bash
# Health probes
curl http://localhost:8090/healthz
curl http://localhost:8090/readyz

# Signup (returns access + refresh token)
curl -X POST http://localhost:8090/api/v1/auth/signup \
  -H 'Content-Type: application/json' \
  -d '{
    "email": "alice@example.com",
    "name": "Alice",
    "password": "Str0ng!Password#2026",
    "workspace_name": "ACME Inc",
    "workspace_slug": "acme-inc"
  }'
# 201 → {"data": {"user": {...}, "tenant": {...}, "tokens": {"access_token": "...", "refresh_token": "..."}}, "request_id": "..."}

# Login (same tokens as signup)
curl -X POST http://localhost:8090/api/v1/auth/login \
  -H 'Content-Type: application/json' \
  -d '{"email": "alice@example.com", "password": "Str0ng!Password#2026"}'
# 200 → {"data": {"tokens": {...}}, ...}

# Get current principal (requires Bearer token)
curl http://localhost:8090/api/v1/me \
  -H 'Authorization: Bearer <access_token>'
# 200 → {"data": {"user": {...}, "membership": {...}}, ...}

# Workspace details
curl http://localhost:8090/api/v1/tenant \
  -H 'Authorization: Bearer <access_token>'

# Create a project (with Idempotency-Key for replay protection)
curl -X POST http://localhost:8090/api/v1/projects \
  -H 'Authorization: Bearer <access_token>' \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: demo-project-1' \
  -d '{"name": "Apollo", "description": "Launch sequence"}'
# 201 → {"data": {"id": "...", "name": "Apollo", ...}, ...}

# Replay the same create with the same Idempotency-Key — get the cached response
curl -X POST http://localhost:8090/api/v1/projects \
  -H 'Authorization: Bearer <access_token>' \
  -H 'Content-Type: application/json' \
  -H 'Idempotency-Key: demo-project-1' \
  -d '{"name": "Apollo", "description": "Launch sequence"}'
# 201 → (exact same response, with X-Idempotency-Replay: true header)

# List projects
curl http://localhost:8090/api/v1/projects \
  -H 'Authorization: Bearer <access_token>'

# Refresh token rotation (single-use refresh tokens)
curl -X POST http://localhost:8090/api/v1/auth/refresh \
  -H 'Content-Type: application/json' \
  -d '{"refresh_token": "<refresh_token>"}'
# 200 → {"data": {"tokens": {"access_token": "...", "refresh_token": "..."}}, ...}

# Reusing the old refresh token fails (family revocation on theft detection)
curl -X POST http://localhost:8090/api/v1/auth/refresh \
  -H 'Content-Type: application/json' \
  -d '{"refresh_token": "<old_refresh_token>"}'
# 401 → {"error": {"code": "auth.refresh_reused", "message": "refresh token reuse detected; session revoked, log in again", ...}, ...}

# Audit trail (admin only; newest first)
curl 'http://localhost:8090/api/v1/tenant/audit?limit=10' \
  -H 'Authorization: Bearer <admin_access_token>'
# 200 → {"data": {"events": [...], "total": N}, ...}

# Prometheus metrics (optional token guard)
curl http://localhost:8090/metrics | head
```

### Using examples.http (REST Client / VS Code REST Client)

Open `api/examples.http` in VS Code with the REST Client extension, or use JetBrains
IDEs (IntelliJ, GoLand) natively. Each request is executable inline.

### Using Postman

```bash
# Import the collection
open api/postman_collection.json
# or use Ctrl+K and paste the file path
```

## Configuration

All environment variables have defaults. See `env.example` for the full reference.

| Variable | Default | Purpose |
|---|---|---|
| `APP_ADDR` | `:8090` | Listen address |
| `APP_SERVICE_NAME` | `mini-saas-api` | Service identity in responses |
| `APP_ENV_FILE` | `.env` | Path to environment file (optional) |
| `APP_MAX_BODY_BYTES` | `1048576` (1 MiB) | Request body size limit; 0 disables |
| `APP_CORS_ALLOWED_ORIGINS` | empty (allows `*`) | Comma-separated origin list; **required in production** |
| `APP_TLS_ENABLED` | `false` | Enable TLS |
| `APP_TLS_CERT_FILE` | — | TLS certificate path |
| `APP_TLS_KEY_FILE` | — | TLS key path |
| `APP_JWT_SECRET` | `dev-secret-do-not-use...` | JWT signing key; **must be ≥32 bytes in production** |
| `APP_JWT_ACCESS_TTL` | `15m` | Access token lifetime |
| `APP_JWT_REFRESH_TTL` | `168h` (7 days) | Refresh token lifetime |
| `APP_TENANT_RPS` | `50` | Per-tenant requests/second (token bucket) |
| `APP_TENANT_BURST` | `100` | Per-tenant burst capacity |
| `APP_TENANT_QUOTA_PER_MINUTE` | `600` | Per-tenant fixed-window quota |
| `APP_DATA_DIR` | `.data` | Directory for embedded KV store (JWT keys, refresh tokens) |
| `APP_METRICS_TOKEN` | empty (no guard) | Optional bearer token to guard `/metrics` endpoint |

## Tests

```bash
cd use-cases/mini-saas-api && go test -race -timeout 30s ./...
```

The test suite covers:
- **Authentication flow** (`auth_flow_test.go`): signup, login, token rotation, reuse detection, family revocation
- **Multi-tenancy flow** (`tenant_flow_test.go`): tenant isolation, membership, RBAC enforcement, last-owner invariant
- **Resource CRUD flow** (`project_flow_test.go`): create, list, update, delete, idempotency deduplication
- **Observability** (`observability_test.go`): metrics emission on requests, counter increment
- **Config** (`config_test.go`): environment variable precedence, defaults, validation
- **Handlers** (`internal/handler/health_test.go`): liveness/readiness probes, component status
- **Domain models** (`internal/domain/*/service_test.go`): user hashing, tenant isolation, access control
- **Race safety**: all tests run with `-race` detector to catch concurrent access bugs

Tests do **not** require external dependencies (database, cache, message queue). All storage is
in-memory; failures to reach external services do not block tests. See `testhelper_test.go`
for the test fixture pattern.

## How to extend

### Add a new user-facing endpoint

1. **Write the handler** in `internal/handler/new_feature.go`:
   ```go
   func (h MyHandler) DoSomething(w http.ResponseWriter, r *http.Request) {
       // Parse request, call domain service, write response
       logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, data, nil))
   }
   ```

2. **Add domain logic** in `internal/domain/myfeature/service.go` (create if needed):
   ```go
   type MyService struct {
       repo MyRepository
   }
   func (s *MyService) DoSomething(ctx context.Context, ...) (Result, error) {
       // Implement your business logic
   }
   ```

3. **Wire in `internal/app/routes.go`**:
   ```go
   myH := handler.MyHandler{Repo: a.MyRepo, Logger: logger}
   v1.post("/my-endpoint", authed(idem(http.HandlerFunc(myH.DoSomething))))
   ```

4. **Add tests** in `internal/handler/handler_test.go` or `internal/domain/myfeature/service_test.go`

5. **Run validation**:
   ```bash
   go test -race -timeout 30s ./...
   ```

### Swap in-memory storage for a real database

Replace the `MemoryStore` implementations in `internal/domain/{user,tenantspace,project}/store.go`
with your database driver. The `Repository` interface stays the same:

```go
type UserRepository interface {
    Create(ctx context.Context, u User) error
    ByID(ctx context.Context, id string) (User, error)
    ByEmail(ctx context.Context, email string) (User, error)
}
```

Wire the new store in `internal/app/routes.go` (or in `New()` when constructing the app).
**No handler changes required.**

### Add admin-only endpoints

Wrap with `requireAdmin`:

```go
v1.patch("/admin/tenant", authed(requireAdmin(idem(http.HandlerFunc(adminH.UpdateTenant)))))
```

`requireAdmin` checks that the principal's role is `owner` or `admin` before the handler
runs. See `internal/handler/guard.go` for the implementation.

### Enforce custom per-tenant rate limits

Modify `uniformRateLimits` in `internal/app/routes.go`:

```go
limiter: tenantcore.NewTokenBucketRateLimiter(uniformRateLimits{
    rps:   getTenantRPS(principal.TenantID),  // look up from config or store
    burst: getTenantBurst(principal.TenantID),
}),
```

The tenant context (`x/tenant`) carries this configuration and the runtime limits per request.

### Log custom audit events

```go
a.Audit.Record(ctx, audit.Entry{
    Event:     "project_shared",
    Actor:     principal.UserID,
    Resource:  "projects/" + projectID,
    Timestamp: time.Now(),
})
```

See `internal/domain/audit/audit.go` for the audit entry schema.

### Add metrics to a custom endpoint

The `x/observability` collector is wired globally in `internal/app/app.go` and feeds every
HTTP request through `httpmetrics` middleware. Custom domain logic can record its own metrics:

```go
a.Collector.Counter("my_operation", 1, prometheus.Labels{"operation": "sync"})
```

See the `x/observability` module docs for the full API.

## Production notes

Before deploying, read `PRODUCTION_CHECKLIST.md`. Key items:

- **Secrets**: Set `APP_JWT_SECRET` to at least 32 cryptographically random bytes; generate with `openssl rand -hex 16`
- **Auth**: Verify `RequireAuth` guards every protected endpoint; test token expiry and refresh flow
- **CORS**: Set `APP_CORS_ALLOWED_ORIGINS` to an explicit list; never use `*` in production
- **Timeouts**: Adjust `core.ReadTimeout`, `core.WriteTimeout`, `core.IdleTimeout` for your workload
- **Rate limiting**: Tune `APP_TENANT_RPS`, `APP_TENANT_BURST`, `APP_TENANT_QUOTA_PER_MINUTE` per tenant tier
- **Observability**: Wire a real Prometheus collector or OpenTelemetry exporter; `/metrics` is read-only by default
- **Persistence**: Implement database-backed repositories for `user`, `tenantspace`, and `project` stores
- **Migrations**: Design schema migration strategy before accepting real data
- **Audit**: Ensure audit trail is durably written (e.g., to a separate audit log service)
- **Metrics token**: Set `APP_METRICS_TOKEN` to guard `/metrics` from public access

## Related references

- **Canonical app layout**: `reference/standard-service/` — learn the bootstrap pattern
- **Multi-tenancy patterns**: `reference/with-tenant/` — tenant resolution, quota, session lifecycle
- **CRUD resource wiring**: `reference/with-rest/` — REST controller conventions
- **Observability setup**: `reference/with-observability/` — Prometheus + OTEL integration
- **Module docs**: `docs/modules/` — package-specific primers for `core`, `contract`, `middleware`, `security`, `store`, `health`, `log`, `metrics`, `x/rest`, `x/tenant`, `x/observability`
- **Style guide**: `docs/reference/canonical-style-guide.md` — handler signatures, error handling, dependency direction
- **Security**: `docs/modules/security/` — JWT, password, RBAC, authorization patterns

## Milestone

This use-case is actively developed in `tasks/milestones/active/M-025-mini-saas-api/`.
