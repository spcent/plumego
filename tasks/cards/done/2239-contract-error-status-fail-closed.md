# Card 2239

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/errors.go
- contract/errors_test.go
Depends On: 2238

Goal:
Ensure `WriteError` and `ErrorBuilder.Build` cannot produce successful HTTP statuses for error responses.

Scope:
- Normalize non-zero statuses below 400 to `500 Internal Server Error` on the error path.
- Preserve zero-value default behavior.
- Add regression coverage for builder and writer inputs using 2xx/3xx statuses.

Non-goals:
- Do not change success response helpers.
- Do not change the error envelope.
- Do not add new public constructors.

Files:
- `contract/errors.go`
- `contract/errors_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this hardens the existing error path.

Done Definition:
- Error responses fail closed to a 5xx status when callers provide non-error statuses.
- Existing error model tests continue to pass.

Outcome:
- Normalized non-zero statuses below 400 to 500 on the API error path.
- Added writer and builder regression tests for 2xx/3xx inputs.
- Validation run: `go test -timeout 20s ./contract/...`; `go vet ./contract/...`.
