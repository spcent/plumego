# Default Values

> **Package**: `github.com/spcent/plumego/config`

Defaults are supplied explicitly by the caller.

## Basic defaults

```go
addr := cfg.GetString("app_addr", ":8080")
debug := cfg.GetBool("app_debug", false)
timeout := cfg.GetDuration("app_timeout", 30*time.Second)
```

## Derived defaults

```go
env := cfg.GetString("app_env", "development")
writeTimeout := 30 * time.Second
if env == "production" {
    writeTimeout = 15 * time.Second
}
```

## Centralize defaults near startup

```go
type AppConfig struct {
    Addr    string
    Debug   bool
    Timeout time.Duration
}

func loadAppConfig(cfg *config.Manager) AppConfig {
    return AppConfig{
        Addr:    cfg.GetString("app_addr", ":8080"),
        Debug:   cfg.GetBool("app_debug", false),
        Timeout: cfg.GetDuration("app_timeout", 30*time.Second),
    }
}
```

## Guidance

- Use defaults for optional behavior.
- Use validators for required values.
- Keep defaults in one place so startup behavior is easy to inspect.
