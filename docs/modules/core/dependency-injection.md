# Dependency Injection

> **Module context**: `core`

## v1 Canonical Position

`core.App` does not expose a built-in DI container API in current v1 surface.

Not available in current public API:

- `app.DI()`
- `core.WithDIContainer(...)`
- component dependency graph declarations through `Dependencies() []reflect.Type`

Use constructor-based dependency injection in your application code instead.

---

## Recommended Pattern

Wire dependencies explicitly in `main` (or a dedicated wiring package), then inject them into handlers/components via struct fields.

```go
type UserService struct {
    repo UserRepo
}

type UserHandler struct {
    svc *UserService
}

func NewUserHandler(svc *UserService) *UserHandler {
    return &UserHandler{svc: svc}
}

func registerRoutes(app *core.App, h *UserHandler) {
    app.Get("/users/:id", h.Get)
}
```

This keeps dependencies explicit, testable, and aligned with Plumego's stdlib-first style.

---

## Migration Note

If legacy examples mention `app.DI()` or `WithDIContainer`, treat them as historical and migrate to constructor-based wiring.
