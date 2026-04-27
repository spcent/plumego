# Card 0563

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: contract
Owned Files:
- contract/context_bind.go
- contract/context_test.go
Depends On: 2258

Goal:
Reject ambiguous `BindJSON` calls with more than one `BindOptions` value.

Scope:
- Return a binding error wrapping `ErrInvalidParam` when multiple option structs are provided.
- Keep zero-option and one-option behavior unchanged.
- Add focused regression coverage.

Non-goals:
- Do not change `BindOptions` fields.
- Do not add option merging behavior.
- Do not change JSON decode semantics.

Files:
- `contract/context_bind.go`
- `contract/context_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this hardens existing optional-argument behavior.

Done Definition:
- `BindJSON(dst, opt1, opt2)` fails explicitly instead of silently ignoring `opt2`.
- Existing bind tests continue to pass.

Outcome:
- `BindJSON` now rejects multiple `BindOptions` values with an error wrapping `ErrInvalidParam`.
- Added coverage proving ambiguous options fail before reading the request body.
