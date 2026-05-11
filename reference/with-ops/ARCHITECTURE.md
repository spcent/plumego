# Architecture Notes — with-ops

This document explains the structural choices in this feature demo.

---

## Why this demo is non-canonical

`reference/standard-service` uses `core.App` as the application root, which
wires stable-root middleware and provides `Get/Post/Put/Delete` helpers.

This demo bypasses `core.App` intentionally. `x/ops` registers its own route
set via `opsHandler.RegisterRoutes(r *router.Router)`, which requires access
to the underlying router. The demo uses `router.NewRouter()` directly to keep
the wiring simple and explicit for a focused capability demo.

This is the correct choice for a demo; production services that combine `x/ops`
with full `core.App` wiring should inject the router or use `core.App.AddRoute`.

---

## Structure

```
main.go   — router, middleware chain, ops handler, http.ListenAndServe
```

All wiring lives in `main.go`. There is no `internal/` sub-package because:
- The demo has no domain model.
- There is no configuration struct beyond two environment variables.
- The ops handler is self-contained; it holds all ops-specific state.

---

## Key components

### `x/ops` handler

```go
opsHandler := ops.New(ops.Options{
    Enabled: true,
    Auth:    ops.AuthConfig{Token: token},
    Hooks:   ops.Hooks{QueueStats: ...},
    Logger:  logger,
})
if err := opsHandler.RegisterRoutes(r); err != nil { ... }
```

The ops handler owns its route surface. All `/ops/*` routes are registered
by `RegisterRoutes`, not listed individually in `main.go`. The token is the
only auth mechanism — see the production checklist for hardening.

### Metrics collector

```go
collector := metrics.NewBaseMetricsCollector()
```

`metrics.BaseMetricsCollector` is a stable-root type. The `/metrics` endpoint
serves its in-process stats. `httpmetrics.Middleware(collector)` populates
it from every request.

### Middleware chain

```go
handler := middleware.NewChain(
    requestid.Middleware(),
    recoveryMw,
    httpmetrics.Middleware(collector),
    accesslogMw,
).Build(r)
```

Middleware is stacked manually. Order matters: request ID is first so all
downstream middleware can read it; metrics runs before access log so the
log includes accurate timing.

---

## Shutdown

This demo calls `http.ListenAndServe` directly and does not implement graceful
shutdown. For production, use `http.Server` with `Shutdown(ctx)` and a signal
handler (see `reference/standard-service/PRODUCTION_CHECKLIST.md`).

---

## What this demo does not demonstrate

- Graceful shutdown
- TLS
- `core.App` integration (see above for rationale)
- Custom ops hook implementations beyond `QueueStats`
