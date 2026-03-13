# Router

## Overview

The `router` package provides a trie-based HTTP router built on `net/http`. Unlike `http.ServeMux`, it supports path parameters (`:id`), wildcard routes (`*path`), per-method dispatch, route groups with shared prefixes, middleware scoping, and named routes with reverse URL generation. It implements `http.Handler` so it plugs into any Go HTTP server.

## Quick Example

```go
package main

import (
	"net/http"
	"github.com/spcent/plumego/router"
	"github.com/spcent/plumego/contract"
)

func main() {
	r := router.NewRouter()
	r.Get("/users/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		id, _ := contract.Param(req, "id")
		w.Write([]byte("user " + id))
	}))
	http.ListenAndServe(":8080", r)
}
```

## API Reference

### Route Registration

Each method accepts a path and `http.Handler`, returns `error`. `*Func` variants accept `http.HandlerFunc`; `*Ctx` variants accept `contract.CtxHandlerFunc`.

| Method | HTTP Verb | `*Func` | `*Ctx` |
|--------|-----------|---------|--------|
| `Get` | GET | `GetFunc` | `GetCtx` |
| `Post` | POST | `PostFunc` | `PostCtx` |
| `Put` | PUT | `PutFunc` | `PutCtx` |
| `Delete` | DELETE | `DeleteFunc` | `DeleteCtx` |
| `Patch` | PATCH | `PatchFunc` | `PatchCtx` |
| `Any` | all | `AnyFunc` | `AnyCtx` |

Additional: `Options`, `Head`, `Handle(method, path, handler)`, `HandleFunc(method, path, handlerFunc)`.

### Path Parameters and Wildcards

- **Parameters**: `:name` captures a single segment -- `/users/:id` matches `/users/42`.
- **Wildcards**: `*name` captures the rest of the path -- `/files/*path` matches `/files/a/b.txt`.

Access values with `contract.Param(req, "id")` or `contract.RequestContextFrom(ctx).Params`.

### Route Groups

`Group(prefix)` returns a sub-router sharing the prefix. Middleware via `Use` applies only to that group.

```go
api := r.Group("/api/v1")
api.Use(authMiddleware)
api.Get("/users", listHandler)     // GET /api/v1/users
api.Get("/users/:id", showHandler) // GET /api/v1/users/:id
```

### Named Routes

```go
r.AddRouteWithOptions("GET", "/users/:id", h, router.WithRouteName("user.show"))
r.URL("user.show", "id", "42") // "/users/42"
```

Metadata options: `WithRouteName`, `WithRouteTags`, `WithRouteDescription`, `WithRouteDeprecated`.

### Static Files

```go
r.Static("/static", "./public")            // serve directory
r.StaticFS("/assets", http.FS(embeddedFS)) // embedded FS
r.StaticSPA("/", "./dist", "index.html")   // SPA fallback
```

## Usage Patterns

### RESTful Resource

`Resource(path, controller)` registers CRUD routes from a `ResourceController`.

```go
r.Resource("/posts", &PostController{})
// GET /posts, POST /posts, GET/PUT/DELETE/PATCH /posts/:id, OPTIONS, HEAD
```

### Modular Registration

Implement `RouteRegistrar` to organize routes by domain.

```go
type OrderRoutes struct{}
func (o *OrderRoutes) Register(r *router.Router) { r.Get("/orders", o.list) }
r.Register(&OrderRoutes{})
```

`NewRouter(opts...)` accepts `WithLogger` and `WithMethodNotAllowed`. Call `Freeze()` to reject late registrations, `Routes()` to inspect all routes.