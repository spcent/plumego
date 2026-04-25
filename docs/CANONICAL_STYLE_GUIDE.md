# Plumego Canonical Style Guide

Scope: `core`, `router`, `middleware`, official docs, code generation, AI-agent workflows.

---

## 1. Core Principles

- **stdlib first** — stay close to `http.Handler`, `*http.Request`, `http.ResponseWriter`, `httptest`
- **one obvious way** — one bootstrap, one route style, one handler shape, one decode path, one error shape, one test style
- **explicit over implicit** — no hidden binding, no context service-locator, no magical response wrappers, no import-order behavior
- **small-step refactorability** — narrow boundaries, stable interfaces, shallow call paths, minimal indirection
- **single canonical path** — the reference app defines structure

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
`x/ai`, `x/ops`, `x/tenant`, `x/websocket`, `x/webhook`, `x/scheduler`, and sibling `x/*` packages are capability layers, not the core learning path. They must not define the primary coding style.

### Reference and Templates
`reference/standard-service` is the only canonical application layout.

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
```

This layout should be demonstrated first in `reference/standard-service` and copied by templates and scaffolds, not reinvented per example.

- `cmd/` — startup only
- `internal/httpapp/` — HTTP wiring only
- `internal/domain/` — business logic
- `internal/platform/` — optional app-local infra adapters only when the behavior does not already belong to a stable Plumego package
- success and error writes should go directly through `contract.WriteResponse` / `contract.WriteError` from handlers instead of app-local JSON/error helper families
- Do not mix routing, domain logic, persistence, and transport helper policy in one package

---

## 4. Canonical Bootstrap

```go
func main() {
    cfg := core.DefaultConfig()
    app := core.New(cfg, core.AppDependencies{})

    if err := app.Use(RequestID(), Recovery(), RequestLogger()); err != nil {
        log.Fatal(err)
    }

    if err := registerRoutes(app); err != nil {
        log.Fatal(err)
    }

    if err := app.Prepare(); err != nil {
        log.Fatal(err)
    }
    srv, err := app.Server()
    if err != nil {
        log.Fatal(err)
    }
    defer app.Shutdown(context.Background())

    if err := srv.ListenAndServe(); err != nil {
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
func registerRoutes(app *core.App) error {
    if err := app.Get("/healthz", healthHandler); err != nil {
        return err
    }
    if err := app.Post("/users", createUserHandler); err != nil {
        return err
    }
    return nil
}
```

- One method + one path + one handler per line
- Route registration errors must be returned to the caller; do not log and continue
- Use `app.Any(...)` for catch-all method registration; do not introduce overlapping helper aliases for the same ANY-route behavior
- Use `app.AddRoute(..., router.WithRouteName(...))` for named routes; keep one canonical metadata path instead of named-route helper aliases
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
        _ = contract.WriteError(w, r, contract.NewErrorBuilder().
            Type(contract.TypeValidation).
            Code("invalid_json").
            Message("invalid request body").
            Build())
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
    _ = contract.WriteError(w, r, contract.NewErrorBuilder().
        Type(contract.TypeRequired).
        Code("missing_id").
        Message("missing route parameter id").
        Build())
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
contract.WriteError(w, r, contract.NewErrorBuilder().
    Type(contract.TypeValidation).
    Code("invalid_json").
    Message("invalid request body").
    Build())
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
contract.WriteResponse(w, r, http.StatusCreated, CreateUserResponse{ID: id}, nil)
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
    app := core.New(core.DefaultConfig(), core.AppDependencies{})
    if err := app.Get("/healthz", healthHandler); err != nil {
        t.Fatal(err)
    }

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

## 14. Prompt Contracts for AI Agents

High-quality Codex tasks should specify:

- goal
- in-scope paths
- out-of-scope paths
- owning module or entrypoint
- public API and dependency policy
- required tests
- validation commands
- done definition

When the request is ambiguous or broad:

- first run an analysis-only pass
- split large work into small reversible cards
- delay code edits until module ownership and scope are explicit

When the task changes or removes an exported symbol:

- enumerate callers first with `rg`
- update every caller in the same change
- verify zero residual references before handoff

Review requests are different from implementation requests:

- findings first
- no code changes unless explicitly requested
- prioritize boundaries, regressions, concurrency, and missing tests

## 15. Rules for AI Agents

Default behavior:
- Prefer stdlib-shaped solutions
- Follow this guide over package convenience APIs when both are possible
- Preserve existing canonical patterns when editing nearby files
- Treat broad bucket names (`utils`, `validator`, `rest`, `pubsub`, `tenant`) as migration debt, not expansion targets

## 16. Agent-First Repo Rules

- Start app-structure work from `reference/standard-service`
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

---

## 17. `contract` Package Rules

`contract` owns **transport primitives only**: request/response envelopes, error
types, HTTP writing helpers, context key accessors, and binding helpers.

### 17.1 Scope boundary

The following categories **do not belong** in `contract` and must live in `x/*`:

| Concern | Correct home |
|---|---|
| Tracing infrastructure (Tracer, Span, Collector, Sampler) | `x/observability` |
| Session lifecycle (SessionStore, SessionValidator, RefreshManager) | `x/tenant` or `x/security` |
| Metrics collection | `x/observability` |
| Business-domain validation rules | caller / domain package |

`contract` may export context keys and ID accessor functions for tracing and
auth, but must not own the full instrumentation or session management subsystem.

### 17.2 Context accessor naming

All context accessor pairs in `contract` use the **With/From** pattern:

```go
// correct
func WithFoo(ctx context.Context, v Foo) context.Context
func FooFromContext(ctx context.Context) Foo

// wrong — mixed prefix order
func ContextWithFoo(ctx context.Context, v Foo) context.Context  // ← forbidden
func FooFromContext(ctx context.Context) Foo
```

Context key types are always **unexported** zero-value structs and always
**inlined** at the call site — no package-level variable:

```go
type fooContextKey struct{}

// correct
context.WithValue(ctx, fooContextKey{}, v)

// wrong
var fooKey fooContextKey            // ← unnecessary variable
context.WithValue(ctx, fooKey, v)
```

### 17.3 Error construction path

One canonical path:

```go
err := contract.NewErrorBuilder().
    Type(contract.TypeValidation).
    Message("validation failed").
    Build()

_ = contract.WriteError(w, r, err)
```

Use `Type(...)` when a known `ErrorType` exists, then override `Code`, `Status`,
or `Category` only when the transport contract truly needs a deviation.

**Do not add** `Ctx.ErrorJSON`, `HandleError`, `SafeExecute`, or new
`NewXxxError(...)` convenience layers on top of this path.

### 17.4 Success response path

One canonical path:

```go
// preferred
contract.WriteResponse(w, r, http.StatusOK, data, meta)
```

`WriteJSON` is a low-level raw-payload writer. It is not the success path.
Use `WriteResponse` from stdlib-shaped handlers.

Do not invent per-feature envelope shapes.

### 17.5 Deprecated API policy

When replacing an API in `contract`:

1. Mark the old symbol: `// Deprecated: Use Foo instead.` (first line of doc comment).
2. **Do not** keep it as a wrapper for more than one minor release.
3. Replace all callers in the same PR that introduces the new API, then delete
   the old symbol — zero dead wrappers at merge time.

Exception: if external callers exist and cannot be migrated in one PR, document
the deprecation deadline in the commit message and a task card.

Change strategy:
1. Keep public behavior unchanged
2. Align local code to canonical style
3. Extract helpers only when repetition justifies it
4. Do not refactor unrelated files opportunistically

Conflict rule: if existing code conflicts with this guide — preserve behavior first, avoid mixing styles in the touched area, move toward canonical when diff is clear and bounded.

---

## 18. Compatibility APIs

- Must not appear in canonical docs
- Must be labeled clearly
- New features must not build on compatibility-only surfaces
- Migration always moves toward canonical, never away

---

## 19. Review Checklist

- Request flow obvious from route to response?
- Code stays close to `net/http` semantics?
- Single visible source for each input?
- JSON decoding explicit or via one transparent helper?
- Middleware transport-only?
- Dependencies explicit in structs or constructors?
- Response and error shapes consistent with framework conventions?
- A new contributor can trace control flow in a few minutes?

---

## 20. Canonical Examples

### Create endpoint

```go
package handlers

import (
    "encoding/json"
    "net/http"

    "github.com/spcent/plumego/contract"
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
        _ = contract.WriteError(w, r, contract.NewErrorBuilder().
            Type(contract.TypeValidation).
            Code("invalid_json").
            Message("invalid request body").
            Build())
        return
    }

    if req.Name == "" {
        _ = contract.WriteError(w, r, contract.NewErrorBuilder().
            Type(contract.TypeRequired).
            Code("missing_name").
            Message("name is required").
            Build())
        return
    }

    id, err := h.Service.Create(req.Name, req.Email)
    if err != nil {
        _ = contract.WriteError(w, r, contract.NewErrorBuilder().
            Type(contract.TypeInternal).
            Code("create_user_failed").
            Message("failed to create user").
            Build())
        return
    }

    _ = contract.WriteResponse(w, r, http.StatusCreated, CreateUserResponse{ID: id}, nil)
}
```

### Route wiring

```go
func registerRoutes(app *core.App, userHandler handlers.UserHandler) error {
    if err := app.Post("/users", userHandler.Create); err != nil {
        return err
    }
    return nil
}
```

### Test

```go
func TestCreateUser(t *testing.T) {
    app := core.New(core.DefaultConfig(), core.AppDependencies{})
    h := handlers.UserHandler{Service: stubUserService{ID: "u_123"}}
    if err := app.Post("/users", h.Create); err != nil {
        t.Fatal(err)
    }

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
