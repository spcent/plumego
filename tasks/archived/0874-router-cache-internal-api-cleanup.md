# Card 0874

Milestone: Router stable readiness
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: done
Primary Module: router
Owned Files: router/cache.go, router/dispatch.go, router/cache_coverage_test.go, router/test_helpers_test.go
Depends On: 0729-router-error-response-contract

Goal:
Remove misleading internal cache parameters and keep test-only helpers out of
production code.

Scope:
- Simplify `matchCache.Lookup` to the data it actually uses.
- Move cache-capacity router helpers used only by tests/benchmarks into test
  files.
- Keep cache behavior and metrics unchanged.

Non-goals:
- Replacing the LRU implementation.
- Exposing cache configuration publicly.
- Changing route matching behavior.

Files:
- router/cache.go
- router/dispatch.go
- router/cache_coverage_test.go
- router/test_helpers_test.go

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Not required.

Done Definition:
- Production cache API comments and parameters match actual behavior.
- Test-only cache helpers are no longer in production files.
- Router tests, race tests, and vet pass.

Outcome:
- Simplified `matchCache.Lookup` to accept only the concrete cache key it uses.
- Updated dispatch and cache coverage tests for the new internal signature.
- Moved cache-capacity router helpers used only by tests and benchmarks into
  `router/test_helpers_test.go`.

Validation:
- `go test -timeout 20s ./router/...`
- `go test -race -timeout 60s ./router/...`
- `go vet ./router/...`
