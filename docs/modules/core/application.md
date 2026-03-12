# Application Creation and Configuration

> **Package**: `github.com/spcent/plumego/core`

Canonical usage guide for creating and configuring `core.App` in v1.

---

## Create an App

```go
app := core.New(core.WithLogger(plumelog.NewGLogger()))
```

With options:

```go
app := core.New(
    core.WithAddr(":8080"),
    core.WithLogger(plumelog.NewGLogger()),
    core.WithDebug(),
    core.WithDevTools(),
    core.WithServerTimeouts(30*time.Second, 5*time.Second, 30*time.Second, 60*time.Second),
)
```

---

## Register Routes

`core.App` uses the standard library handler shape.

```go
app.Get("/users", listUsers)
app.Post("/users", createUser)
app.Put("/users/:id", updateUser)
app.Delete("/users/:id", deleteUser)
app.Patch("/users/:id", patchUser)
app.Any("/webhook", handleWebhook)
```

If you need explicit error handling for duplicate/frozen route registration:

```go
if err := app.AddRoute(http.MethodGet, "/users/:id", http.HandlerFunc(getUser)); err != nil {
    log.Fatalf("add route: %v", err)
}
if err := app.AddRouteWithName(http.MethodGet, "/users/:id", "users.show", http.HandlerFunc(getUser)); err != nil {
    log.Fatalf("add named route: %v", err)
}
```

For `HEAD`, `OPTIONS`, route groups, reverse routing, and metadata, use the router directly:

```go
api := app.Router().Group("/api/v1")
api.Get("/health", http.HandlerFunc(health))
api.Options("/health", http.HandlerFunc(optionsHealth))
```

---

## Register Middleware

Register middleware explicitly before `Prepare()`:

```go
if err := app.Use(
    observability.RequestID(),
    observability.Tracing(nil),
    observability.HTTPMetrics(nil),
    observability.AccessLog(app.Logger()),
    recovery.Recovery(app.Logger()),
); err != nil {
    log.Fatalf("register middleware: %v", err)
}
```

Group-level middleware goes on `app.Router().Group(...)`:

```go
api := app.Router().Group("/api")
api.Use(authMiddleware)
api.Get("/profile", http.HandlerFunc(profileHandler))
```

---

## Components, Runners, Shutdown Hooks

At construction time:

```go
app := core.New(
    core.WithComponent(componentA),
    core.WithRunner(runnerA),
    core.WithShutdownHook(func(ctx context.Context) error {
        return cleanup()
    }),
)
```

At runtime (before prepare freeze):

```go
_ = app.Register(runnerB)
_ = app.OnShutdown(func(ctx context.Context) error { return nil })
```

---

## Prepare, Start, and Serve

Canonical flow is explicit:

```go
ctx := context.Background()

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

`app.Run(ctx)` is still available as a convenience wrapper, but the explicit lifecycle above is the canonical path.

`core.App` also implements `http.Handler`, so you can run it under your own server once prepared:

```go
app := core.New(core.WithAddr(":8080"))
app.Get("/health", healthHandler)

if err := app.Prepare(); err != nil {
    log.Fatal(err)
}
if err := app.Start(context.Background()); err != nil {
    log.Fatal(err)
}
defer app.Shutdown(context.Background())

log.Fatal(http.ListenAndServe(":8080", app))
```

---

## Production Baseline Example

```go
app := core.New(
    core.WithAddr(":8080"),
    core.WithServerTimeouts(30*time.Second, 5*time.Second, 30*time.Second, 60*time.Second),
    core.WithMaxHeaderBytes(1<<20),
    core.WithMethodNotAllowed(true),
)

if err := app.Use(
    observability.RequestID(),
    observability.Tracing(nil),
    observability.HTTPMetrics(nil),
    observability.AccessLog(app.Logger()),
    recovery.Recovery(app.Logger()),
); err != nil {
    log.Fatalf("register middleware: %v", err)
}

app.Get("/health", func(w http.ResponseWriter, r *http.Request) {
    _ = contract.WriteResponse(w, r, http.StatusOK, map[string]string{"status": "ok"}, nil)
})

ctx := context.Background()
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

---

## API Notes

- `WithDebug` has no boolean parameter.
- `WithDebug` no longer auto-mounts devtools.
- Use `WithDevTools` to expose debug routes.
- `WithServerTimeouts` parameters are `(read, readHeader, write, idle)`.
- Middleware configuration is done via `app.Use(...)`, not `With*` middleware options.
- `core.App` now exposes public `Shutdown(ctx)` and `Server()` for explicit lifecycle control.

For the authoritative option list, see [configuration-options.md](configuration-options.md).
