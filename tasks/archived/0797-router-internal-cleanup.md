# Card 0797

Milestone: Router stable readiness
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: done
Primary Module: router
Owned Files: router/router.go, router/matcher.go, router/dispatch.go, router/cache.go, router/cache_coverage_test.go
Depends On: 0721-router-static-root-and-doc-contract

Goal:
Remove dead internal code and stale cache wording left from prior router optimization passes.

Scope:
- Remove unused internal constants and functions.
- Keep test-only cache helpers only if still used.
- Update stale comments that mention parameterized cache lookup.

Non-goals:
- Public API changes.
- Router snapshot changes.
- New performance work.

Files:
- router/router.go
- router/matcher.go
- router/dispatch.go
- router/cache.go
- router/cache_coverage_test.go

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Not required.

Done Definition:
- No unused internal router helpers remain from the identified set.
- Router tests and vet pass.

Outcome:
- Removed unused router internals left from earlier matching/cache passes:
  `defaultMaxParams`, `minCacheCapacity`, `newRouteMatcher`, and
  `findRouteNodeLocked`.
- Updated stale cache lookup wording to reflect concrete request-path lookup.

Validation:
- `go test -timeout 20s ./router/...`
- `go test -race -timeout 60s ./router/...`
- `go vet ./router/...`
