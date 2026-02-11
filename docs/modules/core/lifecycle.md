# Application Lifecycle Management

> **Package**: `github.com/spcent/plumego/core`

This document covers the complete lifecycle of a Plumego application, from creation to shutdown, including component orchestration and graceful termination.

---

## Table of Contents

- [Lifecycle Overview](#lifecycle-overview)
- [Boot Sequence](#boot-sequence)
- [Shutdown Sequence](#shutdown-sequence)
- [Signal Handling](#signal-handling)
- [Component Lifecycle](#component-lifecycle)
- [Runner Lifecycle](#runner-lifecycle)
- [Error Handling](#error-handling)
- [Best Practices](#best-practices)

---

## Lifecycle Overview

A Plumego application follows an explicit lifecycle:

```
┌─────────────┐
│   New()     │  Create application
└──────┬──────┘
       │
       ▼
┌─────────────┐
│  Configure  │  Add routes, middleware, components
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Boot()    │  Start server + components
└──────┬──────┘
       │
       ▼
┌─────────────┐
│   Running   │  Handle requests
└──────┬──────┘
       │
       ▼ (SIGINT/SIGTERM)
┌─────────────┐
│  Shutdown() │  Graceful termination
└─────────────┘
```

---

## Boot Sequence

### 1. Application Creation

```go
app := core.New(options...)
```

**What happens**:
- HTTP server is configured
- Router is initialized
- Middleware chain is built
- DI container is created
- Components are registered (not started)

### 2. Route Registration

```go
app.Get("/users", handler)
app.Post("/users", handler)
```

**What happens**:
- Routes are added to the router's trie
- Middleware can be attached per-route
- Path parameters are validated

### 3. Component Registration

```go
app := core.New(
    core.WithComponent(&websocket.Component{}),
)
```

**What happens**:
- Components are added to registry
- Dependencies are resolved via DI container
- Topological sort determines start order

### 4. Boot Call

```go
if err := app.Boot(); err != nil {
    log.Fatal(err)
}
```

**What happens** (in order):

1. **Dependency Resolution**
   - DI container resolves all component dependencies
   - Topological sort determines start order
   - Circular dependencies are detected and rejected

2. **Component Start**
   - `RegisterRoutes()` called for each component
   - `RegisterMiddleware()` called for each component
   - `Start(ctx)` called in dependency order
   - If any `Start()` fails, previously started components are stopped

3. **Runner Start**
   - All registered runners are started concurrently
   - Each runner runs in its own goroutine
   - Errors are logged but don't block boot

4. **HTTP Server Start**
   - `http.Server.ListenAndServe()` or `ListenAndServeTLS()` is called
   - Server blocks until shutdown signal

5. **Signal Listening**
   - Goroutine listens for SIGINT/SIGTERM
   - On signal, triggers graceful shutdown

---

## Shutdown Sequence

### Automatic Shutdown

`Boot()` blocks until a shutdown signal (SIGINT or SIGTERM) is received:

```go
app.Boot() // Blocks until Ctrl+C
```

### Manual Shutdown

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := app.Shutdown(ctx); err != nil {
    log.Printf("Shutdown error: %v", err)
}
```

### Shutdown Steps (in order)

1. **HTTP Server Shutdown**
   - `http.Server.Shutdown(ctx)` is called
   - New connections are rejected
   - Existing connections are allowed to complete
   - Context timeout applies

2. **Component Stop**
   - `Stop(ctx)` called for each component in **reverse** start order
   - Each component has the remaining context timeout
   - Errors are logged but don't stop the shutdown

3. **Runner Stop**
   - `Stop(ctx)` called for each runner
   - Runners stop concurrently
   - Context timeout applies

4. **Final Cleanup**
   - All goroutines are awaited
   - Resources are released

---

## Signal Handling

### Default Signals

Plumego automatically handles:
- **SIGINT** (Ctrl+C)
- **SIGTERM** (kill)

```go
app.Boot() // Automatically handles signals
```

### Custom Signal Handling

```go
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM, syscall.SIGHUP)

go func() {
    sig := <-sigCh
    log.Printf("Received signal: %v", sig)

    if sig == syscall.SIGHUP {
        // Handle reload
        log.Println("Reloading configuration...")
    } else {
        // Shutdown
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        app.Shutdown(ctx)
    }
}()

app.Boot()
```

### Graceful vs Forceful Shutdown

```go
// Graceful: Wait up to 10 seconds for connections to close
ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
defer cancel()
if err := app.Shutdown(ctx); err != nil {
    log.Printf("Graceful shutdown failed: %v", err)
    // Force shutdown
    os.Exit(1)
}
```

---

## Component Lifecycle

### Component Interface

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

### Lifecycle Hooks

```go
type MyComponent struct {
    db *sql.DB
}

// Called during boot, before Start
func (c *MyComponent) RegisterRoutes(r *router.Router) {
    r.Get("/my-endpoint", c.handleRequest)
}

// Called during boot, before Start
func (c *MyComponent) RegisterMiddleware(m *middleware.Registry) {
    m.Use(c.myMiddleware)
}

// Called during boot, after route/middleware registration
func (c *MyComponent) Start(ctx context.Context) error {
    var err error
    c.db, err = sql.Open("postgres", "...")
    if err != nil {
        return fmt.Errorf("failed to open database: %w", err)
    }
    return nil
}

// Called during shutdown, in reverse order
func (c *MyComponent) Stop(ctx context.Context) error {
    if c.db != nil {
        return c.db.Close()
    }
    return nil
}

// Called by health checks
func (c *MyComponent) Health() (string, health.HealthStatus) {
    if c.db == nil {
        return "mycomponent", health.StatusDown
    }
    if err := c.db.Ping(); err != nil {
        return "mycomponent", health.StatusDown
    }
    return "mycomponent", health.StatusUp
}

// Declares dependencies (started before this component)
func (c *MyComponent) Dependencies() []reflect.Type {
    return []reflect.Type{
        reflect.TypeOf(&DatabaseComponent{}),
    }
}
```

### Start Order

Components are started in topological order based on dependencies:

```go
// Component A has no dependencies
type ComponentA struct {}
func (c *ComponentA) Dependencies() []reflect.Type { return nil }

// Component B depends on A
type ComponentB struct {}
func (c *ComponentB) Dependencies() []reflect.Type {
    return []reflect.Type{reflect.TypeOf(&ComponentA{})}
}

// Component C depends on B
type ComponentC struct {}
func (c *ComponentC) Dependencies() []reflect.Type {
    return []reflect.Type{reflect.TypeOf(&ComponentB{})}
}

app := core.New(
    core.WithComponent(&ComponentC{}), // Registered in any order
    core.WithComponent(&ComponentA{}),
    core.WithComponent(&ComponentB{}),
)

// Start order: A → B → C
// Stop order:  C → B → A
```

### Error Handling

If a component fails to start, all previously started components are stopped:

```go
func (c *MyComponent) Start(ctx context.Context) error {
    if err := c.initialize(); err != nil {
        // This component and all previously started components will be stopped
        return fmt.Errorf("initialization failed: %w", err)
    }
    return nil
}
```

---

## Runner Lifecycle

### Runner Interface

```go
type Runner interface {
    Start(ctx context.Context) error
    Stop(ctx context.Context) error
}
```

### Example Runner

```go
type BackgroundWorker struct {
    done chan struct{}
}

func (w *BackgroundWorker) Start(ctx context.Context) error {
    w.done = make(chan struct{})

    go func() {
        ticker := time.NewTicker(1 * time.Minute)
        defer ticker.Stop()

        for {
            select {
            case <-ticker.C:
                w.doWork()
            case <-w.done:
                return
            case <-ctx.Done():
                return
            }
        }
    }()

    return nil
}

func (w *BackgroundWorker) Stop(ctx context.Context) error {
    close(w.done)
    return nil
}

func (w *BackgroundWorker) doWork() {
    // Background task logic
}
```

### Registration

```go
app := core.New(
    core.WithRunner(&BackgroundWorker{}),
)
```

### Lifecycle Timing

Runners are started **after** all components:

```
Boot():
  1. Start Components (in dependency order)
  2. Start Runners (concurrently)
  3. Start HTTP Server

Shutdown():
  1. Stop HTTP Server
  2. Stop Components (reverse order)
  3. Stop Runners (concurrently)
```

---

## Error Handling

### Boot Errors

```go
if err := app.Boot(); err != nil {
    log.Fatalf("Failed to start: %v", err)
}
```

**Possible errors**:
- Port already in use
- TLS certificate invalid
- Component start failure
- Dependency resolution failure

### Shutdown Errors

```go
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()

if err := app.Shutdown(ctx); err != nil {
    if errors.Is(err, context.DeadlineExceeded) {
        log.Println("Shutdown timed out, forcing exit")
    } else {
        log.Printf("Shutdown error: %v", err)
    }
}
```

**Possible errors**:
- Context deadline exceeded
- Component stop failure
- HTTP server shutdown failure

---

## Best Practices

### ✅ Do

1. **Handle Boot Errors**
   ```go
   if err := app.Boot(); err != nil {
       log.Fatalf("Boot failed: %v", err)
   }
   ```

2. **Set Appropriate Timeouts**
   ```go
   app := core.New(
       core.WithShutdownTimeout(10 * time.Second),
   )
   ```

3. **Implement Health Checks**
   ```go
   func (c *MyComponent) Health() (string, health.HealthStatus) {
       // Check component health
       return "mycomponent", health.StatusUp
   }
   ```

4. **Clean Up Resources**
   ```go
   func (c *MyComponent) Stop(ctx context.Context) error {
       if c.conn != nil {
           c.conn.Close()
       }
       return nil
   }
   ```

5. **Use Context in Long Operations**
   ```go
   func (c *MyComponent) Start(ctx context.Context) error {
       select {
       case <-c.initialize():
           return nil
       case <-ctx.Done():
           return ctx.Err()
       }
   }
   ```

### ❌ Don't

1. **Ignore Errors**
   ```go
   app.Boot() // ❌ Ignoring error
   ```

2. **Block in Start()**
   ```go
   func (c *MyComponent) Start(ctx context.Context) error {
       // ❌ Don't block here
       for {
           c.doWork()
       }
       return nil
   }
   ```

3. **Panic in Lifecycle Methods**
   ```go
   func (c *MyComponent) Start(ctx context.Context) error {
       panic("something went wrong") // ❌ Return error instead
   }
   ```

4. **Leak Goroutines**
   ```go
   func (c *MyComponent) Start(ctx context.Context) error {
       go func() {
           // ❌ No way to stop this goroutine
           for {
               c.doWork()
           }
       }()
       return nil
   }
   ```

5. **Depend on Start Order Without Declaring**
   ```go
   // ❌ Implicit dependency
   func (c *ComponentB) Start(ctx context.Context) error {
       // Assumes ComponentA is already started
       return nil
   }

   // ✅ Explicit dependency
   func (c *ComponentB) Dependencies() []reflect.Type {
       return []reflect.Type{reflect.TypeOf(&ComponentA{})}
   }
   ```

---

## Examples

### Basic Lifecycle

```go
package main

import (
    "log"
    "github.com/spcent/plumego/core"
)

func main() {
    app := core.New(core.WithAddr(":8080"))

    app.Get("/", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Hello World"))
    })

    log.Println("Starting server...")
    if err := app.Boot(); err != nil {
        log.Fatalf("Failed to start: %v", err)
    }
    log.Println("Server stopped")
}
```

### With Component

```go
package main

import (
    "context"
    "log"
    "github.com/spcent/plumego/core"
)

type MyComponent struct {
    started bool
}

func (c *MyComponent) RegisterRoutes(r *router.Router) {
    r.Get("/status", c.handleStatus)
}

func (c *MyComponent) RegisterMiddleware(m *middleware.Registry) {}

func (c *MyComponent) Start(ctx context.Context) error {
    log.Println("MyComponent: Starting...")
    c.started = true
    return nil
}

func (c *MyComponent) Stop(ctx context.Context) error {
    log.Println("MyComponent: Stopping...")
    c.started = false
    return nil
}

func (c *MyComponent) Health() (string, health.HealthStatus) {
    if c.started {
        return "mycomponent", health.StatusUp
    }
    return "mycomponent", health.StatusDown
}

func (c *MyComponent) Dependencies() []reflect.Type {
    return nil
}

func (c *MyComponent) handleStatus(w http.ResponseWriter, r *http.Request) {
    w.Write([]byte("Component is running"))
}

func main() {
    app := core.New(
        core.WithAddr(":8080"),
        core.WithComponent(&MyComponent{}),
    )

    if err := app.Boot(); err != nil {
        log.Fatalf("Boot failed: %v", err)
    }
}
```

### With Graceful Shutdown

```go
package main

import (
    "context"
    "log"
    "os"
    "os/signal"
    "syscall"
    "time"

    "github.com/spcent/plumego/core"
)

func main() {
    app := core.New(
        core.WithAddr(":8080"),
        core.WithShutdownTimeout(10 * time.Second),
    )

    app.Get("/", func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(5 * time.Second) // Simulate slow request
        w.Write([]byte("OK"))
    })

    // Custom signal handling
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

    go func() {
        sig := <-sigCh
        log.Printf("Received signal: %v, shutting down...", sig)

        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()

        if err := app.Shutdown(ctx); err != nil {
            log.Printf("Shutdown error: %v", err)
            os.Exit(1)
        }
        log.Println("Shutdown complete")
        os.Exit(0)
    }()

    log.Println("Server starting...")
    if err := app.Boot(); err != nil {
        log.Fatalf("Boot failed: %v", err)
    }
}
```

---

## Troubleshooting

### Server Won't Stop

**Problem**: Server hangs during shutdown

**Solution**: Check for:
1. Long-running requests
2. Goroutines not respecting context
3. Components blocking in `Stop()`

```go
// Increase timeout
app := core.New(
    core.WithShutdownTimeout(30 * time.Second),
)
```

### Component Start Failure

**Problem**: Component fails to start, boot aborted

**Solution**: Check component `Start()` implementation:
```go
func (c *MyComponent) Start(ctx context.Context) error {
    // Add detailed logging
    log.Println("Starting MyComponent...")

    if err := c.initialize(); err != nil {
        log.Printf("Initialization failed: %v", err)
        return err
    }

    log.Println("MyComponent started successfully")
    return nil
}
```

### Circular Dependencies

**Problem**: `panic: circular dependency detected`

**Solution**: Review component dependencies:
```go
// ❌ Circular: A → B → A
type ComponentA struct {}
func (c *ComponentA) Dependencies() []reflect.Type {
    return []reflect.Type{reflect.TypeOf(&ComponentB{})}
}

type ComponentB struct {}
func (c *ComponentB) Dependencies() []reflect.Type {
    return []reflect.Type{reflect.TypeOf(&ComponentA{})}
}

// ✅ Fixed: A → B (no circular dependency)
type ComponentA struct {}
func (c *ComponentA) Dependencies() []reflect.Type {
    return nil
}
```

---

## Next Steps

- **[Components](components.md)** - Build custom components
- **[Dependency Injection](dependency-injection.md)** - Manage dependencies
- **[Runners](runners.md)** - Background services
- **[Application](application.md)** - Application configuration

---

**Related**:
- [Application Configuration](application.md)
- [Component System](components.md)
- [Health Checks](../health/)
