# Config Module

> **Package**: `github.com/spcent/plumego/config` | **Stability**: High

## Overview

The current `config` package is built around explicit `Manager + Source` composition.

Recommended path:

- create a `config.Manager`
- add one or more sources explicitly
- call `Load(ctx)`
- inject the manager into the parts of the app that need configuration

The package also still exposes global convenience helpers, but those are not the preferred application wiring style.

## Canonical Usage

```go
package main

import (
    "context"
    "log"
    "time"

    "github.com/spcent/plumego/config"
    plumelog "github.com/spcent/plumego/log"
)

func main() {
    cfg := config.NewManager(plumelog.NewGLogger())

    if err := cfg.AddSource(config.NewEnvSource("")); err != nil {
        log.Fatal(err)
    }
    if err := cfg.AddSource(config.NewFileSource(".env", config.FormatEnv, false)); err != nil {
        log.Fatal(err)
    }

    if err := cfg.Load(context.Background()); err != nil {
        log.Fatal(err)
    }

    _ = cfg.GetString("app_addr", ":8080")
    _ = cfg.GetBool("app_debug", false)
    _ = cfg.GetDuration("app_shutdown_timeout", 5*time.Second)
}
```

## Sources

### Environment variables

```go
source := config.NewEnvSource("")
```

If a prefix is provided, matching variables are loaded and the prefix is stripped:

```go
source := config.NewEnvSource("APP_")
```

### File source

```go
source := config.NewFileSource("config.json", config.FormatJSON, false)
```

Supported formats today:

- `config.FormatEnv`
- `config.FormatJSON`

`config.FormatTOML` exists as a constant but is not implemented yet.

### File watching

```go
source := config.NewFileSource(".env", config.FormatEnv, true).WithWatchInterval(2 * time.Second)
```

Then start watchers explicitly:

```go
if err := cfg.StartWatchers(context.Background()); err != nil {
    log.Fatal(err)
}
```

## Access Patterns

### Simple getters

```go
addr := cfg.GetString("app_addr", ":8080")
debug := cfg.GetBool("app_debug", false)
workers := cfg.GetInt("worker_count", 4)
ratio := cfg.GetFloat("sample_ratio", 0.1)
timeout := cfg.GetDuration("app_timeout", 30*time.Second)
```

### Validated accessors

```go
port, err := cfg.Int("port", 8080, &config.Range{Min: 1, Max: 65535})
if err != nil {
    log.Fatal(err)
}
```

### Decode into a struct

```go
type AppConfig struct {
    AppAddr string `config:"app_addr"`
    Debug   bool   `config:"app_debug"`
}

var appConfig AppConfig
if err := cfg.Unmarshal(&appConfig); err != nil {
    log.Fatal(err)
}
```

## Explicit `.env` loading

If you need to mutate process environment variables from a file, do it explicitly:

```go
if err := config.LoadEnvFile(".env", false); err != nil {
    log.Fatal(err)
}
```

That operation writes to process-global environment state. Keep it in `main`, not inside library startup code.

## Global helpers

Package-level helpers still exist:

- `config.GetString(...)`
- `config.GetBool(...)`
- `config.GetInt(...)`
- `config.GetDuration(...)`
- `config.InitDefault()`

They are convenience APIs backed by a global manager. Prefer instance-based wiring for explicit DI.
