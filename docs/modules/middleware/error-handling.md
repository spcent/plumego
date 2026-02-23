# Error Handling in Middleware

> **Package**: `github.com/spcent/plumego/middleware`

Best practices for handling errors in middleware.

---

## Error Strategies

### 1. Early Return

```go
func authMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        token := r.Header.Get("Authorization")
        if token == "" {
            http.Error(w, "Unauthorized", 401)
            return // Stop execution
        }

        next.ServeHTTP(w, r)
    })
}
```

### 2. Recovery Middleware

```go
func recoveryMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if err := recover(); err != nil {
                log.Printf("Panic: %v", err)
                http.Error(w, "Internal Server Error", 500)
            }
        }()

        next.ServeHTTP(w, r)
    })
}
```

### 3. Error Context

```go
func errorContextMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Wrap writer to capture errors
        ew := &errorWriter{ResponseWriter: w}

        next.ServeHTTP(ew, r)

        if ew.err != nil {
            log.Printf("Error: %v", ew.err)
        }
    })
}

type errorWriter struct {
    http.ResponseWriter
    err error
}

func (w *errorWriter) WriteHeader(code int) {
    if code >= 400 {
        w.err = fmt.Errorf("HTTP %d", code)
    }
    w.ResponseWriter.WriteHeader(code)
}
```

---

## Structured Errors

```go
import "github.com/spcent/plumego/contract"

func validationMiddleware(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        if !isValid(r) {
            err := contract.NewValidationError("request", "invalid format")
            contract.WriteError(w, r, err)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

---

## Best Practices

### ✅ Do

1. **Return After Error**
   ```go
   if err != nil {
       http.Error(w, err.Error(), 500)
       return // Don't call next
   }
   ```

2. **Log Errors**
   ```go
   if err != nil {
       log.Printf("Error: %v", err)
       http.Error(w, "Internal error", 500)
       return
   }
   ```

3. **Use Structured Errors**
   ```go
   contract.WriteError(w, r, err)
   ```

### ❌ Don't

1. **Don't Ignore Errors**
   ```go
   // ❌
   validateRequest(r) // Ignoring error

   // ✅
   if err := validateRequest(r); err != nil {
       http.Error(w, err.Error(), 400)
       return
   }
   ```

2. **Don't Expose Internal Errors**
   ```go
   // ❌
   http.Error(w, dbError.Error(), 500)

   // ✅
   log.Printf("DB error: %v", dbError)
   http.Error(w, "Internal error", 500)
   ```

---

**Next**: [Testing Middleware](testing-middleware.md)
