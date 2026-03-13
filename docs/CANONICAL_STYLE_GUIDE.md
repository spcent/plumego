# Plumego Canonical Style Guide

Scope: `core`, `router`, `middleware`, official examples, docs, code generation, AI-agent workflows.

---

## 1. Core Principles

- **stdlib first** — stay close to `http.Handler`, `*http.Request`, `http.ResponseWriter`, `httptest`
- **one obvious way** — one bootstrap, one route style, one handler shape, one decode path, one error shape, one test style
- **explicit over implicit** — no hidden binding, no context service-locator, no magical response wrappers, no import-order behavior
- **small-step refactorability** — narrow boundaries, stable interfaces, shallow call paths, minimal indirection
- **single canonical path** — the reference app defines structure, templates mirror it, examples only demonstrate capabilities

When convenience conflicts with predictability, choose predictability.

---

## 2. Package Roles

### `core`
App construction, lifecycle, route registration entry points, middleware attachment, server startup.
Must stay a kernel and must not become a feature catalog or generic plugin container.

### `router`
Route matching, params, grouping, route tree/lookup, static mounting.
Not allowed: repositories, validators, JSON writers, business response wrappers, service construction.

### `middleware`
Transport-layer cross-cutting only: logging, recovery, timeout, request-id, CORS, auth adapters, rate-limiting, tracing, metrics.
Not allowed: service injection, ORM lookup, business DTO assembly, hidden request binding, domain-policy branching.
Prefer narrow packages such as `requestid`, `tracing`, `accesslog`, `recovery`, `timeout`.

### Extension packages
`ai`, `ops`, `tenant`, websocket, webhook, scheduler — capability layers, not the core learning path. Must not define primary coding style.

### Reference and Templates
`reference/standard-service` is the only canonical application layout.
`templates/standard-service` must mirror that layout for generated projects.
`examples/` are proof-of-capability sandboxes and must not become architectural authority.

---

## 3. Application Structure

```text
cmd/myservice/main.go
internal/httpapp/app.go
internal/httpapp/routes.go
internal/httpapp/handlers/health.go
internal/httpapp/handlers/user_create.go
internal/httpapp/middleware/logging.go
internal/domain/user/service.go
internal/domain/user/repository.go
internal/platform/httpjson/response.go
internal/platform/httperr/error.go
```

This layout should be demonstrated first in `reference/standard-service` and copied by templates and scaffolds, not reinvented per example.

- `cmd/` — startup only
- `internal/httpapp/` — HTTP wiring only
- `internal/domain/` — business logic
- `internal/platform/` — reusable transport/infra helpers
- Do not mix routing, domain logic, persistence, and response helpers in one package

---

## 4. Canonical Bootstrap

```go
func main() {
    app := core.New()

    app.Use(RequestID())
    app.Use(Recovery())
    app.Use(RequestLogger())

    registerRoutes(app)

    if err := http.ListenAndServe(":8080", app); err != nil {
        log.Fatal(err)
    }
}
```

Rules:
- One visible construction site
- Global middleware attached explicitly near startup
- One `registerRoutes` call per bounded area
- No `init()` side-effect registration, no hidden auto-registration

---

## 5. Route Registration

```go
func registerRoutes(app *core.App) {
    app.Get("/healthz", healthHandler)
    app.Post("/users", createUserHandler)
}
```

- One method + one path + one handler per line
- Static registration; no reflection or discovery
- Must remain grep-friendly — path and handler discoverable by search
- Groups: path prefix and shared middleware only; not for hidden policy or service injection

---

## 6. Handler Style

Canonical signature:

```go
func(w http.ResponseWriter, r *http.Request)
```

Handler responsibilities (transport only):
1. Read request inputs
2. Validate transport-level shape
3. Call a service
4. Translate outcome to transport response

Not in handlers: raw SQL, repo construction, context service-lookup, config loading, transaction orchestration, response envelope invention.

---

## 7. Request Decoding

```go
type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

func createUserHandler(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        WriteError(w, BadRequest("invalid_json", "invalid request body"))
        return
    }
    // transport-level validation here
}
```

Thin helper `DecodeJSON(r, &req)` acceptable if: reads exactly one source, returns plain Go errors, no hidden side effects.

Not canonical: middleware-first binding into context, mixed-source auto-binding, magic key DTO retrieval.

---

## 8. Parameter Access

```go
// Route param
id := Param(r, "id")
if id == "" {
    WriteError(w, BadRequest("missing_id", "missing route parameter id"))
    return
}

// Query param
r.URL.Query().Get("page")

// Header
r.Header.Get("Authorization")
```

Data source must be visible at the read site — reader must see whether value came from path, query, header, or body.

---

## 9. Middleware Style

```go
type Middleware func(http.Handler) http.Handler
```

`next` must be called exactly once. No pre/post/error lifecycle splits for ordinary middleware.

Not canonical: business DTO construction, service injection into context, domain success/failure decisions, mutating domain semantics.

---

## 10. Error Model

```go
type ErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}

// One write path:
WriteError(w, BadRequest("invalid_json", "invalid request body"))
```

Rules:
- Structured errors only (no ad hoc `http.Error` in JSON APIs)
- Explicit or predictably derived status codes
- Stable machine-readable error codes
- Logs may be richer; responses must not leak internals
- Identical error class → identical shape across modules

---

## 11. Success Responses

```go
WriteJSON(w, http.StatusCreated, CreateUserResponse{ID: id})
```

- One response helper, used consistently
- No envelope proliferation (`{ "success": true, "data": ... }` etc.) unless there is one global envelope documented once and used everywhere
- Meaningful HTTP status codes set explicitly

---

## 12. Dependency Injection

```go
type UserHandler struct {
    Service user.Service
}

func (h UserHandler) Create(w http.ResponseWriter, r *http.Request) { ... }
```

Not canonical:
```go
svc := MustServiceFromContext(r.Context(), "userService")  // obscures shape, fragile
```

Route wiring file must make clear: who constructs the handler, what its dependencies are, which routes use it.

---

## 13. Testing Style

```go
func TestHealth(t *testing.T) {
    app := core.New()
    app.Get("/healthz", healthHandler)

    req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
    rec := httptest.NewRecorder()
    app.ServeHTTP(rec, req)

    if rec.Code != http.StatusOK {
        t.Fatalf("expected 200, got %d", rec.Code)
    }
}
```

Rules:
- `httptest` first
- Table-driven for pure transport cases
- Domain mocks behind interfaces
- Middleware order visible in the test file
- No full-framework bootstrap for simple route tests
- No hidden global state between tests

---

## 14. Rules for AI Agents

Default behavior:
- Prefer stdlib-shaped solutions
- Follow this guide over package convenience APIs when both are possible
- Preserve existing canonical patterns when editing nearby files
- Read `reference/standard-service` before using `examples/` as shape guidance
- Treat broad bucket names (`utils`, `validator`, `rest`, `pubsub`, `tenant`) as migration debt, not expansion targets

## 15. Agent-First Repo Rules

- Start app-structure work from `reference/standard-service`, not from examples
- Keep each change centered on one primary module when possible
- Prefer adding a new focused extension package over widening a stable root
- Do not introduce new catch-all middleware buckets or protocol family roots
- Keep scaffolds, docs, and the reference app synchronized
- Avoid hidden context-based dependency flow
- Avoid new wrapper abstractions unless justified

Do not introduce:
- New response helper families
- New handler signatures
- New route registration idioms
- New service locator patterns
- New implicit DTO flow through middleware
- New business logic inside middleware
- New package roles that blur core boundaries

Change strategy:
1. Keep public behavior unchanged
2. Align local code to canonical style
3. Extract helpers only when repetition justifies it
4. Do not refactor unrelated files opportunistically

Conflict rule: if existing code conflicts with this guide — preserve behavior first, avoid mixing styles in the touched area, move toward canonical when diff is clear and bounded.

---

## 15. Compatibility APIs

- Must not appear in canonical docs or examples
- Must be labeled clearly
- New features must not build on compatibility-only surfaces
- Migration always moves toward canonical, never away

---

## 16. Review Checklist

- Request flow obvious from route to response?
- Code stays close to `net/http` semantics?
- Single visible source for each input?
- JSON decoding explicit or via one transparent helper?
- Middleware transport-only?
- Dependencies explicit in structs or constructors?
- Response and error shapes consistent with framework conventions?
- A new contributor can trace control flow in a few minutes?

---

## 17. Canonical Examples

### Create endpoint

```go
package handlers

import (
    "encoding/json"
    "net/http"
)

type CreateUserRequest struct {
    Name  string `json:"name"`
    Email string `json:"email"`
}

type CreateUserResponse struct {
    ID string `json:"id"`
}

type UserService interface {
    Create(name, email string) (string, error)
}

type UserHandler struct {
    Service UserService
}

func (h UserHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req CreateUserRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        WriteError(w, BadRequest("invalid_json", "invalid request body"))
        return
    }

    if req.Name == "" {
        WriteError(w, BadRequest("missing_name", "name is required"))
        return
    }

    id, err := h.Service.Create(req.Name, req.Email)
    if err != nil {
        WriteError(w, Internal("create_user_failed", "failed to create user"))
        return
    }

    WriteJSON(w, http.StatusCreated, CreateUserResponse{ID: id})
}
```

### Route wiring

```go
func registerRoutes(app *core.App, userHandler handlers.UserHandler) {
    app.Post("/users", userHandler.Create)
}
```

### Test

```go
func TestCreateUser(t *testing.T) {
    app := core.New()
    h := handlers.UserHandler{Service: stubUserService{ID: "u_123"}}
    app.Post("/users", h.Create)

    body := strings.NewReader(`{"name":"alice","email":"a@example.com"}`)
    req := httptest.NewRequest(http.MethodPost, "/users", body)
    req.Header.Set("Content-Type", "application/json")
    rec := httptest.NewRecorder()

    app.ServeHTTP(rec, req)

    if rec.Code != http.StatusCreated {
        t.Fatalf("expected 201, got %d", rec.Code)
    }
}
```

---

## 18. Forbidden Patterns (Canonical Code)

- Mixing `Get`, `GetCtx`, `GetHandler` styles in one example
- Binding request DTOs in middleware, reading from context in CRUD handlers
- Using `router` for repositories or resource controllers
- Retrieving services from request context maps
- Exposing platform capabilities through `core.App` when ordinary composition suffices
- Introducing new response helper families for a single feature
- Teaching multiple equally valid bootstraps in first-party docs
- Hiding route registration behind import side effects
- Placing business logic in middleware
- Blurring transport and domain package boundaries
- Using official examples to showcase compatibility-only APIs

---

## 19. Final Rule

If a reviewer cannot understand how a request is handled within a few minutes by reading only the route registration, middleware, and handler file — the code is not canonical enough.

Plumego's strength is one stable, explicit, stdlib-aligned style, not many styles.
