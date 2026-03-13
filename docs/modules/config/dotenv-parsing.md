# .env File Parsing

> Legacy note: the historical public `config/` root has been removed.

If an application wants `.env` support, keep it in app-owned startup code.

## Rule

Loading `.env` mutates process environment state. Do it explicitly in `main` or your app-local `internal/config`, never inside shared library bootstrap.

## Example

```go
func Load() (Config, error) {
    cfg := Defaults()

    envFile := cfg.Core.EnvFile
    if v := os.Getenv("APP_ENV_FILE"); v != "" {
        envFile = v
    }

    if envFile != "" {
        if _, err := os.Stat(envFile); err == nil {
            if err := loadEnvFile(envFile); err != nil {
                return cfg, err
            }
        }
    }

    return cfg, applyEnv(&cfg)
}
```

## File Conventions

```text
.env.example
.env.local
.env.test
```

Do not make library behavior depend on implicit file discovery.
