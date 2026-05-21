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
    Create(name string) Item
    Get(id string) (Item, bool)
}

type ItemHandler struct {
    Repo ItemRepository
}

// routes.go wires the concrete implementation
items := handler.ItemHandler{Repo: item.NewMemoryStore()}
```

Handlers with no external dependencies use a zero-field struct (`APIHandler{}`).
Both shapes are valid; choose based on whether the handler needs injected state.

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
