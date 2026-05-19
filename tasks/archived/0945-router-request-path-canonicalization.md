# Card 0945

Milestone: Router stable readiness
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: done
Primary Module: router
Owned Files: router/path.go, router/router_contract_test.go, docs/modules/router/README.md
Depends On: 0735-router-duplicate-param-name-contract

Goal:
Align request-path leading-slash canonicalization with registered route path
canonicalization.

Scope:
- Collapse repeated leading slashes in request paths before matching/cache key
  construction.
- Preserve rejection of internal double slashes.
- Add cache-size and match tests for `/path` and `//path` equivalence.
- Update router docs.

Non-goals:
- Redirecting canonical paths.
- Collapsing internal repeated slashes.
- Changing registered route path storage.

Files:
- router/path.go
- router/router_contract_test.go
- docs/modules/router/README.md

Tests:
- go test -timeout 20s ./router/...
- go test -race -timeout 60s ./router/...
- go vet ./router/...

Docs Sync:
- Required.

Done Definition:
- Repeated leading slash requests match and cache like single leading slash
  requests.
- Internal double slash requests remain unmatched.
- Router tests, race tests, and vet pass.

Outcome:
- Request path normalization now collapses repeated leading slashes while
  preserving internal empty segments.
- Added coverage proving `/users/:id` and `//users/:id` match and share one
  cache entry.
- Updated router docs to align request path and registration path leading-slash
  canonicalization.

Validation:
- `go test -timeout 20s ./router/...`
- `go test -race -timeout 60s ./router/...`
- `go vet ./router/...`
