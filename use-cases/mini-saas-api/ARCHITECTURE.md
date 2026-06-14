# Architecture — mini-saas-api

## Canonical Layout

This service follows the explicit wiring pattern established by `reference/standard-service`:

- **`main.go`**: Process entrypoint only. Exactly four operations: load config → construct app → register routes → start server.
- **`internal/config`**: Configuration loading from environment, flags, and `.env` file. Never imported by handlers or domain.
- **`internal/app`**: HTTP and dependency wiring. Middleware order, route registration, extension configuration. All extension wiring lives here.
- **`internal/handler`**: HTTP adaptation layer. Request parsing, validation, response writing. Handlers are pure `func(http.ResponseWriter, *http.Request)`.
- **`internal/domain`**: Business logic and data models, isolated from HTTP concerns. Packages for user, tenants, projects, access control, audit, and sessions.

## Ownership

| Package | Owns | Coupling |
|---|---|---|
| `main.go` | Signal handling, config loading, app startup, graceful shutdown | Calls: `config.Load()`, `app.New()`, `app.RegisterRoutes()`, `app.Start()` |
| `internal/config` | Environment/flag parsing, defaults, validation | Used by: `main`, `internal/app` |
| `internal/app/app.go` | Core app construction, middleware ordering, dependency initialization | Imports: stable roots only; calls: `core.New()`, all middleware; wires all domain services |
| `internal/app/routes.go` | Route registration, per-route middleware wrapping, handler wiring | Imports: handlers, domain, extensions (`x/tenant`, `x/rest`, `x/observability`); registers all public routes |
| `internal/handler/{auth,tenant,members,projects,audit,health}` | HTTP adaptation: parse requests, call services, write responses | Imports: `contract`, domain services via interfaces, app logger |
| `internal/domain/user` | User account model, bcrypt hashing, authentication, storage interface | Called by: handlers; calls: password service |
| `internal/domain/tenantspace` | Tenant (workspace) model, membership, RBAC, storage interface | Called by: handlers; calls: access control service |
| `internal/domain/project` | Project resource model, quota enforcement, storage interface | Called by: handlers (via `x/rest` controller); uses tenant context |
| `internal/domain/access` | RBAC role hierarchy (Owner > Admin > Member), permission checks | Called by: handlers, tenantspace service; never from middleware |
| `internal/domain/session` | Token pair issuance, refresh rotation, family revocation, family tree tracking | Called by: auth handler; calls: token issuer (JWT + opaque refresh via KV) |
| `internal/domain/audit` | Append-only event log, queryable by tenant | Called by: handlers (post-mutation); writes to in-memory append queue |
| `internal/domain/ident` | Random 128-bit hex ID generation | Called by: all domain services needing identifiers |
| `internal/platform/idemstore` | In-memory implementation of `store/idempotency` contract | Called by: idempotency middleware in handlers |

## Dependency Direction (one-way, verified)

```
main.go
  ↓
internal/config
  ↓
internal/app (constructs services)
  ↓                 ↘
internal/handler ↔ internal/domain/*
  ↑
internal/platform/* (pluggable implementations)
```

**Rules enforced:**
- `main.go` only imports `config` and `app`
- `internal/config` is never imported by handlers or domain
- `internal/handler` never imports `app` or `config` (dependencies injected as constructor params)
- `internal/domain/*` packages have no intra-domain imports except `ident` (used everywhere) and `audit` (used by handlers)
- All `x/*` extension imports live only in `internal/app`

## Module Usage (Stable + Beta)

| Layer | Stable roots | Beta extensions |
|---|---|---|
| **Transport** | `core` (app, routing), `router` (path params, groups), `contract` (response envelope, error builder) | — |
| **Middleware** | `middleware/*` (requestid, security, cors, recovery, accesslog, bodylimit, timeout, abuseguard), `health` (probes) | — |
| **Security** | `security/password`, `security/authn`, `security/jwt`, `store/kv` (state file), `store/idempotency` (replay protection) | — |
| **Observability** | `log` (structured logging), `metrics` (noop collector interface) | `x/observability` (Prometheus collector, exporter) |
| **Multi-tenancy** | — | `x/tenant/core`, `x/tenant/resolve`, `x/tenant/ratelimit`, `x/tenant/quota` |
| **Resource CRUD** | — | `x/rest` (resource controller conventions) |

## Global Middleware Stack

Ordered by `internal/app/app.go.New()`. Order matters; each layer wraps the next:

```
requestid (stamps correlation ID before any logging)
  ↓
securityheaders (X-Frame-Options, X-Content-Type-Options, Referrer-Policy)
  ↓
cors (preflight handling; default: allow all)
  ↓
recovery (panic → 500 response)
  ↓
accesslog (structured log of every request/response)
  ↓
bodylimit (reject oversized bodies; default 1 MiB)
  ↓
httpmetrics (Prometheus HTTP latency instrumentation)
  ↓
timeout (per-request wall-clock limit; default 30 s)
  ↓
[routes and per-route middleware]
```

**Rationale:**
- `requestid` first so every log and error carries correlation ID
- `recovery` before `accesslog` so panics appear as 500 in logs
- `bodylimit` early to reject garbage quickly
- `httpmetrics` before `timeout` to measure handler time only
- `timeout` innermost (closest to handler) to cover only handler execution

## Per-Route Middleware

Visible in `routes.go`; composed from building blocks:

```go
// Global auth guard: extract JWT, verify signature, inject principal into context
requireAuth := handler.RequireAuth(jwtManager, logger)

// RBAC enforcement: check principal.Role ≥ admin
requireAdmin := handler.RequireRole(access.RoleAdmin, logger)

// Idempotency-Key replay protection: request deduplication + same-payload validation
idem := handler.Idempotent(idemStore, 24*time.Hour, logger)

// Tenant chain (x/tenant, beta): resolve tenant from principal → rate-limit → quota enforce
tenantChain := middleware.NewChain(
    tenantresolve.Middleware(...),   // lift tenantID into context
    tenantratelimit.Middleware(...), // per-tenant token bucket
    tenantquota.Middleware(...),     // per-tenant fixed-window
)

// Compose: authenticated + tenant-aware + idempotent
authed := func(h http.Handler) http.Handler { return requireAuth(tenantChain.Build(h)) }

// Register: one method, one path, one handler per line
v1.post("/projects", authed(idem(http.HandlerFunc(projectsH.Create))))
```

Public endpoints (auth, health) skip `requireAuth` and `tenantChain`.

## Request Flow (Authenticated Route)

Example: `POST /api/v1/projects` with JWT, tenant context, idempotency dedup, and RBAC check.

```
HTTP Request
  ↓
requestid → securityheaders → cors → recovery → accesslog → bodylimit → httpmetrics → timeout
  ↓
RequireAuth (JWT signature + expiry check)
  → Extract principal {UserID, TenantID, Role} into context
  → On failure: 401 Unauthorized
  ↓
x/tenant resolve (lift TenantID from principal into x/tenant context)
  ↓
x/tenant ratelimit (consume per-tenant token bucket; default 50 rps, burst 100)
  → On limit: 429 Too Many Requests
  ↓
x/tenant quota (consume per-tenant fixed-window quota; default 600 req/min)
  → On limit: 429 Too Many Requests
  ↓
Idempotency-Key dedup (lookup by key; if cached, return stored response)
  → On match: 201 with X-Idempotency-Replay: true header
  → On conflict (key exists with different payload): 409 Conflict
  ↓
RequireRole(admin) (check principal.Role ≥ admin)
  → On failure: 403 Forbidden
  ↓
Handler (projectsH.Create)
  ↓ (parse JSON request, validate, call service)
  ↓
Domain Service (projectService.Create)
  ↓ (business logic: check quota, generate ID, call repository)
  ↓
Repository (projectStore.Create with TenantID)
  ↓ (store in memory with tenant isolation)
  ↓
Audit (a.Audit.Record "project_created")
  ↓ (append to audit log; failures don't block response)
  ↓
contract.WriteResponse (201, {project data}, nil)
  ↓ (idempotency middleware caches response)
  ↓
HTTP Response (201 + JSON body)
```

## Error Handling

All errors use the canonical `contract.ErrorBuilder` envelope. Mapping lives in `handler/errors.go`:

```go
// Domain error → contract error type
user.ErrWeakPassword → contract.TypeValidation (code: "user.weak_password")
user.ErrNotFound → contract.TypeNotFound (code: "user.not_found")
session.ErrReused → contract.TypeUnauthorized (code: "auth.refresh_reused")
tenantspace.ErrLastOwner → contract.TypeConflict (code: "tenant.last_owner")
```

**Error response shape:**
```json
{
  "error": {
    "code": "tenant.last_owner",
    "message": "cannot demote or remove the last owner",
    "category": "conflict",
    "type": "conflict_error",
    "details": {}
  },
  "request_id": "01HX..."
}
```

**Error handling strategy:**
- Domain packages define sentinel errors (`errors.Is()` safe)
- Handlers catch and map to contract types + human messages
- Untyped errors → 500 Internal Error (logged, never leaked)
- Secrets never log (JWT, password hashes, refresh tokens, Idempotency-Key)

## Response Envelope

All success responses use `contract.WriteResponse`:

```go
logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, data, meta))
```

**Success response shape:**
```json
{
  "data": { "id": "...", "name": "...", ... },
  "meta": { "total": 42, "limit": 20, "offset": 0 },
  "request_id": "01HX..."
}
```

- `data`: the primary result (omitted if `nil`)
- `meta`: pagination/summary info (omitted if `nil`)
- `request_id`: correlation ID (always present)
- No 204 No Content responses; use 200 with `data: null` or omit `data`

## Extension Wiring Points

### `x/tenant` (beta) — Multi-tenancy

**Wired in:** `internal/app/routes.go` (per-route middleware chain)

**Components:**
1. **Resolve** — Extract `TenantID` from `principal.TenantID` and store in `x/tenant` context
2. **Rate Limit** — Token-bucket limiter with uniform config (all tenants same limits)
3. **Quota** — Fixed-window request quota enforcer (e.g., 600 req/min per tenant)

**Data flow:**
- Request enters with JWT containing `tenantID` claim
- Resolve middleware lifts it into `x/tenant` context
- Handlers call `x/tenant.TenantFromContext(r.Context())` to get tenant ID
- All repository methods accept explicit `tenantID` parameter
- Cross-tenant lookups return 404 (not 403) to hide existence

**Customization:** Replace `uniformRateLimits` and `uniformQuota` structs in `routes.go` with dynamic lookups:
```go
limiter: tenantcore.NewTokenBucketRateLimiter(dynamicLimits{
    getTenantRPS: func(tid string) int { /* lookup from config or db */ },
}),
```

### `x/rest` (beta) — Resource CRUD

**Wired in:** `internal/handler/projects.go` (controller) and `internal/app/routes.go` (routes)

**Pattern:** Handler embeds `*x/rest.Controller[Project]`; registers standard CRUD routes:
```go
projectsC := handler.NewProjectsController(service, audit, logger)
v1.get("/projects", authed(projectsC.Index))
v1.post("/projects", authed(idem(projectsC.Create)))
v1.get("/projects/:id", authed(projectsC.Show))
v1.put("/projects/:id", authed(idem(projectsC.Update)))
v1.delete("/projects/:id", authed(requireAdmin(idem(projectsC.Delete))))
```

**Customization:** Replace `x/rest` with hand-rolled handlers if you need non-standard CRUD (e.g., partial list fields, custom update semantics).

### `x/observability` (beta) — Prometheus Metrics

**Wired in:**
- `internal/app/app.go`: Construct `PrometheusCollector`; pass to `httpmetrics` middleware
- `internal/app/routes.go`: Mount `PrometheusExporter` on `/metrics`

**Metrics emitted automatically:**
- `http_requests_total` — counter by method, path, status
- `http_request_duration_seconds` — histogram of request latency
- Custom metrics can be recorded via `a.Collector.Counter()`, `a.Collector.Gauge()`, etc.

**Customization:** Replace `NewPrometheusCollector` with a `metrics.Collector` adapter for OpenTelemetry, Datadog, or any metrics backend.

## Security Boundaries

### Authentication

- Auth is **per-route**, not global. Public endpoints (signup, login, health, metrics) have no auth guard.
- JWT verification happens in `handler.RequireAuth` middleware.
- On invalid token: 401 Unauthorized (no principal context, request rejected immediately).
- On expired token: 401 Unauthorized (client must refresh via `/api/v1/auth/refresh`).
- Token format: `Bearer <JWT>` in `Authorization` header; signature verified with `APP_JWT_SECRET`.

### Authorization (RBAC)

- RBAC enforcement happens **in handlers and domain**, not middleware.
- `handler.RequireRole(access.RoleAdmin, ...)` wraps handlers that require admin.
- Role lattice: Owner > Admin > Member (evaluated with `access.Can(fromRole, toRole)`)
- Last-owner invariant: Cannot demote/remove the last owner of a workspace (checked in domain).
- On permission denied: 403 Forbidden (principal exists but lacks role).

### Tenant Isolation

- Every repository method accepts explicit `tenantID` parameter (enforced at compile time).
- Cross-tenant queries return 404 (not 403) to avoid leaking existence.
- Example: `projectStore.ByID(ctx, tenantID, projectID)` — if project is in a different tenant, returns `ErrNotFound`.
- Tenant ID extracted from JWT claim; validated against `x/tenant` context.

### Secrets

**Never logged:**
- `APP_JWT_SECRET` — HMAC signing key
- Password hashes — stored in-memory; never serialized to logs
- Refresh tokens — single-use opaque tokens; hashed in KV store
- `Idempotency-Key` header — request dedup key; excluded from audit

**Log filtering:** `accesslog` middleware and error handlers skip sensitive fields.

## Extension Stability

- **Stable roots** (all imported): `core`, `router`, `contract`, `middleware/*`, `security/*`, `store/*`, `health`, `log`, `metrics` — v1 compatibility guarantee
- **Beta** (carefully used): `x/tenant` (per-route resolve, ratelimit, quota; all implemented), `x/rest` (resource controller pattern), `x/observability` (Prometheus; can be swapped)
- **Experimental**: None used (none needed for this use-case)

See `docs/concepts/extension-maturity.md` for stability definitions and migration paths.
