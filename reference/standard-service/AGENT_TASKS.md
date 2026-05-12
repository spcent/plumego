# Agent Tasks — standard-service

Operating guide for AI coding agents working in this reference application.

Read `AGENTS.md` (repository root) for the full operating contract.
Read `ARCHITECTURE.md` (this directory) for layout rationale.

---

## Zone Classification

### Safe to modify

| Path | What you can do |
|---|---|
| `internal/handler/api.go` | Add, change, or remove handler methods |
| `internal/handler/health.go` | Update health check logic |
| `internal/handler/handler_test.go` | Add or update handler tests |
| `internal/config/config.go` | Add config fields, change defaults |
| `main.go` | Only the four wiring calls; do not add logic |

### Restricted — requires preflight + reviewer note

| Path | Constraint |
|---|---|
| `internal/app/app.go` | Middleware order is load-bearing; changing it affects all routes |
| `internal/app/routes.go` | Adding a route means adding a public contract; confirm the route is intentional |

### Frozen — do not touch

| Constraint | Reason |
|---|---|
| No `x/*` imports | This service depends only on stable roots. Adding `x/*` breaks the canonical reference. |
| No global variables | Constructor injection only. |
| No `init()` functions | Explicit wiring only. |

---

## Common Task Recipes

### Add a new endpoint

1. Add a handler method to `internal/handler/api.go` (or a new file in `internal/handler/`).
2. Register the route in `internal/app/routes.go`.
3. Write a test in `internal/handler/handler_test.go`.
4. Run validation (see below).

Handler shape:
```go
func (h APIHandler) MyEndpoint(w http.ResponseWriter, r *http.Request) {
    // read → process → write
    _ = contract.WriteResponse(w, r, http.StatusOK, response, nil)
}
```

Route registration shape:
```go
if err := a.Core.Get("/api/my-endpoint", http.HandlerFunc(api.MyEndpoint)); err != nil {
    return err
}
```

### Add a config field

1. Add the field to `AppConfig` in `internal/config/config.go`.
2. Set a safe default in `Defaults()`.
3. Read from environment in `applyEnv()`.
4. Optionally expose as a flag in `applyFlags()`.

### Add middleware

1. Import the middleware package in `internal/app/app.go`.
2. Add it to `app.Use(...)` in `New()` at the correct position.
3. Document why the order was chosen in a short inline comment.

Middleware that runs on all routes goes in `app.Use`. Middleware that applies
to a specific route group wraps the handler in `routes.go`.

---

## Validation Commands

After any change to this service, run in order:

```bash
# 1. Module tests
go test -race -timeout 30s ./reference/standard-service/...

# 2. Boundary check — must not import x/*
go run ./internal/checks/dependency-rules

# 3. Module manifests
go run ./internal/checks/module-manifests

# 4. Full gates (for cross-module or release-relevant changes)
make gates
```

---

## Non-goals

Do not use this reference app to:

- Demonstrate `x/*` extensions (use the corresponding `with-<capability>` demo)
- Implement business logic beyond the minimal echo/status/greet endpoints
- Show authentication patterns (those belong in `reference/with-tenant` or `x/security` examples)
- Prototype new features (this is a stable reference, not a sandbox)

If a task requires changes that violate these non-goals, create a new
`reference/with-<capability>` demo instead of modifying this file.
