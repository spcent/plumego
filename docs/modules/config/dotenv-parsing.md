# .env File Parsing

> **Package**: `github.com/spcent/plumego/config`

Guide to `.env` file format and parsing.

---

## File Format

```bash
# Comments start with #
APP_ADDR=:8080
APP_DEBUG=true

# Quotes are optional
DB_URL="postgres://localhost/db"
API_KEY='secret-key'

# Empty values
OPTIONAL_KEY=

# Spaces around = are trimmed
PORT = 8080
```

---

## Loading

### Default Location

```go
// Loads .env from current directory
cfg := config.Load()
```

### Custom Location

```go
cfg := config.LoadFrom(".env.production")
cfg := config.LoadFrom("/etc/app/.env")
```

---

## Multiple Environments

```bash
.env                 # Default (committed)
.env.local          # Local overrides (gitignored)
.env.production     # Production
.env.staging        # Staging
.env.test           # Testing
```

### .gitignore

```bash
.env.local
.env.*.local
```

---

## Loading Priority

1. OS environment variables (highest)
2. .env file
3. Defaults in code (lowest)

```go
// If APP_ADDR is set in environment, it overrides .env
cfg := config.Load()
addr := cfg.Get("APP_ADDR", ":8080")
```

---

## Best Practices

### ✅ Do

1. **Commit .env.example**
   ```bash
   # .env.example (template)
   APP_ADDR=:8080
   DB_URL=postgres://localhost/db
   ```

2. **Ignore .env.local**
   ```bash
   # .gitignore
   .env.local
   ```

### ❌ Don't

1. **Don't Commit Secrets**
   ```bash
   # ❌ Don't commit
   .env.production  # Contains secrets
   ```

---

**Next**: [Validation](validation.md)
