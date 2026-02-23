# Core Module

> **Package**: `github.com/spcent/plumego/core`
> **Stability**: High - Core API is stable
> **Go Version**: 1.24+

The `core` package is the heart of Plumego, providing application lifecycle management, dependency injection, component orchestration, and configuration options.

---

## Overview

The core module provides:

- **Application Lifecycle**: Explicit `New()` → configuration → `Boot()` → `Shutdown()` flow
- **Dependency Injection**: Built-in DI container for managing component dependencies
- **Component System**: Pluggable architecture for modules (WebSocket, Webhook, Observability, etc.)
- **Runner System**: Background service management with lifecycle coordination
- **Configuration**: Functional options pattern for clean, type-safe configuration
- **Graceful Shutdown**: Connection draining with configurable timeout

---

## Quick Start

```go
package main

import (
    "github.com/spcent/plumego/core"
    "log"
)

func main() {
    // Create application with functional options
    app := core.New(
        core.WithAddr(":8080"),
        core.WithDebug(),
        core.WithRecommendedMiddleware(), // RequestID + Logging + Recovery
    )

    // Define routes
    app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("pong"))
    })

    // Boot and block until shutdown signal
    if err := app.Boot(); err != nil {
        log.Fatal(err)
    }
}
```

---

## Core Concepts

### 1. Application (`*App`)

The `App` struct is your application instance. It orchestrates:
- HTTP server configuration
- Route registration
- Middleware chain
- Component lifecycle
- Graceful shutdown

**Creation**:
```go
app := core.New(options...)
```

### 2. Component Interface

Components are self-contained modules that can:
- Register routes
- Register middleware
- Start/stop background services
- Report health status
- Declare dependencies

```go
type Component interface {
    RegisterRoutes(r *router.Router)
    RegisterMiddleware(m *middleware.Registry)
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Health() (name string, status health.HealthStatus)
    Dependencies() []reflect.Type
}
```

### 3. Runner Interface

Runners are lightweight background services:
```go
type Runner interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

### 4. Dependency Injection

The DI container manages:
- Automatic dependency resolution
- Topological sorting (start order)
- Singleton lifecycle
- Type-safe registration

### 5. Functional Options

Configuration uses the functional options pattern:
```go
type Option func(*App)

func WithAddr(addr string) Option { ... }
```

---

## Documentation Structure

| Document | Description |
|----------|-------------|
| **[Application](application.md)** | App creation, configuration, routing |
| **[Lifecycle](lifecycle.md)** | Boot, shutdown, signal handling |
| **[Components](components.md)** | Component system, built-in components |
| **[Dependency Injection](dependency-injection.md)** | DI container usage |
| **[Configuration Options](configuration-options.md)** | All available `core.With*` options |
| **[Runners](runners.md)** | Background service management |
| **[Testing](testing.md)** | Testing utilities and patterns |

---

## Key Types

```go
// App is the main application instance
type App struct {
    server     *http.Server
    router     *router.Router
    middleware *middleware.Chain
    di         *di.Container
    components []Component
    runners    []Runner
    // ... internal fields
}

// Option configures the application
type Option func(*App)

// Component is a pluggable module
type Component interface {
    RegisterRoutes(r *router.Router)
    RegisterMiddleware(m *middleware.Registry)
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
    Health() (name string, status health.HealthStatus)
    Dependencies() []reflect.Type
}

// Runner is a background service
type Runner interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

---

## Usage Examples

### Minimal Application

```go
app := core.New()
app.Get("/", func(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Hello World"))
})
app.Boot()
```

### Production Configuration

```go
app := core.New(
    // Server
    core.WithAddr(":8080"),
    core.WithServerTimeouts(30*time.Second, 5*time.Second, 30*time.Second, 60*time.Second),
    core.WithMaxBodyBytes(10 << 20), // 10 MiB

    // Security
    core.WithSecurityHeadersEnabled(true),
    core.WithAbuseGuardEnabled(true),

    // Middleware
    core.WithRecommendedMiddleware(),
    core.WithCORS(cors.DefaultConfig()),

    // Components
    core.WithComponent(&websocket.Component{}),
    core.WithComponent(&webhook.Component{}),

    // TLS
    core.WithTLS("cert.pem", "key.pem"),
    core.WithHTTP2(true),
)
```

### With Custom Component

```go
type MyComponent struct {}

func (c *MyComponent) RegisterRoutes(r *router.Router) {
    r.Get("/custom", c.handleCustom)
}

func (c *MyComponent) Start(ctx context.Context) error {
    // Initialize resources
    return nil
}

func (c *MyComponent) Stop(ctx context.Context) error {
    // Cleanup
    return nil
}

// ... other Component methods

app := core.New(
    core.WithComponent(&MyComponent{}),
)
```

---

## Common Patterns

### 1. Graceful Shutdown

```go
app := core.New(
    core.WithShutdownTimeout(10 * time.Second),
)

// Boot() blocks until SIGINT/SIGTERM
if err := app.Boot(); err != nil {
    log.Fatal(err)
}
// Connections are drained before exit
```

### 2. Environment-Based Configuration

```go
import "github.com/spcent/plumego/config"

cfg := config.Load()

app := core.New(
    core.WithAddr(cfg.Get("APP_ADDR", ":8080")),
    core.WithDebug(cfg.GetBool("APP_DEBUG", false)),
)
```

### 3. Component Composition

```go
app := core.New(
    core.WithComponent(&websocket.Component{
        Secret: os.Getenv("WS_SECRET"),
    }),
    core.WithComponent(&webhook.InComponent{
        GitHubSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
    }),
    core.WithComponent(&devtools.Component{
        Enabled: config.IsDebug(),
    }),
)
```

### 4. Testing Applications

```go
func TestApp(t *testing.T) {
    app := core.New(
        core.WithAddr(":0"), // Random port
    )
    app.Get("/test", handler)

    go app.Boot()
    defer app.Shutdown(context.Background())

    // Make HTTP requests to app.Addr()
}
```

---

## Built-in Components

Plumego includes several production-ready components:

| Component | Package | Description |
|-----------|---------|-------------|
| **DevTools** | `core/components/devtools` | Development utilities (routes, profiling) |
| **Observability** | `core/components/observability` | Metrics, tracing, health checks |
| **Ops** | `core/components/ops` | Operational endpoints (health, metrics) |
| **Tenant** | `core/components/tenant` | Multi-tenancy management |
| **Webhook (In)** | `core/components/webhook` | Inbound webhook receivers |
| **WebSocket** | `core/components/websocket` | WebSocket hub with JWT auth |

See [Components](components.md) for detailed documentation.

---

## Configuration Reference

### Server Options

```go
core.WithAddr(":8080")
core.WithServerTimeouts(read, write, idle, shutdown time.Duration)
core.WithMaxBodyBytes(10 << 20)
core.WithTLS("cert.pem", "key.pem")
core.WithHTTP2(true)
```

### Middleware Options

```go
core.WithRecommendedMiddleware()  // RequestID + Logging + Recovery
core.WithRequestID()
core.WithLogging()
core.WithRecovery()
core.WithCORS(config)
core.WithSecurityHeadersEnabled(true)
core.WithAbuseGuardEnabled(true)
```

### Component & Runner Options

```go
core.WithComponent(component Component)
core.WithRunner(runner Runner)
```

### Feature Options

```go
core.WithDebug()
core.WithShutdownTimeout(duration)
core.WithTenantConfigManager(manager)
core.WithAIProvider(provider)
```

See [Configuration Options](configuration-options.md) for complete reference.

---

## Error Handling

```go
// Boot returns error if server fails to start
if err := app.Boot(); err != nil {
    log.Fatalf("Failed to boot: %v", err)
}

// Shutdown returns error if graceful shutdown fails
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
if err := app.Shutdown(ctx); err != nil {
    log.Printf("Shutdown error: %v", err)
}
```

---

## Best Practices

### ✅ Do

- Use functional options for configuration
- Enable recommended middleware in production
- Configure appropriate timeouts
- Register components before calling `Boot()`
- Handle `Boot()` errors
- Use graceful shutdown

### ❌ Don't

- Modify `App` after `Boot()` is called
- Use global mutable state
- Ignore shutdown errors
- Skip timeout configuration in production
- Register routes after `Boot()`

---

## Performance Considerations

1. **Connection Pooling**: HTTP/2 is enabled by default
2. **Keep-Alive**: Idle timeout controls connection reuse
3. **Body Limits**: Prevent memory exhaustion with `WithMaxBodyBytes()`
4. **Shutdown Timeout**: Balance between graceful and fast shutdown
5. **Middleware Order**: Place cheap middleware early in chain

---

## Security Considerations

1. **TLS**: Always use TLS in production (`WithTLS()`)
2. **Timeouts**: Prevent slowloris attacks with read/write timeouts
3. **Body Limits**: Prevent DoS with `WithMaxBodyBytes()`
4. **Security Headers**: Enable with `WithSecurityHeadersEnabled()`
5. **Abuse Guard**: Rate limiting with `WithAbuseGuardEnabled()`

---

## Troubleshooting

### Server won't start

```go
// Check port availability
app := core.New(core.WithAddr(":8080"))
if err := app.Boot(); err != nil {
    // Port in use or permission denied
}
```

### Components not starting

```go
// Check component dependencies
type MyComponent struct {}

func (c *MyComponent) Dependencies() []reflect.Type {
    return []reflect.Type{
        reflect.TypeOf(&OtherComponent{}),
    }
}
```

### Shutdown hangs

```go
// Increase shutdown timeout
app := core.New(
    core.WithShutdownTimeout(30 * time.Second),
)
```

---

## Next Steps

- **[Application](application.md)** - Deep dive into app creation and routing
- **[Lifecycle](lifecycle.md)** - Understand startup and shutdown sequence
- **[Components](components.md)** - Build pluggable components
- **[Dependency Injection](dependency-injection.md)** - Manage component dependencies
- **[Configuration Options](configuration-options.md)** - Complete options reference

---

## Related Modules

- **[Router](../router/)** - HTTP routing and path parameters
- **[Middleware](../middleware/)** - Request processing chain
- **[Contract](../contract/)** - Request context and error handling
- **[Config](../config/)** - Environment variable management

---

**Stability**: High - Breaking changes require major version bump
**Maintainers**: Plumego Core Team
**Last Updated**: 2026-02-11
