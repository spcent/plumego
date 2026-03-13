# Migration Guide

> Legacy note: the historical public `config/` root has been removed.

This guide is for migrating from the old Plumego `config/` root to the current app-owned model.

## Target State

- application config lives in your own `internal/config`
- startup parses env, optional `.env`, and flags explicitly
- Plumego receives concrete values through constructors and options

## Replace This

The removed public config root and its package-global helpers.

## With This

```text
cmd/myapp/main.go
internal/config/config.go
```

## Migration Steps

1. Create an app-local `Config` struct.
2. Move defaults into `Defaults()`.
3. Replace package-global reads with explicit env and flag parsing.
4. Add `Validate(cfg)` for required values and cross-field checks.
5. Pass resolved values into `core.New(...)` and `x/*` constructors.

## Example

Before:

```go
cfg := loadConfigFromApp()
addr := cfg.Core.Addr
debug := cfg.Core.Debug
```

After:

```go
cfg, err := appcfg.Load(os.Args[1:])
if err != nil {
    log.Fatal(err)
}

app := core.New(
    core.WithAddr(cfg.Core.Addr),
)
if cfg.Core.Debug {
    app = core.New(
        core.WithAddr(cfg.Core.Addr),
        core.WithDebug(),
    )
}
```

## Notes For Repo Contributors

Code inside this repository may use `plumego/internal/config` where appropriate. That is an internal implementation detail, not a public import path for external users.
