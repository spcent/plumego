# Card 2257

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: core
Owned Files:
- core/app_helpers.go
- core/http_handler.go
- core/app_test.go
Depends On: 2256

Goal:
Avoid mutating partially initialized apps while discovering they are unusable.

Scope:
- Check app initialization before handler preparation in the `Prepare` path.
- Keep nil receiver behavior distinct from zero-value behavior.
- Return a structured `app not initialized` unavailable response from zero-value `ServeHTTP`.
- Add focused regression coverage.

Non-goals:
- Do not make zero-value `App` supported.
- Do not change normal `New(...)` behavior.
- Do not change router matching behavior.

Files:
- `core/app_helpers.go`
- `core/http_handler.go`
- `core/app_test.go`

Tests:
- `go test -timeout 20s ./core/...`
- `go vet ./core/...`

Docs Sync:
- None expected; this hardens public entrypoints.

Done Definition:
- Zero-value `Prepare` fails before handler preparation side effects.
- Zero-value `ServeHTTP` reports app initialization failure.

Outcome:
- `Prepare` now checks initialization before lazy handler preparation.
- Zero-value `ServeHTTP` now writes a structured `app not initialized` unavailable error.
- Regression coverage asserts zero-value `Prepare` has no handler-preparation side effects.
