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

Before boot freeze, components can also be mounted explicitly:

```go
if err := app.MountComponent(componentA); err != nil {
    log.Fatal(err)
}
```

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

- `x/devtools`
- `x/devtools/pubsubdebug`
- `core/components/observability`
- `x/ops`
- `x/webhook`
- `x/websocket`

Notes:

- Devtools should be mounted explicitly from `x/devtools`, for example `app.MountComponent(xdevtools.NewAppComponent(app))`.
- WebSocket should be mounted explicitly from `x/websocket`.
- Tenant capabilities should be wired through `x/tenant/*`, not `core/components/*`.

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
