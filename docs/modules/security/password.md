# Password Security

> **Package**: `github.com/spcent/plumego/security/password` | **Stability**: Critical

## Overview

The `password/` package provides secure password hashing and verification using industry-standard bcrypt algorithm. Bcrypt is specifically designed for password hashing with built-in salt generation and configurable computational cost.

**Key Features**:
- **Bcrypt hashing**: Industry-standard, resistant to rainbow tables
- **Automatic salting**: Unique salt per password
- **Configurable cost**: Adjustable computational difficulty
- **Timing-safe comparison**: Prevents timing attacks
- **Strength validation**: Password complexity requirements

## Quick Start

### Basic Usage

```go
import "github.com/spcent/plumego/security/password"

// Hash password during registration
plaintext := "user-password-123"
hash, err := password.Hash(plaintext)
if err != nil {
    log.Fatal(err)
}

// Store hash in database (NOT the plaintext password!)
db.Exec("INSERT INTO users (email, password_hash) VALUES (?, ?)", email, hash)

// Verify password during login
storedHash := getUserPasswordHash(userID)
ok := password.Verify(plaintext, storedHash)
if !ok {
    http.Error(w, "Invalid credentials", http.StatusUnauthorized)
    return
}

// User authenticated successfully
```

## Core Functions

### Hash

Generate bcrypt hash from plaintext password:

```go
func Hash(password string) (string, error)
```

**Example**:
```go
hash, err := password.Hash("my-secure-password")
if err != nil {
    return err
}
// hash = "$2a$12$..."
```

**Parameters**:
- `password`: Plaintext password (string)

**Returns**:
- `hash`: Bcrypt hash string (e.g., `$2a$12$...`)
- `error`: Error if hashing fails

**Cost**: Default cost is 12 (good balance between security and performance)

### HashWithCost

Generate hash with custom cost:

```go
func HashWithCost(password string, cost int) (string, error)
```

**Example**:
```go
// Higher cost = more secure but slower
hash, err := password.HashWithCost("my-password", 14)
```

**Cost Guidelines**:
- **10**: Fast, minimum acceptable (testing only)
- **12**: Default, good for production (recommended)
- **14**: High security, slower (~4x cost of 12)
- **16**: Very high security, very slow (~16x cost of 12)

### Verify

Verify plaintext password against hash:

```go
func Verify(password, hash string) bool
```

**Example**:
```go
ok := password.Verify("user-input", storedHash)
if !ok {
    return errors.New("invalid password")
}
```

**Returns**:
- `true`: Password matches hash
- `false`: Password does not match or hash is invalid

**Security**: Uses timing-safe comparison to prevent timing attacks

### IsStrong

Check password strength:

```go
func IsStrong(password string) bool
```

**Example**:
```go
if !password.IsStrong("weak") {
    return errors.New("password too weak")
}
```

**Requirements**:
- Minimum 8 characters
- At least one lowercase letter (a-z)
- At least one uppercase letter (A-Z)
- At least one digit (0-9)

**Optional**: Customize strength requirements (see Advanced section)

## Registration Flow

### Complete Registration Handler

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

    // 1. Validate email format
    if !input.ValidateEmail(req.Email) {
        http.Error(w, "Invalid email", http.StatusBadRequest)
        return
    }

    // 2. Check password strength
    if !password.IsStrong(req.Password) {
        http.Error(w, "Password too weak", http.StatusBadRequest)
        return
    }

    // 3. Hash password (this is slow ~100-300ms)
    hash, err := password.Hash(req.Password)
    if err != nil {
        log.Printf("Password hashing failed: %v", err)
        http.Error(w, "Registration failed", http.StatusInternalServerError)
        return
    }

    // 4. Store user with hash
    userID, err := db.CreateUser(req.Email, hash)
    if err != nil {
        log.Printf("User creation failed: %v", err)
        http.Error(w, "Registration failed", http.StatusInternalServerError)
        return
    }

    // 5. Return success
    w.WriteHeader(http.StatusCreated)
    json.NewEncoder(w).Encode(map[string]string{
        "user_id": userID,
        "message": "Registration successful",
    })
}
```

## Login Flow

### Complete Login Handler

```go
func handleLogin(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // 1. Get user by email
    user, err := db.GetUserByEmail(req.Email)
    if err != nil {
        // Don't reveal whether email exists
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // 2. Verify password (timing-safe)
    ok := password.Verify(req.Password, user.PasswordHash)
    if !ok {
        // Log failed login attempt
        log.Printf("Failed login attempt for user: %s", user.ID)

        // Don't reveal whether password was wrong
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // 3. Issue JWT tokens
    access, refresh, err := issueTokens(user.ID)
    if err != nil {
        http.Error(w, "Login failed", http.StatusInternalServerError)
        return
    }

    // 4. Return tokens
    json.NewEncoder(w).Encode(map[string]string{
        "access_token":  access,
        "refresh_token": refresh,
        "token_type":    "Bearer",
    })
}
```

## Password Reset Flow

### Request Password Reset

```go
func handleRequestPasswordReset(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email string `json:"email"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // Get user
    user, err := db.GetUserByEmail(req.Email)
    if err != nil {
        // Don't reveal whether email exists
        w.WriteHeader(http.StatusOK)
        json.NewEncoder(w).Encode(map[string]string{
            "message": "If email exists, reset link sent",
        })
        return
    }

    // Generate reset token (JWT with short expiry)
    resetToken, err := generateResetToken(user.ID)
    if err != nil {
        http.Error(w, "Reset failed", http.StatusInternalServerError)
        return
    }

    // Send email with reset link
    resetLink := fmt.Sprintf("https://example.com/reset?token=%s", resetToken)
    sendEmail(user.Email, "Password Reset", resetLink)

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "If email exists, reset link sent",
    })
}
```

### Complete Password Reset

```go
func handleResetPassword(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Token       string `json:"token"`
        NewPassword string `json:"new_password"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // 1. Verify reset token
    claims, err := verifyResetToken(req.Token)
    if err != nil {
        http.Error(w, "Invalid or expired token", http.StatusUnauthorized)
        return
    }

    // 2. Check password strength
    if !password.IsStrong(req.NewPassword) {
        http.Error(w, "Password too weak", http.StatusBadRequest)
        return
    }

    // 3. Hash new password
    hash, err := password.Hash(req.NewPassword)
    if err != nil {
        http.Error(w, "Reset failed", http.StatusInternalServerError)
        return
    }

    // 4. Update password
    err = db.UpdateUserPassword(claims.Subject, hash)
    if err != nil {
        http.Error(w, "Reset failed", http.StatusInternalServerError)
        return
    }

    // 5. Invalidate all user sessions (optional but recommended)
    invalidateUserSessions(claims.Subject)

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Password reset successful",
    })
}
```

## Password Strength Validation

### Default Strength Check

```go
// Built-in strength validator
if !password.IsStrong("MyPass123") {
    return errors.New("password must be at least 8 characters with mixed case and numbers")
}
```

### Custom Strength Rules

```go
import "unicode"

func validatePasswordStrength(password string) error {
    if len(password) < 12 {
        return errors.New("password must be at least 12 characters")
    }

    var (
        hasUpper   bool
        hasLower   bool
        hasNumber  bool
        hasSpecial bool
    )

    for _, char := range password {
        switch {
        case unicode.IsUpper(char):
            hasUpper = true
        case unicode.IsLower(char):
            hasLower = true
        case unicode.IsNumber(char):
            hasNumber = true
        case unicode.IsPunct(char) || unicode.IsSymbol(char):
            hasSpecial = true
        }
    }

    if !hasUpper {
        return errors.New("password must contain uppercase letter")
    }
    if !hasLower {
        return errors.New("password must contain lowercase letter")
    }
    if !hasNumber {
        return errors.New("password must contain number")
    }
    if !hasSpecial {
        return errors.New("password must contain special character")
    }

    return nil
}
```

### Common Password Blacklist

```go
var commonPasswords = map[string]bool{
    "password":   true,
    "12345678":   true,
    "qwerty":     true,
    "abc123":     true,
    "monkey":     true,
    "1234567890": true,
    "password123": true,
}

func isCommonPassword(password string) bool {
    return commonPasswords[strings.ToLower(password)]
}

func validatePassword(password string) error {
    if isCommonPassword(password) {
        return errors.New("password is too common")
    }

    if !password.IsStrong(password) {
        return errors.New("password too weak")
    }

    return nil
}
```

## Security Best Practices

### ✅ DO: Store Hash, Not Password

```go
// ✅ CORRECT
hash, _ := password.Hash("user-password")
db.Exec("INSERT INTO users (email, password_hash) VALUES (?, ?)", email, hash)

// ❌ NEVER DO THIS
db.Exec("INSERT INTO users (email, password) VALUES (?, ?)", email, "user-password")
```

### ✅ DO: Use Timing-Safe Verification

```go
// ✅ CORRECT: Verify() uses timing-safe comparison internally
ok := password.Verify(userInput, storedHash)

// ❌ BAD: Don't implement your own verification
if bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(userInput)) == nil {
    // Vulnerable to timing attacks
}
```

### ✅ DO: Never Log Passwords

```go
// ❌ NEVER DO THIS
log.Printf("User password: %s", password)
log.Printf("Password hash: %s", hash)
log.Printf("Login failed for %s with password %s", email, password)

// ✅ CORRECT
log.Printf("Password verification failed for user: %s", userID)
log.Printf("Failed login attempt from IP: %s", clientIP)
```

### ✅ DO: Rate Limit Login Attempts

```go
import "github.com/spcent/plumego/security/abuse"

var loginGuard = abuse.NewGuard(abuse.Config{
    RequestsPerSecond: 1,     // 1 login per second
    BurstSize:         5,     // Allow burst of 5
    CleanupInterval:   time.Minute,
})

func handleLogin(w http.ResponseWriter, r *http.Request) {
    clientIP := getClientIP(r)

    if !loginGuard.Allow(clientIP) {
        http.Error(w, "Too many login attempts", http.StatusTooManyRequests)
        return
    }

    // Proceed with login...
}
```

### ✅ DO: Don't Reveal User Existence

```go
// ❌ BAD: Reveals whether email exists
user, err := db.GetUserByEmail(email)
if err != nil {
    return "Email not found"
}
if !password.Verify(password, user.PasswordHash) {
    return "Invalid password"
}

// ✅ GOOD: Same error message for both cases
user, err := db.GetUserByEmail(email)
if err != nil || !password.Verify(password, user.PasswordHash) {
    return "Invalid credentials"
}
```

## Performance Considerations

### Hashing is Slow (By Design)

Bcrypt is intentionally slow to prevent brute-force attacks:

```go
// Default cost=12 takes ~100-300ms
start := time.Now()
hash, _ := password.Hash("my-password")
fmt.Printf("Hashing took: %v\n", time.Since(start))
// Output: Hashing took: 250ms
```

### Offload to Background Workers

For high-traffic registration endpoints:

```go
type RegistrationJob struct {
    Email    string
    Password string
}

var registrationQueue = make(chan RegistrationJob, 100)

// Background worker
func processRegistrations() {
    for job := range registrationQueue {
        hash, err := password.Hash(job.Password)
        if err != nil {
            log.Printf("Hashing failed: %v", err)
            continue
        }

        if err := db.CreateUser(job.Email, hash); err != nil {
            log.Printf("User creation failed: %v", err)
        }
    }
}

// Handler
func handleRegister(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email    string `json:"email"`
        Password string `json:"password"`
    }

    // Validate and enqueue
    registrationQueue <- RegistrationJob{
        Email:    req.Email,
        Password: req.Password,
    }

    w.WriteHeader(http.StatusAccepted)
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Registration in progress",
    })
}
```

### Verification is Also Slow

Verification takes same time as hashing. For login, this is acceptable as rate limiting prevents abuse.

## Testing

### Unit Tests

```go
func TestPasswordHashing(t *testing.T) {
    plaintext := "my-secure-password"

    // Test hashing
    hash, err := password.Hash(plaintext)
    if err != nil {
        t.Fatalf("Hash failed: %v", err)
    }

    // Hash should not equal plaintext
    if hash == plaintext {
        t.Error("Hash should not equal plaintext")
    }

    // Test verification (correct password)
    if !password.Verify(plaintext, hash) {
        t.Error("Verify failed for correct password")
    }

    // Test verification (wrong password)
    if password.Verify("wrong-password", hash) {
        t.Error("Verify succeeded for wrong password")
    }
}

func TestPasswordStrength(t *testing.T) {
    tests := []struct {
        password string
        strong   bool
    }{
        {"weak", false},                  // Too short
        {"WeakPass", false},              // No numbers
        {"weakpass123", false},           // No uppercase
        {"WEAKPASS123", false},           // No lowercase
        {"StrongPass123", true},          // Valid
        {"MyP@ssw0rd!", true},            // Valid with special chars
    }

    for _, tt := range tests {
        t.Run(tt.password, func(t *testing.T) {
            strong := password.IsStrong(tt.password)
            if strong != tt.strong {
                t.Errorf("IsStrong(%q) = %v, want %v", tt.password, strong, tt.strong)
            }
        })
    }
}
```

## Troubleshooting

### Hash Too Long for Database

**Problem**: Database column too short for bcrypt hash

**Solution**: Use VARCHAR(255) or TEXT:
```sql
ALTER TABLE users MODIFY COLUMN password_hash VARCHAR(255);
```

### Verification Always Fails

**Problem**: Hash corrupted during storage/retrieval

**Checks**:
- Database column type supports full hash length
- No encoding issues (UTF-8 vs ASCII)
- Hash not truncated when reading from DB

### Hashing Too Slow

**Problem**: Registration takes too long

**Solutions**:
- Reduce cost from 12 to 10 (testing only!)
- Use background workers for production
- Consider horizontal scaling

## Related Documentation

- [Security Overview](README.md) — Security module overview
- [JWT Token Management](jwt.md) — Authentication with tokens
- [Abuse Guard](abuse.md) — Rate limiting for login endpoints
- [Input Validation](input.md) — Email format validation
- [Best Practices](best-practices.md) — Security best practices

## Reference Implementation

See examples:
- `examples/reference/` — User registration and login
- `examples/multi-tenant-saas/` — Multi-tenant authentication
