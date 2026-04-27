# Card 0519

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/errors.go
- contract/response.go
- contract/errors_test.go
Depends On:

Goal:
Prevent invalid HTTP status values from reaching response writers through the contract error and JSON write paths.

Scope:
- Normalize invalid status values before `WriteError`, `WriteJSON`, or `ErrorBuilder.Build` can commit headers.
- Reuse one small status-validation path instead of leaving `validateAPIError` detached from the write path.
- Add focused regression coverage for invalid status inputs.

Non-goals:
- Do not change the error response envelope.
- Do not add new public constructors or error helper families.
- Do not change existing exported symbol names.

Files:
- `contract/errors.go`
- `contract/response.go`
- `contract/errors_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this hardens existing behavior without changing documented API shape.

Done Definition:
- Invalid statuses no longer panic response writing.
- Builder output uses a safe status/category/code fallback when callers provide invalid status values.
- Existing error shape tests continue to pass.

Outcome:
- Added private status normalization shared by `WriteJSON`, `WriteError`, and `ErrorBuilder.Build`.
- Preserved zero-value `APIError` defaults while failing closed for non-zero invalid status values.
- Validation run: `go test -timeout 20s ./contract/...`; `go vet ./contract/...`.
