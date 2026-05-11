# Architecture Notes — with-gateway

This document explains the structural choices in this feature demo. It
extends `reference/standard-service` with `x/gateway` and follows the
same layout discipline.

---

## What this demo adds

`reference/standard-service` establishes the canonical layout. This demo adds
exactly one extension capability: a reverse-proxy gateway (`x/gateway`). Everything
else remains identical to the standard service shape.

```
main.go
internal/
  config/   config.go
  app/      app.go, routes.go
```

---

## `app.go` — one additional dependency

`app.New` constructs the `x/gateway.GatewayProxy` alongside the stable-root
dependencies. The proxy is a field on `App`, not a global:

```go
type App struct {
    Core  *core.App
    Cfg   config.Config
    Proxy *gateway.GatewayProxy   // explicit, caller-owned
}
```

The proxy is created with an explicit target list from config:

```go
proxy, err := gateway.NewGateway(gateway.GatewayConfig{
    Targets: []string{cfg.GatewayBackend},
})
```

Backend selection is explicit configuration, not service discovery. Discovery
mechanisms (DNS SRV, Consul, etcd) are decoupled from the proxy constructor
and belong outside this demo.

---

## `routes.go` — proxy is a route handler

The proxy is registered as a wildcard route handler. This keeps the forwarding
boundary visible in the route list:

```go
if err := a.Core.Get("/proxy/*", http.HandlerFunc(a.Proxy.ServeHTTP)); err != nil {
    return err
}
```

The wildcard suffix (`*`) captures the full downstream path. Stripping or
rewriting the path prefix is a configuration concern, not a framework concern.

---

## Shutdown sequence

```
SIGTERM / context cancel
  → a.Core.Shutdown(ctx)   — drain in-flight HTTP and proxy requests
```

`x/gateway` does not require a separate shutdown call. In-flight proxy requests
are drained when the HTTP server stops accepting new connections.

---

## What this demo does not demonstrate

- Load balancing across multiple backends (pass multiple `Targets`)
- Circuit breaking or retry policies
- Request/response header rewriting
- Backend health checking and automatic failover
- TLS verification of upstream connections (enabled by default in production)

For production use, configure `GatewayConfig` with timeout, TLS, and retry
settings appropriate to your upstream services.
