# Card 0534

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/errors.go
- contract/errors_test.go
Depends On: 2243

Goal:
Make normalized API errors complete by filling a safe default message when callers omit one.

Scope:
- Fill `APIError.Message` from `http.StatusText(status)` when missing.
- Keep caller-provided messages unchanged.
- Add focused coverage for zero-value and builder-created errors.

Non-goals:
- Do not change error envelope shape.
- Do not change error code/category normalization rules.
- Do not add new public error helpers.

Files:
- `contract/errors.go`
- `contract/errors_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this completes existing normalization behavior.

Done Definition:
- `WriteError` never emits an empty message for normalized errors.
- `NewErrorBuilder().Build()` returns an error with a non-empty message.

Outcome:
- `normalizeAPIError` now fills missing messages from the normalized HTTP status text.
- Added coverage for zero-value `APIError` writes and default builder output.
