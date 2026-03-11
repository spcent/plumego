# Middleware Best Practices

> Applies to `core.App.Use(...)`, `router.Use(...)`, and group middleware.

---

## 1. Keep Middleware Transport-Only

Middleware should handle cross-cutting HTTP concerns only:

- observability
- authn/authz checks
- rate/concurrency limits
- security headers
- CORS
- request shaping and validation at transport boundary

Do not embed business logic or persistence orchestration in middleware.

---

## 2. Make Order Explicit

Register in intentional order and test it.

```go
_ = app.Use(
    observability.RequestID(),
    observability.Tracing(nil),
    observability.HTTPMetrics(nil),
    observability.AccessLog(app.Logger()),
    recovery.Recovery(app.Logger()),
)
```

Use group middleware for narrower scope:

```go
api := app.Router().Group("/api")
api.Use(authMiddleware)
```

---

## 3. Fail Closed on Security Decisions

For auth/security/limits middleware:

- deny on verification errors
- never bypass checks silently
- return structured transport errors
- avoid leaking sensitive internals

---

## 4. Avoid Hidden Dependencies

Inject dependencies explicitly when constructing middleware.

```go
func AuthMiddleware(verifier TokenVerifier) middleware.Middleware {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            // explicit verifier usage
            next.ServeHTTP(w, r)
        })
    }
}
```

No context-based service locator patterns.

---

## 5. Keep Panic/Error Paths Covered

At minimum, test:

- normal flow
- downstream error flow
- panic flow (recovery)
- order invariants

Use `go test -race` for middleware touching shared state.

---

## 6. Scope Heavy Middleware Carefully

Do not mount expensive middleware globally if only some routes need it.

Prefer:

```go
public := app.Router().Group("/public")
private := app.Router().Group("/api")
private.Use(authMiddleware)
```

---

## 7. Preserve Observability Context

If middleware rewrites context/request, preserve existing trace/request metadata.

- keep request ID stable across chain
- avoid dropping trace/span context
- keep log fields consistent

---

## 8. Production Checklist

- middleware order documented in code
- panic recovery enabled
- auth and rate controls fail closed
- security headers and CORS scoped correctly
- no secrets in logs
- tests include negative/error paths
