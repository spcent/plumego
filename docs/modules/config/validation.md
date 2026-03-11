# Configuration Validation

> **Package**: `github.com/spcent/plumego/config`

Validation in the current package is explicit. Use manager accessors with validators or define a reusable schema.

## Per-value validation

```go
apiKey, err := cfg.String("api_key", "", &config.Required{}, &config.MinLength{Min: 20})
if err != nil {
    log.Fatal(err)
}

port, err := cfg.Int("port", 8080, &config.Range{Min: 1, Max: 65535})
if err != nil {
    log.Fatal(err)
}

baseURL, err := cfg.String("api_base_url", "", &config.URL{})
if err != nil {
    log.Fatal(err)
}
```

## Schema validation

```go
schema := config.NewConfigSchema().
    AddField("app_env", &config.OneOf{Values: []string{"development", "staging", "production"}}).
    AddField("jwt_secret", &config.Required{}, &config.MinLength{Min: 32}).
    AddField("port", &config.Range{Min: 1, Max: 65535})

if err := schema.Validate(cfg.GetAll()); err != nil {
    log.Fatal(err)
}
```

## Manual validation for cross-field rules

```go
func validate(cfg *config.Manager) error {
    if cfg.GetBool("tls_enabled", false) {
        if cfg.GetString("tls_cert_file", "") == "" || cfg.GetString("tls_key_file", "") == "" {
            return fmt.Errorf("tls_cert_file and tls_key_file are required when tls_enabled=true")
        }
    }
    return nil
}
```
