# Card 0525

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/errors.go
- contract/errors_test.go
Depends On: 2234

Goal:
Make `ErrorBuilder` safe for zero-value use and prevent built `APIError` details from aliasing the builder's mutable map.

Scope:
- Lazily initialize `ErrorBuilder` details before writes.
- Copy details when building/normalizing an `APIError`.
- Add focused regression tests for zero-value builders and post-build builder mutation.

Non-goals:
- Do not add new error constructors.
- Do not deep-copy arbitrary values stored in `Details`.
- Do not change the error response envelope.

Files:
- `contract/errors.go`
- `contract/errors_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this preserves the documented builder API.

Done Definition:
- `var b ErrorBuilder; b.Detail(...).Build()` does not panic.
- Mutating the builder after `Build` does not mutate previously built details.
- Existing error model tests continue to pass.

Outcome:
- Added lazy details map initialization for zero-value `ErrorBuilder` use.
- Copied details during API error normalization so built errors do not alias builder state.
- Validation run: `go test -timeout 20s ./contract/...`; `go vet ./contract/...`.
