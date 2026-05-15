# Card 1424

Milestone: v1-breaking-normalization
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: active
Primary Module: router
Owned Files:
- router/*
Depends On:
- 1421

Goal:
- Normalize route metadata and path-rule handling so v1 routing behavior has a
  single documented error model.

Scope:
- Enumerate path normalization, route metadata, reverse routing, and conflict
  behavior before editing.
- Remove deprecated registration or metadata compatibility paths.
- Prefer explicit initialization-time errors or documented fail-fast behavior
  over unclear internal panics.
- Add focused negative tests for invalid metadata/path/param cases.
- Keep router free of response writing, auth policy, and business validation.

Non-goals:
- Do not change handler signatures.
- Do not add middleware behavior.
- Do not introduce non-stdlib dependencies.

Files:
- router/registration.go
- router/metadata.go
- router/path.go
- router/router.go
- router/*_test.go

Tests:
- go test -timeout 20s ./router
- go vet ./router
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update router docs if invalid-path or metadata behavior changes.

Done Definition:
- Removed compatibility paths have no callers.
- Negative path/metadata tests cover the new v1 behavior.
- Router tests and dependency checks pass.

Outcome:
- Removed the route/group relative-path compatibility path for v1:
  `AddRoute` now rejects non-empty paths that do not start with `/`, and
  `Group` fail-fast panics for non-empty prefixes that do not start with `/`.
- Preserved root/group-root registration via the empty path and kept repeated
  leading slash canonicalization for explicit absolute paths.
- Updated router regression tests and module docs to describe the explicit v1
  path model.
- Validation:
  - `go test -timeout 20s ./router`
  - `go vet ./router`
  - `go run ./internal/checks/dependency-rules`
  - `go build ./...`
