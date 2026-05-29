# AGENTS.md - reference/with-rest

Operational guide for agents working in `reference/with-rest`.

This file is intentionally short. It compresses the local rules agents need for
routine changes in this feature demo. The repository root `AGENTS.md` remains
authoritative; when a task becomes architectural, security-sensitive, or
cross-module, fall back to the root workflow and the matching
`specs/task-routing.yaml` entry.

## 1. Minimal Context

For changes confined to this directory, load only:

- repository root `AGENTS.md`, if not already loaded
- this file
- the touched Go files and their tests

Read these local docs only when they answer a concrete question:

- `ARCHITECTURE.md`: layout rationale and resource controller wiring
- `AGENT_TASKS.md`: adding a resource, adding validation, adding a plain route
- `PRODUCTION_CHECKLIST.md`: replacing the in-memory repository, adding auth

## 2. Purpose And Boundaries

`reference/with-rest` shows how to add `x/rest` resource controllers to a
service that follows Plumego's canonical application shape. Everything not
related to resource controllers is identical to `reference/standard-service`.

Hard rules:

- No `x/*` imports beyond `x/rest` and `x/validate`.
- No new third-party dependencies beyond `go-playground/validator` (already in `go.mod`).
- No hidden globals, `init()` registration, reflection routing, or controller scanning.
- Keep `main.go` thin: load config, construct app, register routes, start.
- Keep middleware wiring explicit in `internal/app/app.go`.
- Keep all public routes explicit in `internal/app/routes.go`.
- Do not put HTTP logic in `internal/domain/`.

## 3. Package Ownership

- `main.go`: process entrypoint only.
- `internal/config`: app-local config loading and defaults.
- `internal/app/app.go`: core app construction, controller wiring, middleware order.
- `internal/app/routes.go`: route table — resource routes registered individually.
- `internal/handler`: plain HTTP handlers (non-resource-controller endpoints).
- `internal/domain/user`: User model and in-memory repository implementing `rest.Repository[User]`.

## 4. Change Patterns

Add a resource:

1. Add model and repository in `internal/domain/<name>/`.
2. Create `DBResourceController` in `app.New` (`internal/app/app.go`).
3. Register each action explicitly in `RegisterRoutes` (`internal/app/routes.go`).
4. Add focused tests next to changed behavior.

Add a plain HTTP handler:

1. Implement in `internal/handler/`.
2. Register in `routes.go` alongside resource routes.

## 5. Validation

For docs-only changes:

```bash
git diff --check
```

For Go changes inside this service:

```bash
cd reference/with-rest && go test -race -timeout 30s ./...
go run ./internal/checks/dependency-rules
go run ./internal/checks/module-manifests
```

## 6. Review Focus

When reviewing or optimizing this service, check:

- all routes are registered explicitly in `routes.go`, no auto-wiring
- no `x/*` imports beyond `x/rest` and `x/validate`
- domain package has no HTTP imports
- response envelope uses `contract.WriteResponse` / `contract.WriteError` only
