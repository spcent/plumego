# Default Values

> Legacy note: the historical public `config/` root has been removed.

Defaults should be explicit, local, and reviewable.

## Recommended Pattern

```go
func Defaults() Config {
    return Config{
        Core: core.AppConfig{
            Addr: ":8080",
            Debug: false,
            EnvFile: ".env",
        },
        EnableDocs: true,
        EnableMetrics: true,
    }
}
```

## Guidance

- keep defaults in a single `Defaults()` function
- keep derived defaults near startup
- use defaults for optional behavior only
- validate required values after env and flag overrides

## Derived Defaults

```go
func ApplyEnvironmentDefaults(cfg *Config, env string) {
    if env == "production" {
        cfg.EnableDocs = false
    }
}
```
