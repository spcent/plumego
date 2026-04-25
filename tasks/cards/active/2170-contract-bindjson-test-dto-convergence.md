# Card 2170: Contract BindJSON Test DTO Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: active
Primary Module: contract
Owned Files:
- `contract/context_test.go`
Depends On: none

Goal:
Make `BindJSON` context tests bind into typed request DTOs instead of generic
maps where the test payload shape is fixed.

Problem:
Several `contract/context_test.go` cases use `map[string]any` as the bind
target even though the tests only exercise a known `name` JSON payload or error
path. This weakens the assertion shape and differs from the canonical typed DTO
style used by handlers.

Scope:
- Introduce a local typed JSON request DTO for these bind tests.
- Replace fixed-shape map bind targets in body-size, body-cache, and JSON error
  cases.

Non-goals:
- Do not change `BindJSON` behavior or public APIs.
- Do not alter validation, query binding, or response helpers.
- Do not add dependencies.

Files:
- `contract/context_test.go`

Tests:
- `go test -race -timeout 60s ./contract/...`
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
No docs change required; this is test-only request DTO convergence.

Done Definition:
- Fixed-shape `BindJSON` tests no longer bind into `map[string]any`.
- The listed validation commands pass.

Outcome:
