# Configuration Best Practices

> **Package**: `github.com/spcent/plumego/config`

## Recommended rules

1. Build configuration with `config.Manager`, not implicit globals.
2. Keep `.env` loading explicit in `main`.
3. Validate configuration before starting the app.
4. Inject config into constructors; do not fetch config ad hoc from random call sites.
5. Never commit secrets.

## Good patterns

### Instance-based config wiring

```go
cfg := config.NewManager(plumelog.NewGLogger())
_ = cfg.AddSource(config.NewEnvSource(""))
if err := cfg.Load(context.Background()); err != nil {
    log.Fatal(err)
}

svc := service.New(service.Config{
    BaseURL: cfg.GetString("api_base_url", "http://localhost:8080"),
    Timeout: cfg.GetDuration("api_timeout", 5*time.Second),
})
```

### Early validation

```go
port, err := cfg.Int("port", 8080, &config.Range{Min: 1, Max: 65535})
if err != nil {
    log.Fatal(err)
}
_ = port
```

### Explicit secret handling

```go
secret, err := cfg.String("jwt_secret", "", &config.Required{}, &config.MinLength{Min: 32})
if err != nil {
    log.Fatal(err)
}
_ = secret
```

## Avoid

### Hidden global reads everywhere

```go
// Avoid spreading package-level config lookups through business code.
// Prefer reading once at the edge and injecting the result.
```

### Mixing environment mutation with library startup

```go
// Do not hide config.LoadEnvFile(...) inside app bootstrap helpers.
```

### Empty defaults for required values

```go
// Prefer validated accessors over silently accepting empty strings.
```
