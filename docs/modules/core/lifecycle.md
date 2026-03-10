# Lifecycle

> **Package**: `github.com/spcent/plumego/core`

This page documents the actual v1 lifecycle behavior of `core.App`.

---

## Lifecycle Overview

`core.App` lifecycle is:

1. `core.New(options...)`
2. Register middleware/routes/components/runners/hooks (mutable phase)
3. `app.Boot()`
4. OS signal (`SIGINT`/`SIGTERM`) triggers graceful shutdown path

During boot, configuration is frozen and further mutation is rejected.

---

## What `Boot()` Does

`Boot()` runs the boot sequence in order:

1. Load `.env` file if configured (`WithEnvPath`)
2. Enable debug env flag when debug mode is on
3. Mount components (`RegisterMiddleware`, `RegisterRoutes`)
4. Freeze router
5. Build HTTP handler from router + middleware registry
6. Start components
7. Start runners
8. Start HTTP server (`ListenAndServe` / `ListenAndServeTLS`)
9. Wait for shutdown signal and perform graceful stop

If a start step fails, already-started runners/components are stopped.

---

## Graceful Shutdown Behavior

When a shutdown signal is received while `Boot()` is running:

1. readiness is marked not-ready
2. server shutdown runs with configured timeout (`WithShutdownTimeout`)
3. active connection drain tracking is used during shutdown
4. runners stop in reverse start order
5. components stop in reverse start order
6. shutdown hooks run in reverse registration order

Timeout fallback is `5s` when shutdown timeout is not positive.

---

## Signals

`Boot()` handles:

- `SIGINT`
- `SIGTERM`

No extra application code is required for the default signal path.

---

## Important API Notes

- There is no public `app.Shutdown(ctx)` method in current v1 API.
- There is no public `app.Server()` or `app.Addr()` getter in current v1 API.

If you need full manual server lifecycle control, run with the standard library server and `app` as `http.Handler`:

```go
app := core.New(core.WithAddr(":8080"))
app.Get("/health", healthHandler)

log.Fatal(http.ListenAndServe(":8080", app))
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

Before boot freeze (runtime registration):

```go
_ = app.Register(runnerB)
_ = app.OnShutdown(func(ctx context.Context) error { return nil })
```

---

## Testing Lifecycle Logic

For most cases, prefer `httptest` + `app.ServeHTTP` instead of `Boot()`.  
Use `Boot()` tests only for lifecycle-specific behavior (signals/components/runners).

Quality gates:

```bash
go test -timeout 20s ./...
go test -race -timeout 60s ./...
go vet ./...
```
