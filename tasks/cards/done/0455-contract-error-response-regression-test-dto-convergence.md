# Card 0455: contract Error Response Regression Test DTO Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P2
State: done
Primary Module: contract
Owned Files:
- `contract/active_cards_regression_test.go`
Depends On: none

Goal:
Make the contract request-id error response regression test decode a typed
response shape.

Problem:
`TestWriteErrorUsesTopLevelRequestIDAndTypedFields` decodes the fixed
`WriteError` response envelope through `map[string]any`. This test exists to
lock stable contract fields, so it should assert a typed response shape.

Scope:
- Replace the generic map decode with a local typed response struct.
- Keep all request-id, type, severity, and nested omission assertions.

Non-goals:
- Do not change contract behavior or public APIs.
- Do not add dependencies.

Files:
- `contract/active_cards_regression_test.go`

Tests:
- `go test -race -timeout 60s ./contract/...`
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
No docs change required; this is test-only contract tightening.

Done Definition:
- The fixed error response regression test no longer decodes through
  `map[string]any`.
- The listed validation commands pass.

Outcome:
- Replaced the fixed `WriteError` response map decode with a typed assertion
  struct.
- Preserved request-id promotion, type, severity, and nested omission checks.

Validation:
- `go test -race -timeout 60s ./contract/...`
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`
