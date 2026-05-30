# AGENTS.md - reference/with-tenant

Operational guide for agents working in `reference/with-tenant`.

This file is intentionally short. The repository root `AGENTS.md` remains
authoritative; when a task becomes architectural, security-sensitive, or
cross-module, fall back to the root workflow and the matching
`specs/task-routing.yaml` entry.

## 1. Minimal Context

For changes confined to this directory, load only:

- repository root `AGENTS.md`, if not already loaded
- this file
- the touched Go files and their tests

Read `reference/standard-service` first — this service has the same core shape
and extends it with per-tenant middleware.

## 2. Purpose And Boundaries

`reference/with-tenant` demonstrates per-tenant rate limiting, quota enforcement,
and policy evaluation using `x/tenant`. The tenant middleware chain (resolve →
policy → quota → ratelimit) is applied per-route, not globally, so non-tenant
endpoints incur no overhead.

Hard rules:

- `x/tenant/*` packages are allowed. No other `x/*` imports.
- No new third-party dependencies.
- No hidden globals, `init()` registration, or reflection routing.
- Keep `main.go` thin: load config, construct app, register routes, start.
- Keep middleware wiring explicit in `internal/app/app.go`.
- Keep all routes explicit in `internal/app/routes.go`; apply the tenant middleware
  chain per-route in `routes.go`, not globally.
- Preserve `func(http.ResponseWriter, *http.Request)` handler shape.
- Use `contract.WriteResponse` for success and `contract.WriteError` with
  `contract.NewErrorBuilder()` for errors.
- Tenant configuration (rate limits, quotas, allowed plans) is seeded in-memory
  for demo; do not add external storage without a clear task requirement.

## 3. Package Ownership

- `main.go`: process entrypoint only.
- `internal/config`: config loading, defaults, and environment (APP_ADDR).
- `internal/app/app.go`: middleware chain (requestid → recovery), graceful shutdown.
- `internal/app/routes.go`: tenant middleware chain assembled per-route; ModelHandler
  registered under the chain.
- `internal/handler`: ModelsHandler with ServeHTTP() — lists models for the resolved tenant.

Dependency direction:

```text
main.go → config → app → handler, x/tenant modules
handler → contract, router
```

Handlers must not import `internal/app` or `internal/config`.

## 4. Change Patterns

Add a tenant-aware endpoint:

1. Implement the handler in `internal/handler`.
2. In `routes.go`, wrap the handler with the existing tenant middleware chain before
   registering the route.
3. Add a focused test that sets a tenant ID header and asserts the response.

Change tenant rate limits or quotas:

1. Update the in-memory `TenantConfig` seed in `routes.go`.
2. Add a test that exercises the updated limits.

Add a non-tenant endpoint:

1. Register the route in `routes.go` **without** the tenant middleware chain.
2. The route will bypass tenant resolution and quota enforcement.

## 5. Validation

For docs-only changes:

```bash
git diff --check
```

For Go changes inside this service:

```bash
cd reference/with-tenant && go test -race -timeout 30s ./...
go run ./internal/checks/dependency-rules
```

## 6. Review Focus

When reviewing or optimizing this service, check:

- Tenant middleware chain is applied per-route, not globally.
- `x/tenant/*` only — no other `x/*` packages.
- A missing or invalid tenant ID must return 401/403, never pass to the handler.
- Quota and rate-limit rejections use the correct HTTP status (429 for rate limit,
  403 for quota/policy denial).
- Response envelope uses `contract.WriteResponse` / `contract.WriteError`.
