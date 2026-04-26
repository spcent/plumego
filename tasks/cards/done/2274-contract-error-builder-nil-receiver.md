# Card 2274

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: contract
Owned Files:
- contract/errors.go
- contract/errors_test.go
Depends On: 2273

Goal:
Make nil `*ErrorBuilder` receiver calls fail soft like zero-value builders.

Scope:
- Allow fluent builder methods to be called on a nil builder receiver without panicking.
- Preserve default normalization from `Build`.
- Add focused regression coverage.

Non-goals:
- Do not change `APIError` fields.
- Do not add new builder methods.
- Do not change normal `NewErrorBuilder` behavior.

Files:
- `contract/errors.go`
- `contract/errors_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this hardens existing builder ergonomics.

Done Definition:
- Nil builder chains produce normalized errors instead of panics.
- Existing zero-value builder tests continue to pass.

Outcome:
- Added an internal builder receiver guard used by all fluent methods and `Build`.
- Added regression coverage for nil builder chaining and default nil builder output.
