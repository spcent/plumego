# Lifecycle

> **Package**: `github.com/spcent/plumego/core`

This page documents the actual v1 lifecycle behavior of `core.App`.

---

## Lifecycle Overview

`core.App` lifecycle is:

1. `core.New(options...)`
2. Register middleware/routes/components/runners/hooks (mutable phase)
3. `app.Prepare()`
4. `app.Start(ctx)`
5. Serve via `app.Server()` or `app.Run(ctx)`
6. `app.Shutdown(ctx)`

During `Prepare()`, configuration is frozen and further mutation is rejected.

---

## What `Prepare()` and `Start()` Do

`Prepare()` runs the setup sequence in order:

1. Mount explicit components (`RegisterMiddleware`, `RegisterRoutes`)
2. Freeze router
3. Build HTTP handler from router + middleware registry
4. Construct `*http.Server`

`Start(ctx)` then:

1. Starts logger lifecycle hooks (if implemented)
2. Starts components
3. Starts runners
4. Marks readiness ready

If a start step fails, already-started runners/components are stopped.

---

## Graceful Shutdown Behavior

When `Shutdown(ctx)` is called:

1. readiness is marked not-ready
2. server shutdown runs with configured timeout (`WithShutdownTimeout`)
3. active connection drain tracking is used during shutdown
4. runners stop in reverse start order
5. components stop in reverse start order
6. shutdown hooks run in reverse registration order

Timeout fallback is `5s` when shutdown timeout is not positive.

---

## Signals

`Run(ctx)` handles:

- `SIGINT`
- `SIGTERM`

If you use `Prepare()` / `Start()` / `Server()` directly, signal handling is your responsibility.

---

## Important API Notes

- `WithEnvPath` no longer loads `.env` automatically.
- `WithDebug` no longer auto-mounts devtools.
- `x/devtools` must be mounted explicitly via `app.MountComponent(...)`.

Manual server lifecycle control is now a first-class path:

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

srv, err := app.Server()
if err != nil {
    log.Fatal(err)
}

log.Fatal(srv.ListenAndServe())
```

---

## Lifecycle Hooks

At construction:

```go
app := core.New(
    core.WithRunner(runnerA),
    core.WithComponent(componentA),
    core.WithShutdownHook(func(ctx context.Context) error {
        return cleanup()
    }),
)
```

Before prepare freeze (runtime registration):

```go
_ = app.Register(runnerB)
_ = app.OnShutdown(func(ctx context.Context) error { return nil })
```

---

## Testing Lifecycle Logic

For most cases, prefer `httptest` + `app.ServeHTTP` instead of full lifecycle startup.  
Use `Prepare()` / `Start()` / `Shutdown()` tests for lifecycle-specific behavior.

Quality gates:

```bash
go test -timeout 20s ./...
go test -race -timeout 60s ./...
go vet ./...
```
