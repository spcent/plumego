# Agent Tasks — with-ops

Operating guide for AI coding agents working in this feature demo.

Read `AGENTS.md` (repository root) for the full operating contract.
Read `ARCHITECTURE.md` (this directory) for layout rationale.
Read `docs/modules/x-ops/README.md` for the `x/observability/ops` API.

---

## Zone Classification

### Safe to modify

| Path | What you can do |
|---|---|
| `main.go` (ops hooks) | Implement or update `ops.Hooks` functions |
| `main.go` (metrics route) | Update the `/metrics` response shape |

### Restricted — requires preflight + reviewer note

| Path | Constraint |
|---|---|
| `main.go` (middleware chain) | Middleware order is load-bearing; changing it affects all routes |
| `main.go` (ops options) | Changing `Auth` config or disabling ops affects the security posture |

### Frozen

- Do not mount `x/observability/devtools`. This demo explicitly excludes it.
- Do not expose `/ops/*` routes without the `ops.AuthConfig` token check.
- Do not add `x/*` imports beyond `x/observability/ops` without justification.

---

## Common Task Recipes

### Implement a new ops hook

Add the hook function in `ops.Hooks` inside `main.go`:

```go
opsHandler := ops.New(ops.Options{
    // ...
    Hooks: ops.Hooks{
        QueueStats: func(ctx context.Context, queue string) (ops.QueueStats, error) {
            // call your real queue backend
            return ops.QueueStats{Queue: queue, Queued: count, UpdatedAt: time.Now()}, nil
        },
    },
})
```

See `docs/modules/x-ops/README.md` for the full list of available hooks.

### Add a custom ops-adjacent route

Register the route on `r` before calling `opsHandler.RegisterRoutes`:

```go
if err := r.AddRoute(http.MethodGet, "/status", http.HandlerFunc(statusHandler)); err != nil {
    log.Fatalf("register status route: %v", err)
}
```

Custom routes go through the same middleware chain as all other routes.

### Change the ops token

Update the `OPS_TOKEN` environment variable. Do not hard-code the token in
source. Use a secret manager in production environments.

---

## Validation Commands

```bash
# Module tests
go test -race -timeout 30s ./reference/with-ops/...

# Boundary checks
go run ./internal/checks/dependency-rules
go run ./internal/checks/module-manifests

# Full gates (cross-module or release-relevant changes)
make gates
```

---

## Non-goals

- Do not refactor this demo to use `core.App`. The direct router pattern is
  intentional for `x/observability/ops` route registration (see ARCHITECTURE.md).
- Do not add domain models or business logic.
- Do not mount `x/observability/devtools` — it is intentionally excluded from this demo.
- Do not use this demo as a starting point for services without reading the
  production checklist on ops token security and network restriction.
