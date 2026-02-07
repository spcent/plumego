# XSS Prevention in Plumego

This document explains the XSS (Cross-Site Scripting) prevention model in Plumego and how to properly use the provided security utilities.

## Table of Contents

1. [Understanding the Security Model](#understanding-the-security-model)
2. [HTML Escaping Utilities](#html-escaping-utilities)
3. [Middleware Infrastructure](#middleware-infrastructure)
4. [Best Practices](#best-practices)
5. [CodeQL False Positives](#codeql-false-positives)

## Understanding the Security Model

Plumego follows a **defense-in-depth** approach to XSS prevention:

1. **Input Validation**: Validate and sanitize user input at system boundaries
2. **Output Escaping**: Escape data when rendering HTML, JavaScript, or URLs
3. **Content Security Policy**: Use security headers to restrict script execution
4. **Middleware Separation**: Keep response processing separate from HTML generation

### Key Principle

**Middleware infrastructure passes through response data; handlers generate HTML.**

- **Middleware** (logging, caching, compression) captures and forwards responses without modification
- **Handlers** are responsible for generating safe HTML using proper escaping

## HTML Escaping Utilities

Plumego provides comprehensive escaping utilities in `utils/html.go`:

### HTMLEscape

Escapes HTML special characters to prevent XSS when embedding user content in HTML.

```go
import "github.com/spcent/plumego/utils"

// Unsafe - user input directly in HTML
fmt.Fprintf(w, "<div>%s</div>", userInput)

// Safe - escaped user input
fmt.Fprintf(w, "<div>%s</div>", utils.HTMLEscape(userInput))
```

**Escapes:**
- `<` → `&lt;`
- `>` → `&gt;`
- `&` → `&amp;`
- `"` → `&#34;`
- `'` → `&#39;`

**When to use:**
- Embedding user content in HTML tags
- Displaying user-provided text on web pages
- Rendering form data back to users

### JSEscape

Escapes strings for safe embedding in JavaScript contexts.

```go
// Unsafe - user input in JavaScript
fmt.Fprintf(w, "<script>var name = '%s';</script>", userInput)

// Safe - escaped user input
fmt.Fprintf(w, "<script>var name = '%s';</script>", utils.JSEscape(userInput))
```

**When to use:**
- Embedding data in `<script>` tags
- Inline event handlers (onclick, etc.) - though these should be avoided
- JavaScript string literals

### URLQueryEscape

Escapes strings for safe use in URL query parameters.

```go
// Unsafe - user input in URL
fmt.Fprintf(w, "<a href='/search?q=%s'>Search</a>", userInput)

// Safe - escaped user input
fmt.Fprintf(w, "<a href='/search?q=%s'>Search</a>", utils.URLQueryEscape(userInput))
```

**When to use:**
- Building URLs with user-provided parameters
- Redirect locations
- API endpoints with query strings

### SanitizeHTML

Basic HTML sanitization that removes dangerous tags and attributes.

```go
// Remove dangerous HTML while preserving safe tags
safe := utils.SanitizeHTML(userHTML)
fmt.Fprintf(w, "<div>%s</div>", safe)
```

**Removes:**
- Script tags (`<script>`, `<iframe>`, `<object>`, `<embed>`)
- Event handlers (`onclick`, `onerror`, `onload`, etc.)
- Dangerous protocols (`javascript:`, `data:`)

**⚠️ Warning:** This is a basic sanitizer. For production rich-text content, use a robust library like [bluemonday](https://github.com/microcosm-cc/bluemonday).

**When to use:**
- Simple HTML content from semi-trusted sources
- Markdown-to-HTML output
- User comments with limited formatting

## Middleware Infrastructure

Plumego's middleware infrastructure is designed to be **XSS-safe by default** through architectural separation:

### Response Processing Flow

```
Handler (generates HTML)
    ↓
Middleware A (records metrics)
    ↓
Middleware B (caches response)
    ↓
Middleware C (compresses)
    ↓
Client
```

### Why Middleware Doesn't Introduce XSS

1. **No HTML Generation**: Middleware only captures, forwards, or transforms existing responses
2. **No User Input Injection**: User input comes from handlers, which are responsible for escaping
3. **Content-Type Aware**: Cache middleware only stores safe content types

### Middleware Examples

#### Safe: Response Recording
```go
// middleware/observability/logging.go
func (r *responseRecorder) Write(p []byte) (int, error) {
    // p contains response data from upstream handler
    // This middleware only records bytes, doesn't modify content
    n, err := r.ResponseWriter.Write(p)
    r.bytes += n
    return n, err
}
```

#### Safe: Response Caching
```go
// store/cache/cache.go
func (m *Middleware) ServeHTTP(w http.ResponseWriter, r *http.Request, next http.Handler) {
    // Capture response from handler
    recorder := httptest.NewRecorder()
    next.ServeHTTP(recorder, r)

    // Pass through response
    w.Write(recorder.Body.Bytes())

    // Only cache safe content types
    if isSafeContentType(recorder.Header().Get("Content-Type")) {
        cache.Set(key, recorder.Body.Bytes())
    }
}
```

#### Unsafe: Direct HTML Generation
```go
// DON'T DO THIS in middleware
func UnsafeMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // This injects user input into HTML - UNSAFE!
        fmt.Fprintf(w, "<div>User: %s</div>", r.URL.Query().Get("name"))
        next.ServeHTTP(w, r)
    })
}

// DO THIS instead
func SafeMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Pass through without HTML generation
        next.ServeHTTP(w, r)
    })
}
```

## Best Practices

### 1. Escape at the Output Boundary

Always escape data just before outputting to HTML:

```go
// ✅ Good - escape at output
func HandleUser(w http.ResponseWriter, r *http.Request) {
    name := r.URL.Query().Get("name")
    // ... business logic ...
    fmt.Fprintf(w, "<h1>Hello, %s</h1>", utils.HTMLEscape(name))
}

// ❌ Bad - escaping too early
func HandleUser(w http.ResponseWriter, r *http.Request) {
    name := utils.HTMLEscape(r.URL.Query().Get("name"))
    // ... business logic that might need raw name ...
    fmt.Fprintf(w, "<h1>Hello, %s</h1>", name)
}
```

### 2. Use Template Auto-Escaping

Go's `html/template` package provides automatic context-aware escaping:

```go
import "html/template"

tmpl := template.Must(template.New("page").Parse(`
    <h1>Hello, {{.Name}}</h1>
    <script>var user = {{.Name | js}};</script>
    <a href="/user?id={{.ID | urlquery}}">Profile</a>
`))

tmpl.Execute(w, userData)
```

### 3. Validate Input at System Boundaries

```go
import "github.com/spcent/plumego/security/input"

func HandleEmail(w http.ResponseWriter, r *http.Request) {
    email := r.FormValue("email")

    // Validate format
    if !input.ValidateEmail(email) {
        http.Error(w, "Invalid email", http.StatusBadRequest)
        return
    }

    // Safe to use (but still escape in HTML)
    fmt.Fprintf(w, "<p>Email: %s</p>", utils.HTMLEscape(email))
}
```

### 4. Use Content Security Policy

```go
app := core.New(
    core.WithSecurityHeadersEnabled(true),
)

// Sets headers like:
// Content-Security-Policy: default-src 'self'
// X-Content-Type-Options: nosniff
// X-Frame-Options: DENY
```

### 5. Don't Generate HTML in Middleware

```go
// ✅ Good - middleware passes through
func LoggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        log.Printf("Request: %s %s", r.Method, r.URL.Path)
        next.ServeHTTP(w, r)
    })
}

// ❌ Bad - middleware generates HTML
func BadMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // DON'T inject HTML in middleware
        fmt.Fprintf(w, "<div>Path: %s</div>", r.URL.Path)
        next.ServeHTTP(w, r)
    })
}
```

## CodeQL False Positives

### Why CodeQL Flags Middleware

Static analysis tools like CodeQL may flag middleware code that writes response data:

```go
// Flagged by CodeQL as "Reflected XSS"
w.Write(responseBytes)
```

However, these are **false positives** because:

1. `responseBytes` comes from upstream handlers, not user input
2. The middleware doesn't inject user data into HTML contexts
3. XSS protection is the handler's responsibility

### Suppressing False Positives

Plumego includes CodeQL configuration and inline comments to document these false positives:

```go
// SECURITY NOTE: This is middleware infrastructure that captures response data
// from upstream handlers for metrics/logging purposes. It does not inject user
// input into HTML contexts and therefore does not introduce XSS vulnerabilities.
// XSS protection should be implemented in handlers that generate HTML using utils/html.go.
w.Write(responseBytes)
```

### Actual XSS Vulnerabilities

CodeQL would correctly flag this as XSS:

```go
// Real XSS vulnerability - user input directly in HTML
func UnsafeHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "<div>%s</div>", r.URL.Query().Get("name"))
    //                               ^^^^^^^^^^^^^^^^^^^^^^^^^^^^^^
    //                               User input not escaped - UNSAFE!
}
```

Fix:
```go
func SafeHandler(w http.ResponseWriter, r *http.Request) {
    fmt.Fprintf(w, "<div>%s</div>", utils.HTMLEscape(r.URL.Query().Get("name")))
    //                               ^^^^^^^^^^^^^^^
    //                               Properly escaped - SAFE
}
```

## Security Checklist

When writing handlers that generate HTML:

- [ ] Use `utils.HTMLEscape()` for user content in HTML
- [ ] Use `utils.JSEscape()` for data in JavaScript contexts
- [ ] Use `utils.URLQueryEscape()` for URL parameters
- [ ] Consider `html/template` for complex HTML generation
- [ ] Validate input at system boundaries
- [ ] Enable security headers via `core.WithSecurityHeadersEnabled(true)`
- [ ] Never trust user input - always escape or validate
- [ ] Review middleware to ensure it doesn't generate HTML
- [ ] Test with malicious inputs (`<script>`, `javascript:`, etc.)

## Additional Security Layers

Plumego provides multiple security features:

- **SSRF Protection**: `net/http/security.go` - Prevents server-side request forgery
- **Input Validation**: `security/input/` - Email, URL, phone validation
- **JWT Security**: `security/jwt/` - Token verification with key rotation
- **WebSocket Validation**: `net/websocket/validation.go` - Message sanitization
- **Cache Security**: Control character rejection, tenant isolation

## References

- [OWASP XSS Prevention Cheat Sheet](https://cheatsheetseries.owasp.org/cheatsheets/Cross_Site_Scripting_Prevention_Cheat_Sheet.html)
- [OWASP DOM-based XSS Prevention](https://cheatsheetseries.owasp.org/cheatsheets/DOM_based_XSS_Prevention_Cheat_Sheet.html)
- [Content Security Policy](https://developer.mozilla.org/en-US/docs/Web/HTTP/CSP)
- [Go html/template](https://pkg.go.dev/html/template)
- [bluemonday HTML Sanitizer](https://github.com/microcosm-cc/bluemonday)

## Questions?

If you discover a security vulnerability, please follow the responsible disclosure process in [SECURITY.md](../SECURITY.md).
