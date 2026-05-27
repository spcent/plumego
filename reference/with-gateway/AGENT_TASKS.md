# Agent Tasks — with-gateway

Operating guide for AI coding agents working in this feature demo.

Read `AGENTS.md` (repository root) for the full operating contract.
Read `ARCHITECTURE.md` (this directory) for layout rationale.
Read `docs/modules/x/gateway/README.md` for the `x/gateway` API.

---

## Zone Classification

### Safe to modify

| Path | What you can do |
|---|---|
| `internal/config/config.go` | Add config fields, change defaults |
| `internal/app/routes.go` | Add routes, wrap proxy with middleware |

### Restricted — requires preflight + reviewer note

| Path | Constraint |
|---|---|
| `internal/app/app.go` | Proxy construction and backend config are load-bearing |
| `main.go` | Owns process signal context and top-level wiring only |

### Frozen

- Do not expose the backend address as a query parameter or request header.
- Do not add `x/*` imports beyond `x/gateway` without justification.
- Do not hard-code backend URLs; read from `config.Config`.

---

## Common Task Recipes

### Add a backend target

In `internal/config/config.go`, extend the config to support multiple targets:

```go
type AppConfig struct {
    // ...
    GatewayBackends []string `env:"GATEWAY_BACKENDS"`
}
```

In `internal/app/app.go`, pass the full list:

```go
proxy, err := gateway.NewGateway(gateway.GatewayConfig{
    Targets: cfg.GatewayBackends,
})
```

### Change the proxy route prefix

In `internal/app/routes.go`, update the wildcard path:

```go
if err := a.Core.Get("/api/v1/*", http.HandlerFunc(a.Proxy.ServeHTTP)); err != nil {
    return err
}
```

Ensure the backend expects the path as forwarded, or configure prefix
stripping in `GatewayConfig`.

### Add a plain HTTP route alongside the proxy

Register additional routes before or after the proxy wildcard in `routes.go`:

```go
if err := a.Core.Get("/healthz", http.HandlerFunc(healthHandler)); err != nil {
    return err
}
if err := a.Core.Get("/proxy/*", http.HandlerFunc(a.Proxy.ServeHTTP)); err != nil {
    return err
}
```

Route matching is first-match; register specific routes before the wildcard.

---

## Validation Commands

```bash
# Module tests
go test -race -timeout 30s ./reference/with-gateway/...

# Boundary checks
go run ./internal/checks/dependency-rules
go run ./internal/checks/module-manifests

# Full gates (cross-module or release-relevant changes)
make gates
```

---

## Non-goals

- Do not implement service discovery inside this demo.
- Do not add business logic or domain models to this demo.
- Do not use this demo to prototype authentication patterns against backends.
- Do not add room-based or pub/sub routing — that belongs in `with-websocket`.
