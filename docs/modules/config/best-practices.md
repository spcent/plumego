# Configuration Best Practices

> Legacy note: the historical public `config/` root has been removed.

## Rules

1. Own configuration in the application, not in Plumego core packages.
2. Read configuration once near startup, then inject concrete values.
3. Keep defaults, env parsing, flags, and validation in one place.
4. Fail fast on invalid configuration.
5. Never log secrets or silently fall back for required credentials.

## Good Pattern

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

## Avoid

- scattered `os.Getenv(...)` calls across business code
- middleware reading config ad hoc from request context
- hidden `.env` loading inside library startup
- process-global config singletons in new application code
- empty-string defaults for required secrets
