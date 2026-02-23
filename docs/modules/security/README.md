# Security Module

> **Package Path**: `github.com/spcent/plumego/security` | **Stability**: Critical | **Priority**: P1

## Overview

The `security/` package provides essential cryptographic operations, authentication mechanisms, and security utilities for protecting Plumego applications. It implements industry-standard security practices and follows the principle of "fail closed" on verification errors.

**Security is NOT optional** — This module contains security-critical code that protects against common vulnerabilities including:
- Authentication bypass
- Password cracking
- Rate limiting abuse
- XSS, Clickjacking, MIME sniffing
- Email/URL injection attacks

### Subpackages

| Subpackage | Purpose | Key Features |
|------------|---------|--------------|
| **jwt/** | JWT token management | Sign, verify, refresh, key rotation |
| **password/** | Password operations | Bcrypt hashing, strength validation |
| **abuse/** | Anti-abuse protection | Rate limiting, IP blocking, anomaly detection |
| **headers/** | Security headers | CSP, HSTS, X-Frame-Options, etc. |
| **input/** | Input validation | Email, URL, phone number validation |

## Quick Start

### JWT Authentication

```go
import "github.com/spcent/plumego/security/jwt"

// Create JWT manager with secret
secret := []byte("your-secret-key-at-least-32-bytes")
manager := jwt.NewManager(secret)

// Issue access token
claims := jwt.Claims{
    Subject:   "user-123",
    ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
}
accessToken, err := manager.Issue(claims, jwt.TokenTypeAccess)
if err != nil {
    log.Fatal(err)
}

// Verify token
verified, err := manager.Verify(accessToken, jwt.TokenTypeAccess)
if err != nil {
    // Token invalid or expired
    http.Error(w, "Unauthorized", http.StatusUnauthorized)
    return
}

userID := verified.Subject
```

### Password Hashing

```go
import "github.com/spcent/plumego/security/password"

// Hash password during registration
hash, err := password.Hash("user-password")
if err != nil {
    log.Fatal(err)
}
// Store hash in database

// Verify password during login
ok := password.Verify("user-password", hash)
if !ok {
    http.Error(w, "Invalid credentials", http.StatusUnauthorized)
    return
}
```

### Abuse Guard

```go
import "github.com/spcent/plumego/security/abuse"

// Create abuse guard with rate limits
guard := abuse.NewGuard(abuse.Config{
    RequestsPerSecond: 10,
    BurstSize:         20,
    CleanupInterval:   time.Minute,
})

// Check if client is allowed
clientIP := "192.168.1.1"
if !guard.Allow(clientIP) {
    http.Error(w, "Too many requests", http.StatusTooManyRequests)
    return
}
```

### Security Headers

```go
import "github.com/spcent/plumego/security/headers"

// Apply security headers to response
policy := headers.DefaultPolicy()
headers.Apply(w, policy)

// Response will include:
// Content-Security-Policy: default-src 'self'
// X-Frame-Options: DENY
// X-Content-Type-Options: nosniff
// Referrer-Policy: strict-origin-when-cross-origin
```

### Input Validation

```go
import "github.com/spcent/plumego/security/input"

// Validate email
if !input.ValidateEmail("user@example.com") {
    return errors.New("invalid email format")
}

// Validate URL
if !input.ValidateURL("https://example.com") {
    return errors.New("invalid URL")
}

// Validate phone (E.164 format)
if !input.ValidatePhone("+12025551234") {
    return errors.New("invalid phone number")
}
```

## Core Principles

### 1. Defense in Depth

Plumego security employs multiple layers of protection:

- **Cryptographic layer**: JWT signatures, bcrypt hashing
- **Network layer**: Rate limiting, IP blocking
- **Application layer**: Input validation, security headers
- **Data layer**: Timing-safe comparisons, secure random generation

### 2. Fail Closed

**CRITICAL**: All security checks must fail closed:

```go
// ✅ Good: Fails closed
token, err := manager.Verify(tokenString, jwt.TokenTypeAccess)
if err != nil {
    return http.StatusUnauthorized, err
}
// Proceed with authenticated request

// ❌ Bad: Fails open
token, _ := manager.Verify(tokenString, jwt.TokenTypeAccess)
// Ignoring error allows invalid tokens through
```

### 3. Timing-Safe Operations

Use timing-safe comparisons for secrets to prevent timing attacks:

```go
import "crypto/subtle"

// ✅ Timing-safe comparison
if subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
    return errors.New("invalid signature")
}

// ❌ Vulnerable to timing attacks
if provided == expected {
    return errors.New("invalid signature")
}
```

### 4. Never Log Secrets

**NEVER** log sensitive data:

```go
// ❌ NEVER DO THIS
log.Printf("User token: %s", token)
log.Printf("Password: %s", password)
log.Printf("API key: %s", apiKey)

// ✅ Log safely
log.Printf("Token verification failed for user: %s", userID)
log.Printf("Invalid credentials for user: %s", username)
```

## Security Checklist

Use this checklist when implementing authentication/authorization:

- [ ] JWT tokens signed with strong secret (≥32 bytes)
- [ ] Access tokens have short expiry (≤15 minutes)
- [ ] Refresh tokens have longer expiry (7-30 days)
- [ ] Passwords hashed with bcrypt (cost ≥12)
- [ ] Rate limiting applied to login endpoints
- [ ] Security headers configured correctly
- [ ] User input validated before processing
- [ ] Secrets never logged or exposed in errors
- [ ] HTTPS enforced in production
- [ ] CORS configured with explicit origins

## Module Documentation

Detailed documentation for each subpackage:

- **[JWT Token Management](jwt.md)** — Signing, verification, refresh tokens, key rotation
- **[Password Security](password.md)** — Bcrypt hashing, strength validation, secure comparison
- **[Abuse Guard](abuse.md)** — Rate limiting, IP blocking, anomaly detection
- **[Security Headers](headers.md)** — CSP, HSTS, XSS protection, frame options
- **[Input Validation](input.md)** — Email, URL, phone number validation
- **[Best Practices](best-practices.md)** — OWASP Top 10, security patterns, common pitfalls

## Integration with Middleware

Security features integrate with Plumego middleware:

```go
import (
    "github.com/spcent/plumego/core"
    "github.com/spcent/plumego/middleware/auth"
    "github.com/spcent/plumego/middleware/security"
    "github.com/spcent/plumego/security/jwt"
)

secret := []byte("your-secret-key")
manager := jwt.NewManager(secret)

app := core.New(
    // Security headers middleware
    core.WithSecurityHeadersEnabled(true),

    // Abuse guard (rate limiting)
    core.WithAbuseGuardEnabled(true),

    // JWT authentication middleware
    core.WithMiddleware(auth.JWT(manager)),
)
```

## Performance Considerations

### JWT Verification

JWT verification is cryptographically expensive. Cache verified tokens:

```go
type TokenCache struct {
    cache map[string]*jwt.Claims
    mu    sync.RWMutex
}

func (c *TokenCache) Verify(token string, manager *jwt.Manager) (*jwt.Claims, error) {
    c.mu.RLock()
    if claims, ok := c.cache[token]; ok {
        c.mu.RUnlock()
        return claims, nil
    }
    c.mu.RUnlock()

    // Not cached, verify
    claims, err := manager.Verify(token, jwt.TokenTypeAccess)
    if err != nil {
        return nil, err
    }

    c.mu.Lock()
    c.cache[token] = claims
    c.mu.Unlock()

    return claims, nil
}
```

### Bcrypt Hashing

Bcrypt is intentionally slow. Use background workers for registration:

```go
type RegistrationJob struct {
    Email    string
    Password string
}

func processRegistration(job RegistrationJob) error {
    // Hash in background worker, not request path
    hash, err := password.Hash(job.Password)
    if err != nil {
        return err
    }

    return db.CreateUser(job.Email, hash)
}
```

## Common Vulnerabilities

### ❌ Hardcoded Secrets

```go
// NEVER do this
secret := []byte("my-secret-key")
manager := jwt.NewManager(secret)
```

### ✅ Use Environment Variables

```go
import "github.com/spcent/plumego/config"

cfg := config.Load()
secret := []byte(cfg.Get("JWT_SECRET", ""))
if len(secret) < 32 {
    log.Fatal("JWT_SECRET must be at least 32 bytes")
}
manager := jwt.NewManager(secret)
```

### ❌ Weak Password Requirements

```go
// Too weak
if len(password) < 6 {
    return errors.New("password too short")
}
```

### ✅ Use Password Validator

```go
import "github.com/spcent/plumego/security/password"

if !password.IsStrong(password) {
    return errors.New("password must be at least 8 characters with mixed case and numbers")
}
```

## Testing Security Code

### Unit Testing

```go
func TestJWTVerification(t *testing.T) {
    secret := []byte("test-secret-at-least-32-bytes-long")
    manager := jwt.NewManager(secret)

    claims := jwt.Claims{Subject: "user-123"}
    token, err := manager.Issue(claims, jwt.TokenTypeAccess)
    if err != nil {
        t.Fatal(err)
    }

    // Verify valid token
    verified, err := manager.Verify(token, jwt.TokenTypeAccess)
    if err != nil {
        t.Fatalf("expected valid token, got error: %v", err)
    }
    if verified.Subject != "user-123" {
        t.Errorf("expected subject user-123, got %s", verified.Subject)
    }

    // Verify invalid token
    _, err = manager.Verify("invalid-token", jwt.TokenTypeAccess)
    if err == nil {
        t.Error("expected error for invalid token")
    }
}
```

### Security Testing

Test negative cases:

```go
func TestSecurityNegativeCases(t *testing.T) {
    tests := []struct {
        name  string
        input string
        valid bool
    }{
        {"Valid email", "user@example.com", true},
        {"Missing @", "userexample.com", false},
        {"XSS attempt", "<script>alert('xss')</script>@example.com", false},
        {"SQL injection", "'; DROP TABLE users--@example.com", false},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            valid := input.ValidateEmail(tt.input)
            if valid != tt.valid {
                t.Errorf("ValidateEmail(%q) = %v, want %v", tt.input, valid, tt.valid)
            }
        })
    }
}
```

## Related Documentation

- [Middleware: Authentication](../middleware/auth.md) — JWT middleware integration
- [Middleware: Security Headers](../middleware/security.md) — Security headers middleware
- [Middleware: Rate Limiting](../middleware/ratelimit.md) — Rate limiting patterns
- [Best Practices Guide](best-practices.md) — Security best practices and OWASP Top 10
- [SECURITY.md](../../../SECURITY.md) — Security policy and vulnerability disclosure

## Reference Implementation

See complete working examples:

- `examples/reference/` — Full-featured app with JWT auth
- `examples/multi-tenant-saas/` — Multi-tenant security patterns
- `examples/api-gateway/` — API gateway security

---

**Next Steps**:
1. Read [JWT Token Management](jwt.md) for authentication setup
2. Read [Password Security](password.md) for user credential handling
3. Read [Best Practices](best-practices.md) for security guidelines
