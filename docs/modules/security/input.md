# Input Validation

> **Package**: `github.com/spcent/plumego/security/input` | **Stability**: High

## Overview

The `input/` package provides validation functions for common input types to prevent injection attacks, data corruption, and application errors. Input validation is the **first line of defense** against malicious data.

**Key Validators**:
- **Email**: RFC 5322 compliant email validation
- **URL**: RFC 3986 compliant URL validation
- **Phone**: E.164 international phone number format
- **Alphanumeric**: Letters and numbers only
- **Sanitization**: Remove dangerous characters

## Quick Start

### Basic Usage

```go
import "github.com/spcent/plumego/security/input"

// Validate email
if !input.ValidateEmail("user@example.com") {
    return errors.New("invalid email format")
}

// Validate URL
if !input.ValidateURL("https://example.com/path") {
    return errors.New("invalid URL")
}

// Validate phone (E.164 format)
if !input.ValidatePhone("+12025551234") {
    return errors.New("invalid phone number")
}

// Validate alphanumeric
if !input.ValidateAlphanumeric("abc123") {
    return errors.New("must contain only letters and numbers")
}
```

## Email Validation

### ValidateEmail

Validates email addresses according to RFC 5322.

```go
func ValidateEmail(email string) bool
```

**Examples**:
```go
// ✅ Valid emails
input.ValidateEmail("user@example.com")         // true
input.ValidateEmail("user.name@example.com")    // true
input.ValidateEmail("user+tag@example.com")     // true
input.ValidateEmail("user@sub.example.com")     // true

// ❌ Invalid emails
input.ValidateEmail("invalid")                  // false
input.ValidateEmail("@example.com")             // false
input.ValidateEmail("user@")                    // false
input.ValidateEmail("user @example.com")        // false (space)
input.ValidateEmail("<script>@example.com")     // false (XSS attempt)
```

### Email Validation in Registration

```go
func handleRegister(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Validate email format
    if !input.ValidateEmail(req.Email) {
        http.Error(w, "Invalid email format", http.StatusBadRequest)
        return
    }

    // Normalize email (lowercase)
    req.Email = strings.ToLower(req.Email)

    // Check for disposable email domains (optional)
    if isDisposableEmail(req.Email) {
        http.Error(w, "Disposable emails not allowed", http.StatusBadRequest)
        return
    }

    // Proceed with registration...
}
```

### Custom Email Validation

```go
import "net/mail"

func validateEmailStrict(email string) error {
    // Basic format validation
    if !input.ValidateEmail(email) {
        return errors.New("invalid email format")
    }

    // Parse email parts
    addr, err := mail.ParseAddress(email)
    if err != nil {
        return errors.New("invalid email format")
    }

    // Check length constraints
    if len(addr.Address) > 254 {
        return errors.New("email too long")
    }

    // Check for consecutive dots
    if strings.Contains(addr.Address, "..") {
        return errors.New("invalid email format")
    }

    // Check domain part
    parts := strings.Split(addr.Address, "@")
    if len(parts) != 2 {
        return errors.New("invalid email format")
    }

    domain := parts[1]
    if len(domain) < 3 || !strings.Contains(domain, ".") {
        return errors.New("invalid email domain")
    }

    return nil
}
```

## URL Validation

### ValidateURL

Validates URLs according to RFC 3986.

```go
func ValidateURL(urlStr string) bool
```

**Examples**:
```go
// ✅ Valid URLs
input.ValidateURL("https://example.com")                // true
input.ValidateURL("http://example.com/path")            // true
input.ValidateURL("https://example.com:8080/path")      // true
input.ValidateURL("https://user:pass@example.com")      // true
input.ValidateURL("https://example.com?key=value")      // true

// ❌ Invalid URLs
input.ValidateURL("not-a-url")                          // false
input.ValidateURL("javascript:alert('xss')")            // false (XSS attempt)
input.ValidateURL("file:///etc/passwd")                 // false (local file)
input.ValidateURL("ftp://example.com")                  // false (only HTTP/HTTPS)
```

### URL Validation in API

```go
func handleCreateWebhook(w http.ResponseWriter, r *http.Request) {
    var req struct {
        URL    string `json:"url"`
        Events []string `json:"events"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Validate URL
    if !input.ValidateURL(req.URL) {
        http.Error(w, "Invalid webhook URL", http.StatusBadRequest)
        return
    }

    // Additional checks
    parsedURL, err := url.Parse(req.URL)
    if err != nil {
        http.Error(w, "Invalid webhook URL", http.StatusBadRequest)
        return
    }

    // Only allow HTTPS in production
    if os.Getenv("APP_ENV") == "production" && parsedURL.Scheme != "https" {
        http.Error(w, "Webhook URL must use HTTPS", http.StatusBadRequest)
        return
    }

    // Block internal/private IPs
    if isPrivateIP(parsedURL.Hostname()) {
        http.Error(w, "Cannot use private IP addresses", http.StatusBadRequest)
        return
    }

    // Proceed with webhook creation...
}
```

### Safe URL Parsing

```go
import "net/url"

func validateAndParseURL(urlStr string) (*url.URL, error) {
    // Basic validation
    if !input.ValidateURL(urlStr) {
        return nil, errors.New("invalid URL format")
    }

    // Parse URL
    parsed, err := url.Parse(urlStr)
    if err != nil {
        return nil, err
    }

    // Only allow HTTP/HTTPS
    if parsed.Scheme != "http" && parsed.Scheme != "https" {
        return nil, errors.New("URL must use HTTP or HTTPS")
    }

    // Block localhost
    if parsed.Hostname() == "localhost" || parsed.Hostname() == "127.0.0.1" {
        return nil, errors.New("localhost URLs not allowed")
    }

    // Block private IP ranges
    if isPrivateIP(parsed.Hostname()) {
        return nil, errors.New("private IPs not allowed")
    }

    return parsed, nil
}

func isPrivateIP(hostname string) bool {
    ip := net.ParseIP(hostname)
    if ip == nil {
        return false
    }

    // Check for private IP ranges
    private := []string{
        "10.0.0.0/8",
        "172.16.0.0/12",
        "192.168.0.0/16",
        "127.0.0.0/8",
    }

    for _, cidr := range private {
        _, ipNet, _ := net.ParseCIDR(cidr)
        if ipNet.Contains(ip) {
            return true
        }
    }

    return false
}
```

## Phone Number Validation

### ValidatePhone

Validates phone numbers in E.164 international format.

```go
func ValidatePhone(phone string) bool
```

**E.164 Format**: `+[country code][subscriber number]`

**Examples**:
```go
// ✅ Valid phones (E.164)
input.ValidatePhone("+12025551234")     // US: +1 202 555 1234
input.ValidatePhone("+442071234567")    // UK: +44 20 7123 4567
input.ValidatePhone("+81312345678")     // JP: +81 3 1234 5678

// ❌ Invalid phones
input.ValidatePhone("2025551234")       // false (missing +)
input.ValidatePhone("+1 202 555 1234")  // false (spaces)
input.ValidatePhone("+1-202-555-1234")  // false (dashes)
input.ValidatePhone("(202) 555-1234")   // false (not E.164)
```

### Phone Number Normalization

```go
import "regexp"

func normalizePhone(phone string) (string, error) {
    // Remove all non-digit characters except +
    re := regexp.MustCompile(`[^\d+]`)
    normalized := re.ReplaceAllString(phone, "")

    // Ensure starts with +
    if !strings.HasPrefix(normalized, "+") {
        // Assume US number if no country code
        normalized = "+1" + normalized
    }

    // Validate normalized phone
    if !input.ValidatePhone(normalized) {
        return "", errors.New("invalid phone number")
    }

    return normalized, nil
}

// Usage
phone, err := normalizePhone("(202) 555-1234")
// Returns: "+12025551234"
```

### Phone Number in Registration

```go
func handleRegisterWithPhone(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Phone string `json:"phone"`
        Code  string `json:"verification_code"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Normalize phone number
    phone, err := normalizePhone(req.Phone)
    if err != nil {
        http.Error(w, "Invalid phone number format", http.StatusBadRequest)
        return
    }

    // Verify code (sent via SMS)
    if !verifyPhoneCode(phone, req.Code) {
        http.Error(w, "Invalid verification code", http.StatusUnauthorized)
        return
    }

    // Create user with phone number...
}
```

## Alphanumeric Validation

### ValidateAlphanumeric

Ensures input contains only letters and numbers.

```go
func ValidateAlphanumeric(s string) bool
```

**Examples**:
```go
// ✅ Valid
input.ValidateAlphanumeric("abc123")        // true
input.ValidateAlphanumeric("ABC")           // true
input.ValidateAlphanumeric("123")           // true

// ❌ Invalid
input.ValidateAlphanumeric("abc-123")       // false (dash)
input.ValidateAlphanumeric("abc 123")       // false (space)
input.ValidateAlphanumeric("abc_123")       // false (underscore)
input.ValidateAlphanumeric("<script>")      // false (XSS attempt)
```

### Username Validation

```go
func validateUsername(username string) error {
    // Length check
    if len(username) < 3 || len(username) > 20 {
        return errors.New("username must be 3-20 characters")
    }

    // Alphanumeric check
    if !input.ValidateAlphanumeric(username) {
        return errors.New("username must contain only letters and numbers")
    }

    // Must start with letter
    if !unicode.IsLetter(rune(username[0])) {
        return errors.New("username must start with a letter")
    }

    // Check reserved usernames
    reserved := map[string]bool{
        "admin": true, "root": true, "system": true,
    }
    if reserved[strings.ToLower(username)] {
        return errors.New("username is reserved")
    }

    return nil
}
```

## String Sanitization

### Sanitize Functions

```go
// Remove all non-alphanumeric characters
func Sanitize(s string) string

// Remove dangerous characters (< > " ' & /)
func SanitizeHTML(s string) string

// Remove SQL special characters
func SanitizeSQL(s string) string
```

**Examples**:
```go
// Sanitize general input
input.Sanitize("hello<script>")
// Returns: "helloscript"

// Sanitize HTML input
input.SanitizeHTML("<b>Hello</b>")
// Returns: "bHellob" (tags escaped)

// Sanitize SQL input (use parameterized queries instead!)
input.SanitizeSQL("admin' OR '1'='1")
// Returns: "admin OR 11"
```

### Safe Display Output

```go
import "html"

func handleComment(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Content string `json:"content"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Sanitize HTML before storing
    safeContent := html.EscapeString(req.Content)

    // Store in database
    db.Exec("INSERT INTO comments (content) VALUES (?)", safeContent)

    w.WriteHeader(http.StatusCreated)
}
```

## Injection Attack Prevention

### SQL Injection

**❌ NEVER** concatenate user input into SQL:

```go
// ❌ VULNERABLE
query := fmt.Sprintf("SELECT * FROM users WHERE email = '%s'", email)
rows, _ := db.Query(query)
```

**✅ USE** parameterized queries:

```go
// ✅ SAFE
rows, err := db.Query("SELECT * FROM users WHERE email = ?", email)
```

### XSS (Cross-Site Scripting)

**❌ NEVER** trust user input in HTML:

```go
// ❌ VULNERABLE
fmt.Fprintf(w, "<div>%s</div>", userInput)
```

**✅ ESCAPE** user input:

```go
// ✅ SAFE
import "html/template"

tmpl := template.Must(template.New("page").Parse("<div>{{.}}</div>"))
tmpl.Execute(w, userInput) // Auto-escapes
```

### Command Injection

**❌ NEVER** pass user input to shell commands:

```go
// ❌ VULNERABLE
cmd := exec.Command("sh", "-c", "echo "+userInput)
```

**✅ USE** argument lists:

```go
// ✅ SAFE
cmd := exec.Command("echo", userInput)
```

## Validation Middleware

### Request Validation Middleware

```go
type ValidationRule struct {
    Field     string
    Validator func(string) bool
    Message   string
}

func validateRequest(rules []ValidationRule) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Parse form
            if err := r.ParseForm(); err != nil {
                http.Error(w, "Invalid request", http.StatusBadRequest)
                return
            }

            // Validate each field
            for _, rule := range rules {
                value := r.Form.Get(rule.Field)
                if !rule.Validator(value) {
                    http.Error(w, rule.Message, http.StatusBadRequest)
                    return
                }
            }

            next.ServeHTTP(w, r)
        })
    }
}

// Usage
rules := []ValidationRule{
    {Field: "email", Validator: input.ValidateEmail, Message: "Invalid email"},
    {Field: "phone", Validator: input.ValidatePhone, Message: "Invalid phone"},
}

app.Post("/register", handleRegister, validateRequest(rules))
```

## Testing

### Unit Tests

```go
func TestEmailValidation(t *testing.T) {
    tests := []struct {
        email string
        valid bool
    }{
        {"user@example.com", true},
        {"user.name@example.com", true},
        {"user+tag@example.com", true},
        {"invalid", false},
        {"@example.com", false},
        {"user@", false},
        {"<script>@example.com", false},
    }

    for _, tt := range tests {
        t.Run(tt.email, func(t *testing.T) {
            valid := input.ValidateEmail(tt.email)
            if valid != tt.valid {
                t.Errorf("ValidateEmail(%q) = %v, want %v", tt.email, valid, tt.valid)
            }
        })
    }
}
```

## Related Documentation

- [Security Overview](README.md) — Security module overview
- [Password Security](password.md) — Password validation
- [Validator Module](../validator/) — Request validation framework
- [Best Practices](best-practices.md) — Security best practices

## Reference Implementation

See examples:
- `examples/reference/` — Input validation in registration/login
- `examples/multi-tenant-saas/` — Email validation for multi-tenant users
