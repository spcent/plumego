# Environment Variables Reference

> **Package**: `github.com/spcent/plumego/config`

Complete reference for all standard Plumego environment variables.

---

## Server Configuration

```bash
# Listen address
APP_ADDR=:8080

# Debug mode
APP_DEBUG=false

# Shutdown timeout (milliseconds)
APP_SHUTDOWN_TIMEOUT_MS=5000
```

---

## HTTP Timeouts

```bash
# Read timeout (milliseconds)
APP_READ_TIMEOUT_MS=30000

# Write timeout (milliseconds)
APP_WRITE_TIMEOUT_MS=30000

# Idle timeout (milliseconds)
APP_IDLE_TIMEOUT_MS=60000
```

---

## Request Limits

```bash
# Max request body size (bytes)
APP_MAX_BODY_BYTES=10485760

# Max concurrent requests
APP_MAX_CONCURRENCY=256
```

---

## TLS Configuration

```bash
# Enable TLS
TLS_ENABLED=false

# Certificate file path
TLS_CERT_FILE=cert.pem

# Key file path
TLS_KEY_FILE=key.pem

# Enable HTTP/2
APP_ENABLE_HTTP2=true
```

---

## Database

```bash
# Database URL
DB_URL=postgres://user:pass@localhost:5432/dbname

# Connection pool size
DB_MAX_CONNECTIONS=25

# Connection timeout
DB_CONNECT_TIMEOUT_MS=5000
```

---

## Security

```bash
# JWT secret (min 32 bytes)
JWT_SECRET=your-secret-key-here

# JWT expiration
JWT_EXPIRATION=24h

# WebSocket secret
WS_SECRET=your-ws-secret-here
```

---

## Webhooks

```bash
# GitHub webhook secret
GITHUB_WEBHOOK_SECRET=your-github-secret

# Stripe webhook secret
STRIPE_WEBHOOK_SECRET=your-stripe-secret

# Webhook queue size
WEBHOOK_QUEUE_SIZE=2048

# Webhook workers
WEBHOOK_WORKERS=8
```

---

## Usage

```go
cfg := config.Load()

app := core.New(
    core.WithAddr(cfg.Get("APP_ADDR", ":8080")),
    core.WithServerTimeouts(
        cfg.GetDuration("APP_READ_TIMEOUT", 30*time.Second),
        cfg.GetDuration("APP_WRITE_TIMEOUT", 30*time.Second),
        cfg.GetDuration("APP_IDLE_TIMEOUT", 60*time.Second),
        cfg.GetDuration("APP_SHUTDOWN_TIMEOUT", 5*time.Second),
    ),
    core.WithMaxBodyBytes(cfg.GetInt64("APP_MAX_BODY_BYTES", 10<<20)),
)
```

---

**Next**: [.env Parsing](dotenv-parsing.md)
