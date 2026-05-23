# Agent Tasks — standard-service

Operating guide for AI coding agents working in this reference application.

Read `AGENTS.md` (repository root) for the full operating contract.
Read `ARCHITECTURE.md` (this directory) for layout rationale.

---

## Zone Classification

### Safe to modify

| Path | What you can do |
|---|---|
| `internal/handler/api.go` | Add, change, or remove handler methods |
| `internal/handler/health.go` | Update health check logic or readiness checker interface |
| `internal/handler/items.go` | Extend item HTTP endpoints or add new handler files with DI |
| `internal/domain/item/item.go` | Change sample item model or in-memory repository behavior |
| `internal/domain/item/item_test.go` | Add or update domain model tests |
| `internal/handler/handler_test.go` | Add or update handler tests |
| `internal/app/app_test.go` | Update route shape assertions or middleware wiring tests |
| `internal/config/config.go` | Add config fields, change defaults |
| `main.go` | Only the four wiring calls; do not add logic |

### Restricted — requires preflight + reviewer note

| Path | Constraint |
|---|---|
| `internal/app/app.go` | Middleware order is load-bearing; changing it affects all routes |
| `internal/app/routes.go` | Adding a route means adding a public contract; confirm the route is intentional |

### Frozen — do not touch

| Constraint | Reason |
|---|---|
| No `x/*` imports | This service depends only on stable roots. Adding `x/*` breaks the canonical reference. |
| No global variables | Constructor injection only. |
| No `init()` functions | Explicit wiring only. |

---

## Common Task Recipes

### Add a new endpoint

1. Add a handler method to an existing file in `internal/handler/`, or create a new file there.
2. Register the route in `internal/app/routes.go`.
3. Write a test in `internal/handler/handler_test.go`.
4. Run validation (see below).

**Handler with no dependencies** (zero-field struct):
```go
func (h APIHandler) MyEndpoint(w http.ResponseWriter, r *http.Request) {
    _ = contract.WriteResponse(w, r, http.StatusOK, response, nil)
}
```

**Handler with injected dependencies** (declare the interface in the handler,
	wire the concrete implementation from the owning domain package in `routes.go`):
```go
// handler/widgets.go
type WidgetRepository interface {
    Get(id string) (Widget, bool)
}
type WidgetHandler struct{ Repo WidgetRepository }

func (h WidgetHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    id := router.Param(r, "id")
    widget, ok := h.Repo.Get(id)
    if !ok {
        _ = contract.WriteError(w, r, contract.NewErrorBuilder().
            Type(contract.TypeNotFound).Detail("id", id).Message("not found").Build())
        return
    }
    _ = contract.WriteResponse(w, r, http.StatusOK, widget, nil)
}

// internal/domain/widget/store.go
func NewMemoryStore() *MemoryStore { /* ... */ }

// app/routes.go
widgets := handler.WidgetHandler{Repo: widget.NewMemoryStore()}
if err := a.Core.Get("/api/v1/widgets/:id", http.HandlerFunc(widgets.GetByID)); err != nil {
    return err
}
```

**POST with request body decode**:
```go
func (h WidgetHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req struct{ Name string `json:"name"` }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        _ = contract.WriteError(w, r, contract.NewErrorBuilder().
            Type(contract.TypeBadRequest).Message("body must be valid JSON").Build())
        return
    }
    // validate, call repo, write 201
}
```

### Add a readiness check

A `health.ComponentChecker` represents one dependency (database, cache, downstream service).
Implement the interface and register it in `routes.go`.

```go
// Implement health.ComponentChecker for your dependency.
// This type lives in the domain or infrastructure layer, not in handler/.
type dbChecker struct{ db *sql.DB }

func (c *dbChecker) Name() string              { return "database" }
func (c *dbChecker) Check(ctx context.Context) error { return c.db.PingContext(ctx) }

// Wire it in routes.go alongside the other handler dependencies.
health := handler.HealthHandler{
    ServiceName: a.Cfg.App.ServiceName,
    Checkers: []health.ComponentChecker{
        &dbChecker{db: myDB},
    },
}
```

`GET /readyz` probes each checker in order; the first error returns 503
TypeUnavailable with `detail.component` set to the checker's Name() and
`detail.reason` set to the error message.

### Add a DELETE or LIST endpoint

Extend the repository interface in the handler file and add a method to both the
interface and the domain store, then register the route.

```go
// 1. Extend the interface in handler/widgets.go
type WidgetRepository interface {
    Get(id string) (Widget, bool)
    List() []Widget
    Delete(id string) bool
}

// 2. Implement in internal/domain/widget/store.go
func (s *MemoryStore) List() []Widget { ... }
func (s *MemoryStore) Delete(id string) bool { ... }

// 3. Add handler methods in handler/widgets.go
func (h WidgetHandler) List(w http.ResponseWriter, r *http.Request) {
    _ = contract.WriteResponse(w, r, http.StatusOK, h.Repo.List(), nil)
}

func (h WidgetHandler) Delete(w http.ResponseWriter, r *http.Request) {
    id := router.Param(r, "id")
    if !h.Repo.Delete(id) {
        _ = contract.WriteError(w, r, contract.NewErrorBuilder().
            Type(contract.TypeNotFound).Detail("id", id).Message("not found").Build())
        return
    }
    w.WriteHeader(http.StatusNoContent)
}

// 4. Register both verbs in app/routes.go
if err := a.Core.Get("/api/v1/widgets", http.HandlerFunc(widgets.List)); err != nil {
    return err
}
if err := a.Core.Delete("/api/v1/widgets/:id", http.HandlerFunc(widgets.Delete)); err != nil {
    return err
}
```

### Validate multiple fields and return the first error

`contract.WriteError` writes one error per response. The canonical pattern is
fail-fast: validate fields in priority order and return on the first failure.
This keeps error responses simple and predictable.

```go
func (h WidgetHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Name  string `json:"name"`
        Color string `json:"color"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        _ = contract.WriteError(w, r, contract.NewErrorBuilder().
            Type(contract.TypeBadRequest).
            Code("widget.create.invalid_json").
            Message("request body must be valid JSON").
            Build())
        return
    }
    if req.Name == "" {
        _ = contract.WriteError(w, r, contract.NewErrorBuilder().
            Type(contract.TypeRequired).
            Code("widget.name.required").
            Detail("field", "name").
            Message("name is required").
            Build())
        return
    }
    if req.Color == "" {
        _ = contract.WriteError(w, r, contract.NewErrorBuilder().
            Type(contract.TypeRequired).
            Code("widget.color.required").
            Detail("field", "color").
            Message("color is required").
            Build())
        return
    }
    // all fields valid — proceed
}
```

Each error carries a `Detail("field", "<name>")` annotation so clients know
exactly which field to highlight. Error codes are namespaced
(`<resource>.<field>.<reason>`) for machine-readable disambiguation. Write one
constant per distinct error code at the top of the handler file.

### Add a config field

1. Add the field to `AppConfig` in `internal/config/config.go`.
2. Set a safe default in `Defaults()`.
3. Read from `.env` in `applyEnvMap()` when local file support is needed.
4. Read from environment in `applyEnv()`.
5. Optionally expose as a flag in `applyFlags()`.
6. Use the field in `routes.go` or pass it to a handler constructor.

### Add middleware

1. Import the middleware package in `internal/app/app.go`.
2. Add it to `app.Use(...)` in `New()` at the correct position.
3. Document why the order was chosen in a short inline comment.

Middleware that runs on all routes goes in `app.Use`. Middleware that applies
to a specific route group wraps the handler in `routes.go`.

---

## Validation Commands

After any change to this service, run in order:

```bash
# 1. Module tests (standard-service is its own Go module; cd in first)
cd reference/standard-service && go test -race -timeout 30s ./...

# 2. Boundary check — must not import x/*   (run from repo root)
go run ./internal/checks/dependency-rules

# 3. Module manifests                        (run from repo root)
go run ./internal/checks/module-manifests

# 4. Full gates (for cross-module or release-relevant changes)
make gates
```

---

## Non-goals

Do not use this reference app to:

- Demonstrate `x/*` extensions (use the corresponding `with-<capability>` demo)
- Implement business logic beyond the minimal echo/status/greet endpoints
- Show authentication patterns (those belong in `reference/with-tenant` or `x/security` examples)
- Prototype new features (this is a stable reference, not a sandbox)

If a task requires changes that violate these non-goals, create a new
`reference/with-<capability>` demo instead of modifying this file.
