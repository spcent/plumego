# Middleware Best Practices

> **Package**: `github.com/spcent/plumego/middleware`

Patterns and anti-patterns for effective middleware usage.

---

## General Principles

### ✅ Do

1. **Keep Middleware Focused**
   ```go
   // ✅ Single responsibility
   func loggingMiddleware(next http.Handler) http.Handler { ... }
   func authMiddleware(next http.Handler) http.Handler { ... }

   // ❌ Too many responsibilities
   func superMiddleware(next http.Handler) http.Handler {
       // Logging + Auth + Rate limiting + ...
   }
   ```

2. **Make Middleware Composable**
   ```go
   // ✅ Can be combined
   app.Use(middleware1)
   app.Use(middleware2)

   // ❌ Tightly coupled
   func middleware1And2(next http.Handler) http.Handler { ... }
   ```

3. **Return Early on Errors**
   ```go
   // ✅ Stop execution
   if !authorized {
       http.Error(w, "Unauthorized", 401)
       return
   }

   // ❌ Continue despite error
   if !authorized {
       log.Println("Unauthorized")
   }
   next.ServeHTTP(w, r) // Still executes!
   ```

4. **Use Context for Passing Data**
   ```go
   // ✅ Use context
   ctx := context.WithValue(r.Context(), "user", user)
   r = r.WithContext(ctx)

   // ❌ Global variables
   var globalUser User
   globalUser = user
   ```

---

## Performance

### ✅ Do

1. **Order Middleware by Cost**
   ```go
   // ✅ Cheap first, expensive last
   app.Use(requestID)    // Fast
   app.Use(rateLimit)    // Fast
   app.Use(auth)         // Medium
   app.Use(logging)      // Medium
   app.Use(compression)  // Expensive
   ```

2. **Use Conditional Execution**
   ```go
   // ✅ Skip expensive work when possible
   func middleware(next http.Handler) http.Handler {
       return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
           if r.URL.Path == "/health" {
               next.ServeHTTP(w, r)
               return
           }
           // Expensive work
           next.ServeHTTP(w, r)
       })
   }
   ```

3. **Reuse Resources**
   ```go
   // ✅ Create once
   var pool = sync.Pool{...}

   func middleware(next http.Handler) http.Handler {
       return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
           obj := pool.Get()
           defer pool.Put(obj)
           next.ServeHTTP(w, r)
       })
   }
   ```

### ❌ Don't

1. **Don't Create Resources Per Request**
   ```go
   // ❌ Creates limiter every request
   func rateLimitMiddleware(next http.Handler) http.Handler {
       return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
           limiter := rate.NewLimiter(10, 20) // Expensive!
           next.ServeHTTP(w, r)
       })
   }

   // ✅ Create once
   func rateLimitMiddleware(limit int) func(http.Handler) http.Handler {
       limiter := rate.NewLimiter(rate.Limit(limit), limit*2)
       return func(next http.Handler) http.Handler {
           return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
               next.ServeHTTP(w, r)
           })
       }
   }
   ```

---

## Security

### ✅ Do

1. **Validate Early**
   ```go
   // ✅ Reject invalid requests early
   app.Use(rateLimitMiddleware)
   app.Use(authMiddleware)
   app.Use(businessLogic)
   ```

2. **Don't Leak Information**
   ```go
   // ✅ Generic error
   http.Error(w, "Unauthorized", 401)

   // ❌ Leaks details
   http.Error(w, "Invalid token: "+err.Error(), 401)
   ```

3. **Use Timing-Safe Comparisons**
   ```go
   // ✅ Timing-safe
   if subtle.ConstantTimeCompare([]byte(token1), []byte(token2)) == 1 { ... }

   // ❌ Timing attack vulnerable
   if token1 == token2 { ... }
   ```

---

## Error Handling

### ✅ Do

1. **Log Errors**
   ```go
   if err != nil {
       log.Printf("Error: %v", err)
       http.Error(w, "Internal error", 500)
       return
   }
   ```

2. **Use Recovery Middleware**
   ```go
   app.Use(recoveryMiddleware)
   ```

3. **Handle Timeouts**
   ```go
   func timeoutMiddleware(d time.Duration) func(http.Handler) http.Handler {
       return func(next http.Handler) http.Handler {
           return http.TimeoutHandler(next, d, "Request timeout")
       }
   }
   ```

### ❌ Don't

1. **Don't Ignore Errors**
   ```go
   // ❌
   doSomething(r)

   // ✅
   if err := doSomething(r); err != nil {
       http.Error(w, err.Error(), 500)
       return
   }
   ```

2. **Don't Swallow Panics**
   ```go
   // ❌ Silent failure
   defer func() {
       recover()
   }()

   // ✅ Log and handle
   defer func() {
       if err := recover(); err != nil {
           log.Printf("Panic: %v", err)
           http.Error(w, "Internal error", 500)
       }
   }()
   ```

---

## Common Anti-Patterns

### ❌ 1. Modifying Request After Next

```go
// ❌ Too late
next.ServeHTTP(w, r)
r.Header.Set("X-Custom", "value") // Won't work

// ✅ Before next
r.Header.Set("X-Custom", "value")
next.ServeHTTP(w, r)
```

### ❌ 2. Calling Next Multiple Times

```go
// ❌ Double execution
next.ServeHTTP(w, r)
next.ServeHTTP(w, r)

// ✅ Call once
next.ServeHTTP(w, r)
```

### ❌ 3. Blocking in Middleware

```go
// ❌ Blocks all requests
func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        time.Sleep(10 * time.Second) // Bad!
        next.ServeHTTP(w, r)
    })
}

// ✅ Use timeout middleware
app.Use(timeoutMiddleware(30 * time.Second))
```

### ❌ 4. Storing Mutable State

```go
// ❌ Race condition
var counter int
func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        counter++ // Race!
        next.ServeHTTP(w, r)
    })
}

// ✅ Use atomic
var counter atomic.Int64
func middleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        counter.Add(1)
        next.ServeHTTP(w, r)
    })
}
```

---

## Recommended Stacks

### Development

```go
app := core.New(
    core.WithDebug(),
    core.WithRequestID(),
    core.WithLogging(),
    core.WithRecovery(),
)
```

### Production

```go
app := core.New(
    core.WithRequestID(),
    core.WithSecurityHeadersEnabled(true),
    core.WithAbuseGuardEnabled(true),
    core.WithCORS(cors.Config{...}),
    core.WithRecommendedMiddleware(),
)
```

### High-Performance API

```go
app := core.New(
    core.WithRequestID(),
    core.WithComponent(&ratelimit.Component{...}),
    core.WithComponent(&cache.Component{...}),
    core.WithComponent(&compression.Component{...}),
    core.WithRecovery(),
)
```

---

## Summary

**Key Principles**:
1. Single responsibility per middleware
2. Composable and reusable
3. Order by cost (cheap first)
4. Early returns on errors
5. Use context for data passing
6. Log errors, don't expose details
7. Test thoroughly

---

**Related**: [Middleware Overview](README.md) | [Custom Middleware](custom-middleware.md)
