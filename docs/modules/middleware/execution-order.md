# Middleware Execution Order

> **Package**: `github.com/spcent/plumego/middleware`

Understanding middleware execution order is critical for correct application behavior.

---

## Basic Flow

```
Request → M1 → M2 → M3 → Handler → M3 → M2 → M1 → Response
         ↓    ↓    ↓       ↓       ↑    ↑    ↑
       before before before      after after after
```

### Example

```go
app.Use(middleware1) // First
app.Use(middleware2) // Second
app.Use(middleware3) // Third

// Execution:
// 1. middleware1 (before)
// 2. middleware2 (before)
// 3. middleware3 (before)
// 4. handler
// 5. middleware3 (after)
// 6. middleware2 (after)
// 7. middleware1 (after)
```

---

## Recommended Order

### Production Stack

```go
app := core.New()

// 1. Request Identity (fast, sets context)
app.Use(requestIDMiddleware)

// 2. Rate Limiting (fast, rejects early)
app.Use(rateLimitMiddleware)

// 3. Authentication (medium, validates credentials)
app.Use(authMiddleware)

// 4. Logging (medium, records request)
app.Use(loggingMiddleware)

// 5. Request Validation (medium)
app.Use(validationMiddleware)

// 6. Business Logic Handler
// ... handlers ...

// 7. Recovery (catches panics from all above)
app.Use(recoveryMiddleware)
```

### Why This Order?

1. **RequestID First**: Sets trace ID for all subsequent logging
2. **Rate Limit Early**: Reject excessive requests before expensive operations
3. **Auth Before Logging**: Log authenticated user info
4. **Validation After Auth**: Only validate authorized requests
5. **Recovery Last**: Catch panics from any middleware

---

## Common Patterns

### Security First

```go
app.Use(securityHeadersMiddleware) // Set security headers
app.Use(corsMiddleware)            // CORS validation
app.Use(authMiddleware)            // Authentication
app.Use(authorizationMiddleware)   // Authorization
```

### Performance Optimized

```go
app.Use(cacheMiddleware)        // Serve from cache if possible
app.Use(rateLimitMiddleware)    // Reject if over limit
app.Use(compressionMiddleware)  // Compress response
```

### Observability Stack

```go
app.Use(requestIDMiddleware)
app.Use(tracingMiddleware)
app.Use(metricsMiddleware)
app.Use(loggingMiddleware)
```

---

## Group-Level Order

```go
// Global middleware
app.Use(requestIDMiddleware)
app.Use(loggingMiddleware)

// Group middleware
api := app.Group("/api",
    authMiddleware,      // Group-specific
    rateLimitMiddleware,
)

// Per-route middleware
api.Get("/admin",
    adminMiddleware(handler), // Route-specific
)

// Execution for /api/admin:
// 1. requestIDMiddleware (global)
// 2. loggingMiddleware (global)
// 3. authMiddleware (group)
// 4. rateLimitMiddleware (group)
// 5. adminMiddleware (route)
// 6. handler
```

---

## Anti-Patterns

### ❌ Wrong Order

```go
// ❌ Recovery after all middleware (won't catch their panics)
app.Use(middleware1)
app.Use(middleware2)
app.Use(recoveryMiddleware)

// ✅ Recovery wraps all
app.Use(recoveryMiddleware)
app.Use(middleware1)
app.Use(middleware2)
```

```go
// ❌ Expensive operation before rate limit
app.Use(compressionMiddleware) // Expensive
app.Use(rateLimitMiddleware)   // Too late

// ✅ Cheap checks first
app.Use(rateLimitMiddleware)
app.Use(compressionMiddleware)
```

---

## Testing Order

```go
func TestMiddlewareOrder(t *testing.T) {
    var order []string

    m1 := func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            order = append(order, "m1-before")
            next.ServeHTTP(w, r)
            order = append(order, "m1-after")
        })
    }

    m2 := func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            order = append(order, "m2-before")
            next.ServeHTTP(w, r)
            order = append(order, "m2-after")
        })
    }

    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        order = append(order, "handler")
    })

    // Apply middleware
    wrapped := m1(m2(handler))

    // Execute
    wrapped.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

    // Verify order
    expected := []string{
        "m1-before",
        "m2-before",
        "handler",
        "m2-after",
        "m1-after",
    }

    if !reflect.DeepEqual(order, expected) {
        t.Errorf("Order = %v, want %v", order, expected)
    }
}
```

---

**Next**: [Built-in Middleware](built-in-middleware.md)
