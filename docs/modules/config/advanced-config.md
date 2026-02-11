# Advanced Configuration

> **Package**: `github.com/spcent/plumego/config`

Advanced configuration patterns and techniques.

---

## Struct-Based Config

```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
}

type ServerConfig struct {
    Addr         string
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
}

type DatabaseConfig struct {
    URL          string
    MaxConns     int
    ConnTimeout  time.Duration
}

func LoadAppConfig() Config {
    cfg := config.Load()

    return Config{
        Server: ServerConfig{
            Addr:         cfg.Get("APP_ADDR", ":8080"),
            ReadTimeout:  cfg.GetDuration("READ_TIMEOUT", 30*time.Second),
            WriteTimeout: cfg.GetDuration("WRITE_TIMEOUT", 30*time.Second),
        },
        Database: DatabaseConfig{
            URL:         cfg.MustGet("DB_URL"),
            MaxConns:    cfg.GetInt("DB_MAX_CONNS", 25),
            ConnTimeout: cfg.GetDuration("DB_TIMEOUT", 5*time.Second),
        },
    }
}
```

---

## Feature Flags

```go
type Features struct {
    EnableMetrics bool
    EnableTracing bool
    EnableCache   bool
}

func LoadFeatures(cfg *config.Config) Features {
    return Features{
        EnableMetrics: cfg.GetBool("FEATURE_METRICS", true),
        EnableTracing: cfg.GetBool("FEATURE_TRACING", false),
        EnableCache:   cfg.GetBool("FEATURE_CACHE", true),
    }
}

// Usage
features := LoadFeatures(cfg)
if features.EnableMetrics {
    // Enable metrics
}
```

---

## Environment-Specific Defaults

```go
func getDefaults(env string) map[string]string {
    defaults := map[string]string{
        "APP_ADDR": ":8080",
    }

    switch env {
    case "production":
        defaults["APP_DEBUG"] = "false"
        defaults["LOG_LEVEL"] = "info"
    case "development":
        defaults["APP_DEBUG"] = "true"
        defaults["LOG_LEVEL"] = "debug"
    }

    return defaults
}
```

---

## Configuration Reloading

```go
type ReloadableConfig struct {
    mu     sync.RWMutex
    config Config
}

func (rc *ReloadableConfig) Get() Config {
    rc.mu.RLock()
    defer rc.mu.RUnlock()
    return rc.config
}

func (rc *ReloadableConfig) Reload() {
    rc.mu.Lock()
    defer rc.mu.Unlock()
    rc.config = LoadAppConfig()
}

// Usage
rc := &ReloadableConfig{config: LoadAppConfig()}

// Reload on signal
go func() {
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGHUP)
    for range sigCh {
        rc.Reload()
        log.Println("Configuration reloaded")
    }
}()
```

---

**Next**: [Best Practices](best-practices.md)
