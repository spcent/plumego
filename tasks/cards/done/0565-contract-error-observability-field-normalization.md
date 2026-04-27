# Card 0565

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/errors.go
- contract/errors_test.go
Depends On: 2274

Goal:
Normalize optional error observability fields so invalid type or severity values are not emitted.

Scope:
- Drop unknown `ErrorType` values during normalization.
- Drop invalid `ErrorSeverity` values during normalization.
- Keep valid type and severity values unchanged.
- Add focused coverage for builder and direct `APIError` inputs.

Non-goals:
- Do not change canonical type metadata.
- Do not add new error types or severities.
- Do not change status/code/category normalization.

Files:
- `contract/errors.go`
- `contract/errors_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this aligns normalization with internal validation.

Done Definition:
- `WriteError` never emits unknown error type or severity strings.
- Existing valid type/severity tests continue to pass.

Outcome:
- `normalizeAPIError` now drops unknown error types and invalid severities.
- Added builder and direct `APIError` coverage for invalid optional observability fields.
