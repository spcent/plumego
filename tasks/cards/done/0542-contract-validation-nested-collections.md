# Card 0542

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/validation.go
- contract/active_cards_regression_test.go
Depends On: 2251

Goal:
Validate nested structs inside slices and arrays.

Scope:
- Traverse slice and array elements when they are structs or pointers to structs.
- Use indexed field names such as `Items[0].Name`.
- Preserve programmer-error propagation for malformed nested tags.
- Add focused regression coverage.

Non-goals:
- Do not validate map values.
- Do not change scalar slice validation.
- Do not add new validation rules.

Files:
- `contract/validation.go`
- `contract/active_cards_regression_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this completes existing nested validation behavior.

Done Definition:
- Nested collection elements are validated with stable indexed field names.
- Existing depth-limit and nested struct tests continue to pass.

Outcome:
- `ValidateStruct` now walks nested structs and pointers to structs inside slices and arrays.
- Added coverage for indexed field names and nested programmer-error propagation.
