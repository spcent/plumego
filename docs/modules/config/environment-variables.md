# Environment Variables Reference

> **Package**: `github.com/spcent/plumego/config`

These are common application keys used in Plumego examples and app wiring. The config package does not auto-apply them to `core`; you read them explicitly and pass them into constructors/options.

## Common keys

```bash
APP_ADDR=:8080
APP_DEBUG=false
APP_SHUTDOWN_TIMEOUT_MS=5000
APP_READ_TIMEOUT_MS=30000
APP_WRITE_TIMEOUT_MS=30000
APP_IDLE_TIMEOUT_MS=60000
TLS_ENABLED=false
TLS_CERT_FILE=cert.pem
TLS_KEY_FILE=key.pem
APP_ENABLE_HTTP2=true
```

## Reading them explicitly

```go
cfg := config.NewManager(plumelog.NewGLogger())
_ = cfg.AddSource(config.NewEnvSource(""))
if err := cfg.Load(context.Background()); err != nil {
    log.Fatal(err)
}

app := core.New(
    core.WithAddr(cfg.GetString("app_addr", ":8080")),
    core.WithServerTimeouts(
        cfg.GetDurationMs("app_read_timeout_ms", 30000),
        5*time.Second,
        cfg.GetDurationMs("app_write_timeout_ms", 30000),
        cfg.GetDurationMs("app_idle_timeout_ms", 60000),
    ),
    core.WithShutdownTimeout(cfg.GetDurationMs("app_shutdown_timeout_ms", 5000)),
)

if cfg.GetBool("app_debug", false) {
    app = core.New(
        core.WithAddr(cfg.GetString("app_addr", ":8080")),
        core.WithDebug(),
    )
}
```

## Notes

- `config.NewEnvSource("")` normalizes keys to snake_case, so `APP_ADDR` becomes `app_addr`.
- Duration values ending in `_MS` are typically easier to read with `GetDurationMs(...)`.
- Secrets should come from real environment injection or external secret management, not committed files.
