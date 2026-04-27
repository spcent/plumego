# Card 0059

Priority: P0

Goal:
- Audit every exported symbol in all stable library roots and remove or unexport
  implementation details that leaked into the public API surface before the v1
  freeze.

Scope:
- `router`: unexport `CacheEntry`, `PatternCacheEntry`, `RouteMatcher`,
  `NewRouteMatcher`, `IsParameterized` — all internal implementation details
  with no external users.
- `metrics`: remove `MetricsMiddleware` and `MetricsHandler` — middleware
  helpers that violate the module boundary (middleware belongs in
  `middleware/httpmetrics`); no external users exist.
- Verify no external callers are broken (`go build ./...`).

Non-goals:
- Do not change the behaviour of the router or metrics at runtime.
- Do not move or rename exported symbols that have documented external users.
- Do not audit `x/*` extension packages in this card.

Files:
- `router/cache.go`
- `router/matcher.go`
- `router/pool.go`
- `router/dispatch.go`
- `router/cache_coverage_test.go`
- `metrics/adapter.go`
- `metrics/adapter_test.go`

Tests:
- `go test ./router/... ./metrics/...`
- `go vet ./router/... ./metrics/...`

Docs Sync:
- No doc changes required; unexported symbols are not documented.

Done Definition:
- `router` package exports no internal cache or matcher implementation types.
- `metrics` package exports no HTTP middleware functions.
- All tests pass.
