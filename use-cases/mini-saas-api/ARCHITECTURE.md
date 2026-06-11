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
| `x/tenant` | `internal/app/routes.go` (per-route middleware chain) | Tenant resolution from JWT claim, rate limit, quota |
| `x/rest` | `internal/app/routes.go` (resource controller) | Projects CRUD |
| `x/observability` | `internal/app/app.go` (collector) + `routes.go` (/metrics route) | Prometheus metrics |

## Security Boundaries

- Auth middleware (JWT verification → principal context) is per-route, not global.
- RBAC is enforced in handler layer, calling `internal/domain/access`. Not in middleware.
- Tenant isolation: every repository method accepts an explicit `tenantID` argument.
  Cross-tenant lookups return 404 (not 403) to avoid existence leaks.
- Secrets never log: JWT secret, password hashes, refresh tokens, and `Idempotency-Key`
  values are excluded from all log fields.
