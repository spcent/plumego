# Application Creation and Configuration

> **Package**: `github.com/spcent/plumego/core`

Canonical usage guide for creating and configuring `core.App` in v1.

---

## Create an App

```go
app := core.New()
```

With options:

```go
app := core.New(
    core.WithAddr(":8080"),
    core.WithDebug(),
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

For `HEAD`, `OPTIONS`, route groups, reverse routing, and metadata, use the router directly:

```go
api := app.Router().Group("/api/v1")
api.Get("/health", http.HandlerFunc(health))
api.Options("/health", http.HandlerFunc(optionsHealth))
```

---

## Register Middleware

Register middleware explicitly before `Boot()`:

```go
if err := app.Use(
    observability.RequestID(),
    observability.Logging(app.Logger(), nil, nil),
    recovery.RecoveryMiddleware,
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

At runtime (before boot freeze):

```go
_ = app.Register(runnerB)
_ = app.OnShutdown(func(ctx context.Context) error { return nil })
```

---

## Boot and Serve

`app.Boot()` builds the middleware/router handler, starts the HTTP server, and blocks until shutdown.

```go
if err := app.Boot(); err != nil {
    log.Fatal(err)
}
```

`core.App` also implements `http.Handler`, so you can run it under your own server:

```go
app := core.New(core.WithAddr(":8080"))
app.Get("/health", healthHandler)

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
    observability.Logging(app.Logger(), nil, nil),
    recovery.RecoveryMiddleware,
); err != nil {
    log.Fatalf("register middleware: %v", err)
}

app.Get("/health", func(w http.ResponseWriter, r *http.Request) {
    contract.WriteResponse(w, r, http.StatusOK, map[string]string{"status": "ok"})
})

if err := app.Boot(); err != nil {
    log.Fatal(err)
}
```

---

## API Notes

- `WithDebug` has no boolean parameter.
- `WithServerTimeouts` parameters are `(read, readHeader, write, idle)`.
- Middleware configuration is done via `app.Use(...)`, not `With*` middleware options.
- `core.App` does not expose public `Shutdown(ctx)`, `Addr()`, or `Server()` methods in v1 API.

For the authoritative option list, see [configuration-options.md](configuration-options.md).
