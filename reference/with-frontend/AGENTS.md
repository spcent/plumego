# AGENTS.md - reference/with-frontend

Operational guide for agents working in `reference/with-frontend`.

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
and extends it with static asset serving.

## 2. Purpose And Boundaries

`reference/with-frontend` demonstrates how to serve a Single-Page Application
alongside a JSON API from the same process using `x/frontend`. API routes take
precedence over the catch-all SPA route. The demo uses an in-memory `fstest.MapFS`;
production replaces it with `embed.FS` or `frontend.RegisterFromDir()`.

Hard rules:

- `x/frontend` is the only allowed `x/*` import.
- No new third-party dependencies.
- No hidden globals, `init()` registration, or reflection routing.
- Keep `main.go` thin: load config, construct app, register routes, start.
- Keep middleware wiring explicit in `internal/app/app.go`.
- Keep all routes explicit in `internal/app/routes.go`; register API routes before
  the SPA catch-all so they take precedence.
- Preserve `func(http.ResponseWriter, *http.Request)` handler shape.
- Use `contract.WriteResponse` for success and `contract.WriteError` with
  `contract.NewErrorBuilder()` for errors.

## 3. Package Ownership

- `main.go`: process entrypoint only.
- `internal/config`: config loading, defaults, and environment
  (APP_ADDR, API_PREFIX, UI_PREFIX, ASSETS_DIR).
- `internal/app/app.go`: middleware chain (requestid → recovery), graceful shutdown.
- `internal/app/routes.go`: API routes first, then `frontend.RegisterFS()` for assets.
- `internal/handler`: APIHandler with Status() for the JSON API surface.

Dependency direction:

```text
main.go → config → app → handler, x/frontend
handler → contract, router
```

Handlers must not import `internal/app` or `internal/config`.

## 4. Change Patterns

Swap the demo FS for embedded assets:

1. Add `//go:embed assets/*` and `var assets embed.FS` to `main.go` or a dedicated file.
2. Replace `fstest.MapFS{...}` in `routes.go` with `assets`.
3. Update `ASSETS_DIR` in `env.example` with the correct embed path comment.

Add an API endpoint:

1. Add a method to `APIHandler` (or a new handler struct) in `internal/handler`.
2. Register the route in `routes.go` **before** the `frontend.RegisterFS(...)` call.
3. Add a focused test.

Change asset cache headers:

1. Pass `frontend.Config{...}` options to `frontend.RegisterFS()` in `routes.go`.
2. Adjust `MaxAge` for versioned assets and `NoCache` for index.html.

## 5. Validation

For docs-only changes:

```bash
git diff --check
```

For Go changes inside this service:

```bash
cd reference/with-frontend && go test -race -timeout 30s ./...
go run ./internal/checks/dependency-rules
```

## 6. Review Focus

When reviewing or optimizing this service, check:

- API routes are registered **before** the SPA catch-all in `routes.go`.
- `x/frontend` only — no other `x/*` packages.
- `embed.FS` or `frontend.RegisterFromDir()` replaces `fstest.MapFS` in production.
- Cache-control headers are appropriate: immutable for versioned assets, no-cache for index.html.
- Response envelope uses `contract.WriteResponse` / `contract.WriteError` for API responses.
