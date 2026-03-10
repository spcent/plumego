# Router Middleware Binding

> **Package**: `github.com/spcent/plumego/router`

This page defines how middleware binds at app/router/group/route scope and how order is resolved.

---

## Canonical Middleware Type

```go
type Middleware func(http.Handler) http.Handler
```

Both `core.App.Use(...)` and `router.Router.Use(...)` consume this type.

---

## Binding Scopes

### 1. App-Level (global)

```go
if err := app.Use(
    observability.RequestID(),
    observability.Logging(app.Logger(), nil, nil),
    recovery.RecoveryMiddleware,
); err != nil {
    log.Fatal(err)
}
```

Applies to all requests handled by the app.

### 2. Router-Level

```go
r := app.Router()
r.Use(authnMiddleware)
```

Applies to all routes in this router tree.

### 3. Group-Level

```go
api := app.Router().Group("/api")
api.Use(authzMiddleware)

admin := api.Group("/admin")
admin.Use(adminOnlyMiddleware)
```

Applies only to routes in that group (and child groups).

### 4. Route-Level

Wrap single handler explicitly:

```go
app.Get("/download", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    // ...
}))

app.Get("/download-secure", http.HandlerFunc(
    perRouteMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // ...
    })).ServeHTTP,
))
```

---

## Order Rules

For a route under nested groups:

1. app-level middleware (registration order)
2. router-level middleware (registration order)
3. parent group middleware (registration order)
4. child group middleware (registration order)
5. handler

On response unwind, reverse order applies.

---

## Example: Nested Stack

```go
app := core.New(core.WithAddr(":8080"))

_ = app.Use(observability.RequestID())

r := app.Router()
r.Use(observability.Logging(app.Logger(), nil, nil))

api := r.Group("/api")
api.Use(authnMiddleware)

v1 := api.Group("/v1")
v1.Use(rateLimitMiddleware)

v1.Get("/users", http.HandlerFunc(listUsers))
```

Effective chain for `/api/v1/users`:

`RequestID -> Logging -> authn -> rateLimit -> handler`

---

## Error and Panic Path

Recommended baseline includes panic recovery:

```go
_ = app.Use(
    observability.RequestID(),
    observability.Logging(app.Logger(), nil, nil),
    recovery.RecoveryMiddleware,
)
```

Recovery should be early enough to catch downstream panics.

---

## Common Mistakes

- Registering middleware after app boot starts.
- Mixing business logic into middleware.
- Relying on implicit order instead of explicit registration sequence.
- Applying heavy middleware globally when only subset routes need it.

---

## Validation Strategy

For router middleware changes, keep tests for:

- parent/child group order
- panic path behavior
- error response path
- route matching unaffected by middleware stacking
