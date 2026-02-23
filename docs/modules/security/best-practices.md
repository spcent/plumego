# Security Best Practices

> **Package**: `github.com/spcent/plumego/security` | **Priority**: Critical

## Overview

This guide covers security best practices for Plumego applications, including OWASP Top 10 vulnerabilities, secure coding patterns, and production security checklist.

**Security is NOT optional** — Every application handling user data or operating on the internet must implement these practices.

## Security Principles

### 1. Defense in Depth

Implement multiple layers of security:

```
┌─────────────────────────────────────┐
│ Network Layer                       │ ← Firewall, DDoS protection
├─────────────────────────────────────┤
│ Application Layer                   │ ← Rate limiting, authentication
├─────────────────────────────────────┤
│ Input Validation Layer              │ ← Sanitization, validation
├─────────────────────────────────────┤
│ Business Logic Layer                │ ← Authorization, access control
├─────────────────────────────────────┤
│ Data Layer                          │ ← Encryption, SQL injection prevention
└─────────────────────────────────────┘
```

### 2. Principle of Least Privilege

Grant minimum necessary permissions:

```go
// ❌ BAD: All users can access admin API
func handleAPI(w http.ResponseWriter, r *http.Request) {
    // No role check
    adminData := getAdminData()
    json.NewEncoder(w).Encode(adminData)
}

// ✅ GOOD: Check user role
func handleAPI(w http.ResponseWriter, r *http.Request) {
    claims := getClaimsFromContext(r.Context())

    if claims.Role != "admin" {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    adminData := getAdminData()
    json.NewEncoder(w).Encode(adminData)
}
```

### 3. Fail Closed

Security checks must fail safely:

```go
// ❌ BAD: Fails open (allows on error)
token, _ := manager.Verify(tokenString, jwt.TokenTypeAccess)
// Proceeding with potentially nil/invalid token

// ✅ GOOD: Fails closed (denies on error)
token, err := manager.Verify(tokenString, jwt.TokenTypeAccess)
if err != nil {
    http.Error(w, "Unauthorized", http.StatusUnauthorized)
    return
}
// Proceed only with valid token
```

### 4. Never Trust User Input

Validate and sanitize ALL user input:

```go
func handleCreatePost(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Title   string `json:"title"`
        Content string `json:"content"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // ✅ Validate length
    if len(req.Title) > 200 {
        http.Error(w, "Title too long", http.StatusBadRequest)
        return
    }

    // ✅ Sanitize HTML
    safeContent := html.EscapeString(req.Content)

    // Store sanitized content
    db.Exec("INSERT INTO posts (title, content) VALUES (?, ?)", req.Title, safeContent)
}
```

## OWASP Top 10 (2021)

### A01:2021 – Broken Access Control

**Risk**: Users accessing resources they shouldn't

**Prevention**:

```go
// ✅ Check ownership
func handleDeletePost(w http.ResponseWriter, r *http.Request) {
    postID := plumego.Param(r, "id")
    userID := getUserIDFromContext(r.Context())

    // Verify user owns the post
    post, err := db.GetPost(postID)
    if err != nil {
        http.Error(w, "Not found", http.StatusNotFound)
        return
    }

    if post.AuthorID != userID {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    // Proceed with deletion
    db.DeletePost(postID)
}
```

**Checklist**:
- [ ] Deny by default (explicit allow)
- [ ] Check authorization on every request
- [ ] Verify resource ownership
- [ ] Log access control failures
- [ ] Use RBAC (Role-Based Access Control)

### A02:2021 – Cryptographic Failures

**Risk**: Sensitive data exposed due to weak crypto

**Prevention**:

```go
import "github.com/spcent/plumego/security/password"
import "github.com/spcent/plumego/security/jwt"

// ✅ Strong password hashing
hash, err := password.Hash(plaintext)

// ✅ Strong JWT secret (≥32 bytes)
secret := []byte(os.Getenv("JWT_SECRET"))
if len(secret) < 32 {
    log.Fatal("JWT_SECRET must be at least 32 bytes")
}

// ✅ Use HTTPS in production
app := core.New(
    core.WithTLS("cert.pem", "key.pem"),
)
```

**Checklist**:
- [ ] Use bcrypt for passwords (cost ≥12)
- [ ] Strong secrets (≥32 bytes random)
- [ ] HTTPS enforced in production
- [ ] TLS 1.2+ only
- [ ] Never log secrets

### A03:2021 – Injection

**Risk**: SQL/Command/LDAP injection attacks

**Prevention**:

```go
// ❌ SQL Injection vulnerability
query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email)
rows, _ := db.Query(query)

// ✅ Parameterized queries
rows, err := db.Query("SELECT * FROM users WHERE email = ?", email)

// ❌ Command injection vulnerability
cmd := exec.Command("sh", "-c", "ping "+host)

// ✅ Argument list (no shell)
cmd := exec.Command("ping", host)
```

**Checklist**:
- [ ] Use parameterized queries
- [ ] Validate input before use
- [ ] Escape special characters
- [ ] Use ORM with SQL injection protection
- [ ] Never concatenate user input in commands

### A04:2021 – Insecure Design

**Risk**: Fundamental design flaws

**Prevention**:

```go
// ❌ Weak design: Rate limit after authentication
func handleLogin(w http.ResponseWriter, r *http.Request) {
    // Check credentials first (allows brute force)
    if !checkCredentials(email, password) {
        return
    }

    // Rate limit AFTER (too late!)
    if !rateLimiter.Allow(clientIP) {
        return
    }
}

// ✅ Strong design: Rate limit BEFORE authentication
func handleLogin(w http.ResponseWriter, r *http.Request) {
    clientIP := getClientIP(r)

    // Rate limit BEFORE credentials check
    if !loginGuard.Allow(clientIP) {
        http.Error(w, "Too many attempts", http.StatusTooManyRequests)
        return
    }

    // Now check credentials
    if !checkCredentials(email, password) {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }
}
```

**Checklist**:
- [ ] Threat modeling completed
- [ ] Rate limiting on authentication
- [ ] Account lockout after failed attempts
- [ ] Security requirements documented
- [ ] Code review by security-aware developers

### A05:2021 – Security Misconfiguration

**Risk**: Default configs, verbose errors, unnecessary features

**Prevention**:

```go
// ✅ Disable debug in production
app := core.New(
    core.WithDebug(os.Getenv("APP_ENV") != "production"),
)

// ✅ Custom error messages (don't leak info)
func handleError(w http.ResponseWriter, r *http.Request, err error) {
    log.Printf("Error: %v", err) // Log full error

    // Generic error to client
    http.Error(w, "Internal server error", http.StatusInternalServerError)
}

// ✅ Security headers enabled
app := core.New(
    core.WithSecurityHeadersEnabled(true),
)
```

**Checklist**:
- [ ] Debug mode disabled in production
- [ ] Default credentials changed
- [ ] Error messages don't leak info
- [ ] Security headers configured
- [ ] Unused features disabled

### A06:2021 – Vulnerable Components

**Risk**: Using components with known vulnerabilities

**Prevention**:

```bash
# Check for vulnerabilities
go list -json -m all | nancy sleuth

# Update dependencies
go get -u ./...
go mod tidy

# Vendor dependencies (reproducible builds)
go mod vendor
```

**Checklist**:
- [ ] Dependencies up to date
- [ ] Vulnerability scanning in CI/CD
- [ ] Remove unused dependencies
- [ ] Pin dependency versions
- [ ] Monitor security advisories

### A07:2021 – Authentication Failures

**Risk**: Weak authentication allowing account takeover

**Prevention**:

```go
import "github.com/spcent/plumego/security/jwt"
import "github.com/spcent/plumego/security/password"
import "github.com/spcent/plumego/security/abuse"

// ✅ Strong authentication
func handleLogin(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }

    // 1. Rate limiting
    if !loginGuard.Allow(getClientIP(r)) {
        http.Error(w, "Too many attempts", http.StatusTooManyRequests)
        return
    }

    // 2. Get user
    user, err := db.GetUserByEmail(req.Email)
    if err != nil {
        // Same error as wrong password (don't reveal existence)
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // 3. Verify password (timing-safe)
    if !password.Verify(req.Password, user.PasswordHash) {
        // Log failed attempt
        log.Printf("Failed login: user=%s ip=%s", user.ID, getClientIP(r))
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // 4. Issue tokens
    access, refresh, err := issueTokens(user.ID)
    if err != nil {
        http.Error(w, "Login failed", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{
        "access_token":  access,
        "refresh_token": refresh,
    })
}
```

**Checklist**:
- [ ] Password hashing (bcrypt cost ≥12)
- [ ] Rate limiting on login
- [ ] Account lockout after N failures
- [ ] Strong password requirements
- [ ] MFA/2FA available
- [ ] Session timeout configured
- [ ] Secure password reset flow

### A08:2021 – Software and Data Integrity Failures

**Risk**: Unsigned updates, insecure CI/CD

**Prevention**:

```go
// ✅ Verify webhook signatures
import "crypto/hmac"
import "crypto/sha256"

func verifyGitHubSignature(body []byte, signature string, secret []byte) bool {
    mac := hmac.New(sha256.New, secret)
    mac.Write(body)
    expectedMAC := mac.Sum(nil)

    receivedMAC, err := hex.DecodeString(strings.TrimPrefix(signature, "sha256="))
    if err != nil {
        return false
    }

    return hmac.Equal(receivedMAC, expectedMAC)
}
```

**Checklist**:
- [ ] Webhook signatures verified
- [ ] CI/CD pipeline secured
- [ ] Code signing for releases
- [ ] Dependency checksums verified
- [ ] Immutable builds

### A09:2021 – Security Logging Failures

**Risk**: Insufficient logging to detect breaches

**Prevention**:

```go
import "github.com/spcent/plumego/log"

// ✅ Log security events
func handleLogin(w http.ResponseWriter, r *http.Request) {
    clientIP := getClientIP(r)

    // Log login attempt
    log.Info("Login attempt",
        "ip", clientIP,
        "email", req.Email,
        "user_agent", r.Header.Get("User-Agent"),
    )

    if !password.Verify(req.Password, user.PasswordHash) {
        // Log failure
        log.Warn("Login failed",
            "ip", clientIP,
            "email", req.Email,
            "reason", "invalid_password",
        )
        return
    }

    // Log success
    log.Info("Login successful",
        "ip", clientIP,
        "user_id", user.ID,
    )
}
```

**What to Log**:
- ✅ Authentication events (success/failure)
- ✅ Authorization failures
- ✅ Input validation failures
- ✅ Rate limit violations
- ✅ Security errors
- ❌ **NEVER** log passwords/tokens/secrets

**Checklist**:
- [ ] Log authentication events
- [ ] Log authorization failures
- [ ] Log input validation errors
- [ ] Centralized logging
- [ ] Log retention policy
- [ ] Alerting on suspicious patterns

### A10:2021 – Server-Side Request Forgery (SSRF)

**Risk**: Server makes requests to internal/unintended URLs

**Prevention**:

```go
import "net"

// ✅ Validate and sanitize URLs
func isPrivateIP(hostname string) bool {
    ip := net.ParseIP(hostname)
    if ip == nil {
        return false
    }

    private := []string{
        "10.0.0.0/8",
        "172.16.0.0/12",
        "192.168.0.0/16",
        "127.0.0.0/8",
        "169.254.0.0/16", // Link-local
    }

    for _, cidr := range private {
        _, ipNet, _ := net.ParseCIDR(cidr)
        if ipNet.Contains(ip) {
            return true
        }
    }

    return false
}

func handleFetchURL(w http.ResponseWriter, r *http.Request) {
    targetURL := r.URL.Query().Get("url")

    // Validate URL format
    if !input.ValidateURL(targetURL) {
        http.Error(w, "Invalid URL", http.StatusBadRequest)
        return
    }

    parsed, err := url.Parse(targetURL)
    if err != nil {
        http.Error(w, "Invalid URL", http.StatusBadRequest)
        return
    }

    // Block non-HTTP(S) schemes
    if parsed.Scheme != "http" && parsed.Scheme != "https" {
        http.Error(w, "URL must be HTTP or HTTPS", http.StatusBadRequest)
        return
    }

    // Block private IPs
    if isPrivateIP(parsed.Hostname()) {
        http.Error(w, "Cannot fetch from private IPs", http.StatusBadRequest)
        return
    }

    // Fetch URL (now safe)
    resp, err := http.Get(targetURL)
    // ...
}
```

**Checklist**:
- [ ] Validate all user-provided URLs
- [ ] Block private IP ranges
- [ ] Allow-list domains (if possible)
- [ ] Disable URL redirects
- [ ] Network-level restrictions

## Production Security Checklist

### Pre-Deployment

- [ ] **Secrets Management**
  - [ ] No secrets in code
  - [ ] Environment variables for config
  - [ ] Secrets rotation policy

- [ ] **Authentication**
  - [ ] Strong password requirements
  - [ ] Rate limiting on login
  - [ ] JWT secrets ≥32 bytes
  - [ ] Access tokens ≤15 minutes
  - [ ] Refresh tokens 7-30 days

- [ ] **HTTPS/TLS**
  - [ ] HTTPS enforced
  - [ ] TLS 1.2+ only
  - [ ] HSTS enabled (`max-age=31536000`)
  - [ ] Valid SSL certificate

- [ ] **Security Headers**
  - [ ] CSP configured
  - [ ] X-Frame-Options: DENY
  - [ ] X-Content-Type-Options: nosniff
  - [ ] Referrer-Policy configured

- [ ] **Input Validation**
  - [ ] Email validation
  - [ ] URL validation
  - [ ] SQL parameterized queries
  - [ ] XSS prevention (escape output)

- [ ] **Rate Limiting**
  - [ ] Global rate limits
  - [ ] Per-endpoint limits
  - [ ] Abuse guard enabled

### Post-Deployment

- [ ] **Monitoring**
  - [ ] Security event logging
  - [ ] Failed login alerts
  - [ ] Rate limit alerts
  - [ ] Error rate monitoring

- [ ] **Incident Response**
  - [ ] Security contact configured
  - [ ] Incident response plan
  - [ ] Backup strategy

- [ ] **Regular Audits**
  - [ ] Dependency scanning
  - [ ] Penetration testing
  - [ ] Code review

## Secure Coding Patterns

### Pattern 1: Fail Closed

```go
// ✅ ALWAYS check errors
token, err := manager.Verify(tokenString, jwt.TokenTypeAccess)
if err != nil {
    return http.StatusUnauthorized, err
}
```

### Pattern 2: Constant-Time Comparison

```go
import "crypto/subtle"

// ✅ Timing-safe comparison
if subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
    return errors.New("invalid signature")
}
```

### Pattern 3: Context Timeouts

```go
// ✅ Always set timeouts
ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
defer cancel()

req, _ := http.NewRequestWithContext(ctx, "GET", url, nil)
```

### Pattern 4: Parameterized Queries

```go
// ✅ Use placeholders
db.Query("SELECT * FROM users WHERE email = ?", email)

// ✅ NOT string concatenation
```

### Pattern 5: Escape Output

```go
import "html/template"

// ✅ Auto-escape with templates
tmpl.Execute(w, data)

// ✅ Manual escape
html.EscapeString(userInput)
```

## Security Testing

### Unit Tests

```go
func TestAuthorizationChecks(t *testing.T) {
    // Test unauthorized access
    req := httptest.NewRequest("DELETE", "/posts/123", nil)
    rec := httptest.NewRecorder()

    handleDeletePost(rec, req)

    if rec.Code != http.StatusUnauthorized {
        t.Errorf("Expected 401, got %d", rec.Code)
    }
}
```

### Penetration Testing

Tools:
- **OWASP ZAP**: Web app scanner
- **Burp Suite**: Security testing
- **sqlmap**: SQL injection testing
- **nmap**: Port scanning

## Related Documentation

- [Security Overview](README.md) — Security module overview
- [JWT Token Management](jwt.md) — Authentication tokens
- [Password Security](password.md) — Password hashing
- [Abuse Guard](abuse.md) — Rate limiting
- [Security Headers](headers.md) — HTTP security headers
- [Input Validation](input.md) — Input sanitization

## External Resources

- [OWASP Top 10](https://owasp.org/Top10/)
- [OWASP Cheat Sheet Series](https://cheatsheetseries.owasp.org/)
- [CWE Top 25](https://cwe.mitre.org/top25/)
- [Go Security Checklist](https://github.com/guardrailsio/awesome-golang-security)
