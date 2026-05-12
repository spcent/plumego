# Architecture Notes — with-rest

This document explains the structural choices in this feature demo. It
extends `reference/standard-service` with `x/rest` resource controllers
and follows the same layout discipline.

---

## What this demo adds

`reference/standard-service` establishes the canonical layout. This demo adds
exactly one extension capability: REST resource controllers (`x/rest`).
Everything else remains identical to the standard service shape.

```
main.go
internal/
  config/       config.go
  domain/
    user/       user.go   — User model + Repository
  app/          app.go, routes.go
```

The `internal/domain/user` sub-package holds the domain model and the in-memory
repository. This keeps business types out of the HTTP layer.

---

## `app.go` — one additional dependency

`app.New` constructs the `x/rest.DBResourceController` alongside the stable-root
dependencies. The controller is a field on `App`, not a global:

```go
type App struct {
    Core  *core.App
    Cfg   config.Config
    Users *rest.DBResourceController[user.User]   // explicit, caller-owned
}
```

The controller is created with an explicit spec and repository:

```go
spec := rest.DefaultResourceSpec("users").WithPrefix("/api/users")
users := rest.NewDBResource[user.User](spec, user.NewRepository())
```

`rest.DefaultResourceSpec` chooses standard route shapes (collection + member).
The prefix is the only configuration required to mount the resource.

---

## `routes.go` — resource methods are route handlers

Each REST action is a regular `http.HandlerFunc` registered via `a.Core.*`:

```go
if err := a.Core.Get("/api/users",        http.HandlerFunc(a.Users.Index));  err != nil { ... }
if err := a.Core.Get("/api/users/:id",    http.HandlerFunc(a.Users.Show));   err != nil { ... }
if err := a.Core.Post("/api/users",       http.HandlerFunc(a.Users.Create)); err != nil { ... }
if err := a.Core.Put("/api/users/:id",    http.HandlerFunc(a.Users.Update)); err != nil { ... }
if err := a.Core.Delete("/api/users/:id", http.HandlerFunc(a.Users.Delete)); err != nil { ... }
```

Each handler is explicit and individually inspectable. There is no hidden
auto-wiring: the route surface is the route list.

---

## `internal/domain/user`

`user.Repository` is an in-memory store for demo purposes. It implements
`rest.Repository[User]` — the interface `x/rest` requires to drive
`DBResourceController`. In production, replace `user.Repository` with a
database-backed implementation of the same interface.

---

## Shutdown sequence

```
SIGTERM / context cancel
  → a.Core.Shutdown(ctx)   — drain in-flight HTTP requests
```

`x/rest` controllers hold no background state. Shutdown only requires draining
the HTTP server.

---

## What this demo does not demonstrate

- Pagination or query-parameter filtering (supported by `rest.QueryParams`)
- Input validation via `controller.WithValidator(...)`
- Custom error shapes beyond the default `contract.WriteError` output
- Authentication on write routes

For production use, add a validator and wrap write routes with an auth middleware.
