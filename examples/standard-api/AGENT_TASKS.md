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

**Handler method on APIHandler** (lifecycle dependencies already wired — Logger, ServiceName, Version):
```go
func (h APIHandler) MyEndpoint(w http.ResponseWriter, r *http.Request) {
    logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, response, nil))
}
```

**Handler with injected dependencies** (declare the interface in the handler,
	wire the concrete implementation from the owning domain package in `routes.go`):
```go
// handler/widgets.go
type WidgetRepository interface {
    Get(ctx context.Context, id string) (Widget, bool)
}
type WidgetHandler struct {
    Repo   WidgetRepository
    Logger plumelog.StructuredLogger // pass a.Core.Logger() from routes.go
}

func (h WidgetHandler) GetByID(w http.ResponseWriter, r *http.Request) {
    id := router.Param(r, "id")
    widget, ok := h.Repo.Get(r.Context(), id)
    if !ok {
        logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
            Type(contract.TypeNotFound).Detail("id", id).Message("not found").Build()))
        return
    }
    logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, widget, nil))
}

// internal/domain/widget/store.go
func NewMemoryStore() *MemoryStore { /* ... */ }

// app/routes.go
widgets := handler.WidgetHandler{Repo: widget.NewMemoryStore(), Logger: a.Core.Logger()}
if err := a.Core.Get("/api/v1/widgets/:id", http.HandlerFunc(widgets.GetByID)); err != nil {
    return err
}
```

`logWriteErr` is defined in `internal/handler/write.go`. It logs the error at Warn
level when the write fails (e.g., client disconnect after headers are sent).

**POST with request body decode**:
```go
func (h WidgetHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req struct{ Name string `json:"name"` }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
            Type(contract.TypeBadRequest).Message("body must be valid JSON").Build()))
        return
    }
    // validate, call repo, write 201
}
```

For a handler that owns a fixed request shape, prefer the strict decode path used
by the `items.go` write handlers (Create/Update/Patch) — `decodeJSONStrict(r, &req)`
rejects unknown fields so a client typo is reported instead of silently dropped.
Keep the lenient `json.NewDecoder(r.Body).Decode(...)` form only when the endpoint
must tolerate forward-compatible extra fields.

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
    Logger:      a.Core.Logger(),
    Checkers: []health.ComponentChecker{
        &dbChecker{db: myDB},
    },
}
```

`GET /readyz` probes **all** checkers regardless of prior failures. A 503
TypeUnavailable is returned if any checker fails, with each failing component
name as a detail key and its error message as the value:
`detail.database="connection refused"`. This lets operators see every unhealthy
dependency in a single response.

### Add a DELETE or LIST endpoint

Extend the repository interface in the handler file and add a method to both the
interface and the domain store, then register the route.

```go
// 1. Extend the interface in handler/widgets.go
type WidgetRepository interface {
    Get(ctx context.Context, id string) (Widget, bool)
    List(ctx context.Context) []Widget
    Delete(ctx context.Context, id string) bool
}

// 2. Implement in internal/domain/widget/store.go
func (s *MemoryStore) List(_ context.Context) []Widget { ... }
func (s *MemoryStore) Delete(_ context.Context, id string) bool { ... }

// 3. Add handler methods in handler/widgets.go
func (h WidgetHandler) List(w http.ResponseWriter, r *http.Request) {
    logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusOK, h.Repo.List(r.Context()), nil))
}

func (h WidgetHandler) Delete(w http.ResponseWriter, r *http.Request) {
    id := router.Param(r, "id")
    if !h.Repo.Delete(r.Context(), id) {
        logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
            Type(contract.TypeNotFound).Detail("id", id).Message("not found").Build()))
        return
    }
    // contract.WriteResponse skips the body for 204 (statusDisallowsBody);
    // using it here keeps the response path consistent with all other handlers.
    logWriteErr(h.Logger, contract.WriteResponse(w, r, http.StatusNoContent, nil, nil))
}

// 4. Register both verbs in app/routes.go using routeReg.
// routeReg captures only the first registration error so the table stays flat;
// inspect reg.err once at the end rather than checking each call individually.
wg := newRouteReg(a.Core.Group("/api/v1"))
wg.get("/widgets",        http.HandlerFunc(widgets.List))
wg.delete("/widgets/:id", http.HandlerFunc(widgets.Delete))
return wg.err
```

### Validate multiple required fields — collect all errors in one response

The canonical pattern collects **all** missing required fields before writing an
error response. Clients receive a single actionable reply that names every field
they need to fix, rather than discovering failures one at a time across multiple
round trips.

`ErrorBuilder.Detail(key, value)` adds one entry per failing field. The builder
accumulates details and writes them together in a single 400 response.

```go
// requireWidgetFields checks that name and color are both non-empty.
// It adds a per-field detail to eb for each empty field and returns false
// when at least one is missing — callers write one error for all problems.
func requireWidgetFields(eb *contract.ErrorBuilder, name, color string) (*contract.ErrorBuilder, bool) {
    valid := true
    if name == "" {
        eb = eb.Detail("name", "field is required")
        valid = false
    }
    if color == "" {
        eb = eb.Detail("color", "field is required")
        valid = false
    }
    return eb, valid
}

func (h WidgetHandler) Create(w http.ResponseWriter, r *http.Request) {
    var req struct {
        Name  string `json:"name"`
        Color string `json:"color"`
    }
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        if errors.Is(err, io.EOF) {
            logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
                Type(contract.TypeRequired).
                Code("widget.create.body_required").
                Message("request body is required").
                Build()))
            return
        }
        logWriteErr(h.Logger, contract.WriteError(w, r, contract.NewErrorBuilder().
            Type(contract.TypeBadRequest).
            Code("widget.create.invalid_json").
            Message("request body must be valid JSON").
            Build()))
        return
    }

    // Body/JSON errors short-circuit above; field validation collects all problems.
    eb := contract.NewErrorBuilder().
        Type(contract.TypeRequired).
        Code("widget.fields.required").
        Message("one or more required fields are missing")
    if eb, ok := requireWidgetFields(eb, req.Name, req.Color); !ok {
        logWriteErr(h.Logger, contract.WriteError(w, r, eb.Build()))
        return
    }
    // all fields valid — proceed
}
```

Each detail entry carries the field name as the key so clients know exactly which
field to highlight (`{"details":{"name":"field is required","color":"field is
required"}}`). Error codes are namespaced (`<resource>.<operation>.<reason>`) for
machine-readable disambiguation. Write one constant per distinct error code at the
top of the handler file.

### Add a config field

1. Add the field to `AppConfig` in `internal/config/config.go`.
2. Set a safe default in `Defaults()`.
3. Read from `.env` in `applyEnvMap()` when local file support is needed.
4. Read from environment in `applyEnv()`.
5. Optionally expose as a flag in `applyFlags()` — no extra maintenance needed;
   `filterFlagArgs` picks up the new flag automatically via `fs.VisitAll`.
6. Use the field in `routes.go` or pass it to a handler constructor.

### Add middleware

1. Import the middleware package in `internal/app/app.go`.
2. Add it to `app.Use(...)` in `New()` at the correct position.
3. Document why the order was chosen in a short inline comment.

Middleware that runs on all routes goes in `app.Use`. Middleware that applies
to a specific route group wraps the handler in `routes.go`.

**Route-group middleware** — when every route under a prefix needs the same
middleware (e.g., all `/api/v1/*` routes require authentication), wrap the
group's routes with a shared middleware rather than wrapping each handler:

```go
// routes.go: wrap a whole group with one middleware call.
// auth wraps every handler registered on the returned group — no per-route
// wrapping needed. Read-only routes outside this group are unaffected.
authed := auth.Middleware(cfg.App.AuthSecret)

v1 := newRouteReg(a.Core.Group("/api/v1"))
// All routes on v1 run through authed first.
v1.get("/items",        authed(http.HandlerFunc(items.List)))
v1.post("/items",       authed(http.HandlerFunc(items.Create)))
v1.get("/items/:id",    authed(http.HandlerFunc(items.GetByID)))
v1.put("/items/:id",    authed(http.HandlerFunc(items.Update)))
v1.patch("/items/:id",  authed(http.HandlerFunc(items.Patch)))
v1.delete("/items/:id", authed(http.HandlerFunc(items.Delete)))
```

For mixed cases where only a subset of routes in a group needs additional
protection, compose guards: `authed(writeGuard(http.HandlerFunc(items.Create)))`.
The outermost middleware in the chain runs first on every incoming request.

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
