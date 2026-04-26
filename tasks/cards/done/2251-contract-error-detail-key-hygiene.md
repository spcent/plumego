# Card 2251

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: contract
Owned Files:
- contract/errors.go
- contract/errors_test.go
Depends On: 2250

Goal:
Avoid emitting empty-string detail keys from `ErrorBuilder`.

Scope:
- Ignore empty detail keys in `Detail`.
- Ignore empty keys when merging `Details`.
- Keep caller-provided non-empty keys and values unchanged.
- Add focused coverage for zero-value and normal builders.

Non-goals:
- Do not sanitize or redact detail values.
- Do not deep-copy nested detail values.
- Do not change JSON envelope structure.

Files:
- `contract/errors.go`
- `contract/errors_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this hardens existing builder behavior.

Done Definition:
- Empty detail keys are omitted from built errors.
- Existing detail merge behavior for non-empty keys still works.

Outcome:
- `ErrorBuilder.Detail` now ignores empty keys.
- `ErrorBuilder.Details` now skips empty keys while preserving non-empty detail fields.
