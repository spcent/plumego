# Environment Variables Reference

> Legacy note: the historical public `config/` root has been removed.

These are common application keys used by Plumego reference apps and examples.

## Common Keys

```text
APP_ADDR
APP_DEBUG
APP_ENV_FILE
WS_SECRET
GITHUB_WEBHOOK_SECRET
STRIPE_WEBHOOK_SECRET
WEBHOOK_TRIGGER_TOKEN
ENABLE_DOCS
ENABLE_METRICS
ENABLE_WEBHOOKS
```

## Mapping Pattern

```go
func applyEnv(cfg *Config) {
    if v := os.Getenv("APP_ADDR"); v != "" {
        cfg.Core.Addr = v
    }
    if os.Getenv("APP_DEBUG") == "true" {
        cfg.Core.Debug = true
    }
    if v := os.Getenv("WS_SECRET"); v != "" {
        cfg.WebSocketSecret = v
    }
}
```

## Guidance

- use env vars as inputs to your app config, not as hidden framework globals
- keep names stable and uppercase with underscores
- parse once, then pass typed values into `core.With...` or extension constructors
