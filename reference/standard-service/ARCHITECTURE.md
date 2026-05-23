# Architecture Notes — standard-service

This document explains the directory structure choices in this reference
application. It supplements the README rather than repeating it.

---

## Why this layout

```
main.go
internal/
  config/   config.go
  app/      app.go, routes.go
  domain/
    item/   item.go
  handler/  api.go, health.go, items.go, handler_test.go
```

This is the canonical Plumego application shape. Every scaffold and getting-
started guide produces this structure. Deviating from it without an explicit
reason creates drift that reviewers and agents have to recover from.

---

## `main.go` — startup only

`main.go` does exactly four things:

1. Load configuration (`config.Load`)
2. Construct the application (`app.New`)
3. Register routes (`app.RegisterRoutes`)
4. Start the server with a signal-aware context (`app.Start(ctx)`)

It contains no business logic, no handler code, and no wiring decisions. If
`main.go` grows beyond those four calls, the growth belongs somewhere else.

---

## `internal/config/` — app-local configuration

Configuration lives in its own package to keep `main.go` small and to make
config loading testable in isolation.

`Defaults()` returns safe values for local development. `Load()` layers
environment variables and flags over those defaults. `Validate()` rejects
unusable configurations before the server starts.

**What does not belong here:** framework-level config (e.g., server timeouts,
TLS) is assembled from `core.DefaultConfig` and passed into `core.New`. App-
local config and kernel config are separate structs for this reason.

---

## `internal/app/app.go` — wiring

`app.New` constructs dependencies and attaches middleware in the order they
run. Reading `app.go` answers:

- Which dependencies does this service have?
- What middleware runs on every request?
- In what order?

Middleware order in `app.Use(...)` is the authoritative declaration. The order
is visible and grep-friendly. No middleware is added by the framework without
the caller choosing it.

**What does not belong here:** handler instantiation, route paths, business
logic, database connections. Those either live in `routes.go`, `handler/`, or
are passed in as dependencies.

---

## `internal/app/routes.go` — route map

`RegisterRoutes` is the single file that declares every public HTTP endpoint.
Reading it answers:

- What paths does this service expose?
- Which handler owns each path?
- What middleware runs on specific route groups?

All route registration is explicit. There is no controller scanning, no
annotation routing, and no framework-managed discovery.

**What does not belong here:** handler implementation. A route file contains
registrations, not logic.

---

## `internal/handler/` — transport handlers

Handlers convert HTTP requests to domain responses. They:

- Read request parameters (`r.URL.Query()`, `json.NewDecoder(r.Body).Decode`)
- Call domain logic (passed in through the handler struct's fields)
- Write responses using `contract.WriteResponse` or `contract.WriteError`

They do not own middleware concerns (logging, timeouts, rate limits), and they
do not own persistence concerns (database queries, cache lookups).

Tests live next to the handlers they test (`handler_test.go`).

`internal/domain/item` owns the sample item model and in-memory repository. This
keeps the transport package focused on HTTP adaptation while still showing where
real business and persistence code belongs in an application.

### Constructor injection

Handlers that require external dependencies declare them as interface fields and
receive concrete implementations from `routes.go`. This keeps handlers
independently testable: tests pass a stub; production passes the real store.

```go
// handler declares the interface it needs
type ItemRepository interface {
    Create(ctx context.Context, name string) item.Item
    Get(ctx context.Context, id string) (item.Item, bool)
    List(ctx context.Context) []item.Item
    Delete(ctx context.Context, id string) bool
}

type ItemHandler struct {
    Repo ItemRepository
}

// routes.go wires the concrete implementation
items := handler.ItemHandler{Repo: item.NewMemoryStore()}
```

Handlers with no external dependencies use a zero-field struct (`APIHandler{}`).
Both shapes are valid; choose based on whether the handler needs injected state.

### Readiness checking

`HealthHandler` supports an optional list of `health.ComponentChecker` implementations.
Each checker represents one dependency (database, cache, downstream service).
`GET /readyz` probes them in order; the first failure returns 503 TypeUnavailable.

```go
// Implement health.ComponentChecker for your dependency.
// Name() is used to label the component in success and error responses.
// Check(ctx) should return nil if healthy, error if unhealthy.
type dbChecker struct{ db *sql.DB }

func (c *dbChecker) Name() string              { return "database" }
func (c *dbChecker) Check(ctx context.Context) error { return c.db.PingContext(ctx) }

// routes.go registers one checker per dependency
health := handler.HealthHandler{
    ServiceName: a.Cfg.App.ServiceName,
    Checkers: []health.ComponentChecker{
        &dbChecker{db: myDB},
    },
}
```

The reference wires no checkers because it has no real dependencies. When no
checkers are registered `GET /readyz` returns 200 immediately, which is correct
for a stateless service.

### Route layout — groups and collection + member pairs

Routes that share a common path prefix are registered through a `RouteGroup` so
the prefix is declared once and never repeated:

```go
v1 := a.Core.Group("/api/v1")
v1.Get("/greet",        http.HandlerFunc(api.Greet))
v1.Get("/items",        http.HandlerFunc(items.List))
v1.Post("/items",       http.HandlerFunc(items.Create))
v1.Get("/items/:id",    http.HandlerFunc(items.GetByID))
v1.Put("/items/:id",    writeGuard(http.HandlerFunc(items.Update)))
v1.Delete("/items/:id", http.HandlerFunc(items.Delete))
```

`RouteGroup` is a `core.App` concept: it enforces the same lifecycle and
nil-handler rules as the top-level `Get/Post/Delete` methods. Groups can be
nested (`api := app.Group("/api"); v1 := api.Group("/v1")`).

REST resources follow a consistent two-path pattern:

```
GET    /api/v1/items         → list all items
POST   /api/v1/items         → create an item

GET    /api/v1/items/:id     → fetch one item
PUT    /api/v1/items/:id     → replace an item (idempotent full replacement)
DELETE /api/v1/items/:id     → remove one item
```

The same collection path and member path carry different verbs; the group
prefix keeps the registration readable without any controller scanning or
annotation-based magic.

**PUT vs PATCH**: This reference uses PUT (full replacement). PATCH (partial
update) is semantically more complex: it requires distinguishing "field not
provided" from "field set to null/zero", which typically demands a custom
merge strategy or JSON Merge Patch (RFC 7396). Use PUT when the client owns
the full resource representation; reach for PATCH only when partial updates
are a hard product requirement and the merge semantics are well-defined.

### Error accumulation in route registration (`routeReg`)

`routes.go` uses a small local helper (`routeReg`) to avoid per-call error
checks when registering routes. It wraps `routeAdder` and records only the
first error:

```go
v1 := newRouteReg(a.Core.Group("/api/v1"))
v1.get("/items",     http.HandlerFunc(items.List))   // error captured internally
v1.post("/items",    writeGuard(http.HandlerFunc(items.Create)))
v1.get("/items/:id", http.HandlerFunc(items.GetByID))
return v1.err                                         // single check at the end
```

This keeps the route table visually flat — one line per route — without
sacrificing error propagation. The pattern is correct because route
registration errors (duplicate path, nil handler) are always programming
mistakes; there is at most one at a time and the first one is the one to fix.

---

## Dependency direction

```
main.go
  → internal/config
  → internal/app          (imports internal/config, internal/handler)
      → internal/domain/item
      → internal/handler  (imports contract, router, internal/domain/item)
```

`internal/handler` has no upward imports. It does not import `internal/app`
or `internal/config`. This keeps handlers independently testable.

---

## What this service intentionally excludes

`reference/standard-service` depends only on stable root packages. It does
not import `x/*`. This is a design constraint, not an oversight.

If you need to add a capability from `x/*`, use the corresponding
`reference/with-<capability>` demo as the starting point instead of modifying
this reference.
