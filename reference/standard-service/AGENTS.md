# AGENTS.md - reference/standard-service

Operational guide for agents working in `reference/standard-service`.

This file is intentionally short. It compresses the local rules agents need for
routine code optimization in this reference app. The repository root
`AGENTS.md` remains authoritative; when a task becomes architectural, security
sensitive, public-API-changing, or cross-module, fall back to the root workflow
and the matching `specs/task-routing.yaml` entry.

## 1. Minimal Context

For changes confined to this directory, do not load the full control plane by
default.

Start with:

- repository root `AGENTS.md`, if not already loaded
- `specs/task-routing.yaml` only far enough to select `app_bootstrap` or
  `http_endpoint`
- this file
- the touched Go files and their tests

Read these local docs only when they answer a concrete question:

- `ARCHITECTURE.md`: ownership, dependency direction, or layout ambiguity
- `AGENT_TASKS.md`: adding an endpoint, config field, middleware, or readiness
  checker
- `PRODUCTION_CHECKLIST.md`: production hardening, security, observability, or
  lifecycle changes
- `README.md`: user-facing reference behavior or scaffold alignment

There is no `module.yaml` in this submodule. Treat `go.mod` as the local module
boundary and keep validation scoped unless the change touches root packages.

## 2. Purpose And Boundaries

`reference/standard-service` is the canonical Plumego application shape. It
teaches explicit wiring, stable-root-only dependencies, standard-library HTTP
handlers, and the default app layout.

Hard rules:

- No `x/*` imports in this service.
- No new third-party dependencies.
- No hidden globals, `init()` registration, reflection routing, or controller
  scanning.
- Keep `main.go` thin: load config, construct app, register routes, start.
- Keep middleware wiring explicit in `internal/app/app.go`.
- Keep all public routes explicit in `internal/app/routes.go`.
- Preserve `func(http.ResponseWriter, *http.Request)` handler shape.
- Use `contract.WriteResponse` for success and `contract.WriteError` with
  `contract.NewErrorBuilder()` for errors.
- Decode JSON request bodies with `json.NewDecoder(r.Body).Decode(...)`.

If a requested capability belongs to an extension, create or use a
`reference/with-<capability>` app instead of expanding this canonical reference.

## 3. Package Ownership

- `main.go`: process entrypoint only.
- `internal/config`: app-local config loading, defaults, flags, environment,
  and validation. Keep kernel config in `Config.Core`.
- `internal/app/app.go`: core app construction, stable middleware order, and
  graceful shutdown.
- `internal/app/routes.go`: route table and handler dependency wiring.
- `internal/handler`: HTTP adaptation, request parsing, response writing, and
  handler-local DTOs.
- `internal/domain/item`: sample domain model (`item.go`), repository interface and in-memory implementation (`store.go`), and service interface and implementation (`service.go`). The three-file split demonstrates the canonical model → repository → service layering described in the root style guide §3.

Dependency direction is one way:

```text
main.go -> internal/config -> internal/app -> internal/handler
                                      \-> internal/domain/item
internal/handler -> contract/router and internal/domain/item
```

Handlers must not import `internal/app` or `internal/config`.

## 4. Change Patterns

Add or change an endpoint:

1. Implement handler logic in `internal/handler`.
2. Put domain behavior or storage in `internal/domain/<name>`.
3. Declare handler dependencies as interfaces in the handler package.
4. Wire concrete dependencies in `internal/app/routes.go`.
5. Register one method, one path, one handler per visible route line.
6. Add focused handler and domain tests.

Add config:

1. Add the app field in `internal/config.AppConfig`.
2. Set a safe default in `Defaults()`.
3. Support `.env`, process env, and flags only when needed.
4. Validate unusable values before startup.

Add middleware:

1. Wire global middleware in `internal/app/app.go`.
2. Preserve visible order and add a short comment only when order matters.
3. For route-specific middleware, wrap the handler in `routes.go`.

Do not remove endpoints, status codes, JSON fields, error types, or sample
teaching behavior during optimization unless the task explicitly asks for that
contract change.

## 5. Validation

For docs-only changes:

```bash
git diff --check
```

For Go changes inside this service:

```bash
cd reference/standard-service && go test -race -timeout 30s ./...
```

Also run from the repository root when route shape, middleware wiring, imports,
or reference layout changed:

```bash
go run ./internal/checks/dependency-rules
go run ./internal/checks/reference-layout
go run ./internal/checks/module-manifests
```

Use full `make gates` only for cross-module, release-relevant, public API, or
root-package changes.

## 6. Review Focus

When reviewing or optimizing this service, prioritize:

- accidental `x/*` imports or new dependencies
- hidden wiring replacing explicit route or middleware declarations
- response envelope drift away from `contract`
- handler logic that belongs in domain or app wiring
- route/status/JSON contract regressions without tests
- shutdown, context, and middleware-order regressions
