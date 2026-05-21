# Architecture Notes — with-websocket

This document explains the structural choices in this feature demo. It
extends `reference/standard-service` with `x/websocket` and follows the
same layout discipline.

---

## What this demo adds

`reference/standard-service` establishes the canonical layout. This demo adds
exactly one extension capability: a WebSocket server (`x/websocket`). Everything
else remains identical to the standard service shape.

```
main.go
internal/
  config/   config.go
  app/      app.go, routes.go
```

---

## `app.go` — one additional dependency

`app.New` constructs the `x/websocket.Server` alongside the stable-root
dependencies. The WebSocket server is a field on `App`, not a global:

```go
type App struct {
    Core *core.App
    Cfg  config.Config
    WS   *websocket.Server   // explicit, caller-owned
}
```

This keeps the WebSocket lifecycle visible. `Start(ctx)` shuts down
`a.WS.Shutdown(ctx)` before `a.Core.Shutdown(ctx)` when the caller-owned context
is canceled.

**What not to do:** Do not store the WebSocket server in a package-level
variable. Do not call `websocket.Enable()` or any self-registering helper.

---

## `routes.go` — upgrade is a route handler

The WebSocket upgrade is registered as a normal route via `ws.RegisterRoutes`.
This keeps the upgrade visible in the route list, not hidden inside the
framework:

```go
if err := a.WS.RegisterRoutes(a.Core); err != nil {
    return err
}
```

The route is explicit. It can be inspected, replaced, or wrapped with middleware
the same way any other route can be.

---

## Shutdown sequence

```
main.run signal context cancel
  → a.WS.Shutdown(ctx)    — drain active WebSocket connections
  → a.Core.Shutdown(ctx)  — drain in-flight HTTP requests
```

Shutdown order matters: WebSocket connections must be closed before the HTTP
server stops accepting new connections.

---

## What this demo does not demonstrate

- Room-based broadcast (see `x/websocket` module docs for room API)
- JWT-authenticated upgrades (`wsCfg.AllowUnauthenticated = true` here for simplicity)
- Capacity limits per hub or room
- Custom connection lifecycle hooks

For production use, set `AllowUnauthenticated = false` and wire an auth
middleware or token validator at the upgrade route.
