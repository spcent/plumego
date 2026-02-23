# Configuration Options Reference

> **Package**: `github.com/spcent/plumego/core`

Complete reference for all `core.With*` configuration options available when creating a Plumego application.

---

## Table of Contents

- [Server Options](#server-options)
- [TLS Options](#tls-options)
- [Middleware Options](#middleware-options)
- [Component Options](#component-options)
- [Feature Options](#feature-options)
- [Advanced Options](#advanced-options)
- [Presets](#presets)

---

## Server Options

### WithAddr

Set the server listen address.

```go
func WithAddr(addr string) Option
```

**Example**:
```go
app := core.New(
    core.WithAddr(":8080"),        // Listen on port 8080
    core.WithAddr("0.0.0.0:3000"), // Listen on all interfaces
    core.WithAddr("127.0.0.1:80"), // Listen on localhost only
)
```

**Default**: `:8080`

**Environment Variable**: `APP_ADDR`

---

### WithServerTimeouts

Configure HTTP server timeouts.

```go
func WithServerTimeouts(read, write, idle, shutdown time.Duration) Option
```

**Parameters**:
- `read`: Maximum duration for reading the entire request (headers + body)
- `write`: Maximum duration before timing out writes of the response
- `idle`: Maximum duration to wait for the next request when keep-alives are enabled
- `shutdown`: Maximum duration to wait during graceful shutdown

**Example**:
```go
app := core.New(
    core.WithServerTimeouts(
        30 * time.Second, // read
        30 * time.Second, // write
        60 * time.Second, // idle
        10 * time.Second, // shutdown
    ),
)
```

**Defaults**:
- Read: `30s` (`APP_READ_TIMEOUT_MS`)
- Write: `30s` (`APP_WRITE_TIMEOUT_MS`)
- Idle: `60s` (`APP_IDLE_TIMEOUT_MS`)
- Shutdown: `5s` (`APP_SHUTDOWN_TIMEOUT_MS`)

---

### WithMaxBodyBytes

Set the maximum request body size.

```go
func WithMaxBodyBytes(n int64) Option
```

**Example**:
```go
app := core.New(
    core.WithMaxBodyBytes(10 << 20), // 10 MiB
    core.WithMaxBodyBytes(1 << 30),  // 1 GiB
)
```

**Default**: `10485760` (10 MiB) (`APP_MAX_BODY_BYTES`)

---

### WithMaxConcurrency

Set the maximum number of concurrent requests.

```go
func WithMaxConcurrency(n int) Option
```

**Example**:
```go
app := core.New(
    core.WithMaxConcurrency(256),   // Default
    core.WithMaxConcurrency(1000),  // High traffic
)
```

**Default**: `256` (`APP_MAX_CONCURRENCY`)

---

### WithShutdownTimeout

Set the graceful shutdown timeout.

```go
func WithShutdownTimeout(d time.Duration) Option
```

**Example**:
```go
app := core.New(
    core.WithShutdownTimeout(10 * time.Second),
)
```

**Default**: `5s` (`APP_SHUTDOWN_TIMEOUT_MS`)

---

### WithServer

Use a custom `http.Server` instance.

```go
func WithServer(server *http.Server) Option
```

**Example**:
```go
server := &http.Server{
    Addr:         ":8080",
    ReadTimeout:  30 * time.Second,
    WriteTimeout: 30 * time.Second,
}

app := core.New(
    core.WithServer(server),
)
```

---

## TLS Options

### WithTLS

Enable TLS with certificate and key files.

```go
func WithTLS(certFile, keyFile string) Option
```

**Example**:
```go
app := core.New(
    core.WithTLS("cert.pem", "key.pem"),
)
```

**Environment Variables**:
- `TLS_CERT_FILE`
- `TLS_KEY_FILE`
- `TLS_ENABLED`

---

### WithHTTP2

Enable or disable HTTP/2 support.

```go
func WithHTTP2(enabled bool) Option
```

**Example**:
```go
app := core.New(
    core.WithHTTP2(true), // Enable HTTP/2
)
```

**Default**: `true` (`APP_ENABLE_HTTP2`)

**Note**: Requires TLS for HTTP/2

---

## Middleware Options

### WithRecommendedMiddleware

Add recommended middleware stack (RequestID + Logging + Recovery).

```go
func WithRecommendedMiddleware() Option
```

**Example**:
```go
app := core.New(
    core.WithRecommendedMiddleware(),
)

// Equivalent to:
app := core.New(
    core.WithRequestID(),
    core.WithLogging(),
    core.WithRecovery(),
)
```

---

### WithRequestID

Add request ID middleware.

```go
func WithRequestID() Option
```

**Example**:
```go
app := core.New(
    core.WithRequestID(),
)
```

**Adds**: `X-Request-ID` header to all requests

---

### WithLogging

Add request logging middleware.

```go
func WithLogging() Option
```

**Example**:
```go
app := core.New(
    core.WithLogging(),
)
```

**Logs**: Method, path, status, duration, IP

---

### WithRecovery

Add panic recovery middleware.

```go
func WithRecovery() Option
```

**Example**:
```go
app := core.New(
    core.WithRecovery(),
)
```

**Behavior**: Catches panics, logs error, returns 500

---

### WithCORS

Add CORS middleware with configuration.

```go
func WithCORS(config cors.Config) Option
```

**Example**:
```go
import "github.com/spcent/plumego/middleware/cors"

app := core.New(
    core.WithCORS(cors.Config{
        AllowOrigins: []string{"https://example.com"},
        AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
        AllowHeaders: []string{"Authorization", "Content-Type"},
        MaxAge:       3600,
    }),
)
```

---

### WithSecurityHeadersEnabled

Enable security headers middleware.

```go
func WithSecurityHeadersEnabled(enabled bool) Option
```

**Example**:
```go
app := core.New(
    core.WithSecurityHeadersEnabled(true),
)
```

**Adds**:
- `X-Frame-Options: DENY`
- `X-Content-Type-Options: nosniff`
- `X-XSS-Protection: 1; mode=block`
- `Strict-Transport-Security: max-age=31536000`
- `Content-Security-Policy: default-src 'self'`

---

### WithAbuseGuardEnabled

Enable rate limiting and abuse prevention.

```go
func WithAbuseGuardEnabled(enabled bool) Option
```

**Example**:
```go
app := core.New(
    core.WithAbuseGuardEnabled(true),
)
```

---

### WithMiddlewareChain

Use a custom middleware chain.

```go
func WithMiddlewareChain(chain *middleware.Chain) Option
```

**Example**:
```go
import "github.com/spcent/plumego/middleware"

chain := middleware.NewChain().
    Use(middleware.RequestID).
    Use(middleware.Logging).
    Use(customMiddleware)

app := core.New(
    core.WithMiddlewareChain(chain),
)
```

---

## Component Options

### WithComponent

Register a component.

```go
func WithComponent(component Component) Option
```

**Example**:
```go
import "github.com/spcent/plumego/core/components/websocket"

app := core.New(
    core.WithComponent(&websocket.Component{
        Secret: os.Getenv("WS_SECRET"),
    }),
)
```

**See**: [Components](components.md) for full documentation

---

### WithRunner

Register a background runner.

```go
func WithRunner(runner Runner) Option
```

**Example**:
```go
app := core.New(
    core.WithRunner(&BackgroundWorker{}),
)
```

**See**: [Runners](runners.md) for full documentation

---

## Feature Options

### WithDebug

Enable debug mode.

```go
func WithDebug(enabled ...bool) Option
```

**Example**:
```go
app := core.New(
    core.WithDebug(),       // Enable
    core.WithDebug(true),   // Enable
    core.WithDebug(false),  // Disable
)
```

**Default**: `false` (`APP_DEBUG`)

**Effects**:
- Verbose logging
- Detailed error messages
- Development endpoints

---

### WithTenantConfigManager

Set the tenant configuration manager for multi-tenancy.

```go
func WithTenantConfigManager(manager tenant.ConfigManager) Option
```

**Example**:
```go
import "github.com/spcent/plumego/tenant"

configMgr := tenant.NewInMemoryConfigManager()

app := core.New(
    core.WithTenantConfigManager(configMgr),
)
```

**See**: [Tenant Module](../tenant/) for details

---

### WithTenantMiddleware

Enable tenant middleware with options.

```go
func WithTenantMiddleware(opts TenantMiddlewareOptions) Option
```

**Example**:
```go
app := core.New(
    core.WithTenantMiddleware(core.TenantMiddlewareOptions{
        HeaderName:      "X-Tenant-ID",
        AllowMissing:    false,
        QuotaManager:    quotaMgr,
        PolicyEvaluator: policyEval,
    }),
)
```

---

### WithAIProvider

Set the AI provider for AI gateway features.

```go
func WithAIProvider(provider ai.Provider) Option
```

**Example**:
```go
import "github.com/spcent/plumego/ai/provider"

claude := provider.NewClaudeProvider(os.Getenv("ANTHROPIC_API_KEY"))

app := core.New(
    core.WithAIProvider(claude),
)
```

**See**: [AI Module](../ai/) for details

---

### WithSessionManager

Set the session manager for AI conversations.

```go
func WithSessionManager(manager ai.SessionManager) Option
```

**Example**:
```go
import "github.com/spcent/plumego/ai/session"

sessionMgr := session.NewInMemoryManager()

app := core.New(
    core.WithSessionManager(sessionMgr),
)
```

---

## Advanced Options

### WithRouter

Use a custom router.

```go
func WithRouter(router *router.Router) Option
```

**Example**:
```go
import "github.com/spcent/plumego/router"

customRouter := router.New(
    router.WithCaseSensitive(true),
    router.WithStrictSlash(false),
)

app := core.New(
    core.WithRouter(customRouter),
)
```

---

### WithDIContainer

Use a custom DI container.

```go
func WithDIContainer(container *di.Container) Option
```

**Example**:
```go
import "github.com/spcent/plumego/core/di"

container := di.New()
// Pre-register services
container.Register(&MyService{})

app := core.New(
    core.WithDIContainer(container),
)
```

---

## Presets

### Production Preset

```go
func ProductionConfig(addr string) []core.Option {
    return []core.Option{
        core.WithAddr(addr),
        core.WithServerTimeouts(30*time.Second, 30*time.Second, 60*time.Second, 10*time.Second),
        core.WithMaxBodyBytes(10 << 20),
        core.WithSecurityHeadersEnabled(true),
        core.WithAbuseGuardEnabled(true),
        core.WithRecommendedMiddleware(),
        core.WithCORS(cors.DefaultConfig()),
    }
}

app := core.New(ProductionConfig(":443")...)
```

### Development Preset

```go
func DevelopmentConfig(addr string) []core.Option {
    return []core.Option{
        core.WithAddr(addr),
        core.WithDebug(true),
        core.WithRecommendedMiddleware(),
        core.WithComponent(&devtools.Component{Enabled: true}),
    }
}

app := core.New(DevelopmentConfig(":8080")...)
```

### Testing Preset

```go
func TestConfig() []core.Option {
    return []core.Option{
        core.WithAddr(":0"), // Random port
        core.WithServerTimeouts(5*time.Second, 5*time.Second, 5*time.Second, 1*time.Second),
    }
}

app := core.New(TestConfig()...)
```

---

## Option Composition

### Combining Options

```go
baseOptions := []core.Option{
    core.WithAddr(":8080"),
    core.WithRecommendedMiddleware(),
}

tlsOptions := []core.Option{
    core.WithTLS("cert.pem", "key.pem"),
    core.WithHTTP2(true),
}

app := core.New(append(baseOptions, tlsOptions...)...)
```

### Conditional Options

```go
opts := []core.Option{
    core.WithAddr(":8080"),
}

if config.EnableTLS {
    opts = append(opts, core.WithTLS("cert.pem", "key.pem"))
}

if config.EnableMetrics {
    opts = append(opts, core.WithComponent(&observability.Component{}))
}

app := core.New(opts...)
```

### Environment-Based Configuration

```go
import "github.com/spcent/plumego/config"

cfg := config.Load()

opts := []core.Option{
    core.WithAddr(cfg.Get("APP_ADDR", ":8080")),
    core.WithDebug(cfg.GetBool("APP_DEBUG", false)),
    core.WithMaxBodyBytes(cfg.GetInt64("APP_MAX_BODY_BYTES", 10<<20)),
}

if cfg.GetBool("TLS_ENABLED", false) {
    opts = append(opts,
        core.WithTLS(
            cfg.Get("TLS_CERT_FILE", "cert.pem"),
            cfg.Get("TLS_KEY_FILE", "key.pem"),
        ),
    )
}

app := core.New(opts...)
```

---

## Complete Example

```go
package main

import (
    "os"
    "time"

    "github.com/spcent/plumego/config"
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/core/components/devtools"
    "github.com/spcent/plumego/core/components/observability"
    "github.com/spcent/plumego/core/components/websocket"
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
        core.WithMaxBodyBytes(10 << 20),
        core.WithMaxConcurrency(256),

        // TLS
        core.WithTLS(
            cfg.Get("TLS_CERT_FILE", "cert.pem"),
            cfg.Get("TLS_KEY_FILE", "key.pem"),
        ),
        core.WithHTTP2(true),

        // Security
        core.WithSecurityHeadersEnabled(true),
        core.WithAbuseGuardEnabled(true),

        // Middleware
        core.WithRecommendedMiddleware(),
        core.WithCORS(cors.Config{
            AllowOrigins: []string{cfg.Get("CORS_ORIGIN", "*")},
            AllowMethods: []string{"GET", "POST", "PUT", "DELETE"},
            AllowHeaders: []string{"Authorization", "Content-Type"},
            MaxAge:       3600,
        }),

        // Components
        core.WithComponent(&devtools.Component{
            Enabled: cfg.GetBool("APP_DEBUG", false),
        }),
        core.WithComponent(&observability.Component{
            MetricsEnabled: cfg.GetBool("METRICS_ENABLED", true),
            TracingEnabled: cfg.GetBool("TRACING_ENABLED", false),
        }),
        core.WithComponent(&websocket.Component{
            Secret: os.Getenv("WS_SECRET"),
        }),

        // Features
        core.WithDebug(cfg.GetBool("APP_DEBUG", false)),
    )

    app.Boot()
}
```

---

## Environment Variables Reference

| Variable | Default | Option |
|----------|---------|--------|
| `APP_ADDR` | `:8080` | `WithAddr` |
| `APP_DEBUG` | `false` | `WithDebug` |
| `APP_READ_TIMEOUT_MS` | `30000` | `WithServerTimeouts` |
| `APP_WRITE_TIMEOUT_MS` | `30000` | `WithServerTimeouts` |
| `APP_IDLE_TIMEOUT_MS` | `60000` | `WithServerTimeouts` |
| `APP_SHUTDOWN_TIMEOUT_MS` | `5000` | `WithShutdownTimeout` |
| `APP_MAX_BODY_BYTES` | `10485760` | `WithMaxBodyBytes` |
| `APP_MAX_CONCURRENCY` | `256` | `WithMaxConcurrency` |
| `APP_ENABLE_HTTP2` | `true` | `WithHTTP2` |
| `TLS_ENABLED` | `false` | - |
| `TLS_CERT_FILE` | - | `WithTLS` |
| `TLS_KEY_FILE` | - | `WithTLS` |

See `env.example` in the root directory for a complete list.

---

## Next Steps

- **[Application](application.md)** - Using options to configure your app
- **[Components](components.md)** - Built-in and custom components
- **[Middleware](../middleware/)** - Middleware configuration
- **[Config Module](../config/)** - Environment variable management

---

**Related**:
- [Application Configuration](application.md)
- [Component System](components.md)
- [Config Module](../config/)
