# Card 0124

Milestone:
Recipe: specs/change-recipes/test-boundary.yaml
Priority: P1
State: done
Primary Module: middleware
Owned Files:
  - middleware/conformance/runtime_invariants_test.go
  - middleware/conformance_test.go
  - middleware/layer_boundary_test.go
Depends On: 0123

Goal:
Keep stable middleware conformance tests independent from extension packages.
The current conformance suite imports `x/tenant`, coupling stable middleware
validation to extension package state.

Scope:
- Remove `x/tenant` imports from stable middleware conformance tests.
- Preserve middleware error-schema and next-call invariant coverage using stable fixtures.
- Leave extension integration tests in extension-owned packages.

Non-goals:
- Do not remove extension behavior tests from their owning packages.
- Do not weaken stable middleware invariants.
- Do not change production code.

Files:
- `middleware/conformance/runtime_invariants_test.go`
- `middleware/conformance_test.go`
- `middleware/layer_boundary_test.go`

Tests:
- `go test -timeout 20s ./middleware/...`
- `go run ./internal/checks/dependency-rules`
- `go vet ./middleware/...`

Docs Sync:
- Not required.

Done Definition:
- `rg -n 'github.com/spcent/plumego/x/' middleware --glob '*_test.go'` does not report conformance imports.
- Middleware tests and boundary checks pass.

Outcome:
- Removed `x/tenant` imports and tenant-specific cases from stable middleware
  conformance tests.
- Kept stable error schema coverage through auth, bodylimit, ratelimit,
  recovery, and canonical fixture middleware.
- Left `middleware/layer_boundary_test.go` allowlist strings unchanged; those
  are boundary assertions, not package imports in conformance tests.
- Validated with:
  - `go test -timeout 20s ./middleware/...`
  - `go run ./internal/checks/dependency-rules`
  - `go vet ./middleware/...`
  - `rg -n 'github.com/spcent/plumego/x/' middleware/conformance_test.go middleware/conformance --glob '*_test.go'`
