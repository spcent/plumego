# Components

> **Package**: `github.com/spcent/plumego/core`

Components are pluggable units that can register routes/middleware and participate in app lifecycle.

---

## Component Interface

```go
type Component interface {
    RegisterRoutes(r *router.Router)
    RegisterMiddleware(m *middleware.Registry)
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Health() (name string, status health.HealthStatus)
}
```

Use `core.BaseComponent` to embed no-op defaults for optional methods.

---

## Register Components

At app construction:

```go
app := core.New(
    core.WithComponent(componentA),
    core.WithComponents(componentB, componentC),
)
```

Components can also be added by convenience wrappers before boot freeze (for example websocket helpers), but canonical setup is constructor-time registration.

---

## Lifecycle Order

During `Boot()`:

1. Each component `RegisterMiddleware(...)`
2. Each component `RegisterRoutes(...)`
3. Components `Start(...)` in declared order

During shutdown:

1. Components `Stop(...)` in reverse start order

If a component fails to start, already-started components are stopped in reverse order.

---

## Example Component

```go
type AuditComponent struct {
    core.BaseComponent
    logger log.StructuredLogger
}

func (c *AuditComponent) RegisterRoutes(r *router.Router) {
    r.Get("/audit/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        w.Write([]byte("ok"))
    }))
}

func (c *AuditComponent) Start(ctx context.Context) error {
    c.logger.Info("audit component started")
    return nil
}

func (c *AuditComponent) Stop(ctx context.Context) error {
    c.logger.Info("audit component stopped")
    return nil
}

func (c *AuditComponent) Health() (string, health.HealthStatus) {
    return "audit", health.HealthStatus{Status: health.StatusHealthy}
}
```

---

## Built-in Components

Common component packages:

- `core/components/devtools`
- `core/components/observability`
- `core/components/ops`
- `core/components/webhook`
- `core/components/websocket`
- `core/components/tenant` (experimental surface)

Notes:

- Devtools component is auto-mounted when `WithDebug()` is enabled (unless already present).
- WebSocket can be composed either by `WithComponent(...)` or helper methods (`ConfigureWebSocket*`).

---

## Health Reporting

`Health()` should return a stable component name and a structured health status.

```go
func (c *MyComponent) Health() (string, health.HealthStatus) {
    return "my-component", health.HealthStatus{Status: health.StatusHealthy}
}
```

Avoid expensive checks in `Health()`; keep it fast and side-effect free.

---

## Best Practices

- Keep components focused on transport/capability wiring.
- Avoid hidden global side effects in `Start()`.
- Ensure `Stop()` is idempotent and context-aware.
- Register routes/middleware explicitly and deterministically.
- Prefer constructor-based dependency injection into component structs.
