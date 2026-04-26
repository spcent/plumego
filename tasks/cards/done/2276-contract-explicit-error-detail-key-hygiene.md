# Card 2276

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: contract
Owned Files:
- contract/errors.go
- contract/errors_test.go
Depends On: 2275

Goal:
Apply detail key hygiene to direct `APIError` values, not only `ErrorBuilder`.

Scope:
- Drop empty detail keys while cloning `APIError.Details`.
- Preserve non-empty keys and values.
- Add focused coverage for direct `APIError` writes.

Non-goals:
- Do not redact or transform detail values.
- Do not deep-copy nested values.
- Do not change error envelope shape.

Files:
- `contract/errors.go`
- `contract/errors_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this completes existing detail-key hygiene.

Done Definition:
- Empty detail keys are omitted from direct API errors.
- Existing builder detail tests continue to pass.

Outcome:
- `cloneAnyMap` now drops empty detail keys and collapses all-empty maps to nil.
- Added direct `WriteError` coverage proving empty detail keys are omitted while normal keys remain.
- Verified with `go test -timeout 20s ./contract/...` and `go vet ./contract/...`.
