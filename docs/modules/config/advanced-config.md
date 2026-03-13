# Advanced Configuration

> Legacy note: the historical public `config/` root has been removed.

Advanced configuration in Plumego is application-owned.

## Canonical Pattern

Keep configuration layering in one app-local package, usually `internal/config`:

1. start from explicit defaults
2. load optional `.env` or env vars
3. apply command-line flags
4. validate once
5. pass concrete values into constructors

## Recommended Layout

```text
cmd/myapp/main.go
internal/config/config.go
internal/config/env.go
```

## Example

```go
package config

import (
    "flag"
    "fmt"
    "os"

    "github.com/spcent/plumego/core"
)

type Config struct {
    Core core.AppConfig
    EnableDocs bool
}

func Defaults() Config {
    return Config{
        Core: core.AppConfig{
            Addr: ":8080",
            Debug: false,
        },
        EnableDocs: true,
    }
}

func Load(args []string) (Config, error) {
    cfg := Defaults()

    if v := os.Getenv("APP_ADDR"); v != "" {
        cfg.Core.Addr = v
    }
    if os.Getenv("APP_DEBUG") == "true" {
        cfg.Core.Debug = true
    }

    fs := flag.NewFlagSet("myapp", flag.ContinueOnError)
    fs.StringVar(&cfg.Core.Addr, "addr", cfg.Core.Addr, "listen address")
    fs.BoolVar(&cfg.Core.Debug, "debug", cfg.Core.Debug, "enable debug mode")
    fs.BoolVar(&cfg.EnableDocs, "enable-docs", cfg.EnableDocs, "enable docs")
    if err := fs.Parse(args); err != nil {
        return cfg, err
    }

    if cfg.Core.Addr == "" {
        return cfg, fmt.Errorf("addr is required")
    }
    return cfg, nil
}
```

## Within This Repository

Reference applications inside this repo may use `plumego/internal/config` as an implementation helper. External applications must not import Plumego `internal/*` packages.
