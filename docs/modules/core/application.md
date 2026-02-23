# Application Creation and Configuration

> **Package**: `github.com/spcent/plumego/core`

This document covers how to create, configure, and use the `App` struct, which is the main entry point for all Plumego applications.

---

## Table of Contents

- [Creating an Application](#creating-an-application)
- [Configuration Patterns](#configuration-patterns)
- [Routing](#routing)
- [Handler Types](#handler-types)
- [Application Methods](#application-methods)
- [Advanced Configuration](#advanced-configuration)
- [Examples](#examples)

---

## Creating an Application

### Basic Creation

```go
package main

import "github.com/spcent/plumego/core"

func main() {
    app := core.New()
    app.Boot()
}
```

### With Options

```go
app := core.New(
    core.WithAddr(":8080"),
    core.WithDebug(),
    core.WithRecommendedMiddleware(),
)
```

### Configuration Sources

Applications can be configured from multiple sources:

```go
import "github.com/spcent/plumego/config"

cfg := config.Load() // Loads from .env and environment

app := core.New(
    core.WithAddr(cfg.Get("APP_ADDR", ":8080")),
    core.WithDebug(cfg.GetBool("APP_DEBUG", false)),
    core.WithShutdownTimeout(
        time.Duration(cfg.GetInt("APP_SHUTDOWN_TIMEOUT_MS", 5000)) * time.Millisecond,
    ),
)
```

---

## Configuration Patterns

### 1. Functional Options Pattern

Plumego uses functional options for type-safe, composable configuration:

```go
type Option func(*App)

func WithAddr(addr string) Option {
    return func(a *App) {
        a.addr = addr
    }
}

// Usage
app := core.New(
    WithAddr(":3000"),
    WithDebug(),
)
```

### 2. Builder Pattern (Not Recommended)

```go
// ❌ Don't do this
app := core.New()
app.SetAddr(":8080")
app.SetDebug(true)
app.Boot()

// ✅ Do this instead
app := core.New(
    core.WithAddr(":8080"),
    core.WithDebug(),
)
app.Boot()
```

### 3. Configuration Presets

```go
func ProductionConfig(addr string) []core.Option {
    return []core.Option{
        core.WithAddr(addr),
        core.WithServerTimeouts(30*time.Second, 5*time.Second, 30*time.Second, 60*time.Second),
        core.WithMaxBodyBytes(10 << 20),
        core.WithSecurityHeadersEnabled(true),
        core.WithAbuseGuardEnabled(true),
        core.WithRecommendedMiddleware(),
        core.WithTLS("cert.pem", "key.pem"),
        core.WithHTTP2(true),
    }
}

app := core.New(ProductionConfig(":443")...)
```

---

## Routing

### HTTP Methods

```go
app.Get("/users", listUsers)
app.Post("/users", createUser)
app.Put("/users/:id", updateUser)
app.Patch("/users/:id", patchUser)
app.Delete("/users/:id", deleteUser)
app.Head("/users/:id", headUser)
app.Options("/users", optionsUsers)
```

### Any Method

```go
app.Any("/webhook", handleWebhook)
```

### Route Groups

```go
api := app.Group("/api/v1")
api.Get("/users", listUsers)
api.Post("/users", createUser)

admin := api.Group("/admin", authMiddleware)
admin.Delete("/users/:id", deleteUser)
// Matches: /api/v1/admin/users/:id
```

### Static Files

```go
// Serve from directory
app.Static("/assets", "./public")

// Serve embedded files
//go:embed static/*
var staticFS embed.FS
app.StaticFS("/static", http.FS(staticFS))
```

---

## Handler Types

### 1. Standard Library Handler

```go
app.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    w.Write([]byte("pong"))
})
```

### 2. Context-Aware Handler

```go
import "github.com/spcent/plumego"

app.GetCtx("/health", func(ctx *plumego.Context) {
    ctx.JSON(http.StatusOK, map[string]string{
        "status": "ok",
    })
})
```

### 3. Handler with Path Parameters

```go
app.Get("/users/:id", func(w http.ResponseWriter, r *http.Request) {
    id := plumego.Param(r, "id")
    fmt.Fprintf(w, "User ID: %s", id)
})

// Or with Context
app.GetCtx("/users/:id", func(ctx *plumego.Context) {
    id := ctx.Param("id")
    ctx.JSON(200, map[string]string{"id": id})
})
```

### 4. Middleware as Handler

```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}

app.Get("/protected", authMiddleware(http.HandlerFunc(handler)))
```

---

## Application Methods

### Core Methods

```go
// Boot starts the HTTP server and blocks until shutdown
func (a *App) Boot() error

// Shutdown gracefully shuts down the server
func (a *App) Shutdown(ctx context.Context) error

// Addr returns the server's listen address
func (a *App) Addr() string

// Router returns the underlying router
func (a *App) Router() *router.Router

// Server returns the underlying http.Server
func (a *App) Server() *http.Server
```

### Routing Methods

```go
// HTTP method shortcuts
func (a *App) Get(path string, handler http.HandlerFunc)
func (a *App) Post(path string, handler http.HandlerFunc)
func (a *App) Put(path string, handler http.HandlerFunc)
func (a *App) Patch(path string, handler http.HandlerFunc)
func (a *App) Delete(path string, handler http.HandlerFunc)
func (a *App) Head(path string, handler http.HandlerFunc)
func (a *App) Options(path string, handler http.HandlerFunc)
func (a *App) Any(path string, handler http.HandlerFunc)

// Context-aware variants
func (a *App) GetCtx(path string, handler func(*plumego.Context))
func (a *App) PostCtx(path string, handler func(*plumego.Context))
// ... etc

// Route groups
func (a *App) Group(prefix string, middleware ...func(http.Handler) http.Handler) *router.Group

// Static files
func (a *App) Static(path, dir string)
func (a *App) StaticFS(path string, fs http.FileSystem)
```

### Middleware Methods

```go
// Use adds middleware to the global chain
func (a *App) Use(middleware ...func(http.Handler) http.Handler)

// UseFunc adds middleware functions
func (a *App) UseFunc(fn func(http.Handler) http.Handler)
```

### Component & Runner Methods

```go
// RegisterComponent adds a component
func (a *App) RegisterComponent(c Component) error

// RegisterRunner adds a runner
func (a *App) RegisterRunner(r Runner) error

// DI returns the dependency injection container
func (a *App) DI() *di.Container
```

---

## Advanced Configuration

### Custom HTTP Server

```go
server := &http.Server{
    Addr:         ":8080",
    ReadTimeout:  30 * time.Second,
    WriteTimeout: 30 * time.Second,
    IdleTimeout:  60 * time.Second,
}

app := core.New(
    core.WithServer(server),
)
```

### Custom Router

```go
customRouter := router.New(
    router.WithCaseSensitive(true),
    router.WithStrictSlash(false),
)

app := core.New(
    core.WithRouter(customRouter),
)
```

### Middleware Chain

```go
chain := middleware.NewChain().
    Use(middleware.RequestID).
    Use(middleware.Logging).
    Use(middleware.Recovery)

app := core.New(
    core.WithMiddlewareChain(chain),
)
```

---

## Examples

### Minimal REST API

```go
package main

import (
    "github.com/spcent/plumego"
    "github.com/spcent/plumego/core"
)

func main() {
    app := core.New(
        core.WithAddr(":8080"),
        core.WithRecommendedMiddleware(),
    )

    app.GetCtx("/users", listUsers)
    app.GetCtx("/users/:id", getUser)
    app.PostCtx("/users", createUser)
    app.PutCtx("/users/:id", updateUser)
    app.DeleteCtx("/users/:id", deleteUser)

    app.Boot()
}

func listUsers(ctx *plumego.Context) {
    ctx.JSON(200, []string{"alice", "bob"})
}

func getUser(ctx *plumego.Context) {
    id := ctx.Param("id")
    ctx.JSON(200, map[string]string{"id": id})
}

func createUser(ctx *plumego.Context) {
    var user map[string]string
    if err := ctx.Bind(&user); err != nil {
        ctx.Error(400, "Invalid request")
        return
    }
    ctx.JSON(201, user)
}

func updateUser(ctx *plumego.Context) {
    id := ctx.Param("id")
    ctx.JSON(200, map[string]string{"id": id, "updated": "true"})
}

func deleteUser(ctx *plumego.Context) {
    ctx.Status(204)
}
```

### Production Application

```go
package main

import (
    "log"
    "os"
    "time"

    "github.com/spcent/plumego/config"
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/middleware/cors"
)

func main() {
    cfg := config.Load()

    app := core.New(
        // Server
        core.WithAddr(cfg.Get("APP_ADDR", ":8080")),
        core.WithServerTimeouts(
            30*time.Second, // read
            30*time.Second, // write
            60*time.Second, // idle
            10*time.Second, // shutdown
        ),
        core.WithMaxBodyBytes(10 << 20), // 10 MiB

        // Security
        core.WithTLS(
            cfg.Get("TLS_CERT_FILE", "cert.pem"),
            cfg.Get("TLS_KEY_FILE", "key.pem"),
        ),
        core.WithHTTP2(cfg.GetBool("APP_ENABLE_HTTP2", true)),
        core.WithSecurityHeadersEnabled(true),
        core.WithAbuseGuardEnabled(true),

        // Middleware
        core.WithRecommendedMiddleware(),
        core.WithCORS(cors.Config{
            AllowOrigins: []string{cfg.Get("CORS_ORIGIN", "*")},
            AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
        }),

        // Features
        core.WithDebug(cfg.GetBool("APP_DEBUG", false)),
    )

    // Routes
    setupRoutes(app)

    // Start server
    log.Printf("Starting server on %s", app.Addr())
    if err := app.Boot(); err != nil {
        log.Fatalf("Failed to start: %v", err)
    }
}

func setupRoutes(app *core.App) {
    app.Get("/health", healthCheck)

    api := app.Group("/api/v1")
    api.Get("/users", listUsers)
    // ... more routes
}
```

### Multi-Component Application

```go
package main

import (
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/core/components/devtools"
    "github.com/spcent/plumego/core/components/observability"
    "github.com/spcent/plumego/core/components/websocket"
    "github.com/spcent/plumego/core/components/webhook"
)

func main() {
    app := core.New(
        core.WithAddr(":8080"),
        core.WithRecommendedMiddleware(),

        // Components
        core.WithComponent(&devtools.Component{
            Enabled: true,
        }),
        core.WithComponent(&observability.Component{
            MetricsEnabled: true,
            TracingEnabled: false,
        }),
        core.WithComponent(&websocket.Component{
            Secret: os.Getenv("WS_SECRET"),
        }),
        core.WithComponent(&webhook.InComponent{
            GitHubSecret: os.Getenv("GITHUB_WEBHOOK_SECRET"),
        }),
    )

    app.Boot()
}
```

### Testing Application

```go
package main

import (
    "context"
    "net/http"
    "testing"
    "time"

    "github.com/spcent/plumego/core"
)

func TestApplication(t *testing.T) {
    app := core.New(
        core.WithAddr(":0"), // Random port
    )

    app.Get("/test", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })

    // Start in background
    go func() {
        if err := app.Boot(); err != nil {
            t.Errorf("Boot failed: %v", err)
        }
    }()

    // Wait for server to start
    time.Sleep(100 * time.Millisecond)

    // Make request
    resp, err := http.Get("http://" + app.Addr() + "/test")
    if err != nil {
        t.Fatalf("Request failed: %v", err)
    }
    defer resp.Body.Close()

    if resp.StatusCode != http.StatusOK {
        t.Errorf("Expected 200, got %d", resp.StatusCode)
    }

    // Shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()
    if err := app.Shutdown(ctx); err != nil {
        t.Errorf("Shutdown failed: %v", err)
    }
}
```

---

## Best Practices

### Configuration

✅ **Do**:
- Use functional options for configuration
- Load configuration from environment variables
- Set appropriate timeouts for production
- Enable security features by default
- Use presets for common configurations

❌ **Don't**:
- Hardcode sensitive values (secrets, keys)
- Ignore error returns from `Boot()` or `Shutdown()`
- Modify app configuration after `Boot()`
- Use global mutable state

### Routing

✅ **Do**:
- Register all routes before calling `Boot()`
- Use route groups for API versioning
- Use context-aware handlers for cleaner code
- Follow RESTful conventions

❌ **Don't**:
- Register routes after `Boot()` is called
- Use wildcard routes carelessly
- Mix different API versions in same group
- Forget to handle errors in handlers

### Error Handling

✅ **Do**:
```go
if err := app.Boot(); err != nil {
    log.Fatalf("Failed to start: %v", err)
}
```

❌ **Don't**:
```go
app.Boot() // Ignoring error
```

---

## Common Patterns

### 1. Configuration Factory

```go
type AppConfig struct {
    Addr          string
    Debug         bool
    TLSCert       string
    TLSKey        string
    EnableMetrics bool
}

func NewApp(cfg AppConfig) *core.App {
    opts := []core.Option{
        core.WithAddr(cfg.Addr),
    }

    if cfg.Debug {
        opts = append(opts, core.WithDebug())
    }

    if cfg.TLSCert != "" && cfg.TLSKey != "" {
        opts = append(opts, core.WithTLS(cfg.TLSCert, cfg.TLSKey))
    }

    if cfg.EnableMetrics {
        opts = append(opts, core.WithComponent(&observability.Component{}))
    }

    return core.New(opts...)
}
```

### 2. Graceful Restart

```go
func main() {
    app := core.New(core.WithAddr(":8080"))

    // Setup routes
    setupRoutes(app)

    // Handle signals
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

    go func() {
        <-sigCh
        ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
        defer cancel()
        app.Shutdown(ctx)
    }()

    app.Boot()
}
```

### 3. Health Check Endpoint

```go
app.Get("/health", func(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(http.StatusOK)
    w.Write([]byte(`{"status":"ok","timestamp":"` + time.Now().Format(time.RFC3339) + `"}`))
})
```

---

## Troubleshooting

### Problem: Port Already in Use

```go
// Error: listen tcp :8080: bind: address already in use

// Solution 1: Use different port
app := core.New(core.WithAddr(":8081"))

// Solution 2: Find and kill process using port
// $ lsof -ti:8080 | xargs kill -9
```

### Problem: TLS Certificate Error

```go
// Error: tls: failed to find any PEM data in certificate input

// Solution: Verify certificate files exist and are valid
if _, err := os.Stat("cert.pem"); os.IsNotExist(err) {
    log.Fatal("Certificate file not found")
}
```

### Problem: Timeout During Shutdown

```go
// Increase shutdown timeout
app := core.New(
    core.WithShutdownTimeout(30 * time.Second),
)
```

---

## Next Steps

- **[Lifecycle](lifecycle.md)** - Understand the boot and shutdown sequence
- **[Components](components.md)** - Build pluggable components
- **[Configuration Options](configuration-options.md)** - Complete options reference
- **[Router](../router/)** - Deep dive into routing system

---

**Related**:
- [Lifecycle Management](lifecycle.md)
- [Component System](components.md)
- [Router Module](../router/)
- [Middleware Module](../middleware/)
