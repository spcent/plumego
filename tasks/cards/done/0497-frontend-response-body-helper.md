# Card 0497: Frontend Response Body Helper

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P1
State: done
Primary Module: x/frontend
Owned Files:
- `x/frontend/frontend_test.go`
Depends On: none

Goal:
Consolidate repeated frontend response body containment assertions.

Problem:
`frontend_test.go` repeats `strings.Contains(rec.Body.String(), ...)` checks
for served files, fallback pages, custom error pages, and precompressed assets.
The tests are checking the same HTTP body contract with slightly different
failure messages.

Scope:
- Add a local helper for asserting a recorder body contains an expected
  fragment.
- Use it for straightforward response body checks.

Non-goals:
- Do not change frontend serving behavior, routing, cache headers, or security
  checks.
- Do not rewrite error string or traversal-specific tests.
- Do not add dependencies.

Files:
- `x/frontend/frontend_test.go`

Tests:
- `go test -race -timeout 60s ./x/frontend/...`
- `go test -timeout 20s ./x/frontend/...`
- `go vet ./x/frontend/...`

Docs Sync:
No docs change required; this is test cleanup only.

Done Definition:
- Straightforward response body assertions use a named helper.
- The listed validation commands pass.

Outcome:
- Added `assertBodyContains` for frontend response body checks.
- Validation passed for frontend race tests, normal tests, and vet.
