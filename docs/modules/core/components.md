# Component System

> **Package**: `github.com/spcent/plumego/core`

Components are self-contained, pluggable modules that extend Plumego applications with additional functionality. This document covers how to build, register, and manage components.

---

## Table of Contents

- [Overview](#overview)
- [Component Interface](#component-interface)
- [Creating Components](#creating-components)
- [Built-in Components](#built-in-components)
- [Component Lifecycle](#component-lifecycle)
- [Dependency Management](#dependency-management)
- [Best Practices](#best-practices)
- [Examples](#examples)

---

## Overview

### What is a Component?

A component is a modular piece of functionality that can:
- Register HTTP routes
- Register middleware
- Start and stop background services
- Report health status
- Declare dependencies on other components

### Why Components?

Components provide:
- **Modularity**: Encapsulate related functionality
- **Reusability**: Share components across applications
- **Testability**: Test components in isolation
- **Lifecycle Management**: Automatic start/stop orchestration
- **Dependency Resolution**: Automatic ordering based on dependencies

---

## Component Interface

```go
type Component interface {
    // RegisterRoutes adds routes to the router
    RegisterRoutes(r *router.Router)

    // RegisterMiddleware adds middleware to the registry
    RegisterMiddleware(m *middleware.Registry)

    // Start initializes the component
    Start(ctx context.Context) error

    // Stop cleanly shuts down the component
    Stop(ctx context.Context) error

    // Health returns the component's health status
    Health() (name string, status health.HealthStatus)

    // Dependencies returns types this component depends on
    Dependencies() []reflect.Type
}
```

### Method Details

#### RegisterRoutes

Called during boot to register HTTP endpoints:

```go
func (c *MyComponent) RegisterRoutes(r *router.Router) {
    r.Get("/my-endpoint", c.handleRequest)
    r.Post("/my-endpoint", c.handleCreate)

    // Can also use route groups
    group := r.Group("/my-api")
    group.Get("/status", c.handleStatus)
}
```

#### RegisterMiddleware

Called during boot to register middleware:

```go
func (c *MyComponent) RegisterMiddleware(m *middleware.Registry) {
    m.Use(c.authMiddleware)
    m.Use(c.loggingMiddleware)
}
```

#### Start

Called during boot to initialize resources:

```go
func (c *MyComponent) Start(ctx context.Context) error {
    // Connect to database
    db, err := sql.Open("postgres", c.dsn)
    if err != nil {
        return fmt.Errorf("failed to connect: %w", err)
    }
    c.db = db

    // Start background workers
    go c.worker(ctx)

    return nil
}
```

#### Stop

Called during shutdown to clean up resources:

```go
func (c *MyComponent) Stop(ctx context.Context) error {
    // Close database
    if c.db != nil {
        if err := c.db.Close(); err != nil {
            return fmt.Errorf("failed to close db: %w", err)
        }
    }

    // Stop background workers
    c.stopWorker()

    return nil
}
```

#### Health

Called by health check endpoints:

```go
func (c *MyComponent) Health() (string, health.HealthStatus) {
    if c.db == nil {
        return "mycomponent", health.StatusDown
    }

    if err := c.db.Ping(); err != nil {
        return "mycomponent", health.StatusDown
    }

    return "mycomponent", health.StatusUp
}
```

#### Dependencies

Declares component dependencies for ordering:

```go
func (c *MyComponent) Dependencies() []reflect.Type {
    return []reflect.Type{
        reflect.TypeOf(&DatabaseComponent{}),
        reflect.TypeOf(&CacheComponent{}),
    }
}
```

---

## Creating Components

### Basic Component

```go
package mycomponent

import (
    "context"
    "net/http"
    "reflect"

    "github.com/spcent/plumego/router"
    "github.com/spcent/plumego/middleware"
    "github.com/spcent/plumego/health"
)

type Component struct {
    // Configuration
    enabled bool

    // Internal state
    started bool
}

func New(enabled bool) *Component {
    return &Component{
        enabled: enabled,
    }
}

func (c *Component) RegisterRoutes(r *router.Router) {
    if !c.enabled {
        return
    }

    r.Get("/my-api/status", c.handleStatus)
}

func (c *Component) RegisterMiddleware(m *middleware.Registry) {
    // No middleware
}

func (c *Component) Start(ctx context.Context) error {
    if !c.enabled {
        return nil
    }

    log.Println("MyComponent: Starting...")
    c.started = true
    return nil
}

func (c *Component) Stop(ctx context.Context) error {
    if !c.enabled {
        return nil
    }

    log.Println("MyComponent: Stopping...")
    c.started = false
    return nil
}

func (c *Component) Health() (string, health.HealthStatus) {
    if !c.enabled {
        return "mycomponent", health.StatusDisabled
    }

    if c.started {
        return "mycomponent", health.StatusUp
    }

    return "mycomponent", health.StatusDown
}

func (c *Component) Dependencies() []reflect.Type {
    return nil
}

func (c *Component) handleStatus(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("OK"))
}
```

### Component with Database

```go
package dbcomponent

import (
    "context"
    "database/sql"
    "fmt"

    _ "github.com/lib/pq"
)

type Component struct {
    DSN string
    db  *sql.DB
}

func (c *Component) Start(ctx context.Context) error {
    db, err := sql.Open("postgres", c.DSN)
    if err != nil {
        return fmt.Errorf("failed to open database: %w", err)
    }

    if err := db.PingContext(ctx); err != nil {
        db.Close()
        return fmt.Errorf("failed to ping database: %w", err)
    }

    c.db = db
    return nil
}

func (c *Component) Stop(ctx context.Context) error {
    if c.db != nil {
        return c.db.Close()
    }
    return nil
}

func (c *Component) Health() (string, health.HealthStatus) {
    if c.db == nil {
        return "database", health.StatusDown
    }

    ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
    defer cancel()

    if err := c.db.PingContext(ctx); err != nil {
        return "database", health.StatusDown
    }

    return "database", health.StatusUp
}

func (c *Component) DB() *sql.DB {
    return c.db
}

// ... other methods
```

### Component with Background Worker

```go
package workercomponent

import (
    "context"
    "time"
)

type Component struct {
    interval time.Duration
    done     chan struct{}
}

func New(interval time.Duration) *Component {
    return &Component{
        interval: interval,
        done:     make(chan struct{}),
    }
}

func (c *Component) Start(ctx context.Context) error {
    go c.worker(ctx)
    return nil
}

func (c *Component) Stop(ctx context.Context) error {
    close(c.done)
    return nil
}

func (c *Component) worker(ctx context.Context) {
    ticker := time.NewTicker(c.interval)
    defer ticker.Stop()

    for {
        select {
        case <-ticker.C:
            c.doWork()
        case <-c.done:
            return
        case <-ctx.Done():
            return
        }
    }
}

func (c *Component) doWork() {
    // Background task logic
}

// ... other methods
```

---

## Built-in Components

Plumego includes several production-ready components in `core/components/`:

### DevTools Component

Development utilities and debugging endpoints.

```go
import "github.com/spcent/plumego/core/components/devtools"

app := core.New(
    core.WithComponent(&devtools.Component{
        Enabled: true,
    }),
)
```

**Features**:
- `/debug/routes` - List all registered routes
- `/debug/pprof` - Go profiling endpoints
- `/debug/config` - View configuration

### Observability Component

Metrics, tracing, and monitoring.

```go
import "github.com/spcent/plumego/core/components/observability"

app := core.New(
    core.WithComponent(&observability.Component{
        MetricsEnabled: true,
        TracingEnabled: true,
    }),
)
```

**Features**:
- Prometheus metrics at `/metrics`
- OpenTelemetry tracing
- Request logging
- Performance monitoring

### Ops Component

Operational endpoints for health checks and readiness.

```go
import "github.com/spcent/plumego/core/components/ops"

app := core.New(
    core.WithComponent(&ops.Component{}),
)
```

**Features**:
- `/health` - Health check endpoint
- `/ready` - Readiness probe
- `/version` - Version information

### Tenant Component

Multi-tenancy management.

```go
import "github.com/spcent/plumego/core/components/tenant"

app := core.New(
    core.WithComponent(&tenant.Component{
        ConfigManager: configMgr,
        QuotaManager:  quotaMgr,
    }),
)
```

**Features**:
- Tenant configuration management
- Quota enforcement
- Policy evaluation
- Per-tenant rate limiting

### WebSocket Component

WebSocket hub with JWT authentication.

```go
import "github.com/spcent/plumego/core/components/websocket"

app := core.New(
    core.WithComponent(&websocket.Component{
        Secret: os.Getenv("WS_SECRET"),
    }),
)
```

**Features**:
- WebSocket connections at `/ws`
- JWT token authentication
- Room-based messaging
- Broadcast support

### Webhook Component

Inbound webhook receivers (GitHub, Stripe, etc.).

```go
import "github.com/spcent/plumego/core/components/webhook"

app := core.New(
    core.WithComponent(&webhook.InComponent{
        GitHubSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
        StripeSecret: os.Getenv("STRIPE_WEBHOOK_SECRET"),
    }),
)
```

**Features**:
- `/webhooks/github` - GitHub webhook receiver
- `/webhooks/stripe` - Stripe webhook receiver
- Signature verification
- Event dispatching via pub/sub

---

## Component Lifecycle

### Registration

```go
app := core.New(
    core.WithComponent(&MyComponent{}),
)
```

### Boot Sequence

1. **RegisterRoutes** - All components register routes
2. **RegisterMiddleware** - All components register middleware
3. **Dependency Resolution** - DI container resolves dependencies
4. **Start** - Components start in dependency order

### Shutdown Sequence

1. **HTTP Server Stop** - Server stops accepting new connections
2. **Stop** - Components stop in **reverse** order
3. **Cleanup** - All resources released

### Example Flow

```go
// Registration
app := core.New(
    core.WithComponent(&ComponentA{}), // No dependencies
    core.WithComponent(&ComponentB{}), // Depends on A
)

// Boot
app.Boot()
// 1. ComponentA.RegisterRoutes()
// 2. ComponentA.RegisterMiddleware()
// 3. ComponentB.RegisterRoutes()
// 4. ComponentB.RegisterMiddleware()
// 5. ComponentA.Start()
// 6. ComponentB.Start()

// Shutdown
app.Shutdown(ctx)
// 1. ComponentB.Stop()
// 2. ComponentA.Stop()
```

---

## Dependency Management

### Declaring Dependencies

```go
type ComponentB struct {
    // Needs ComponentA
}

func (c *ComponentB) Dependencies() []reflect.Type {
    return []reflect.Type{
        reflect.TypeOf(&ComponentA{}),
    }
}
```

### Multiple Dependencies

```go
func (c *MyComponent) Dependencies() []reflect.Type {
    return []reflect.Type{
        reflect.TypeOf(&DatabaseComponent{}),
        reflect.TypeOf(&CacheComponent{}),
        reflect.TypeOf(&LoggingComponent{}),
    }
}
```

### Accessing Dependencies

Components can access each other through the DI container:

```go
func (c *ComponentB) Start(ctx context.Context) error {
    // Get dependency from DI container
    var componentA *ComponentA
    if err := app.DI().Resolve(&componentA); err != nil {
        return err
    }

    c.componentA = componentA
    return nil
}
```

### Circular Dependencies

Circular dependencies are **not allowed** and will cause a panic:

```go
// ❌ This will panic
type ComponentA struct {}
func (c *ComponentA) Dependencies() []reflect.Type {
    return []reflect.Type{reflect.TypeOf(&ComponentB{})}
}

type ComponentB struct {}
func (c *ComponentB) Dependencies() []reflect.Type {
    return []reflect.Type{reflect.TypeOf(&ComponentA{})}
}
```

---

## Best Practices

### ✅ Do

1. **Keep Components Focused**
   ```go
   // ✅ Single responsibility
   type CacheComponent struct { ... }
   type DatabaseComponent struct { ... }

   // ❌ Too many responsibilities
   type DataComponent struct {
       cache    *redis.Client
       database *sql.DB
       queue    *amqp.Channel
   }
   ```

2. **Handle Errors Gracefully**
   ```go
   func (c *Component) Start(ctx context.Context) error {
       if err := c.initialize(); err != nil {
           return fmt.Errorf("initialization failed: %w", err)
       }
       return nil
   }
   ```

3. **Implement Health Checks**
   ```go
   func (c *Component) Health() (string, health.HealthStatus) {
       if err := c.check(); err != nil {
           return "mycomponent", health.StatusDown
       }
       return "mycomponent", health.StatusUp
   }
   ```

4. **Clean Up Resources**
   ```go
   func (c *Component) Stop(ctx context.Context) error {
       if c.conn != nil {
           c.conn.Close()
       }
       return nil
   }
   ```

5. **Declare Dependencies Explicitly**
   ```go
   func (c *Component) Dependencies() []reflect.Type {
       return []reflect.Type{reflect.TypeOf(&RequiredComponent{})}
   }
   ```

### ❌ Don't

1. **Don't Block in Start()**
   ```go
   // ❌ Blocks forever
   func (c *Component) Start(ctx context.Context) error {
       for {
           c.doWork()
       }
       return nil
   }

   // ✅ Start goroutine instead
   func (c *Component) Start(ctx context.Context) error {
       go c.worker(ctx)
       return nil
   }
   ```

2. **Don't Panic**
   ```go
   // ❌ Don't panic
   func (c *Component) Start(ctx context.Context) error {
       panic("something went wrong")
   }

   // ✅ Return error
   func (c *Component) Start(ctx context.Context) error {
       return errors.New("something went wrong")
   }
   ```

3. **Don't Ignore Context**
   ```go
   // ❌ Ignores context timeout
   func (c *Component) Stop(ctx context.Context) error {
       time.Sleep(10 * time.Second) // Might exceed context timeout
       return nil
   }

   // ✅ Respects context
   func (c *Component) Stop(ctx context.Context) error {
       select {
       case <-c.cleanup():
           return nil
       case <-ctx.Done():
           return ctx.Err()
       }
   }
   ```

4. **Don't Use Global State**
   ```go
   // ❌ Global variable
   var globalDB *sql.DB

   // ✅ Component field
   type Component struct {
       db *sql.DB
   }
   ```

---

## Examples

### Minimal Component

```go
type MinimalComponent struct {}

func (c *MinimalComponent) RegisterRoutes(r *router.Router) {
    r.Get("/minimal", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })
}

func (c *MinimalComponent) RegisterMiddleware(m *middleware.Registry) {}
func (c *MinimalComponent) Start(ctx context.Context) error { return nil }
func (c *MinimalComponent) Stop(ctx context.Context) error { return nil }
func (c *MinimalComponent) Health() (string, health.HealthStatus) {
    return "minimal", health.StatusUp
}
func (c *MinimalComponent) Dependencies() []reflect.Type { return nil }
```

### Full-Featured Component

```go
package analytics

import (
    "context"
    "database/sql"
    "time"
)

type Component struct {
    DSN      string
    Interval time.Duration

    db     *sql.DB
    done   chan struct{}
    events chan Event
}

type Event struct {
    Type      string
    Timestamp time.Time
    Data      map[string]interface{}
}

func (c *Component) RegisterRoutes(r *router.Router) {
    r.Post("/analytics/track", c.handleTrack)
    r.Get("/analytics/stats", c.handleStats)
}

func (c *Component) RegisterMiddleware(m *middleware.Registry) {
    m.Use(c.trackingMiddleware)
}

func (c *Component) Start(ctx context.Context) error {
    // Connect to database
    db, err := sql.Open("postgres", c.DSN)
    if err != nil {
        return fmt.Errorf("failed to connect: %w", err)
    }
    c.db = db

    // Initialize channels
    c.done = make(chan struct{})
    c.events = make(chan Event, 1000)

    // Start background worker
    go c.worker(ctx)

    return nil
}

func (c *Component) Stop(ctx context.Context) error {
    close(c.done)
    close(c.events)

    if c.db != nil {
        return c.db.Close()
    }

    return nil
}

func (c *Component) Health() (string, health.HealthStatus) {
    if c.db == nil {
        return "analytics", health.StatusDown
    }

    if err := c.db.Ping(); err != nil {
        return "analytics", health.StatusDown
    }

    return "analytics", health.StatusUp
}

func (c *Component) Dependencies() []reflect.Type {
    return nil
}

func (c *Component) worker(ctx context.Context) {
    ticker := time.NewTicker(c.Interval)
    defer ticker.Stop()

    batch := make([]Event, 0, 100)

    for {
        select {
        case event := <-c.events:
            batch = append(batch, event)
            if len(batch) >= 100 {
                c.flushBatch(batch)
                batch = batch[:0]
            }

        case <-ticker.C:
            if len(batch) > 0 {
                c.flushBatch(batch)
                batch = batch[:0]
            }

        case <-c.done:
            c.flushBatch(batch)
            return

        case <-ctx.Done():
            c.flushBatch(batch)
            return
        }
    }
}

func (c *Component) flushBatch(events []Event) {
    // Write events to database
    for _, event := range events {
        _, err := c.db.Exec(
            "INSERT INTO events (type, timestamp, data) VALUES ($1, $2, $3)",
            event.Type, event.Timestamp, event.Data,
        )
        if err != nil {
            log.Printf("Failed to insert event: %v", err)
        }
    }
}

func (c *Component) Track(event Event) {
    select {
    case c.events <- event:
    default:
        log.Println("Event queue full, dropping event")
    }
}

func (c *Component) handleTrack(w http.ResponseWriter, r *http.Request) {
    // Handle incoming track request
}

func (c *Component) handleStats(w http.ResponseWriter, r *http.Request) {
    // Handle stats request
}

func (c *Component) trackingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()
        next.ServeHTTP(w, r)
        duration := time.Since(start)

        c.Track(Event{
            Type:      "http_request",
            Timestamp: time.Now(),
            Data: map[string]interface{}{
                "method":   r.Method,
                "path":     r.URL.Path,
                "duration": duration.Milliseconds(),
            },
        })
    })
}
```

---

## Next Steps

- **[Dependency Injection](dependency-injection.md)** - DI container usage
- **[Lifecycle](lifecycle.md)** - Component lifecycle in detail
- **[Runners](runners.md)** - Lightweight background services
- **[Testing](testing.md)** - Testing components

---

**Related**:
- [Application Configuration](application.md)
- [Lifecycle Management](lifecycle.md)
- [Health Checks](../health/)
