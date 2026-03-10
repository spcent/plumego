# Router module

The `router` package provides trie-based matching with groups, path parameters, wildcard routes, and reverse routing while staying compatible with standard `http.Handler`.

## Handler shape
Use standard handlers:

- `Get`, `Post`, `Put`, `Patch`, `Delete`, `Options`, `Head`, `Any`
- Signature: `http.Handler` (or `http.HandlerFunc`)

Read path params via `contract.Param(r, "id")`.

## Path patterns
- Static: `/docs/index.html`
- Parameters: `/users/:id`, `/teams/:teamID/members/:id`
- Wildcards: `/*filepath`

## Grouping and middleware order
Middleware order is explicit: `global router.Use(...) -> group.Use(...) -> route`.

```go
r := app.Router()
r.Use(observability.RequestID())

api := r.Group("/api")
api.Use(auth.SimpleAuth(os.Getenv("AUTH_TOKEN")))
api.Use(timeout.Timeout(2 * time.Second))

api.Get("/users/:id", http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
    id, _ := contract.Param(req, "id")
    _, _ = w.Write([]byte("user=" + id))
}))
```

Register routes before `app.Boot()`; boot freezes route registration.

## Reverse routing
Assign route names and generate URLs safely:

```go
if err := r.AddRouteWithOptions(router.GET, "/users/:id", http.HandlerFunc(showUser),
    router.WithRouteName("users.show"),
); err != nil {
    log.Fatal(err)
}

u, err := r.URL("users.show", map[string]string{"id": "42"})
if err != nil {
    log.Fatal(err)
}
fmt.Println(u.String()) // /users/42
```

## Method-not-allowed behavior
Enable 405 responses for method mismatch:

```go
app := core.New(core.WithMethodNotAllowed(true))
```

## Where to look in repo
- `router/registration.go`: route/group registration and validation.
- `router/dispatch.go`: match and dispatch path.
- `router/metadata.go`: route meta and reverse routing.
- `router/reverse_routing_group_test.go`: group + reverse-routing boundary tests.
