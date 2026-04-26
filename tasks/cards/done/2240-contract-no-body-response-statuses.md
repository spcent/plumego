# Card 2240

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: contract
Owned Files:
- contract/response.go
- contract/active_cards_regression_test.go
Depends On: 2239

Goal:
Make `WriteJSON` / `WriteResponse` respect HTTP statuses that must not include response bodies.

Scope:
- Do not write JSON bodies for 1xx, 204, or 304 statuses.
- Avoid setting JSON content type when no body is written.
- Preserve status normalization and normal JSON behavior for body-allowed statuses.

Non-goals:
- Do not change the success response envelope for normal statuses.
- Do not change error response behavior.
- Do not add streaming or SSE helpers.

Files:
- `contract/response.go`
- `contract/active_cards_regression_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this aligns helper behavior with HTTP semantics.

Done Definition:
- `WriteJSON` and `WriteResponse` write only headers for no-body statuses.
- Existing success response tests continue to pass.

Outcome:
- Added a success-response status normalizer separate from error-status normalization.
- Made `WriteJSON` / `WriteResponse` skip JSON body and content type for 1xx, 204, and 304 statuses.
- Validation run: `go test -timeout 20s ./contract/...`; `go vet ./contract/...`.
