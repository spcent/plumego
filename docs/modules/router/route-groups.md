# Route Groups

> **Package**: `github.com/spcent/plumego/router`

Route groups let you share path prefixes and middleware stacks across related routes.

---

## Canonical Usage

With `core.App`, create groups from the underlying router:

```go
app := core.New(core.WithAddr(":8080"))
r := app.Router()

api := r.Group("/api")
v1 := api.Group("/v1")

v1.Get("/health", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("ok"))
}))
```

With standalone router:

```go
r := router.NewRouter()
api := r.Group("/api")
api.Get("/ping", http.HandlerFunc(pingHandler))
```

Compile-oriented complete example:

```go
package main

import (
    "context"
    "log"
    "net/http"

    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/router"
)

func main() {
    ctx := context.Background()
    app := core.New(core.WithAddr(":8080"))
    r := app.Router()

    api := r.Group("/api")
    v1 := api.Group("/v1")
    v1.Get("/users/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
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

## Prefix Composition

`Group("/api")` + `Group("/v1")` + `Get("/users")` becomes `/api/v1/users`.

Rules:

- Group prefixes are normalized.
- Nested groups inherit parent prefix.
- Empty route path on a group maps to group root.

---

## Group Middleware

Attach middleware at router or group scope using `Use`.

```go
r := app.Router()

r.Use(observability.RequestID())

api := r.Group("/api")
api.Use(authMiddleware)

admin := api.Group("/admin")
admin.Use(adminOnlyMiddleware)

admin.Get("/stats", http.HandlerFunc(statsHandler))
```

Execution is outer-to-inner, then handler, then unwind.

---

## Parameters in Groups

```go
tenant := app.Router().Group("/tenants/:tenant_id")
tenant.Get("/users/:id", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    tenantID := router.Param(r, "tenant_id")
    userID := router.Param(r, "id")
    _, _ = w.Write([]byte(tenantID + ":" + userID))
}))
```

---

## Reverse Routing with Grouped Paths

Use named routes and `URL(...)`:

```go
r := app.Router()
v1 := r.Group("/api/v1")

_ = v1.AddRouteWithOptions(http.MethodGet, "/users/:id", http.HandlerFunc(userHandler),
    router.WithRouteName("users.show"),
)

url := r.URL("users.show", "id", "42")
// /api/v1/users/42
_ = url
```

---

## Practical Patterns

### Versioned APIs

```go
r := app.Router()
v1 := r.Group("/api/v1")
v2 := r.Group("/api/v2")

v1.Get("/users", http.HandlerFunc(usersV1))
v2.Get("/users", http.HandlerFunc(usersV2))
```

### Public vs Protected

```go
r := app.Router()

public := r.Group("/api/public")
public.Get("/info", http.HandlerFunc(infoHandler))

private := r.Group("/api")
private.Use(authMiddleware)
private.Get("/profile", http.HandlerFunc(profileHandler))
```

---

## Testing Group Behavior

Use `httptest` and assert:

- composed path matching
- middleware order across parent/child groups
- parameter extraction from grouped routes
- reverse route generation on grouped names

Relevant tests in repository include reverse-routing and group boundary coverage.

---

## Notes

- Group creation API is `app.Router().Group(...)` (or `router.NewRouter().Group(...)`).
- Keep middleware transport-layer only; no business/service injection in middleware.
