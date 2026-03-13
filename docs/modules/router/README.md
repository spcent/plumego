# Router Module

> **Package**: `github.com/spcent/plumego/router`  
> **Stability**: High (v1 GA)

The router module provides path matching, parameters, route groups, route metadata, and reverse routing.

---

## Features

- HTTP method routing (`GET/POST/PUT/PATCH/DELETE/HEAD/OPTIONS/ANY`)
- path parameters (`:id`) and wildcard parameters (`*path`)
- nested route groups with shared prefix/middleware
- middleware stacks per router/group
- route metadata and named routes
- reverse routing (`URL(...)`)
- optional 405 behavior (`SetMethodNotAllowed` / `WithMethodNotAllowed`)

---

## Quick Start with `core.App`

```go
ctx := context.Background()

app := core.New(core.WithAddr(":8080"))
r := app.Router()

r.Use(requestid.Middleware())

api := r.Group("/api/v1")
api.Get("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ok"))
}))

api.Get("/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    id := router.Param(r, "id")
    _, _ = w.Write([]byte("user=" + id))
}))

if err := app.Prepare(); err != nil {
    log.Fatal(err)
}
if err := app.Start(ctx); err != nil {
    log.Fatal(err)
}
srv, err := app.Server()
if err != nil {
    log.Fatal(err)
}
defer app.Shutdown(ctx)

log.Fatal(srv.ListenAndServe())
```

Compile-oriented complete example:

```go
package main

import (
    "context"
    "log"
    "net/http"

    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/middleware/requestid"
    "github.com/spcent/plumego/router"
)

func main() {
    ctx := context.Background()
    app := core.New(core.WithAddr(":8080"))
    r := app.Router()
    r.Use(requestid.Middleware())

    api := r.Group("/api/v1")
    api.Get("/users/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
        id := router.Param(req, "id")
        _, _ = w.Write([]byte("user=" + id))
    }))

    if err := app.Prepare(); err != nil {
        log.Fatal(err)
    }
    if err := app.Start(ctx); err != nil {
        log.Fatal(err)
    }
    srv, err := app.Server()
    if err != nil {
        log.Fatal(err)
    }
    defer app.Shutdown(ctx)

    log.Fatal(srv.ListenAndServe())
}
```

---

## Standalone Usage

```go
r := router.NewRouter(router.WithMethodNotAllowed(true))

r.Get("/ping", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("pong"))
}))

log.Fatal(http.ListenAndServe(":8080", r))
```

---

## Groups and Middleware

```go
r := app.Router()

api := r.Group("/api")
api.Use(authMiddleware)

admin := api.Group("/admin")
admin.Use(adminMiddleware)
admin.Get("/stats", http.HandlerFunc(statsHandler))
```

---

## Reverse Routing

```go
r := app.Router()

_ = r.AddRouteWithOptions(http.MethodGet, "/users/:id/files/*path", http.HandlerFunc(fileHandler),
    router.WithRouteName("users.file"),
)

u := r.URL("users.file", "id", "42", "path", "reports/q1.pdf")
// /users/42/files/reports/q1.pdf
_ = u
```

If a required param is missing, `URL(...)` returns an empty string.

---

## Metadata

Attach route metadata when registering routes:

```go
_ = r.AddRouteWithOptions(http.MethodGet, "/users/:id", http.HandlerFunc(userHandler),
    router.WithRouteName("users.show"),
)
```

Use `Routes()`, `NamedRoutes()`, and `Print(...)` for introspection/debug tooling.

---

## Testing Guidance

For router changes, include tests for:

- static/param/wildcard matching
- group prefix composition
- middleware order across nested groups
- reverse routing correctness and edge cases
- method-not-allowed behavior (`405 + Allow`)
