# Card 2253

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: core
Owned Files:
- core/app_helpers.go
- core/lifecycle.go
- core/app_test.go
- core/lifecycle_test.go
Depends On: 2252

Goal:
Finish zero-value `App` lifecycle consistency for `Server` and `Shutdown`.

Scope:
- Return `app not initialized` from zero-value `Server`.
- Return `app not initialized` from zero-value `Shutdown`.
- Keep nil receiver errors distinct from zero-value receiver errors.
- Add focused regression coverage.

Non-goals:
- Do not make zero-value `App` a supported construction path.
- Do not change normal `New(...)` lifecycle behavior.
- Do not add public APIs.

Files:
- `core/app_helpers.go`
- `core/lifecycle.go`
- `core/app_test.go`
- `core/lifecycle_test.go`

Tests:
- `go test -timeout 20s ./core/...`
- `go vet ./core/...`

Docs Sync:
- None expected; this hardens existing public entrypoints.

Done Definition:
- Zero-value lifecycle calls fail predictably without mutating app state.
- Existing core lifecycle tests continue to pass.

Outcome:
- Zero-value `App.Server` and `App.Shutdown` now return wrapped `app not initialized` errors.
- Added a shared initialized-state helper for public entrypoint guards.
