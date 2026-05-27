# Agent Tasks — production-service

Operating guide for AI coding agents working in this reference application.

Read `AGENTS.md` (repository root) for the full operating contract.
Read `README.md` (this directory) for routes, environment variables, and storage notes.
Read `PRODUCTION_CHECKLIST.md` (this directory) for production hardening decisions.

---

## Zone Classification

### Safe to modify

| Path | What you can do |
|---|---|
| `internal/handler/service.go` | Update service metadata response |
| `internal/handler/status.go` | Extend the status report fields |
| `internal/handler/ops.go` | Update ops metrics response shape |
| `internal/handler/profile.go` | Extend profile handler logic |
| `internal/handler/handler_test.go` | Add or update handler tests |
| `internal/domain/tenant/tenant.go` | Extend the profile model |
| `internal/domain/tenant/store.go` | Change file-store serialization or seeded profiles |
| `internal/config/config.go` | Add config fields, change defaults |
| `internal/config/config_test.go` | Add config tests |
| `internal/app/app_test.go` | Update route shape or middleware assertions |

### Restricted — requires preflight + reviewer note

| Path | Constraint |
|---|---|
| `internal/app/app.go` | Middleware order is load-bearing; all 9 layers must stay in documented sequence |
| `internal/app/routes.go` | Adding a route changes the public contract; confirm the route is intentional |
| `main.go` | Owns signal context and top-level wiring only; do not add logic |

### Frozen

| Constraint | Reason |
|---|---|
| No `x/*` imports beyond `x/tenant/resolve` | Stable-root-only production baseline; extension capabilities are explicit add-ons |
| No new globals or `init()` functions | Constructor injection only |
| `APP_API_TOKEN` and `OPS_TOKEN` must have no fallback | Fail-closed by design; adding a default breaks the security guarantee |
| `x/observability/devtools` must not be mounted | Debug routes must not reach production |

---

## Common Task Recipes

### Add a new protected API route

1. Add a handler method in `internal/handler/`.
2. Add bearer-auth guard in `internal/app/routes.go` using `middleware/auth`.
3. Wire handler dependencies from `internal/domain/` or `internal/app/`.
4. Register the route in `routes.go` with explicit verb and path.
5. Add a handler test that checks the 401 path (missing/invalid token) and the
   200 path (valid token).

```go
// routes.go — protected route pattern
profile := handler.ProfileHandler{
    Store:  a.Profiles,
    Logger: a.Core.Logger(),
}
apiToken := middleware.RequireBearer(a.Cfg.App.APIToken)
if err := a.Core.Get("/api/v1/resource/:id",
    http.HandlerFunc(auth.Chain(apiToken, profile.GetResource))); err != nil {
    return err
}
```

### Add a new ops route

1. Implement or extend `internal/handler/ops.go`.
2. Apply the ops bearer guard in `routes.go`.
3. Add a test for the 401 and 200 paths.

```go
// routes.go — ops route pattern
ops := handler.OpsHandler{Metrics: a.Metrics, Logger: a.Core.Logger()}
opsToken := middleware.RequireBearer(a.Cfg.App.OpsToken)
if err := a.Core.Get("/ops/new-stat",
    http.HandlerFunc(auth.Chain(opsToken, ops.NewStat))); err != nil {
    return err
}
```

### Add a config field

1. Add the field to `AppConfig` in `internal/config/config.go`.
2. Set a safe default in `Defaults()` (avoid defaults for secret fields).
3. Read from environment in `applyEnv()`.
4. Validate unusable values (e.g., non-positive limits) in `Validate()`.
5. Use the field in `routes.go` or pass it to a handler/middleware constructor.

**For secret fields** (tokens, keys): do not set a default; let the zero value
mean "disabled" or "fail-closed", and document this in `PRODUCTION_CHECKLIST.md`.

### Change the tenant header name

In `internal/app/routes.go`, update the `x/tenant/resolve` middleware
configuration:

```go
tenantMw := tenantresolve.Middleware(tenantresolve.Config{
    HeaderName: "X-My-Tenant-ID",
})
```

Confirm the new header name is documented in `README.md` and `PRODUCTION_CHECKLIST.md`.

### Replace the noop tracer with a real one

In `internal/app/app.go`, swap `noopTracer{}` with your `tracing.Tracer`
implementation. The `tracing.Middleware` accepts any type that satisfies the
`tracing.Tracer` interface:

```go
// Before:
tracing.Middleware(noopTracer{})

// After (example with x/observability/tracer):
tracer, err := ottracer.New(cfg.App.TraceEndpoint)
if err != nil { return nil, fmt.Errorf("configure tracer: %w", err) }
// ...
tracing.Middleware(tracer)
```

Remove `noopTracer` and `noopSpan` from `app.go` once a real tracer is wired.

---

## Validation Commands

```bash
# Module tests
cd reference/production-service && go test -race -timeout 30s ./...

# Boundary checks (run from repo root)
go run ./internal/checks/dependency-rules
go run ./internal/checks/reference-layout
go run ./internal/checks/module-manifests

# Full gates (cross-module or release-relevant changes)
make gates
```

---

## Non-goals

Do not use this reference app to:

- Demonstrate `x/*` extensions beyond `x/tenant/resolve` (use a `with-*` demo)
- Prototype new auth schemes (extend `middleware/auth` instead)
- Add business logic beyond profile read and service metadata
- Serve as a runtime library or embed into other services

If a task requires `x/*` imports beyond `x/tenant/resolve`, create or extend a
`reference/with-<capability>` demo instead of modifying this baseline.
