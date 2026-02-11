# JWT Token Management

> **Package**: `github.com/spcent/plumego/security/jwt` | **Stability**: Critical

## Overview

The `jwt/` package provides a complete JWT (JSON Web Token) authentication system with support for:

- **Token types**: Access tokens (short-lived), Refresh tokens (long-lived)
- **Signing algorithms**: HS256 (HMAC-SHA256)
- **Key rotation**: Multiple signing keys with graceful rollover
- **Secure defaults**: Strong cryptographic operations, timing-safe comparisons

JWT tokens are used for stateless authentication where the server doesn't need to store session data.

## Quick Start

### Basic Usage

```go
import "github.com/spcent/plumego/security/jwt"

// Create manager with secret (must be ≥32 bytes)
secret := []byte("your-secret-key-at-least-32-bytes-long")
manager := jwt.NewManager(secret)

// Issue access token (short-lived)
claims := jwt.Claims{
    Subject:   "user-123",
    ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
    Custom: map[string]interface{}{
        "email": "user@example.com",
        "role":  "admin",
    },
}
accessToken, err := manager.Issue(claims, jwt.TokenTypeAccess)
if err != nil {
    log.Fatal(err)
}

// Issue refresh token (long-lived)
refreshClaims := jwt.Claims{
    Subject:   "user-123",
    ExpiresAt: time.Now().Add(7 * 24 * time.Hour).Unix(),
}
refreshToken, err := manager.Issue(refreshClaims, jwt.TokenTypeRefresh)
if err != nil {
    log.Fatal(err)
}

// Verify access token
verified, err := manager.Verify(accessToken, jwt.TokenTypeAccess)
if err != nil {
    // Token invalid, expired, or wrong type
    http.Error(w, "Unauthorized", http.StatusUnauthorized)
    return
}

userID := verified.Subject
email := verified.Custom["email"].(string)
```

## Core Types

### Manager

The main JWT manager:

```go
type Manager struct {
    secret []byte
    // ... internal fields
}

func NewManager(secret []byte) *Manager
func (m *Manager) Issue(claims Claims, tokenType TokenType) (string, error)
func (m *Manager) Verify(tokenString string, tokenType TokenType) (*Claims, error)
func (m *Manager) Refresh(refreshToken string) (accessToken, newRefreshToken string, err error)
```

### Claims

JWT claims structure:

```go
type Claims struct {
    Subject   string                 // User ID or subject identifier
    IssuedAt  int64                  // Unix timestamp when token was issued
    ExpiresAt int64                  // Unix timestamp when token expires
    TokenType TokenType              // "access" or "refresh"
    Custom    map[string]interface{} // Custom claims (role, email, etc.)
}
```

### Token Types

```go
type TokenType string

const (
    TokenTypeAccess  TokenType = "access"  // Short-lived (minutes)
    TokenTypeRefresh TokenType = "refresh" // Long-lived (days)
)
```

## Token Lifecycle

### 1. Issue Access + Refresh Tokens

```go
func issueTokenPair(userID string) (access, refresh string, err error) {
    manager := getJWTManager() // Your manager instance

    // Access token: 15 minutes
    accessClaims := jwt.Claims{
        Subject:   userID,
        ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
        Custom: map[string]interface{}{
            "email": "user@example.com",
            "role":  "user",
        },
    }
    access, err = manager.Issue(accessClaims, jwt.TokenTypeAccess)
    if err != nil {
        return "", "", err
    }

    // Refresh token: 7 days
    refreshClaims := jwt.Claims{
        Subject:   userID,
        ExpiresAt: time.Now().Add(7 * 24 * time.Hour).Unix(),
    }
    refresh, err = manager.Issue(refreshClaims, jwt.TokenTypeRefresh)
    if err != nil {
        return "", "", err
    }

    return access, refresh, nil
}
```

### 2. Verify Token in Middleware

```go
func authMiddleware(manager *jwt.Manager) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract token from Authorization header
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, "Missing authorization", http.StatusUnauthorized)
                return
            }

            // Expected format: "Bearer <token>"
            parts := strings.Split(authHeader, " ")
            if len(parts) != 2 || parts[0] != "Bearer" {
                http.Error(w, "Invalid authorization format", http.StatusUnauthorized)
                return
            }

            token := parts[1]
            claims, err := manager.Verify(token, jwt.TokenTypeAccess)
            if err != nil {
                http.Error(w, "Invalid token", http.StatusUnauthorized)
                return
            }

            // Add claims to context
            ctx := context.WithValue(r.Context(), "user_id", claims.Subject)
            ctx = context.WithValue(ctx, "claims", claims)

            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}
```

### 3. Refresh Token Flow

```go
func handleRefresh(w http.ResponseWriter, r *http.Request) {
    var req struct {
        RefreshToken string `json:"refresh_token"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    manager := getJWTManager()

    // Refresh returns new access token AND new refresh token
    newAccess, newRefresh, err := manager.Refresh(req.RefreshToken)
    if err != nil {
        http.Error(w, "Invalid refresh token", http.StatusUnauthorized)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{
        "access_token":  newAccess,
        "refresh_token": newRefresh,
        "token_type":    "Bearer",
        "expires_in":    "900", // 15 minutes in seconds
    })
}
```

## Token Expiration Strategy

### Recommended Expiration Times

| Token Type | Expiration | Use Case |
|------------|------------|----------|
| **Access Token** | 15 minutes | API requests, authenticated operations |
| **Refresh Token** | 7 days | Token renewal without re-authentication |
| **Remember Me** | 30 days | Extended session for trusted devices |

### Implementation

```go
const (
    AccessTokenExpiry  = 15 * time.Minute
    RefreshTokenExpiry = 7 * 24 * time.Hour
    RememberMeExpiry   = 30 * 24 * time.Hour
)

func issueTokens(userID string, rememberMe bool) (access, refresh string, err error) {
    manager := getJWTManager()

    // Access token (always 15 minutes)
    accessClaims := jwt.Claims{
        Subject:   userID,
        ExpiresAt: time.Now().Add(AccessTokenExpiry).Unix(),
    }
    access, err = manager.Issue(accessClaims, jwt.TokenTypeAccess)
    if err != nil {
        return "", "", err
    }

    // Refresh token (7 or 30 days)
    refreshExpiry := RefreshTokenExpiry
    if rememberMe {
        refreshExpiry = RememberMeExpiry
    }

    refreshClaims := jwt.Claims{
        Subject:   userID,
        ExpiresAt: time.Now().Add(refreshExpiry).Unix(),
    }
    refresh, err = manager.Issue(refreshClaims, jwt.TokenTypeRefresh)
    if err != nil {
        return "", "", err
    }

    return access, refresh, nil
}
```

## Custom Claims

### Adding Custom Data

```go
claims := jwt.Claims{
    Subject:   "user-123",
    ExpiresAt: time.Now().Add(15 * time.Minute).Unix(),
    Custom: map[string]interface{}{
        "email":       "user@example.com",
        "role":        "admin",
        "permissions": []string{"read", "write", "delete"},
        "tenant_id":   "tenant-456",
    },
}

token, err := manager.Issue(claims, jwt.TokenTypeAccess)
```

### Extracting Custom Claims

```go
claims, err := manager.Verify(token, jwt.TokenTypeAccess)
if err != nil {
    return err
}

// Type assertions for custom claims
email := claims.Custom["email"].(string)
role := claims.Custom["role"].(string)

// Handle optional claims safely
if tenantID, ok := claims.Custom["tenant_id"].(string); ok {
    // Tenant ID is present
    log.Printf("Request from tenant: %s", tenantID)
}

// Handle array claims
if perms, ok := claims.Custom["permissions"].([]interface{}); ok {
    permissions := make([]string, len(perms))
    for i, p := range perms {
        permissions[i] = p.(string)
    }
}
```

## Key Rotation

### Why Rotate Keys?

Key rotation limits the impact of key compromise:
- Old tokens become invalid after rotation
- Reduces attack window for stolen keys
- Industry best practice for production systems

### Single Key (Simple)

```go
// Create manager with single key
secret := []byte(os.Getenv("JWT_SECRET"))
manager := jwt.NewManager(secret)
```

### Multiple Keys (Advanced)

```go
// Primary key for signing new tokens
primaryKey := []byte(os.Getenv("JWT_SECRET_PRIMARY"))

// Secondary keys for verifying old tokens
secondaryKeys := [][]byte{
    []byte(os.Getenv("JWT_SECRET_SECONDARY_1")),
    []byte(os.Getenv("JWT_SECRET_SECONDARY_2")),
}

// Manager tries keys in order: primary, then secondaries
manager := jwt.NewManagerWithKeys(primaryKey, secondaryKeys)
```

### Rotation Process

1. **Add new key as secondary**:
   ```
   JWT_SECRET_PRIMARY=old-key
   JWT_SECRET_SECONDARY_1=new-key
   ```

2. **Deploy and wait for token expiry** (15 minutes for access tokens)

3. **Promote new key to primary**:
   ```
   JWT_SECRET_PRIMARY=new-key
   JWT_SECRET_SECONDARY_1=old-key
   ```

4. **Deploy and wait for refresh token expiry** (7-30 days)

5. **Remove old key**:
   ```
   JWT_SECRET_PRIMARY=new-key
   ```

## Security Considerations

### Secret Strength

```go
// ❌ BAD: Weak secret
secret := []byte("secret")

// ❌ BAD: Predictable secret
secret := []byte("my-app-jwt-secret")

// ✅ GOOD: Strong random secret (≥32 bytes)
secret := []byte("8kJ9mN2pQ5rS8tU1vW3xY6zA7bC0dE4fG")

// ✅ BEST: Generate with crypto/rand
secret := make([]byte, 32)
if _, err := rand.Read(secret); err != nil {
    log.Fatal(err)
}
```

### Never Log Tokens

```go
// ❌ NEVER DO THIS
log.Printf("Issued token: %s", token)
log.Printf("Token claims: %+v", claims)

// ✅ Log safely (without token content)
log.Printf("Issued %s token for user: %s", tokenType, claims.Subject)
log.Printf("Token verification failed for user: %s", userID)
```

### Validate Token Type

```go
// ❌ BAD: Not checking token type
claims, err := manager.Verify(token, jwt.TokenTypeAccess)

// Refresh token could be used as access token!

// ✅ GOOD: Manager enforces token type
claims, err := manager.Verify(accessToken, jwt.TokenTypeAccess)
// Returns error if token type is "refresh"

claims, err := manager.Verify(refreshToken, jwt.TokenTypeRefresh)
// Returns error if token type is "access"
```

## Error Handling

### Token Verification Errors

```go
claims, err := manager.Verify(token, jwt.TokenTypeAccess)
if err != nil {
    switch {
    case errors.Is(err, jwt.ErrTokenExpired):
        // Token expired, suggest refresh
        return http.StatusUnauthorized, "Token expired"

    case errors.Is(err, jwt.ErrTokenInvalid):
        // Token malformed or signature invalid
        return http.StatusUnauthorized, "Invalid token"

    case errors.Is(err, jwt.ErrTokenTypeMismatch):
        // Wrong token type (e.g., refresh used as access)
        return http.StatusUnauthorized, "Wrong token type"

    default:
        // Unknown error
        return http.StatusInternalServerError, "Token verification failed"
    }
}
```

## Complete Authentication Flow

### Login Endpoint

```go
func handleLogin(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Email      string `json:"email"`
        Password   string `json:"password"`
        RememberMe bool   `json:"remember_me"`
    }

    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Invalid request", http.StatusBadRequest)
        return
    }

    // 1. Verify credentials (check password hash)
    user, err := db.GetUserByEmail(req.Email)
    if err != nil || !password.Verify(req.Password, user.PasswordHash) {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    // 2. Issue token pair
    access, refresh, err := issueTokens(user.ID, req.RememberMe)
    if err != nil {
        http.Error(w, "Token generation failed", http.StatusInternalServerError)
        return
    }

    // 3. Return tokens
    json.NewEncoder(w).Encode(map[string]string{
        "access_token":  access,
        "refresh_token": refresh,
        "token_type":    "Bearer",
        "expires_in":    "900", // 15 minutes
    })
}
```

### Protected Endpoint

```go
func handleProtectedResource(w http.ResponseWriter, r *http.Request) {
    // Extract user ID from context (set by auth middleware)
    userID, ok := r.Context().Value("user_id").(string)
    if !ok {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Use userID for authorization checks
    if !canAccessResource(userID, resourceID) {
        http.Error(w, "Forbidden", http.StatusForbidden)
        return
    }

    // Return protected resource
    json.NewEncoder(w).Encode(map[string]string{
        "message": "Protected data",
        "user_id": userID,
    })
}
```

## Testing

### Unit Tests

```go
func TestJWTLifecycle(t *testing.T) {
    secret := []byte("test-secret-at-least-32-bytes-long!")
    manager := jwt.NewManager(secret)

    // Test issue
    claims := jwt.Claims{
        Subject:   "user-123",
        ExpiresAt: time.Now().Add(time.Hour).Unix(),
    }
    token, err := manager.Issue(claims, jwt.TokenTypeAccess)
    if err != nil {
        t.Fatalf("Issue failed: %v", err)
    }

    // Test verify
    verified, err := manager.Verify(token, jwt.TokenTypeAccess)
    if err != nil {
        t.Fatalf("Verify failed: %v", err)
    }
    if verified.Subject != "user-123" {
        t.Errorf("Expected subject user-123, got %s", verified.Subject)
    }

    // Test wrong token type
    _, err = manager.Verify(token, jwt.TokenTypeRefresh)
    if err == nil {
        t.Error("Expected error for wrong token type")
    }
}

func TestTokenExpiration(t *testing.T) {
    secret := []byte("test-secret-at-least-32-bytes-long!")
    manager := jwt.NewManager(secret)

    // Issue expired token
    claims := jwt.Claims{
        Subject:   "user-123",
        ExpiresAt: time.Now().Add(-time.Hour).Unix(), // Already expired
    }
    token, err := manager.Issue(claims, jwt.TokenTypeAccess)
    if err != nil {
        t.Fatalf("Issue failed: %v", err)
    }

    // Verify should fail
    _, err = manager.Verify(token, jwt.TokenTypeAccess)
    if !errors.Is(err, jwt.ErrTokenExpired) {
        t.Errorf("Expected ErrTokenExpired, got %v", err)
    }
}
```

## Performance Considerations

### Token Caching

For high-traffic APIs, cache verified tokens:

```go
type TokenCache struct {
    cache *lru.Cache // LRU cache
    mgr   *jwt.Manager
}

func (tc *TokenCache) Verify(token string, tokenType jwt.TokenType) (*jwt.Claims, error) {
    // Check cache first
    if cached, ok := tc.cache.Get(token); ok {
        claims := cached.(*jwt.Claims)
        // Check expiration
        if time.Now().Unix() < claims.ExpiresAt {
            return claims, nil
        }
        tc.cache.Remove(token)
    }

    // Not cached or expired, verify
    claims, err := tc.mgr.Verify(token, tokenType)
    if err != nil {
        return nil, err
    }

    // Cache for future requests
    tc.cache.Add(token, claims)
    return claims, nil
}
```

## Related Documentation

- [Security Overview](README.md) — Security module overview
- [Password Security](password.md) — Password hashing and verification
- [Middleware: Authentication](../middleware/auth.md) — JWT middleware integration
- [Best Practices](best-practices.md) — Security best practices

## Reference Implementation

See examples:
- `examples/reference/` — JWT authentication with login/refresh
- `examples/api-gateway/` — API gateway with JWT verification
- `examples/multi-tenant-saas/` — Multi-tenant JWT with tenant claims
