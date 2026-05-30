# AGENTS.md - reference/production-service

Operational guide for agents working in `reference/production-service`.

This file is intentionally short. The repository root `AGENTS.md` remains
authoritative; when a task becomes architectural, security-sensitive, or
cross-module, fall back to the root workflow and the matching
`specs/task-routing.yaml` entry.

## 1. Minimal Context

For changes confined to this directory, load only:

- repository root `AGENTS.md`, if not already loaded
- this file
- the touched Go files and their tests

Read these local docs only when they answer a concrete question:

- `README.md`: feature overview and configuration table
- `PRODUCTION_CHECKLIST.md`: hardening, observability, and lifecycle decisions
- `ARCHITECTURE.md`: ownership, middleware order rationale, and dependency direction

There is no `module.yaml` in this submodule. Treat `go.mod` as the local module
boundary and keep validation scoped unless the change touches root packages.

## 2. Purpose And Boundaries

`reference/production-service` extends `reference/standard-service` with
production-grade hardening: bearer-token auth, per-IP abuse guard, distributed
tracing, HTTP metrics, tenant context injection, and persistent profile storage.
Read `standard-service` first — this service extends that shape rather than
replacing it.

Hard rules:

- No `x/*` imports beyond `x/tenant/resolve` (tenant ID injection only).
- No new third-party dependencies.
- No hidden globals, `init()` registration, reflection routing, or controller scanning.
- Keep `main.go` thin: load config, construct app, register routes, start.
- Keep middleware wiring explicit and ordered in `internal/app/app.go`.
- Keep all public routes explicit in `internal/app/routes.go`.
- Preserve `func(http.ResponseWriter, *http.Request)` handler shape.
- Use `contract.WriteResponse` for success and `contract.WriteError` with
  `contract.NewErrorBuilder()` for errors.
- Fail closed on auth: a missing or invalid bearer token must return 401, never pass.

## 3. Package Ownership

- `main.go`: process entrypoint only.
- `internal/config`: config loading, defaults, flags, environment, and validation.
- `internal/app/app.go`: middleware chain (requestid → security → recovery → accesslog
  → bodylimit → abuseguard → tracing → httpmetrics → timeout), graceful shutdown.
- `internal/app/routes.go`: route table; two auth chains: tenant API and ops.
- `internal/handler`: ServiceHandler (health/root), StatusHandler (deployment info),
  ProfileHandler (tenant-aware profile), OpsHandler (in-process metrics).
- `internal/domain/profile`: profile model and JSON-backed persistent store.
- `internal/auth`: bearer-token guard middleware.

Dependency direction:

```text
main.go → config → app → handler
                    \→ domain/profile
                    \→ auth
handler → contract, router, domain/profile
auth    → contract
```

Handlers must not import `internal/app` or `internal/config`.

## 4. Change Patterns

Add a protected endpoint:

1. Implement handler in `internal/handler`.
2. Wrap with `auth.RequireToken` in `routes.go`, not inside the handler.
3. Register one method, one path, one handler per visible route line.
4. Add focused handler tests using `httptest`.

Add config:

1. Add field to `AppConfig` in `internal/config/config.go`.
2. Set a safe default in `Defaults()`.
3. Parse the env var in `applyEnv()`; validate in `Validate()` when the value is required.
4. Update `env.example` with a comment block.

Change middleware order:

1. Edit the `app.Use(...)` call in `internal/app/app.go`.
2. Preserve the documented order: security and CORS before recovery; recovery before
   logging; abuse guard after body limit; tracing before metrics; timeout innermost.

Do not remove endpoints, status codes, JSON fields, error types, or hardening
decisions without an explicit task requesting a contract change.

## 5. Validation

For docs-only changes:

```bash
git diff --check
```

For Go changes inside this service:

```bash
cd reference/production-service && go test -race -timeout 30s ./...
```

Also run from the repository root when middleware wiring, imports, or layout changed:

```bash
go run ./internal/checks/dependency-rules
go run ./internal/checks/reference-layout
```

## 6. Review Focus

When reviewing or optimizing this service, prioritize:

- Bearer token auth must fail closed: 401 on missing or invalid token, never 200.
- `abuseguard` must be after `bodylimit` so oversized bodies are rejected before a
  rate-limit token is consumed.
- Tracing middleware must be wired before `httpmetrics` so spans include metric tags.
- Accidental `x/*` imports beyond `x/tenant/resolve`.
- Response envelope must use `contract.WriteResponse` / `contract.WriteError`.
- `ProfileHandler.GetProfile` must resolve the tenant ID from context, not headers.
