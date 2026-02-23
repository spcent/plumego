# Default Values

> **Package**: `github.com/spcent/plumego/config`

Managing default values and fallbacks.

---

## Basic Defaults

```go
addr := cfg.Get("APP_ADDR", ":8080")          // Default: :8080
debug := cfg.GetBool("APP_DEBUG", false)      // Default: false
timeout := cfg.GetDuration("TIMEOUT", 30*time.Second)  // Default: 30s
```

---

## Computed Defaults

```go
env := cfg.Get("ENV", "development")

var debug bool
if env == "production" {
    debug = false  // Production default
} else {
    debug = true   // Development default
}
```

---

## Centralized Defaults

```go
type AppConfig struct {
    Addr    string
    Debug   bool
    Timeout time.Duration
}

func LoadConfig() AppConfig {
    cfg := config.Load()

    return AppConfig{
        Addr:    cfg.Get("APP_ADDR", ":8080"),
        Debug:   cfg.GetBool("APP_DEBUG", false),
        Timeout: cfg.GetDuration("TIMEOUT", 30*time.Second),
    }
}
```

---

## Best Practices

### ✅ Do

1. **Always Provide Defaults**
   ```go
   addr := cfg.Get("APP_ADDR", ":8080")
   ```

2. **Use Sensible Defaults**
   ```go
   timeout := cfg.GetDuration("TIMEOUT", 30*time.Second)  // ✅ Reasonable
   ```

### ❌ Don't

1. **Don't Use Empty Defaults for Critical Values**
   ```go
   // ❌ Empty default for critical value
   apiKey := cfg.Get("API_KEY", "")

   // ✅ Require critical values
   apiKey := cfg.MustGet("API_KEY")
   ```

---

**Next**: [Advanced Config](advanced-config.md)
