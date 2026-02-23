# Custom Middleware

> **Package**: `github.com/spcent/plumego/middleware`

Guide to building custom middleware for Plumego applications.

---

## Middleware Signature

```go
type Middleware func(http.Handler) http.Handler
```

### Basic Template

```go
func myMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Before handler
        // - Validate request
        // - Modify request
        // - Check conditions

        // Call next handler
        next.ServeHTTP(w, r)

        // After handler
        // - Modify response
        // - Log results
        // - Clean up resources
    })
}
```

---

## Common Patterns

### 1. Request Validation

```go
func validateHeaderMiddleware(headerName, expectedValue string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            value := r.Header.Get(headerName)
            if value != expectedValue {
                http.Error(w, "Invalid header", http.StatusBadRequest)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// Usage
app.Use(validateHeaderMiddleware("X-API-Version", "v1"))
```

### 2. Request Modification

```go
func addContextMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Add value to context
        ctx := context.WithValue(r.Context(), "requestID", generateID())
        r = r.WithContext(ctx)

        next.ServeHTTP(w, r)
    })
}
```

### 3. Response Modification

```go
func addHeaderMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Add response header
        w.Header().Set("X-Custom-Header", "value")

        next.ServeHTTP(w, r)
    })
}
```

### 4. Timing/Metrics

```go
func timingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        next.ServeHTTP(w, r)

        duration := time.Since(start)
        log.Printf("%s %s took %v", r.Method, r.URL.Path, duration)
    })
}
```

### 5. Conditional Execution

```go
func conditionalMiddleware(condition func(*http.Request) bool) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if condition(r) {
                // Do something
                doWork(r)
            }
            next.ServeHTTP(w, r)
        })
    }
}

// Usage
app.Use(conditionalMiddleware(func(r *http.Request) bool {
    return r.Method == "POST"
}))
```

---

## Advanced Patterns

### Response Writer Wrapper

Capture status code and response size:

```go
type responseWriter struct {
    http.ResponseWriter
    statusCode int
    size       int
}

func (rw *responseWriter) WriteHeader(code int) {
    rw.statusCode = code
    rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
    size, err := rw.ResponseWriter.Write(b)
    rw.size += size
    return size, err
}

func loggingMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        rw := &responseWriter{
            ResponseWriter: w,
            statusCode:     200,
        }

        next.ServeHTTP(rw, r)

        log.Printf("%s %s %d %d bytes", r.Method, r.URL.Path, rw.statusCode, rw.size)
    })
}
```

### Middleware with Configuration

```go
type RateLimiterConfig struct {
    RequestsPerSecond float64
    Burst             int
}

func rateLimitMiddleware(config RateLimiterConfig) func(http.Handler) http.Handler {
    limiter := rate.NewLimiter(rate.Limit(config.RequestsPerSecond), config.Burst)

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            if !limiter.Allow() {
                http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
                return
            }
            next.ServeHTTP(w, r)
        })
    }
}

// Usage
app.Use(rateLimitMiddleware(RateLimiterConfig{
    RequestsPerSecond: 100,
    Burst:             200,
}))
```

### Middleware with State

```go
type MetricsMiddleware struct {
    requestCount  atomic.Int64
    totalDuration atomic.Int64
}

func NewMetricsMiddleware() *MetricsMiddleware {
    return &MetricsMiddleware{}
}

func (m *MetricsMiddleware) Middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        start := time.Now()

        next.ServeHTTP(w, r)

        duration := time.Since(start)
        m.requestCount.Add(1)
        m.totalDuration.Add(int64(duration))
    })
}

func (m *MetricsMiddleware) Stats() (count int64, avgDuration time.Duration) {
    count = m.requestCount.Load()
    if count == 0 {
        return 0, 0
    }
    total := m.totalDuration.Load()
    return count, time.Duration(total / count)
}

// Usage
metrics := NewMetricsMiddleware()
app.Use(metrics.Middleware)

// Later, get stats
count, avg := metrics.Stats()
```

---

## Complete Examples

### Authentication Middleware

```go
package auth

import (
    "context"
    "net/http"
    "strings"

    "github.com/spcent/plumego/security/jwt"
)

type contextKey string

const userContextKey contextKey = "user"

func JWTMiddleware(secret string) func(http.Handler) http.Handler {
    manager := jwt.NewManager(secret)

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Extract token
            authHeader := r.Header.Get("Authorization")
            if authHeader == "" {
                http.Error(w, "Missing authorization header", http.StatusUnauthorized)
                return
            }

            parts := strings.Split(authHeader, " ")
            if len(parts) != 2 || parts[0] != "Bearer" {
                http.Error(w, "Invalid authorization header", http.StatusUnauthorized)
                return
            }

            token := parts[1]

            // Verify token
            claims, err := manager.Verify(token, jwt.TokenTypeAccess)
            if err != nil {
                http.Error(w, "Invalid token", http.StatusUnauthorized)
                return
            }

            // Add user to context
            ctx := context.WithValue(r.Context(), userContextKey, claims.UserID)
            r = r.WithContext(ctx)

            next.ServeHTTP(w, r)
        })
    }
}

// Helper to extract user ID from context
func GetUserID(r *http.Request) string {
    userID, _ := r.Context().Value(userContextKey).(string)
    return userID
}
```

### Caching Middleware

```go
package cache

import (
    "bytes"
    "crypto/sha256"
    "fmt"
    "net/http"
    "sync"
    "time"
)

type CacheEntry struct {
    body       []byte
    headers    http.Header
    statusCode int
    cachedAt   time.Time
}

func HTTPCacheMiddleware(ttl time.Duration) func(http.Handler) http.Handler {
    cache := &sync.Map{}

    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // Only cache GET requests
            if r.Method != http.MethodGet {
                next.ServeHTTP(w, r)
                return
            }

            // Generate cache key
            key := fmt.Sprintf("%x", sha256.Sum256([]byte(r.URL.String())))

            // Check cache
            if cached, ok := cache.Load(key); ok {
                entry := cached.(*CacheEntry)

                // Check if expired
                if time.Since(entry.cachedAt) < ttl {
                    // Serve from cache
                    for k, v := range entry.headers {
                        w.Header()[k] = v
                    }
                    w.Header().Set("X-Cache", "HIT")
                    w.WriteHeader(entry.statusCode)
                    w.Write(entry.body)
                    return
                }

                // Expired, remove from cache
                cache.Delete(key)
            }

            // Cache miss, capture response
            rec := &responseRecorder{
                ResponseWriter: w,
                body:           &bytes.Buffer{},
            }

            next.ServeHTTP(rec, r)

            // Store in cache
            cache.Store(key, &CacheEntry{
                body:       rec.body.Bytes(),
                headers:    rec.Header().Clone(),
                statusCode: rec.statusCode,
                cachedAt:   time.Now(),
            })
        })
    }
}

type responseRecorder struct {
    http.ResponseWriter
    body       *bytes.Buffer
    statusCode int
}

func (r *responseRecorder) Write(b []byte) (int, error) {
    r.body.Write(b)
    return r.ResponseWriter.Write(b)
}

func (r *responseRecorder) WriteHeader(code int) {
    r.statusCode = code
    r.ResponseWriter.WriteHeader(code)
}
```

---

## Best Practices

### ✅ Do

1. **Return Early on Errors**
   ```go
   if !valid {
       http.Error(w, "Invalid request", 400)
       return // Don't call next
   }
   next.ServeHTTP(w, r)
   ```

2. **Use Context for Passing Data**
   ```go
   ctx := context.WithValue(r.Context(), "key", value)
   r = r.WithContext(ctx)
   ```

3. **Make Configurable**
   ```go
   func myMiddleware(config Config) func(http.Handler) http.Handler {
       // Use config
   }
   ```

### ❌ Don't

1. **Don't Modify Request After Next**
   ```go
   // ❌ Too late
   next.ServeHTTP(w, r)
   r.Header.Set("X-Custom", "value") // Won't work!
   ```

2. **Don't Call Next Multiple Times**
   ```go
   // ❌ Double execution
   next.ServeHTTP(w, r)
   next.ServeHTTP(w, r) // Error!
   ```

---

**Next**: [Execution Order](execution-order.md)
