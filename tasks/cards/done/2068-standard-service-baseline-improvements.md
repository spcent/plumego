# Card 2068

Milestone: standalone
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: implementation
Priority: P1
State: done
Primary Module: reference/standard-service
Owned Files:
- `reference/standard-service/internal/app/app.go`
- `reference/standard-service/internal/app/routes.go`
- `reference/standard-service/internal/app/app_test.go`
- `reference/standard-service/internal/handler/guard.go`

## Goal

Improve `reference/standard-service` so it fully demonstrates plumego's stable-root
value rather than leaving three stable-root middleware unshown and placing a middleware
function in the wrong package.

After this card the reference app:
1. Wires `middleware/security/headers`, `middleware/httpmetrics`, and `middleware/cors`
   in the canonical middleware stack — production-critical middleware every service needs
2. Moves `handler/guard.go` → `internal/middleware/guard.go` so per-route middleware
   lives in the right package and the handler package owns only handlers

## Scope

**Task A — Add three middleware to the stack (`internal/app/app.go`)**

Wire in this order (existing positions preserved, new ones inserted):
```
requestid → security → cors → recovery → accesslog → bodylimit → httpmetrics → timeout
```

- `middleware/security.Middleware(security.Config{})` — default policy (X-Frame-Options,
  X-Content-Type-Options, Referrer-Policy); fallible constructor, propagate error
- `cors.Middleware(cors.CORSOptions{})` — permissive defaults for reference/dev; add comment
  pointing to `cors.StrictDefaultOptions` for production
- `httpmetrics.Middleware(metrics.NewNoopCollector())` — noop observer so tests are
  unaffected; add comment showing how to swap in a real collector

Remove the "Extension points" comment block (those points are now wired).

**Task B — Move guard to the right package**

1. Create `reference/standard-service/internal/middleware/guard.go`:
   - Same implementation as `handler/guard.go` — no logic changes
   - Package name: `appmiddleware` (avoids collision with imported `middleware` package)
   - Export: `RequireWriteKey(key string, logger plumelog.StructuredLogger) func(http.Handler) http.Handler`

2. Update `reference/standard-service/internal/app/routes.go`:
   - Replace import `standard-service/internal/handler` guard reference with
     `standard-service/internal/middleware`
   - `writeGuard := appmiddleware.RequireWriteKey(a.Cfg.App.WriteKey, a.Core.Logger())`

3. Delete `reference/standard-service/internal/handler/guard.go`

4. Update `reference/standard-service/internal/handler/handler_test.go`:
   - Tests for `RequireWriteKey` move from `handler_test` → new
     `reference/standard-service/internal/middleware/guard_test.go`

## Non-goals

- No changes to handler logic, domain, config, or test assertions about responses
- No new config fields
- Do not wire a real `metrics.HTTPObserver` implementation (noop is correct for reference)
- Do not change the Hello() endpoint list or the drift test
- No external dependency changes

## Files

- `reference/standard-service/internal/app/app.go`
- `reference/standard-service/internal/app/routes.go`
- `reference/standard-service/internal/handler/guard.go` (delete)
- `reference/standard-service/internal/middleware/guard.go` (new)
- `reference/standard-service/internal/middleware/guard_test.go` (new, moved from handler_test.go)

## Acceptance Tests

- `reference/standard-service/internal/app/app_test.go`: `TestMiddlewareStack` (existing — must still pass)
- `reference/standard-service/internal/middleware/guard_test.go`: `TestRequireWriteKeyEmptyBypass`, `TestRequireWriteKeyCorrectKey`, `TestRequireWriteKeyMissingHeader`, `TestRequireWriteKeyWrongHeader`

## Tests

- Security headers are present in responses (X-Frame-Options, X-Content-Type-Options)
- CORS headers present on cross-origin requests (Access-Control-Allow-Origin)
- httpmetrics noop does not change response shape or timing behavior
- RequireWriteKey behavior identical after move: empty bypass, correct key allowed, wrong key 401

## Docs Sync

- `reference/standard-service/ARCHITECTURE.md` — update middleware stack table
- `reference/standard-service/README.md` — note security/cors/metrics in middleware section

## Validation

```bash
go test -race ./reference/standard-service/...
go vet ./reference/standard-service/...
go run ./internal/checks/dependency-rules
```

## Done Definition

- [ ] All three middleware wired in app.go with correct order and comments
- [ ] guard.go moved to internal/middleware/, handler/guard.go deleted
- [ ] guard_test.go exists in internal/middleware/ with all four tests passing
- [ ] All Validation commands exit 0
- [ ] gofmt -l . produces no output for reference/standard-service
- [ ] ARCHITECTURE.md and README.md updated

## Outcome

All three tasks completed. Three commits on `feat/standard-service-improvements`:

1. Wired `middleware/security/headers` (default policy), `middleware/cors` (permissive dev defaults),
   and `middleware/httpmetrics` (noop collector) into `app.go` in the canonical order:
   `requestid → security → cors → recovery → accesslog → bodylimit → httpmetrics → timeout`.
   Removed the stale "Extension points" comment block.

2. Moved `RequireWriteKey` from `handler/guard.go` → `internal/middleware/guard.go`
   (package `appmiddleware`). Routes.go updated to `appmiddleware.RequireWriteKey`.
   Guard tests moved to `internal/middleware/guard_test.go`. `handler/guard.go` deleted.

3. Updated `ARCHITECTURE.md` (directory tree, new middleware/ section, dependency diagram)
   and `README.md` (middleware stack description, canonical files list).

All tests pass: `go test -race ./...`, `go vet ./...`, `dependency-rules`, `reference-layout`.
