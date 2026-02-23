# Configuration Best Practices

> **Package**: `github.com/spcent/plumego/config`

Best practices for managing application configuration.

---

## General Principles

### ✅ Do

1. **Use Environment Variables for Secrets**
   ```go
   // ✅ From environment
   secret := cfg.MustGet("JWT_SECRET")

   // ❌ Hardcoded
   secret := "my-secret-key"
   ```

2. **Provide Sensible Defaults**
   ```go
   addr := cfg.Get("APP_ADDR", ":8080")
   timeout := cfg.GetDuration("TIMEOUT", 30*time.Second)
   ```

3. **Validate Early**
   ```go
   cfg := config.Load()
   if err := validateConfig(cfg); err != nil {
       log.Fatal(err)
   }
   ```

4. **Document Variables**
   ```bash
   # env.example
   APP_ADDR=:8080  # Server listen address
   DB_URL=postgres://localhost/db  # Database connection URL
   ```

---

## Security

### ✅ Do

1. **Never Commit Secrets**
   ```bash
   # .gitignore
   .env.local
   .env.production
   .env.*.local
   ```

2. **Use Strong Secrets**
   ```bash
   # ✅ Strong (32+ bytes)
   JWT_SECRET=long-random-string-with-32-plus-characters

   # ❌ Weak
   JWT_SECRET=secret
   ```

3. **Rotate Secrets Regularly**
   ```bash
   # Production: Use secret management service
   # AWS Secrets Manager, HashiCorp Vault, etc.
   ```

---

## Organization

### ✅ Do

1. **Group Related Variables**
   ```bash
   # Server
   APP_ADDR=:8080
   APP_DEBUG=false

   # Database
   DB_URL=postgres://localhost/db
   DB_MAX_CONNS=25

   # Cache
   REDIS_URL=redis://localhost:6379
   REDIS_POOL_SIZE=10
   ```

2. **Use Consistent Naming**
   ```bash
   # ✅ Consistent prefix
   APP_ADDR=:8080
   APP_DEBUG=false
   APP_TIMEOUT=30s

   # ❌ Inconsistent
   ADDR=:8080
   DEBUG_MODE=false
   TIMEOUT_SECONDS=30
   ```

---

## Development vs Production

### ✅ Do

1. **Different Configs per Environment**
   ```bash
   # .env (development)
   APP_DEBUG=true
   DB_URL=postgres://localhost/dev

   # .env.production
   APP_DEBUG=false
   DB_URL=postgres://prod-server/prod
   ```

2. **Use ENV Variable to Switch**
   ```go
   env := os.Getenv("ENV")
   if env == "" {
       env = "development"
   }

   var cfg *config.Config
   if env == "production" {
       cfg = config.LoadFrom(".env.production")
   } else {
       cfg = config.Load()
   }
   ```

---

## Anti-Patterns

### ❌ Don't

1. **Don't Use Magic Numbers**
   ```go
   // ❌ Magic number
   timeout := 30 * time.Second

   // ✅ From config
   timeout := cfg.GetDuration("TIMEOUT", 30*time.Second)
   ```

2. **Don't Scatter Config Access**
   ```go
   // ❌ Scattered throughout code
   func handler1() {
       addr := os.Getenv("APP_ADDR")
   }
   func handler2() {
       addr := os.Getenv("APP_ADDR")
   }

   // ✅ Centralized
   type AppConfig struct {
       Addr string
   }
   config := LoadAppConfig()
   ```

3. **Don't Ignore Validation Errors**
   ```go
   // ❌ Silent failure
   port, _ := strconv.Atoi(cfg.Get("PORT", "8080"))

   // ✅ Handle errors
   port := cfg.GetInt("PORT", 8080)
   if port < 1 || port > 65535 {
       log.Fatal("Invalid PORT")
   }
   ```

---

## Summary

**Key Rules**:
1. Environment variables for secrets
2. Sensible defaults for everything
3. Validate early, fail fast
4. Never commit secrets
5. Document all variables
6. Group related config
7. Use consistent naming

---

**Next**: [Migration Guide](migration-guide.md)
