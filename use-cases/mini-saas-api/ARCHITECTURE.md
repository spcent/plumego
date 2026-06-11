# Architecture — mini-saas-api

## Ownership

| Package | Owns |
|---|---|
| `main.go` | Process entrypoint: load config, construct app, register routes, start |
| `internal/config` | Env/flag loading, defaults, validation; never imported by handler or domain |
| `internal/app/app.go` | core.New, global middleware chain, graceful shutdown |
| `internal/app/routes.go` | Route table, handler dependency wiring; the only place x/* extensions are imported |
| `internal/handler` | HTTP adaptation: request parsing, response writing, handler-local DTOs |
| `internal/domain/user` | User model, repository interface, in-memory store, service |
| `internal/domain/tenantspace` | Tenant + membership model, repository, service |
| `internal/domain/project` | Project model, repository, service (consumed by x/rest controller) |
| `internal/domain/access` | RBAC role lattice; called from handlers, never from middleware |
| `internal/domain/audit` | Append-only audit log recorder and query |
| `internal/domain/session` | JWT access-token issuing + opaque refresh tokens with rotation and family revocation (over a narrow KV interface) |
| `internal/domain/ident` | Random 128-bit hex identifiers |
| `internal/platform/idemstore` | In-memory implementation of the stable `store/idempotency` contract |

## Dependency Direction (one-way, enforced)

```
main.go → internal/config → internal/app → internal/handler
                                    ↘ internal/domain/*
internal/handler → internal/domain/*
```

Handlers must not import `internal/app` or `internal/config`.
Domain packages must not import handler packages.
`x/*` imports live only in `internal/app`.

## Extension Wiring Points

| Extension | Wired in | Purpose |
|---|---|---|
| `x/tenant` | `internal/app/routes.go` (per-route chain: resolve → ratelimit → quota) | Tenant context from JWT claim, per-tenant token bucket + fixed-window quota |
| `x/rest` | controller in `internal/handler/projects.go`, routes in `internal/app/routes.go` | Projects CRUD resource controller |
| `x/observability` | `internal/app/app.go` (collector into httpmetrics) + `routes.go` (`/metrics` exporter) | Prometheus metrics |

## Request path (authenticated routes)

```
requestid → securityheaders → cors → recovery → accesslog → bodylimit
→ httpmetrics(Prometheus) → timeout                       (global, app.go)
→ RequireAuth (JWT → principal w/ tenant+role)            (per route)
→ x/tenant resolve → ratelimit → quota                    (per route)
→ [RequireRole(admin)]  [Idempotent]                      (where applicable)
→ handler → domain service → repository
```

## Security Boundaries

- Auth middleware (JWT verification → principal context) is per-route, not global.
- RBAC is enforced in handler layer, calling `internal/domain/access`. Not in middleware.
- Tenant isolation: every repository method accepts an explicit `tenantID` argument.
  Cross-tenant lookups return 404 (not 403) to avoid existence leaks.
- Secrets never log: JWT secret, password hashes, refresh tokens, and `Idempotency-Key`
  values are excluded from all log fields.
