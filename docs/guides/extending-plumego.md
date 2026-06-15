# Authoring Extensions for Plumego

This guide helps you build extensions that are compatible with Plumego and the Go community.

## What is an Extension?

An extension is a package that:
- Builds on top of Plumego's stable roots
- Provides optional capabilities (not required for basic services)
- Maintains stdlib compatibility
- Can be published independently or proposed for inclusion in `x/`

Examples:
- A new storage backend (`x/data/dynamodb`)
- A middleware package (`middleware/ratelimit`)
- A protocol adapter (`x/websocket`)
- An observability exporter (`x/observability/datadog`)

## Design Principles

### 1. Maintain stdlib compatibility

Your extension should work with any `http.Handler`, not just Plumego apps.

✅ Good:
```go
func MyMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Do something
        next.ServeHTTP(w, r)
    })
}

// Works with Plumego:
app.Use(MyMiddleware)

// Works with stdlib:
mux := http.NewServeMux()
handler := MyMiddleware(mux)
```

❌ Bad:
```go
// Plumego-specific coupling
func MyMiddleware(app *core.App) {
    app.UseInternal(...)
}
```

### 2. Keep dependencies minimal

Limit external dependencies, especially in production-critical paths.

✅ Good:
- Stdlib-only for core functionality
- Optional external deps behind a `vendor` tag or separate package

❌ Bad:
- Heavy dependencies for simple functionality
- Tight coupling to specific versions

### 3. Provide clear boundaries

Document what your extension owns and doesn't own.

✅ Good:
```go
// Package myelasticsearch provides Elasticsearch storage backend for Plumego.
// Owns: Elasticsearch connection pooling, query translation, error mapping
// Does NOT own: Business logic, domain models, query building
package myelasticsearch
```

### 4. Accept context.Context everywhere

Use Go's standard context for cancellation and timeouts.

✅ Good:
```go
func (db *Database) GetUser(ctx context.Context, id string) (*User, error) {
    // Respects context cancellation
}
```

❌ Bad:
```go
func (db *Database) GetUser(id string) (*User, error) {
    // No context support
}
```

### 5. Error handling

Return unwrapped errors that consumers can use `errors.Is()` on.

✅ Good:
```go
var ErrNotFound = errors.New("user not found")

func (db *Database) GetUser(id string) (*User, error) {
    if notFound {
        return nil, ErrNotFound
    }
}

// Caller can use:
if errors.Is(err, myelasticsearch.ErrNotFound) { ... }
```

❌ Bad:
```go
func (db *Database) GetUser(id string) (*User, error) {
    if notFound {
        return nil, fmt.Errorf("not found: %s", id)  // Can't use errors.Is
    }
}
```

## Publishing as a Community Extension

If you're publishing independently outside `x/`:

### Step 1: Namespace your package

Use a descriptive name following Go conventions:

✅ Good names:
- `github.com/you/plumego-elasticsearch` (what it's for)
- `github.com/you/plumego-metrics-prometheus` (family-specific)
- `github.com/you/pgx-health` (what it integrates with)

❌ Bad names:
- `plumego` (reserved for main repo)
- `mylib` (too generic)
- `x-something` (reserved for main repo)

### Step 2: Create a community-extension.yaml

Document your extension for discovery:

```yaml
name: Elasticsearch Storage Backend
slug: plumego-elasticsearch
description: >
  Storage backend for Elasticsearch, providing
  query translation and connection pooling.
authors:
  - "Your Name <you@example.com>"
homepage: "https://github.com/you/plumego-elasticsearch"
repository: "https://github.com/you/plumego-elasticsearch.git"
license: "MIT"
go_version: "1.26+"
dependencies: ["github.com/elastic/go-elasticsearch/v8"]
compatibility:
  plumego_min: "v1.1.0"
  plumego_max: "v1.x"
tags: ["storage", "database", "elasticsearch"]
```

### Step 3: Write clear documentation

- **README.md**: What it does, why, and a quick example
- **API examples**: Runnable code showing common patterns
- **Tests**: Unit tests demonstrating usage
- **Godoc**: Clear comments on exported types/functions

### Step 4: Maintain compatibility

- Test against the Plumego versions you claim to support
- Use semantic versioning
- Document breaking changes clearly
- Provide migration guides for major versions

### Step 5: Get discovered

- Document in your README that it's a Plumego extension
- Include the `plumego-` or `plumego_` prefix in package name
- Add topic `plumego` to your GitHub repo
- Consider submitting to Go module discovery sites

## Proposing for Inclusion in `x/`

If you want your extension to be part of the official `x/` family:

### Prerequisites

- Proven utility (used in production by 2+ organizations)
- Stable API across 2+ releases
- Comprehensive test coverage (>80%)
- Clear documentation
- No external runtime dependencies (or approved exceptions)
- Full Plumego design guide adherence

### Proposal Process

1. **Open an issue** describing:
   - What problem it solves
   - Why it belongs in Plumego core
   - Current usage / production evidence
   - Stability timeline

2. **Maintainers review** and may ask for:
   - Design clarifications
   - Test coverage improvements
   - API adjustments for consistency

3. **If approved:**
   - Transfer the package to the Plumego repository
   - Integrate with CI/CD
   - Add documentation to `docs/modules/x/`
   - Update maturity dashboard
   - Release in the next minor version as `experimental`

4. **Progression to beta/ga:**
   - Must show API stability across releases
   - Must have production evidence
   - Must be documented in extension-maturity.yaml

## API Design Checklist

When designing your extension, ensure:

- [ ] **Explicit constructor:** `New()` or similar, no global state
- [ ] **Interfaces, not concrete types:** Allow mocking and swapping
- [ ] **Consistent naming:** Match Plumego style (`Handler`, `Middleware`, etc.)
- [ ] **Context everywhere:** Accept `context.Context` for cancellation
- [ ] **Clear errors:** Return unwrapped errors, define sentinel values
- [ ] **No panics:** Return errors instead
- [ ] **Documentation:** Every exported type/function has a comment
- [ ] **Examples:** Include `example_test.go` or doc examples
- [ ] **Tests:** Unit tests covering main paths and error cases
- [ ] **Logging:** Accept `log.StructuredLogger` interface, not stdlib log
- [ ] **No side effects:** No global state, no hidden registration
- [ ] **Middleware shape:** If you provide middleware, use `func(http.Handler) http.Handler`

## Example: Building a Simple Middleware

Let's build a `x-forwarded-for` header extractor:

```go
// Package trustedproxy provides middleware for reading X-Forwarded-For headers
package trustedproxy

import (
    "context"
    "net"
    "net/http"
)

// Config holds middleware configuration
type Config struct {
    // TrustedCIDRs is a list of CIDR ranges that can set X-Forwarded-For
    TrustedCIDRs []string
}

// Handler returns middleware that extracts client IP from X-Forwarded-For
// if the request comes from a trusted proxy.
func Handler(cfg Config) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            clientIP := r.RemoteAddr
            
            // Check if request from trusted proxy
            if isTrustedProxy(r.RemoteAddr, cfg.TrustedCIDRs) {
                if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
                    clientIP = xff
                }
            }
            
            // Add to request context for handlers to use
            ctx := context.WithValue(r.Context(), "client_ip", clientIP)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// ClientIPFromContext extracts the client IP from request context
func ClientIPFromContext(r *http.Request) string {
    ip, ok := r.Context().Value("client_ip").(string)
    if !ok {
        return r.RemoteAddr
    }
    return ip
}

func isTrustedProxy(addr string, cidrs []string) bool {
    ip, _, _ := net.SplitHostPort(addr)
    for _, cidr := range cidrs {
        _, network, _ := net.ParseCIDR(cidr)
        if network != nil && network.Contains(net.ParseIP(ip)) {
            return true
        }
    }
    return false
}

// example_test.go demonstrates usage
func ExampleHandler() {
    cfg := Config{
        TrustedCIDRs: []string{"10.0.0.0/8"},
    }
    
    handler := Handler(cfg)
    
    // Works with any http.Handler
    app := someApp()
    app.Use(handler)
    
    // Also works with stdlib mux
    mux := http.NewServeMux()
    http.ListenAndServe(":8080", handler(mux))
}
```

## Testing Your Extension

```go
func TestHandler(t *testing.T) {
    cfg := Config{TrustedCIDRs: []string{"10.0.0.0/8"}}
    
    tests := []struct {
        name     string
        remoteIP string
        xffIP    string
        want     string
    }{
        {
            name:     "trusted proxy with header",
            remoteIP: "10.0.0.1",
            xffIP:    "203.0.113.1",
            want:     "203.0.113.1",
        },
        {
            name:     "untrusted proxy ignored",
            remoteIP: "8.8.8.8",
            xffIP:    "203.0.113.1",
            want:     "8.8.8.8",
        },
    }
    
    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            req := httptest.NewRequest("GET", "/", nil)
            req.RemoteAddr = tt.remoteIP + ":12345"
            if tt.xffIP != "" {
                req.Header.Set("X-Forwarded-For", tt.xffIP)
            }
            
            var gotIP string
            Handler(cfg)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                gotIP = ClientIPFromContext(r)
            })).ServeHTTP(httptest.NewRecorder(), req)
            
            if gotIP != tt.want {
                t.Errorf("got %q, want %q", gotIP, tt.want)
            }
        })
    }
}
```

## Version Management

Use semantic versioning:

- **Patch (v1.0.1):** Bug fixes, documentation
- **Minor (v1.1.0):** New features, new behavior (backward-compatible)
- **Major (v2.0.0):** Breaking API changes

Always update your README's Plumego version requirements when you release.

## Getting Help

- Read `docs/reference/canonical-style-guide.md` for Plumego conventions
- Look at `x/` extensions for examples
- Open an issue in the Plumego repo if you have questions
- Check `docs/concepts/extension-boundary.md` for boundary discussions

---

**Ready to build an extension?** Start with the checklist above and the example middleware pattern. Your first extension doesn't need to be in the main repo — publish independently and prove its value first.
