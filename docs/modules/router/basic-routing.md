# Basic Routing

> **Package**: `github.com/spcent/plumego/router`

This page covers canonical route registration for Plumego v1.

## HTTP Methods

`core.App` convenience methods:

```go
app.Get(path, handlerFunc)    // GET
app.Post(path, handlerFunc)   // POST
app.Put(path, handlerFunc)    // PUT
app.Patch(path, handlerFunc)  // PATCH
app.Delete(path, handlerFunc) // DELETE
app.Any(path, handlerFunc)    // ANY
```

For `HEAD` and `OPTIONS`, register through `app.Router()`:

```go
r := app.Router()
r.Head("/users/:id", http.HandlerFunc(headUser))
r.Options("/users", http.HandlerFunc(optionsUsers))
```

## Handler Types

### Standard handler (preferred)

```go
func(w http.ResponseWriter, r *http.Request)
```

### `http.Handler`

```go
type Handler interface {
    ServeHTTP(w http.ResponseWriter, r *http.Request)
}
```

### `contract.Ctx` adapter (explicit)

```go
app.Post("/users", contract.AdaptCtxHandler(func(ctx *contract.Ctx) {
    _ = ctx.Response(http.StatusCreated, map[string]any{"ok": true}, nil)
}, app.Logger()).ServeHTTP)
```

## Route Registration

```go
app := core.New(core.WithAddr(":8080"))

app.Get("/users", listUsers)
app.Post("/users", createUser)
app.Get("/users/:id", getUser)
```

Compile-oriented complete example:

```go
package main

import (
    "context"
    "log"
    "net/http"

    "github.com/spcent/plumego/core"
)

func main() {
    ctx := context.Background()
    app := core.New(core.WithAddr(":8080"))

    app.Get("/users", func(w http.ResponseWriter, r *http.Request) {
        _, _ = w.Write([]byte("list users"))
    })
    app.Post("/users", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusCreated)
    })
    app.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
        _, _ = w.Write([]byte("get user"))
    })

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

Use one method+path registration per line to keep behavior explicit.

## Groups and Middleware

```go
r := app.Router()
api := r.Group("/api")
api.Use(observability.RequestID())
api.Use(auth.SimpleAuth(os.Getenv("AUTH_TOKEN")))

api.Get("/users/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
    id, _ := contract.Param(req, "id")
    _, _ = w.Write([]byte(id))
}))
```

Execution order: parent `Use` -> child group `Use` -> route handler.

## Static Files

Register static serving through router helpers:

```go
r := app.Router()
r.Static("/assets", "./public")
r.StaticFS("/docs", http.FS(embeddedDocsFS))
```

## Reverse Routing

```go
r := app.Router()

if err := r.AddRouteWithOptions(router.GET, "/users/:id", http.HandlerFunc(getUser),
    router.WithRouteName("users.show"),
); err != nil {
    log.Fatal(err)
}

u := r.URL("users.show", "id", "42")
_ = u // /users/42
```

## Best Practices

- Register routes/middleware before `app.Prepare()`.
- Prefer explicit route names for endpoints reused by links/redirects.
- Keep handlers in `net/http` shape and adapt `contract.Ctx` only when needed.
