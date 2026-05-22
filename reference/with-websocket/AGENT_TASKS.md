# Agent Tasks — with-websocket

Operating guide for AI coding agents working in this feature demo.

Read `AGENTS.md` (repository root) for the full operating contract.
Read `ARCHITECTURE.md` (this directory) for layout rationale.
Read `docs/modules/x-websocket/README.md` for the `x/websocket` API.

---

## Zone Classification

### Safe to modify

| Path | What you can do |
|---|---|
| `internal/config/config.go` | Add config fields, change defaults |
| `internal/app/routes.go` | Add routes, add middleware to specific routes |

### Restricted — requires preflight + reviewer note

| Path | Constraint |
|---|---|
| `internal/app/app.go` | Hub configuration and shutdown order are load-bearing |
| `main.go` | Owns process signal context and top-level wiring only |

### Frozen

- Do not remove the WebSocket shutdown before the HTTP server shutdown in `Start(ctx)`.
- Do not add `x/*` imports beyond `x/websocket` without justification.
- Do not set `AllowUnauthenticated = true` in production configurations.

---

## Common Task Recipes

### Add a plain HTTP route alongside WebSocket

Add the handler inline or in a new handler file, then register in `routes.go`:

```go
if err := a.Core.Get("/api/status", http.HandlerFunc(statusHandler)); err != nil {
    return err
}
```

The WebSocket routes are registered by `a.WS.RegisterRoutes(a.Core)` — do not
re-register WebSocket paths manually.

### Change WebSocket hub configuration

In `internal/app/app.go`, `New()`:

```go
wsCfg := websocket.DefaultWebSocketConfig()
wsCfg.Secret = []byte(cfg.WSSecret)
wsCfg.AllowUnauthenticated = false   // require auth in production
wsCfg.AllowedOrigins = []string{"https://app.example.com"}
wsCfg.MaxConnections = 1000
```

All WebSocket configuration is in one place: `app.New`. Do not scatter
configuration across multiple files.

---

## Validation Commands

```bash
# Module tests
go test -race -timeout 30s ./reference/with-websocket/...

# Boundary checks
go run ./internal/checks/dependency-rules
go run ./internal/checks/module-manifests

# Full gates (cross-module or release-relevant changes)
make gates
```

---

## Non-goals

- Do not use this demo as the starting point for room-based broadcast systems
  without first reading `docs/modules/x-websocket/README.md`.
- Do not add business logic or domain models to this demo.
- Do not use this demo to prototype authentication patterns — it intentionally
  leaves auth minimal for clarity.
