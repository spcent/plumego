# Testing Middleware

> **Package**: `github.com/spcent/plumego/middleware`

Guide to testing custom and built-in middleware.

---

## Basic Test

```go
func TestMiddleware(t *testing.T) {
    // Create handler
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("OK"))
    })

    // Apply middleware
    wrapped := myMiddleware(handler)

    // Create test request
    req := httptest.NewRequest("GET", "/test", nil)
    w := httptest.NewRecorder()

    // Execute
    wrapped.ServeHTTP(w, req)

    // Verify
    if w.Code != 200 {
        t.Errorf("Expected 200, got %d", w.Code)
    }
}
```

---

## Table-Driven Tests

```go
func TestAuthMiddleware(t *testing.T) {
    tests := []struct {
        name       string
        token      string
        wantStatus int
    }{
        {"valid token", "valid-token", 200},
        {"invalid token", "invalid-token", 401},
        {"missing token", "", 401},
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(200)
            })

            wrapped := authMiddleware(handler)

            req := httptest.NewRequest("GET", "/", nil)
            if tt.token != "" {
                req.Header.Set("Authorization", "Bearer "+tt.token)
            }

            w := httptest.NewRecorder()
            wrapped.ServeHTTP(w, req)

            if w.Code != tt.wantStatus {
                t.Errorf("Status = %d, want %d", w.Code, tt.wantStatus)
            }
        })
    }
}
```

---

## Testing Middleware Chain

```go
func TestMiddlewareChain(t *testing.T) {
    var order []string

    m1 := func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            order = append(order, "m1")
            next.ServeHTTP(w, r)
        })
    }

    m2 := func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            order = append(order, "m2")
            next.ServeHTTP(w, r)
        })
    }

    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        order = append(order, "handler")
    })

    wrapped := m1(m2(handler))
    wrapped.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))

    expected := []string{"m1", "m2", "handler"}
    if !reflect.DeepEqual(order, expected) {
        t.Errorf("Order = %v, want %v", order, expected)
    }
}
```

---

## Testing with Context

```go
func TestContextMiddleware(t *testing.T) {
    middleware := func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            ctx := context.WithValue(r.Context(), "key", "value")
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }

    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        value := r.Context().Value("key")
        if value != "value" {
            t.Error("Context value not set")
        }
    })

    wrapped := middleware(handler)
    wrapped.ServeHTTP(httptest.NewRecorder(), httptest.NewRequest("GET", "/", nil))
}
```

---

**Next**: [Best Practices](best-practices.md)
