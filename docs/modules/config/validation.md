# Configuration Validation

> Legacy note: the historical public `config/` root has been removed.

Validation should happen once during startup, after defaults, env, and flags have been applied.

## Recommended Pattern

```go
func Validate(cfg Config) error {
    if cfg.Core.Addr == "" {
        return fmt.Errorf("addr is required")
    }
    if cfg.WebSocketSecret != "" && len(cfg.WebSocketSecret) < 32 {
        return fmt.Errorf("ws_secret must be at least 32 bytes")
    }
    if cfg.Core.Debug && cfg.EnableDocs == false {
        return nil
    }
    return nil
}
```

## Cross-Field Validation

```go
if cfg.EnableWebhooks && cfg.WebhookToken == "" {
    return fmt.Errorf("webhook token is required when webhooks are enabled")
}
```

## Guidance

- validate typed config, not raw `os.Getenv(...)` values spread through the codebase
- fail before app bootstrap
- keep security-sensitive checks explicit and easy to review
