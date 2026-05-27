# Agent Tasks — with-rest

Operating guide for AI coding agents working in this feature demo.

Read `AGENTS.md` (repository root) for the full operating contract.
Read `ARCHITECTURE.md` (this directory) for layout rationale.
Read `docs/modules/x/rest/README.md` for the `x/rest` API.

---

## Zone Classification

### Safe to modify

| Path | What you can do |
|---|---|
| `internal/config/config.go` | Add config fields, change defaults |
| `internal/domain/user/user.go` | Change the User model or Repository logic |
| `internal/app/routes.go` | Add routes, add per-route middleware |

### Restricted — requires preflight + reviewer note

| Path | Constraint |
|---|---|
| `internal/app/app.go` | Middleware order and resource construction are load-bearing |
| `main.go` | Owns process signal context and top-level wiring only |

### Frozen

- Do not put HTTP logic in `internal/domain/`.
- Do not add `x/*` imports beyond `x/rest` without justification.
- Do not use global variables or `init()` functions.

---

## Common Task Recipes

### Add a new resource

1. Add a model and repository in `internal/domain/<name>/`.
2. In `app.New` (`internal/app/app.go`), create the controller:
   ```go
   spec := rest.DefaultResourceSpec("<name>").WithPrefix("/api/<name>")
   a.<Name> = rest.NewDBResource[domain.Model](spec, domain.NewRepository())
   ```
3. Add the field to `App`:
   ```go
   type App struct {
       Core   *core.App
       Cfg    config.Config
       Users  *rest.DBResourceController[user.User]
       <Name> *rest.DBResourceController[<domain>.Model]
   }
   ```
4. Register routes in `RegisterRoutes` (`internal/app/routes.go`).

### Add input validation

Call `WithValidator` before storing the controller on `App`:

```go
users := rest.NewDBResource[user.User](spec, user.NewRepository()).
    WithValidator(myValidator)
```

`myValidator` must implement `interface{ Validate(any) error }`.

### Add a plain HTTP route

Register alongside the resource routes in `routes.go`:

```go
if err := a.Core.Get("/api/hello", http.HandlerFunc(hello)); err != nil {
    return err
}
```

---

## Validation Commands

```bash
# Module tests
go test -race -timeout 30s ./reference/with-rest/...

# Boundary checks
go run ./internal/checks/dependency-rules
go run ./internal/checks/module-manifests

# Full gates (cross-module or release-relevant changes)
make gates
```

---

## Non-goals

- Do not add business logic beyond the minimal User CRUD.
- Do not add authentication patterns — those belong in `reference/with-tenant`.
- Do not prototype room-based or streaming patterns here.
- Do not use this demo as a starting point for production services without
  replacing the in-memory repository and adding validation.
