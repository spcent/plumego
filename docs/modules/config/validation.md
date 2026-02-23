# Configuration Validation

> **Package**: `github.com/spcent/plumego/config`

Patterns for validating configuration values.

---

## Required Variables

### MustGet

```go
// Panics if missing
apiKey := cfg.MustGet("API_KEY")
```

### Manual Check

```go
apiKey := cfg.Get("API_KEY", "")
if apiKey == "" {
    log.Fatal("API_KEY is required")
}
```

### Multiple Required

```go
required := []string{"API_KEY", "DB_URL", "JWT_SECRET"}
for _, key := range required {
    if cfg.Get(key, "") == "" {
        log.Fatalf("%s is required", key)
    }
}
```

---

## Range Validation

```go
port := cfg.GetInt("PORT", 8080)
if port < 1 || port > 65535 {
    log.Fatal("PORT must be 1-65535")
}

maxConns := cfg.GetInt("MAX_CONNECTIONS", 100)
if maxConns < 1 || maxConns > 1000 {
    log.Fatal("MAX_CONNECTIONS must be 1-1000")
}
```

---

## Format Validation

```go
import "regexp"

email := cfg.Get("ADMIN_EMAIL", "")
emailRegex := regexp.MustCompile(`^[a-z0-9._%+\-]+@[a-z0-9.\-]+\.[a-z]{2,}$`)
if !emailRegex.MatchString(email) {
    log.Fatal("Invalid ADMIN_EMAIL format")
}
```

---

## Validation Helper

```go
func validateConfig(cfg *config.Config) error {
    // Required
    if cfg.Get("API_KEY", "") == "" {
        return errors.New("API_KEY is required")
    }

    // Range
    port := cfg.GetInt("PORT", 8080)
    if port < 1 || port > 65535 {
        return fmt.Errorf("invalid PORT: %d", port)
    }

    return nil
}

// Usage
cfg := config.Load()
if err := validateConfig(cfg); err != nil {
    log.Fatal(err)
}
```

---

**Next**: [Defaults](defaults.md)
