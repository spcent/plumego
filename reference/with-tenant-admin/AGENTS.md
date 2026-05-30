# AGENTS.md - reference/with-tenant-admin

Operational guide for agents working in `reference/with-tenant-admin`.

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
and extends it with a multi-resource admin API.

## 2. Purpose And Boundaries

`reference/with-tenant-admin` demonstrates a protected administrative API for
managing tenant lifecycle, quotas, and usage using `x/tenant/core`. All routes
are guarded by a static bearer-token middleware. In-memory stores make the service
self-contained for demos and tests; production replaces them with real storage.

Hard rules:

- `x/tenant/core` is the only allowed `x/*` import.
- No new third-party dependencies.
- No hidden globals, `init()` registration, or reflection routing.
- Keep `main.go` thin: load config, construct app, register routes, start.
- Keep middleware wiring explicit in `internal/app/app.go`.
- Keep all routes explicit in `internal/app/routes.go`; apply `auth.RequireAdminToken`
  to every route — no unprotected admin endpoint.
- Preserve `func(http.ResponseWriter, *http.Request)` handler shape.
- Use `contract.WriteResponse` for success and `contract.WriteError` with
  `contract.NewErrorBuilder()` for errors.
- `ADMIN_TOKEN` must be non-empty and validated at startup; the service must refuse
  to start with an empty token.

## 3. Package Ownership

- `main.go`: process entrypoint only.
- `internal/config`: config loading, defaults, and environment (APP_ADDR, ADMIN_TOKEN,
  APP_LOG_LEVEL).
- `internal/app/app.go`: middleware chain (requestid → recovery), graceful shutdown.
- `internal/app/routes.go`: route table; all routes wrapped with `auth.RequireAdminToken`.
- `internal/handler/tenantadmin`: Handler with CreateTenant, GetTenant, SuspendTenant, DeleteTenant.
- `internal/handler/quotaadmin`: Handler with GetQuota, SetQuota, ResetQuota.
- `internal/handler/usage`: Handler with RecordUsage, GetUsageReport.
- `internal/auth`: bearer-token guard middleware.

Dependency direction:

```text
main.go → config → app → handler/*, auth, x/tenant/core
handler/* → contract, router
auth → contract
```

Handlers must not import `internal/app` or `internal/config`.

## 4. Change Patterns

Add a protected admin endpoint:

1. Implement the handler method in the appropriate `internal/handler/<domain>` package.
2. Register the route in `routes.go` wrapped with `auth.RequireAdminToken`.
3. Add a focused handler test; use a valid token in the test request header.

Add a new admin resource:

1. Create `internal/handler/<resource>/handler.go` with a `Handler` struct and
   dependency injection via a `Deps` struct.
2. Wire the in-memory store dependency in `routes.go`.
3. Register routes under `/admin/<resource>/...` with admin auth.

Swap in-memory store for persistent storage:

1. Implement the same store interface from `x/tenant/core` backed by a real database.
2. Wire the new implementation in `routes.go` — handlers stay unchanged.

## 5. Validation

For docs-only changes:

```bash
git diff --check
```

For Go changes inside this service:

```bash
cd reference/with-tenant-admin && go test -race -timeout 30s ./...
go run ./internal/checks/dependency-rules
```

## 6. Review Focus

When reviewing or optimizing this service, check:

- Every route is protected with `auth.RequireAdminToken` — no admin endpoint is public.
- `ADMIN_TOKEN` is validated as non-empty at startup.
- `x/tenant/core` only — no other `x/*` packages.
- Handler dependencies are injected via `Deps` struct; no package-level vars.
- In-memory stores are clearly documented as non-persistent; swappable via interface.
- Response envelope uses `contract.WriteResponse` / `contract.WriteError`.
