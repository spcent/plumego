# Card 2249

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/errors.go
- contract/errors_test.go
Depends On: 2248

Goal:
Align `validateAPIError` with the canonical error-status normalization path.

Scope:
- Treat non-error HTTP statuses as invalid for `APIError`.
- Keep 4xx and 5xx statuses valid.
- Add focused coverage for 2xx and 3xx statuses.

Non-goals:
- Do not change success response status handling.
- Do not change `WriteJSON` or `WriteResponse`.
- Do not export `validateAPIError`.

Files:
- `contract/errors.go`
- `contract/errors_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this aligns an internal validator with existing error-write behavior.

Done Definition:
- `validateAPIError` rejects statuses below 400 and above 599.
- Existing error builder/write normalization tests continue to pass.

Outcome:
- `validateAPIError` now rejects non-error HTTP statuses.
- Added coverage for 2xx and 3xx API error statuses.
