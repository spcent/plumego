# Plumego Canonical Style Guide

Status: Draft  
Audience: Plumego maintainers, contributors, AI agents, code reviewers  
Scope: Canonical coding style for Plumego core-facing web applications and examples  
Applies to: `core`, `router`, `middleware`, official examples, docs, code generation prompts, AI/code-agent workflows

---

# 1. Purpose

This document defines the **single recommended way** to build HTTP applications with Plumego.

Its primary goal is not API completeness. Its goal is to make Plumego:

- easy for humans to read and review,
- easy for AI agents to modify safely,
- easy to refactor in small steps,
- hard to misuse through hidden framework behavior,
- stable enough that examples, generated code, and production code all look the same.

This guide is intentionally strict.

Plumego may support compatibility APIs, adapters, and optional helper packages, but **official examples, docs, scaffolds, and AI-generated code must follow the canonical style in this document**.

---

# 2. Core philosophy

## 2.1 Standard library first

Plumego is a thin layer over Go's standard HTTP model.

Canonical Plumego code should stay close to:

- `http.Handler`
- `http.HandlerFunc`
- `*http.Request`
- `http.ResponseWriter`
- standard `context.Context`
- `httptest`

Plumego adds ergonomics and conventions, but it must not hide the standard library execution model.

## 2.2 One obvious way

For each common task, there should be exactly one recommended pattern.

Examples:

- one recommended app bootstrap style
- one recommended route registration style
- one recommended handler style
- one recommended middleware style
- one recommended JSON request decode style
- one recommended JSON response style
- one recommended error response shape
- one recommended testing style

If Plumego supports multiple ways to do the same thing, only one may be canonical.

## 2.3 Explicit over implicit

Canonical Plumego code must prefer explicit data flow and explicit control flow.

Avoid hidden behavior such as:

- automatic mixed-source binding,
- service lookup through request context,
- magical default response wrappers,
- route registration with side effects not visible in the local file,
- middleware that silently injects business dependencies.

## 2.4 Small-step refactorability

A human or AI agent should be able to change one route, one handler, one middleware, or one response format without needing to understand the entire repository.

That requires:

- narrow package boundaries,
- stable interfaces,
- shallow call paths,
- minimal indirection,
- canonical structure across examples.

---

# 3. Non-goals

This guide does **not** attempt to make Plumego:

- the shortest possible framework for demos,
- the most feature-rich framework,
- a macro-heavy or annotation-driven framework,
- an auto-binding or auto-DI framework,
- a platform where all capabilities are surfaced from `App` directly.

The canonical style optimizes for **predictability**, not maximal convenience.

---

# 4. Canonical package roles

To remain AI-agent friendly, package names must closely match package responsibilities.

## 4.1 `core`

`core` is the main runtime entry point.

It should contain only foundational application primitives such as:

- app construction,
- app lifecycle,
- route registration entry points,
- middleware attachment,
- server startup integration,
- minimal core-facing abstractions.

`core` must not become a feature catalog.

## 4.2 `router`

`router` is for routing concerns only.

Allowed responsibilities:

- route matching,
- route params,
- grouping,
- route tree / lookup,
- route registration internals,
- static route mounting if retained.

Not allowed in canonical routing packages:

- repositories,
- resource controllers,
- validators,
- JSON writers,
- business response wrappers,
- persistence abstractions.

## 4.3 `middleware`

`middleware` is for transport-layer cross-cutting concerns only.

Allowed responsibilities:

- logging,
- recovery,
- timeout,
- request id,
- CORS,
- auth adapters,
- rate limiting adapters,
- tracing adapters.

Not allowed in canonical middleware:

- service injection,
- ORM lookup,
- business DTO assembly,
- hidden request binding required by handlers,
- response rewriting that changes domain semantics.

## 4.4 Extension packages

Packages such as AI, ops, tenant, websocket, webhook, frontend hosting, and scheduler belong to extension or capability layers, not the canonical core learning path.

They may exist and be documented, but they must not define the primary coding style for standard HTTP applications.

---

# 5. Canonical application structure

The canonical file layout for a Plumego HTTP service should be simple and shallow.

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

## 5.1 Directory principles

- `cmd/` contains startup only.
- `internal/httpapp/` contains HTTP wiring only.
- `internal/domain/` contains business logic.
- `internal/platform/` contains reusable transport or infra helpers.
- Do not mix routing, domain logic, persistence, and response helpers in one package.

## 5.2 Canonical bootstrap boundary

`main.go` should do very little:

- build config,
- build dependencies,
- construct app,
- register routes,
- start server.

Business logic must not be assembled inline in `main.go` beyond simple dependency wiring.

---

# 6. Canonical app bootstrap

The canonical style uses a small, explicit app construction flow.

## 6.1 Recommended shape

```go
package main

import (
    "log"
    "net/http"

    "github.com/spcent/plumego/core"
)

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

## 6.2 Canonical bootstrap rules

- Prefer constructing the app in one visible place.
- Prefer `http.ListenAndServe` compatibility.
- Attach global middleware explicitly and near startup.
- Call one route registration function per bounded HTTP area.
- Avoid hidden auto-registration from unrelated packages.

## 6.3 Avoid

- giant startup files with embedded business logic,
- route registration spread across many `init()` functions,
- startup driven by implicit side effects,
- app behavior that depends on package import order.

---

# 7. Canonical route registration

## 7.1 One public route registration style

Canonical code should use **one primary route registration API**.

Recommended style:

```go
func registerRoutes(app *core.App) {
    app.Get("/healthz", healthHandler)
    app.Post("/users", createUserHandler)
}
```

If compatibility APIs exist, they are not canonical.

## 7.2 Route registration rules

- One route, one explicit method, one explicit path, one explicit handler.
- Prefer static registration over reflection or discovery.
- Prefer local visibility over global magical grouping.
- Route registration should remain grep-friendly.

## 7.3 Groups

Route groups may be used for:

- path prefix,
- shared middleware.

Groups must not become a container for hidden business policy, service injection, or version policy that is difficult to trace.

---

# 8. Canonical handler style

## 8.1 Preferred handler shape

Canonical Plumego handlers should follow one stable signature.

The preferred shape for official examples is the simplest long-term style supported by core.

If Plumego continues to support several signatures internally, the canonical guide still chooses one. For canonical docs and examples, assume:

```go
func createUserHandler(w http.ResponseWriter, r *http.Request)
```

or, if Plumego's chosen long-term public direction is a single lightweight context type:

```go
func createUserHandler(ctx *plumego.Context) error
```

But only **one** of these may be canonical at a time.

Until v1 API contraction is complete, official docs should bias toward the style that is closest to `net/http` and easiest to adapt.

## 8.2 Handler responsibilities

A canonical handler should do only transport work:

1. read request inputs,
2. validate transport-level presence and shape,
3. call a service,
4. write a response.

## 8.3 Handler anti-patterns

Do not put these directly in handlers unless unavoidable:

- raw SQL,
- repository construction,
- dependency lookup from context maps,
- config loading,
- complex transaction orchestration,
- hidden retries,
- logging setup,
- response envelope invention per file.

---

# 9. Canonical request decoding

## 9.1 Explicit decode is the default

Canonical Plumego code should decode JSON explicitly.

Recommended pattern:

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

## 9.2 Allowed helper level

A thin helper such as `DecodeJSON(r, &req)` is acceptable if it is transparent and predictable.

## 9.3 Not canonical

The following are **not** canonical for general examples:

- middleware-first binding that stores DTOs into request context,
- mixed binding across path, query, headers, and body automatically,
- auto-validation with hidden translation rules,
- request DTO retrieval through magic keys.

Such helpers may exist, but must be treated as optional compatibility or advanced features.

---

# 10. Canonical parameter access

## 10.1 Route params

Route params should be read explicitly where they are used.

```go
id := Param(r, "id")
if id == "" {
    WriteError(w, BadRequest("missing_id", "missing route parameter id"))
    return
}
```

## 10.2 Query params

Use `r.URL.Query()` or one thin helper. Do not layer multiple param APIs unless there is a compelling reason.

## 10.3 Headers

Read headers directly from `r.Header` or a single thin helper.

The canonical rule is simple: **data source visibility matters**. A reader should see whether the value came from path, query, header, or body.

---

# 11. Canonical middleware style

## 11.1 Middleware must be plain wrapping

Canonical middleware should be a standard wrapping function.

```go
type Middleware func(http.Handler) http.Handler
```

or whatever exact standard Plumego exports that matches this shape.

## 11.2 Middleware responsibilities

Allowed:

- request id,
- logging,
- recovery,
- timeout,
- auth gate,
- rate limiting,
- tracing,
- metrics.

## 11.3 Middleware must not hide business inputs

Not canonical:

- middleware that builds business DTOs required by handlers,
- middleware that injects service objects into context for routine use,
- middleware that silently mutates domain meaning,
- middleware that writes normal business success responses.

## 11.4 `next` semantics must stay simple

Avoid complex middleware semantics such as:

- calling `next` multiple times,
- pre/post/error split lifecycles for ordinary middleware,
- control flow that requires reading framework internals to understand.

---

# 12. Canonical error model

## 12.1 One error response shape

Plumego must expose one canonical transport error structure for JSON APIs.

Recommended baseline:

```go
type ErrorResponse struct {
    Code    string `json:"code"`
    Message string `json:"message"`
}
```

Optional fields such as `details`, `request_id`, or `category` are acceptable if they are framework-wide and stable.

## 12.2 One write path

Canonical code should use one helper to emit errors, for example:

```go
WriteError(w, BadRequest("invalid_json", "invalid request body"))
```

or an equivalent single canonical helper.

## 12.3 Error rules

- transport errors should be structured,
- status code should be explicit or derived predictably,
- error codes should be stable machine-readable strings,
- end-user message should be concise and safe,
- logs may include richer internal details, but responses should not.

## 12.4 Avoid

- ad hoc `http.Error` text in normal JSON APIs,
- file-by-file custom envelopes,
- one middleware returning a different shape from the rest of the framework,
- helpers that silently wrap success and error responses differently per module.

---

# 13. Canonical success responses

## 13.1 Response writing should be minimal and direct

For JSON APIs, use one response helper or direct JSON encoding consistently.

Recommended shape:

```go
WriteJSON(w, http.StatusCreated, CreateUserResponse{ID: id})
```

## 13.2 Avoid response envelope proliferation

Do not invent multiple success wrappers such as:

- `{ "success": true, "data": ... }`
- `{ "code": 0, "result": ... }`
- `{ "status": "ok", "payload": ... }`

unless the project has committed to one global envelope. If there is a global envelope, it must be documented once and used everywhere.

## 13.3 Status codes matter

Canonical code must set meaningful HTTP status codes explicitly for non-trivial cases.

---

# 14. Canonical dependency flow

## 14.1 Constructor injection, not request-context service lookup

Business dependencies should be injected through constructors or structs.

Recommended pattern:

```go
type UserHandler struct {
    Service user.Service
}

func (h UserHandler) Create(w http.ResponseWriter, r *http.Request) {
    // ...
}
```

## 14.2 Avoid service locator patterns

Not canonical:

```go
svc := MustServiceFromContext(r.Context(), "userService")
```

This obscures dependency shape and makes AI-generated changes more fragile.

## 14.3 Route wiring should show dependencies

The file that registers handlers should make dependency ownership obvious.

---

# 15. Canonical testing style

## 15.1 `httptest` first

Canonical Plumego HTTP tests should use `httptest` and standard request/response assertions.

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

## 15.2 Testing rules

- test routes as handlers, not only through mocks,
- prefer table-driven tests for pure transport cases,
- isolate domain mocks or stubs behind interfaces,
- keep setup local and small.

## 15.3 Avoid

- test harnesses that require full framework bootstrapping for simple route tests,
- hidden global app state between tests,
- tests that rely on implicit middleware ordering not visible in the test file.

---

# 16. Canonical documentation rules

## 16.1 Official examples must teach one style

Official docs and examples must not mix canonical and compatibility patterns in the same tutorial.

Each example must be labeled as one of:

- `canonical`
- `compatibility`
- `experimental`

## 16.2 Canonical examples must prefer boring clarity

Examples are not a place to demonstrate every Plumego feature. They are a place to demonstrate the best long-term shape.

## 16.3 README discipline

The repository README and package-level docs must present:

1. the canonical bootstrap,
2. the canonical route registration style,
3. the canonical request decode style,
4. the canonical error write style,
5. the canonical test style.

Capability packs belong below the core path, not above it.

---

# 17. Rules for AI agents and code generation

This section is normative.

## 17.1 Default behavior for agents

AI agents working in Plumego repositories must:

- prefer standard-library-shaped solutions,
- follow this guide over package convenience APIs when both are possible,
- avoid introducing new public helper styles unless explicitly requested,
- preserve existing canonical patterns when editing nearby files,
- avoid hidden context-based dependency flow,
- avoid new wrapper abstractions around simple handlers unless justified.

## 17.2 When multiple APIs exist

If Plumego exposes multiple ways to achieve the same behavior, agents must choose the canonical one from this guide, not the shortest one.

## 17.3 Do not widen the style surface

Agents must not introduce:

- new response helper families,
- new handler signatures,
- new route registration idioms,
- new service locator patterns,
- new implicit DTO flow through middleware,
- new business logic inside middleware,
- new package roles that blur core boundaries.

## 17.4 Change strategy

Agents should prefer small, reversible edits:

1. keep public behavior unchanged,
2. align local code to canonical style,
3. extract helpers only when repeated enough to justify them,
4. do not refactor unrelated files opportunistically.

---

# 18. Compatibility APIs and migration posture

This guide does not require immediate removal of compatibility APIs.

However:

- compatibility APIs must not appear in canonical docs,
- compatibility APIs should be labeled clearly,
- new features should not be built on compatibility-only surfaces,
- migration should always move toward canonical style, not away from it.

If a code path uses a deprecated or compatibility API, maintainers and AI agents should prefer migration when touching that area, unless the change scope must remain minimal.

---

# 19. Review checklist

A PR follows the canonical style if reviewers can answer yes to most of the following:

- Is the request flow obvious from route to response?
- Does the code stay close to `net/http` semantics?
- Is there a single visible way inputs are read?
- Is JSON decoding explicit or done through one transparent helper?
- Are middleware responsibilities transport-only?
- Are dependencies explicit in structs or constructors?
- Is the response format consistent with framework conventions?
- Is the error shape consistent with framework conventions?
- Would an AI agent likely generate similar code next time?
- Would a new contributor understand the control flow quickly?

If the answer is no to several items, the PR is probably widening Plumego's style surface and should be revised.

---

# 20. Canonical examples

## 20.1 Canonical create endpoint

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

## 20.2 Canonical route wiring

```go
func registerRoutes(app *core.App, userHandler handlers.UserHandler) {
    app.Post("/users", userHandler.Create)
}
```

## 20.3 Canonical test

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

# 21. Forbidden patterns in canonical code

The following patterns must not appear in canonical docs, templates, or generated starter code:

- mixing `Get`, `GetCtx`, and `GetHandler` styles in one example,
- binding request DTOs in middleware and reading them from context in ordinary CRUD handlers,
- using `router` as a home for repositories or resource controllers,
- retrieving services from request context maps,
- exposing platform capabilities through `core.App` convenience methods when ordinary composition is enough,
- introducing new response helper families for a single feature,
- teaching multiple equally valid bootstraps in first-party docs,
- hiding route registration behind import side effects,
- placing business logic in middleware,
- introducing package names that blur transport and domain boundaries.

---

# 22. Versioning note

Until Plumego v1 stabilizes its public surface, this guide acts as the forward-compatible style target.

Where the repository currently contains broader or older API surfaces, maintainers should:

- keep this guide as the canonical north star,
- document compatibility paths explicitly,
- move examples and new packages toward canonical style,
- prefer contraction over expansion in core-facing APIs.

---

# 23. Final rule

If a human reviewer or AI agent cannot understand how a request is handled within a few minutes by reading only the route registration, middleware, and handler file, the code is not canonical enough.

Plumego's long-term strength is not that it can support many styles. It is that it can support one stable, explicit, standard-library-aligned style extremely well.
