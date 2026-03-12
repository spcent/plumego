# Configuration Options Reference

> **Package**: `github.com/spcent/plumego/core`
>
> This page is the canonical `core.With*` option list for v1 and reflects `core/options.go`.

---

## Table of Contents

- [Server and Runtime Options](#server-and-runtime-options)
- [TLS Options](#tls-options)
- [Application Behavior Options](#application-behavior-options)
- [Lifecycle Extension Options](#lifecycle-extension-options)
- [Observability Hook Options](#observability-hook-options)
- [Middleware Registration (Canonical)](#middleware-registration-canonical)
- [Common Composition Patterns](#common-composition-patterns)

---

## Server and Runtime Options

### `WithAddr`

Set server listen address.

```go
func WithAddr(address string) Option
```

```go
app := core.New(
    core.WithAddr(":8080"),
)
```

Default: `:8080`

---

### `WithEnvPath`

Set the `.env` file path loaded during boot.

```go
func WithEnvPath(path string) Option
```

```go
app := core.New(
    core.WithEnvPath(".env.prod"),
)
```

Default: `.env`

---

### `WithShutdownTimeout`

Set graceful shutdown timeout.

```go
func WithShutdownTimeout(timeout time.Duration) Option
```

```go
app := core.New(
    core.WithShutdownTimeout(10*time.Second),
)
```

Default: `5s`

---

### `WithServerTimeouts`

Set HTTP server read/read-header/write/idle timeouts.

```go
func WithServerTimeouts(read, readHeader, write, idle time.Duration) Option
```

```go
app := core.New(
    core.WithServerTimeouts(
        30*time.Second, // read
        5*time.Second,  // read header
        30*time.Second, // write
        60*time.Second, // idle
    ),
)
```

Defaults:
- read: `30s`
- read header: `5s`
- write: `30s`
- idle: `60s`

---

### `WithMaxHeaderBytes`

Set max request header size accepted by the HTTP server.

```go
func WithMaxHeaderBytes(bytes int) Option
```

```go
app := core.New(
    core.WithMaxHeaderBytes(1<<20), // 1 MiB
)
```

Default: `1<<20` (1 MiB)

---

### `WithHTTP2`

Enable or disable HTTP/2 support when TLS is enabled.

```go
func WithHTTP2(enabled bool) Option
```

```go
app := core.New(
    core.WithHTTP2(false),
)
```

Default: `true`

---

### `WithRouter`

Use a custom router instance.

```go
func WithRouter(router *router.Router) Option
```

```go
r := router.NewRouter()

app := core.New(
    core.WithRouter(r),
)
```

---

## TLS Options

### `WithTLS`

Enable TLS using cert/key file paths.

```go
func WithTLS(certFile, keyFile string) Option
```

```go
app := core.New(
    core.WithTLS("./certs/server.crt", "./certs/server.key"),
)
```

Equivalent to setting `WithTLSConfig(TLSConfig{Enabled: true, ...})`.

---

### `WithTLSConfig`

Set full TLS configuration.

```go
func WithTLSConfig(tlsConfig TLSConfig) Option
```

```go
app := core.New(
    core.WithTLSConfig(core.TLSConfig{
        Enabled:  true,
        CertFile: "./certs/server.crt",
        KeyFile:  "./certs/server.key",
    }),
)
```

---

## Application Behavior Options

### `WithDebug`

Enable debug mode.

```go
func WithDebug() Option
```

```go
app := core.New(
    core.WithDebug(),
)
```

Notes:
- `WithDebug` has no boolean parameter in v1.
- Debug mode does not auto-mount devtools.

---

### `WithDevTools`

Explicitly mount the devtools component.

```go
func WithDevTools() Option
```

```go
app := core.New(
    core.WithDebug(),
    core.WithDevTools(),
)
```

---

### `WithLogger`

Set a custom structured logger.

`core.New(...)` uses `log.NewNoOpLogger()` by default. Use `WithLogger(...)` whenever you expect `app.Logger()` to emit logs.

```go
func WithLogger(logger log.StructuredLogger) Option
```

```go
app := core.New(
    core.WithLogger(myLogger),
)
```

---

### `WithMethodNotAllowed`

Control 405 responses when path matches but HTTP method does not.

```go
func WithMethodNotAllowed(enabled bool) Option
```

```go
app := core.New(
    core.WithMethodNotAllowed(true),
)
```

---

## Lifecycle Extension Options

### `WithComponent` / `WithComponents`

Register lifecycle components.

```go
func WithComponent(component Component) Option
func WithComponents(components ...Component) Option
```

```go
app := core.New(
    core.WithComponents(componentA, componentB),
)
```

---

### `WithRunner` / `WithRunners`

Register background runners.

```go
func WithRunner(runner Runner) Option
func WithRunners(runners ...Runner) Option
```

```go
app := core.New(
    core.WithRunners(runnerA, runnerB),
)
```

---

### `WithShutdownHook` / `WithShutdownHooks`

Register graceful-shutdown hooks.

```go
func WithShutdownHook(hook ShutdownHook) Option
func WithShutdownHooks(hooks ...ShutdownHook) Option
```

```go
app := core.New(
    core.WithShutdownHook(func(ctx context.Context) error {
        return cleanup()
    }),
)
```

---

## Observability Hook Options

### `WithMetricsCollector`

Inject request metrics collector used by observability middleware.

```go
func WithMetricsCollector(collector metrics.MetricsCollector) Option
```

```go
app := core.New(
    core.WithMetricsCollector(metricsCollector),
)
```

---

### `WithTracer`

Inject tracer used by observability middleware.

```go
func WithTracer(tracer observability.Tracer) Option
```

```go
app := core.New(
    core.WithTracer(tracer),
)
```

---

### `WithHealthManager`

Attach a health manager so core lifecycle updates readiness automatically.

```go
func WithHealthManager(hm health.HealthManager) Option
```

```go
hm, err := health.NewHealthManager(health.HealthCheckConfig{})
if err != nil {
    log.Fatal(err)
}

app := core.New(
    core.WithHealthManager(hm),
)
```

---

## Middleware Registration (Canonical)

Middleware is not configured by `With*` options in v1. Register middleware explicitly before `Boot()`:

```go
app := core.New(core.WithAddr(":8080"))

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

This keeps middleware order explicit and testable.

---

## Common Composition Patterns

### Production baseline

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
```

### Conditional TLS

```go
opts := []core.Option{
    core.WithAddr(":8080"),
}

if enableTLS {
    opts = append(opts, core.WithTLS(certFile, keyFile))
}

app := core.New(opts...)
```

---

## v1 Compatibility Notes

These historical option names are **not** part of the current core v1 API and should not be used in docs/examples:

- `WithRecommendedMiddleware`
- `WithRequestID`
- `WithLogging`
- `WithRecovery`
- `WithCORS`
- `WithTenantConfigManager`
- `WithTenantMiddleware`
- `WithServer`
- `WithMiddlewareChain`

Use explicit middleware wiring (`app.Use(...)`) and tenant middleware composition on router groups instead.
