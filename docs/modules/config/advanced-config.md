# Advanced Configuration

> **Package**: `github.com/spcent/plumego/config`

Advanced patterns in the current config package are built around explicit `Manager`, explicit `Source`, and explicit injection.

## Struct-based loading

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/spcent/plumego/config"
    plumelog "github.com/spcent/plumego/log"
)

type AppConfig struct {
    AppAddr       string        `config:"app_addr"`
    AppDebug      bool          `config:"app_debug"`
    ReadTimeout   time.Duration `config:"read_timeout"`
    WriteTimeout  time.Duration `config:"write_timeout"`
    DatabaseURL   string        `config:"db_url"`
    DatabaseConns int           `config:"db_max_conns"`
}

func main() {
    cfg := config.NewManager(plumelog.NewGLogger())
    _ = cfg.AddSource(config.NewEnvSource(""))
    _ = cfg.AddSource(config.NewFileSource(".env", config.FormatEnv, false))

    if err := cfg.Load(context.Background()); err != nil {
        log.Fatal(err)
    }

    var appConfig AppConfig
    if err := cfg.Unmarshal(&appConfig); err != nil {
        log.Fatal(err)
    }
}
```

## Feature flags

```go
func featureFlags(cfg *config.Manager) struct {
    Metrics bool
    Tracing bool
    Cache   bool
} {
    return struct {
        Metrics bool
        Tracing bool
        Cache   bool
    }{
        Metrics: cfg.GetBool("feature_metrics", true),
        Tracing: cfg.GetBool("feature_tracing", false),
        Cache:   cfg.GetBool("feature_cache", true),
    }
}
```

## Watching file-backed config

```go
source := config.NewFileSource(".env", config.FormatEnv, true).WithWatchInterval(2 * time.Second)
cfg := config.NewManager(plumelog.NewGLogger())

_ = cfg.AddSource(source)
_ = cfg.Load(context.Background())

_ = cfg.Watch("app_debug", func(oldValue, newValue any) {
    log.Printf("app_debug changed: %v -> %v", oldValue, newValue)
})

if err := cfg.StartWatchers(context.Background()); err != nil {
    log.Fatal(err)
}
```

## Source ordering

Sources are merged in registration order, later sources win.

```go
cfg := config.NewManager(plumelog.NewGLogger())
_ = cfg.AddSource(config.NewFileSource("base.env", config.FormatEnv, false))
_ = cfg.AddSource(config.NewFileSource("override.env", config.FormatEnv, false))
_ = cfg.AddSource(config.NewEnvSource(""))
```

This is the recommended way to model environment-specific overrides instead of hiding selection logic inside `config.Load()`.
