# Security Headers

> **Package**: `github.com/spcent/plumego/security/headers` | **Stability**: High

## Overview

The `headers/` package provides security header policies that protect against common web vulnerabilities including:

- **XSS (Cross-Site Scripting)**: Injection of malicious scripts
- **Clickjacking**: UI redress attacks via iframes
- **MIME sniffing**: Content-type confusion attacks
- **MITM (Man-in-the-Middle)**: Protocol downgrade attacks
- **Data leakage**: Information disclosure via Referer header

Security headers are **HTTP response headers** that instruct browsers to enable additional security features.

## Quick Start

### Basic Usage

```go
import "github.com/spcent/plumego/security/headers"

// Apply default security headers
policy := headers.DefaultPolicy()
headers.Apply(w, policy)

// Response includes:
// Content-Security-Policy: default-src 'self'
// X-Frame-Options: DENY
// X-Content-Type-Options: nosniff
// Referrer-Policy: strict-origin-when-cross-origin
// X-XSS-Protection: 1; mode=block
```

### With Middleware

```go
import "github.com/spcent/plumego/core"

app := core.New(
    // Enable security headers globally
    core.WithSecurityHeadersEnabled(true),
)

// All responses will include security headers
```

## Core Types

### Policy

```go
type Policy struct {
    ContentSecurityPolicy   string
    StrictTransportSecurity string
    XFrameOptions           string
    XContentTypeOptions     string
    ReferrerPolicy          string
    XSSProtection           string
    PermissionsPolicy       string
}
```

### Preset Policies

```go
func DefaultPolicy() Policy      // Balanced security for most apps
func StrictPolicy() Policy       // Maximum security (may break features)
func LenientPolicy() Policy      // Minimal restrictions
func CustomPolicy(opts ...Option) Policy // Build custom policy
```

## Security Headers Explained

### Content Security Policy (CSP)

Controls which resources the browser is allowed to load.

**Purpose**: Prevent XSS attacks by restricting script sources

**Example**:
```go
policy := headers.Policy{
    ContentSecurityPolicy: "default-src 'self'; script-src 'self' cdn.example.com; style-src 'self' 'unsafe-inline'",
}
```

**Directives**:
- `default-src`: Fallback for all resource types
- `script-src`: JavaScript sources
- `style-src`: CSS sources
- `img-src`: Image sources
- `font-src`: Font sources
- `connect-src`: AJAX/WebSocket/EventSource connections
- `frame-src`: iframe sources

**Common Values**:
- `'self'`: Same origin only
- `'none'`: Block all
- `'unsafe-inline'`: Allow inline scripts/styles (not recommended)
- `'unsafe-eval'`: Allow eval() (not recommended)
- `https:`: Any HTTPS URL
- `cdn.example.com`: Specific domain

**Example Policies**:

```go
// Strict: Only same-origin resources
"default-src 'self'"

// With CDN: Allow scripts from CDN
"default-src 'self'; script-src 'self' https://cdn.jsdelivr.net"

// With inline styles: Allow inline CSS (for styled components)
"default-src 'self'; style-src 'self' 'unsafe-inline'"

// Full example: Modern web app
"default-src 'self'; script-src 'self' https://cdn.example.com; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' https://fonts.gstatic.com; connect-src 'self' https://api.example.com"
```

### Strict-Transport-Security (HSTS)

Forces browsers to use HTTPS for all future requests.

**Purpose**: Prevent protocol downgrade attacks (HTTP → HTTPS)

**Example**:
```go
policy := headers.Policy{
    StrictTransportSecurity: "max-age=31536000; includeSubDomains; preload",
}
```

**Directives**:
- `max-age=<seconds>`: Duration to remember HTTPS requirement
- `includeSubDomains`: Apply to all subdomains
- `preload`: Eligible for browser preload lists

**Recommendations**:
- Development: Don't use HSTS (certificate issues)
- Staging: `max-age=300` (5 minutes)
- Production: `max-age=31536000; includeSubDomains` (1 year)

### X-Frame-Options

Controls whether page can be embedded in iframe.

**Purpose**: Prevent clickjacking attacks

**Example**:
```go
policy := headers.Policy{
    XFrameOptions: "DENY",
}
```

**Values**:
- `DENY`: Never allow framing
- `SAMEORIGIN`: Allow framing by same origin
- `ALLOW-FROM https://example.com`: Allow specific origin (deprecated)

**Recommendations**:
- Use `DENY` unless you need iframe embedding
- For OAuth/payment flows, use `SAMEORIGIN`

### X-Content-Type-Options

Prevents MIME type sniffing.

**Purpose**: Prevent content-type confusion attacks

**Example**:
```go
policy := headers.Policy{
    XContentTypeOptions: "nosniff",
}
```

**Value**: Always use `nosniff`

**Why**: Prevents browser from interpreting files as different MIME types than declared (e.g., image as JavaScript)

### Referrer-Policy

Controls what information is sent in Referer header.

**Purpose**: Prevent information leakage

**Example**:
```go
policy := headers.Policy{
    ReferrerPolicy: "strict-origin-when-cross-origin",
}
```

**Values**:
- `no-referrer`: Never send Referer
- `same-origin`: Send only for same-origin requests
- `strict-origin`: Send origin only (no path)
- `strict-origin-when-cross-origin`: Full URL for same-origin, origin only for cross-origin (recommended)

### X-XSS-Protection

Legacy XSS filter (mostly replaced by CSP).

**Example**:
```go
policy := headers.Policy{
    XSSProtection: "1; mode=block",
}
```

**Values**:
- `0`: Disable XSS filter
- `1`: Enable XSS filter
- `1; mode=block`: Enable and block page if XSS detected

**Note**: Modern browsers ignore this in favor of CSP.

### Permissions-Policy

Controls browser features and APIs.

**Purpose**: Restrict access to sensitive features

**Example**:
```go
policy := headers.Policy{
    PermissionsPolicy: "geolocation=(), microphone=(), camera=()",
}
```

**Features**:
- `geolocation`: GPS location
- `microphone`: Audio recording
- `camera`: Video recording
- `payment`: Payment Request API
- `usb`: WebUSB API

**Syntax**:
- `feature=()`: Block for all
- `feature=(self)`: Allow for same origin
- `feature=(self "https://example.com")`: Allow for specific origins
- `feature=*`: Allow for all origins (not recommended)

## Preset Policies

### Default Policy

```go
policy := headers.DefaultPolicy()

// Returns:
// ContentSecurityPolicy: "default-src 'self'; script-src 'self'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self'; connect-src 'self'"
// XFrameOptions: "DENY"
// XContentTypeOptions: "nosniff"
// ReferrerPolicy: "strict-origin-when-cross-origin"
// XSSProtection: "1; mode=block"
// PermissionsPolicy: "geolocation=(), microphone=(), camera=()"
```

### Strict Policy

```go
policy := headers.StrictPolicy()

// Returns:
// ContentSecurityPolicy: "default-src 'none'; script-src 'self'; style-src 'self'; img-src 'self'; font-src 'self'; connect-src 'self'"
// StrictTransportSecurity: "max-age=31536000; includeSubDomains; preload"
// XFrameOptions: "DENY"
// XContentTypeOptions: "nosniff"
// ReferrerPolicy: "no-referrer"
// XSSProtection: "1; mode=block"
// PermissionsPolicy: "geolocation=(), microphone=(), camera=(), payment=(), usb=()"
```

### Lenient Policy

```go
policy := headers.LenientPolicy()

// Returns:
// ContentSecurityPolicy: "default-src *; script-src * 'unsafe-inline'; style-src * 'unsafe-inline'"
// XFrameOptions: "SAMEORIGIN"
// XContentTypeOptions: "nosniff"
// ReferrerPolicy: "strict-origin-when-cross-origin"
```

## Custom Policies

### Building Custom Policy

```go
policy := headers.CustomPolicy(
    headers.WithCSP("default-src 'self'; script-src 'self' https://cdn.example.com"),
    headers.WithHSTS("max-age=31536000"),
    headers.WithFrameOptions("SAMEORIGIN"),
    headers.WithNoSniff(),
    headers.WithReferrerPolicy("strict-origin-when-cross-origin"),
)

headers.Apply(w, policy)
```

### Per-Environment Policies

```go
func getSecurityPolicy() headers.Policy {
    env := os.Getenv("APP_ENV")

    switch env {
    case "production":
        return headers.StrictPolicy()

    case "staging":
        policy := headers.DefaultPolicy()
        policy.StrictTransportSecurity = "max-age=300" // 5 minutes
        return policy

    case "development":
        return headers.LenientPolicy()

    default:
        return headers.DefaultPolicy()
    }
}
```

## Integration Patterns

### Global Middleware

```go
func securityHeadersMiddleware(policy headers.Policy) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            headers.Apply(w, policy)
            next.ServeHTTP(w, r)
        })
    }
}

app := core.New()
app.Use(securityHeadersMiddleware(headers.DefaultPolicy()))
```

### Per-Route Policies

```go
// Strict policy for API
apiPolicy := headers.CustomPolicy(
    headers.WithCSP("default-src 'none'"),
    headers.WithFrameOptions("DENY"),
)

// Lenient policy for admin dashboard (needs inline scripts)
adminPolicy := headers.CustomPolicy(
    headers.WithCSP("default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'"),
    headers.WithFrameOptions("SAMEORIGIN"),
)

app.Get("/api/*", handleAPI, securityHeadersMiddleware(apiPolicy))
app.Get("/admin/*", handleAdmin, securityHeadersMiddleware(adminPolicy))
```

## Common Scenarios

### Single Page Application (SPA)

```go
policy := headers.CustomPolicy(
    headers.WithCSP(
        "default-src 'self'; " +
        "script-src 'self' 'unsafe-inline'; " + // React inline scripts
        "style-src 'self' 'unsafe-inline'; " +  // CSS-in-JS
        "img-src 'self' data: https:; " +       // Images and base64
        "font-src 'self' https://fonts.gstatic.com; " +
        "connect-src 'self' https://api.example.com", // API calls
    ),
    headers.WithFrameOptions("DENY"),
    headers.WithNoSniff(),
)
```

### API Server (JSON only)

```go
policy := headers.CustomPolicy(
    headers.WithCSP("default-src 'none'"), // No resources loaded
    headers.WithFrameOptions("DENY"),
    headers.WithNoSniff(),
    headers.WithReferrerPolicy("no-referrer"),
)
```

### Static Site

```go
policy := headers.CustomPolicy(
    headers.WithCSP(
        "default-src 'self'; " +
        "script-src 'self' https://cdn.jsdelivr.net; " +
        "style-src 'self' https://cdn.jsdelivr.net; " +
        "img-src 'self' https:; " +
        "font-src 'self' https://fonts.gstatic.com",
    ),
    headers.WithHSTS("max-age=31536000; includeSubDomains"),
    headers.WithFrameOptions("DENY"),
)
```

### OAuth/Payment Flow

```go
// Allow framing by same origin (for modal flows)
policy := headers.CustomPolicy(
    headers.WithCSP("default-src 'self'; frame-ancestors 'self'"),
    headers.WithFrameOptions("SAMEORIGIN"),
    headers.WithNoSniff(),
)
```

## Testing Security Headers

### Check Headers in Response

```go
func TestSecurityHeaders(t *testing.T) {
    app := setupTestApp()
    req := httptest.NewRequest("GET", "/", nil)
    rec := httptest.NewRecorder()

    app.ServeHTTP(rec, req)

    // Check CSP
    csp := rec.Header().Get("Content-Security-Policy")
    if !strings.Contains(csp, "default-src 'self'") {
        t.Errorf("Missing CSP: %s", csp)
    }

    // Check X-Frame-Options
    xfo := rec.Header().Get("X-Frame-Options")
    if xfo != "DENY" {
        t.Errorf("Expected X-Frame-Options: DENY, got: %s", xfo)
    }

    // Check X-Content-Type-Options
    xcto := rec.Header().Get("X-Content-Type-Options")
    if xcto != "nosniff" {
        t.Errorf("Expected X-Content-Type-Options: nosniff, got: %s", xcto)
    }
}
```

### Online Scanners

Test your deployed application:
- [SecurityHeaders.com](https://securityheaders.com)
- [Mozilla Observatory](https://observatory.mozilla.org)
- [SSL Labs](https://www.ssllabs.com/ssltest/)

## Troubleshooting

### CSP Blocking Resources

**Problem**: CSP blocking legitimate resources

**Solution**: Check browser console for CSP violations, adjust policy:

```go
// Before (too strict)
"default-src 'self'"

// After (allow CDN)
"default-src 'self'; script-src 'self' https://cdn.example.com"
```

### HSTS Certificate Errors

**Problem**: Can't access site after enabling HSTS with invalid certificate

**Solution**: Clear HSTS for domain in browser:
- Chrome: `chrome://net-internals/#hsts`
- Firefox: Delete `SiteSecurityServiceState.txt`

### Iframe Not Loading

**Problem**: X-Frame-Options blocking legitimate iframe

**Solution**: Change from `DENY` to `SAMEORIGIN` or remove if embedding needed:

```go
policy := headers.CustomPolicy(
    headers.WithFrameOptions("SAMEORIGIN"),
)
```

## Security Audit Checklist

- [ ] CSP configured with no `'unsafe-inline'` or `'unsafe-eval'`
- [ ] HSTS enabled in production with `max-age >= 31536000`
- [ ] X-Frame-Options set to `DENY` or `SAMEORIGIN`
- [ ] X-Content-Type-Options set to `nosniff`
- [ ] Referrer-Policy set (recommend `strict-origin-when-cross-origin`)
- [ ] Permissions-Policy restricts unused features
- [ ] HTTPS enforced in production
- [ ] Security headers tested with online scanner
- [ ] CSP violations monitored in production

## Related Documentation

- [Security Overview](README.md) — Security module overview
- [Middleware: Security Headers](../middleware/security.md) — Security headers middleware
- [Best Practices](best-practices.md) — Security best practices

## Reference Implementation

See examples:
- `examples/reference/` — Security headers in production app
- `examples/api-gateway/` — API-specific security headers
