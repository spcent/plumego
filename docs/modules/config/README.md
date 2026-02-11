# Config Module

> **Package**: `github.com/spcent/plumego/config`
> **Stability**: High - Configuration API is stable
> **Go Version**: 1.24+

The `config` package provides environment variable loading, `.env` file parsing, and configuration validation.

---

## Overview

The config module provides:

- **Environment Variables**: Load from OS environment
- **.env Files**: Parse `.env` files (development)
- **Type Conversion**: String, int, bool, float, duration
- **Default Values**: Fallback when variable missing
- **Validation**: Required variables and format validation
- **Zero Dependencies**: Uses only standard library

---

## Quick Start

### Basic Usage

```go
import "github.com/spcent/plumego/config"

// Load configuration
cfg := config.Load()

// Get values with defaults
addr := cfg.Get("APP_ADDR", ":8080")
debug := cfg.GetBool("APP_DEBUG", false)
timeout := cfg.GetDuration("APP_TIMEOUT", 30*time.Second)
```

### With Application

```go
package main

import (
    "github.com/spcent/plumego/config"
    "github.com/spcent/plumego/core"
)

func main() {
    cfg := config.Load()

    app := core.New(
        core.WithAddr(cfg.Get("APP_ADDR", ":8080")),
        core.WithDebug(cfg.GetBool("APP_DEBUG", false)),
    )

    app.Boot()
}
```

---

## Core Functions

### Load

Load configuration from environment and `.env` file:

```go
cfg := config.Load()
```

**Search order**:
1. `.env` file in current directory
2. OS environment variables (override `.env`)

### LoadFrom

Load from specific `.env` file:

```go
cfg := config.LoadFrom(".env.production")
```

### Get Methods

```go
// String
value := cfg.Get("KEY", "default")

// Int
port := cfg.GetInt("PORT", 8080)

// Int64
size := cfg.GetInt64("MAX_SIZE", 10485760)

// Bool
enabled := cfg.GetBool("ENABLED", false)

// Float
rate := cfg.GetFloat("RATE", 1.5)

// Duration
timeout := cfg.GetDuration("TIMEOUT", 30*time.Second)

// Required (panics if missing)
apiKey := cfg.MustGet("API_KEY")
```

---

## Environment Variables

### Standard Variables

See `env.example` for complete list:

```bash
# Server
APP_ADDR=:8080
APP_DEBUG=false
APP_SHUTDOWN_TIMEOUT_MS=5000

# Timeouts
APP_READ_TIMEOUT_MS=30000
APP_WRITE_TIMEOUT_MS=30000
APP_IDLE_TIMEOUT_MS=60000

# Limits
APP_MAX_BODY_BYTES=10485760
APP_MAX_CONCURRENCY=256

# TLS
TLS_ENABLED=false
TLS_CERT_FILE=cert.pem
TLS_KEY_FILE=key.pem

# Features
APP_ENABLE_HTTP2=true
```

### Loading in Application

```go
cfg := config.Load()

app := core.New(
    core.WithAddr(cfg.Get("APP_ADDR", ":8080")),
    core.WithServerTimeouts(
        cfg.GetDuration("APP_READ_TIMEOUT", 30*time.Second),
        cfg.GetDuration("APP_WRITE_TIMEOUT", 30*time.Second),
        cfg.GetDuration("APP_IDLE_TIMEOUT", 60*time.Second),
        cfg.GetDuration("APP_SHUTDOWN_TIMEOUT", 5*time.Second),
    ),
    core.WithMaxBodyBytes(cfg.GetInt64("APP_MAX_BODY_BYTES", 10<<20)),
    core.WithDebug(cfg.GetBool("APP_DEBUG", false)),
)
```

---

## .env Files

### Format

```bash
# Comments
APP_ADDR=:8080
APP_DEBUG=true

# Quotes (optional)
DB_URL="postgres://localhost/db"

# Multiline (not supported)
# Use environment variables instead

# Empty values
OPTIONAL_KEY=
```

### Multiple Environments

```bash
# Development
.env                 # Local development
.env.local          # Local overrides (gitignored)

# Testing
.env.test

# Production
.env.production     # Production values
```

### Loading Priority

1. **Highest**: OS environment variables
2. **Medium**: `.env.local` (gitignored)
3. **Lowest**: `.env` (committed)

---

## Type Conversion

### String

```go
value := cfg.Get("KEY", "default")
```

### Integer

```go
port := cfg.GetInt("PORT", 8080)
```

Supports: `"8080"` → `8080`

### Boolean

```go
enabled := cfg.GetBool("ENABLED", false)
```

Supports:
- `true`: `"true"`, `"1"`, `"yes"`, `"on"`
- `false`: `"false"`, `"0"`, `"no"`, `"off"`, `""`

### Float

```go
rate := cfg.GetFloat("RATE", 1.5)
```

Supports: `"1.5"` → `1.5`

### Duration

```go
timeout := cfg.GetDuration("TIMEOUT", 30*time.Second)
```

Supports:
- Milliseconds: `"5000"` → `5s`
- Duration strings: `"30s"`, `"5m"`, `"1h"`

---

## Validation

### Required Variables

```go
// Panic if missing
apiKey := cfg.MustGet("API_KEY")

// Check multiple required
cfg.RequireAll("API_KEY", "DB_URL", "SECRET_KEY")
```

### Custom Validation

```go
port := cfg.GetInt("PORT", 8080)
if port < 1 || port > 65535 {
    log.Fatal("Invalid port")
}

dbURL := cfg.Get("DB_URL", "")
if dbURL == "" {
    log.Fatal("DB_URL is required")
}
```

---

## Complete Examples

### Basic Application

```go
package main

import (
    "log"
    "github.com/spcent/plumego/config"
    "github.com/spcent/plumego/core"
)

func main() {
    // Load config
    cfg := config.Load()

    // Create app with config
    app := core.New(
        core.WithAddr(cfg.Get("APP_ADDR", ":8080")),
        core.WithDebug(cfg.GetBool("APP_DEBUG", false)),
    )

    // Routes
    app.Get("/", homeHandler)

    // Start
    log.Printf("Starting server on %s", cfg.Get("APP_ADDR", ":8080"))
    app.Boot()
}
```

### With Validation

```go
package main

import (
    "log"
    "github.com/spcent/plumego/config"
    "github.com/spcent/plumego/core"
)

func main() {
    cfg := config.Load()

    // Validate required variables
    required := []string{
        "APP_ADDR",
        "DB_URL",
        "JWT_SECRET",
    }

    for _, key := range required {
        if cfg.Get(key, "") == "" {
            log.Fatalf("Required variable %s is not set", key)
        }
    }

    // Validate ranges
    port := cfg.GetInt("APP_PORT", 8080)
    if port < 1 || port > 65535 {
        log.Fatal("APP_PORT must be between 1 and 65535")
    }

    // Create app
    app := core.New(
        core.WithAddr(cfg.Get("APP_ADDR", ":8080")),
    )

    app.Boot()
}
```

### Environment-Specific Config

```go
package main

import (
    "github.com/spcent/plumego/config"
    "github.com/spcent/plumego/core"
)

func main() {
    // Load environment-specific config
    env := os.Getenv("ENV")
    if env == "" {
        env = "development"
    }

    var cfg *config.Config
    switch env {
    case "production":
        cfg = config.LoadFrom(".env.production")
    case "staging":
        cfg = config.LoadFrom(".env.staging")
    default:
        cfg = config.Load() // .env
    }

    app := core.New(
        core.WithAddr(cfg.Get("APP_ADDR", ":8080")),
        core.WithDebug(cfg.GetBool("APP_DEBUG", false)),
    )

    app.Boot()
}
```

---

## Best Practices

### ✅ Do

1. **Use Defaults**
   ```go
   // ✅ With default
   addr := cfg.Get("APP_ADDR", ":8080")

   // ❌ No default (empty string if missing)
   addr := cfg.Get("APP_ADDR", "")
   ```

2. **Validate Critical Variables**
   ```go
   apiKey := cfg.Get("API_KEY", "")
   if apiKey == "" {
       log.Fatal("API_KEY is required")
   }
   ```

3. **Use Type-Safe Methods**
   ```go
   // ✅ Type-safe
   port := cfg.GetInt("PORT", 8080)

   // ❌ Manual conversion
   portStr := cfg.Get("PORT", "8080")
   port, _ := strconv.Atoi(portStr)
   ```

4. **Document Variables**
   ```go
   // See env.example for all available variables
   ```

### ❌ Don't

1. **Don't Hardcode Secrets**
   ```go
   // ❌ Hardcoded
   secret := "my-secret-key"

   // ✅ From environment
   secret := cfg.MustGet("SECRET_KEY")
   ```

2. **Don't Commit .env.local**
   ```bash
   # .gitignore
   .env.local
   .env.*.local
   ```

---

## Documentation Structure

| Document | Description |
|----------|-------------|
| **[Environment Variables](environment-variables.md)** | Standard variables and usage |
| **[.env Parsing](dotenv-parsing.md)** | .env file format and loading |
| **[Validation](validation.md)** | Configuration validation |
| **[Defaults](defaults.md)** | Default values and fallbacks |
| **[Advanced Config](advanced-config.md)** | Advanced patterns |
| **[Best Practices](best-practices.md)** | Configuration best practices |
| **[Migration Guide](migration-guide.md)** | Migrating from other systems |

---

## Next Steps

- **[Environment Variables](environment-variables.md)** - Complete variable reference
- **[.env Parsing](dotenv-parsing.md)** - .env file details
- **[Validation](validation.md)** - Validation patterns

---

## Related Modules

- **[Core](../core/)** - Application configuration
- **[Security](../security/)** - Secret management

---

**Stability**: High - Breaking changes require major version bump
**Maintainers**: Plumego Core Team
**Last Updated**: 2026-02-11
