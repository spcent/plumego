# Card 0365: Enrich Reference Standard-Service with Canonical Patterns

Priority: P3
State: done
Recipe: specs/change-recipes/fix-bug.yaml
Primary Module: reference/standard-service

## Goal

`reference/standard-service` demonstrates only the minimal happy-path: two GET handlers that
call `contract.WriteResponse`.  New contributors reading it have no example of:

- Structured error responses via `contract.WriteError` / `contract.NewErrorBuilder()`
- Request input: query parameter binding via `contract.BindQuery` or `r.URL.Query().Get()`
- Route grouping with `router.Group` (not shown at all)
- Middleware stack wiring (no middleware is added in the reference)
- Named routes and URL generation via `r.URL`

The style guide (`docs/CANONICAL_STYLE_GUIDE.md`) defines these patterns but the reference
app doesn't demonstrate them, forcing readers to search across X-package examples instead.

## Scope

Add to `reference/standard-service/internal/handler/api.go`:

1. **Error response handler** — a handler that reads a required query param (`?name=`) and
   returns `TypeBadRequest` if missing, or a greeting if present.  Demonstrates
   `contract.NewErrorBuilder().Type(contract.TypeBadRequest)...`.

2. **Query param binding** — use `r.URL.Query().Get("name")` (canonical per style guide §8)
   rather than `contract.BindQuery` (optional helper).  Keep it simple.

Add to `reference/standard-service/internal/app/routes.go`:

3. **Route group** — register the new greeting handler under an `/api/v1` group to show
   `core.Group("/api/v1")` and `group.Get("/greet", ...)`.

4. **Middleware stack** — apply at least `recovery.Recovery(logger)` to the app so new
   contributors see where middleware is attached (in `app.go`, before `RegisterRoutes`).

Do NOT add external dependencies.  All additions must use only stable root packages
(`contract`, `middleware/recovery`, `log`, `core`).

## Non-goals

- Do not add auth, tenant, or any X-package middleware to the reference (that belongs in
  `reference/with-*` variants).
- Do not change the existing `Hello` / `Status` handlers.
- Do not add `BindQuery` — keep query parsing as `r.URL.Query().Get()` per style guide §8.
- Do not add tests to the reference app (it is documentation, not a tested service).

## Files

- `reference/standard-service/internal/handler/api.go` (add `Greet` handler)
- `reference/standard-service/internal/app/routes.go` (add `/api/v1` group + `/greet` route)
- `reference/standard-service/internal/app/app.go` (wire recovery middleware)
- `reference/standard-service/internal/config/config.go` (add logger field if not present)

## Tests

```bash
go build ./reference/standard-service/...
go vet ./reference/standard-service/...
```

Manual verification: `go run ./reference/standard-service/` should start, and:
- `GET /api/v1/greet?name=Alice` → `200 {"message":"hello, Alice"}`
- `GET /api/v1/greet` (no param) → `400` with structured error body

## Done Definition

- `reference/standard-service` builds and vets cleanly.
- A `Greet` handler exists that returns a structured error for missing `name` param.
- Routes are registered using `core.Group` for the `/api/v1` prefix.
- Recovery middleware is applied in `app.go`.
- No X-package imports are introduced.

## Outcome

Completed. Added `Greet` handler in `reference/standard-service/internal/handler/api.go`
demonstrating canonical error response (`TypeRequired` + `Detail("field","name")`)
and query param binding via `r.URL.Query().Get("name")`.  Route registered at
`/api/v1/greet` in `routes.go`.  Middleware stack (recovery, requestid, accesslog)
was already wired in `app.go`.  `core.App` has no `Group()` method so full paths
are used.  `go build ./reference/standard-service/...` and `go vet` pass.
