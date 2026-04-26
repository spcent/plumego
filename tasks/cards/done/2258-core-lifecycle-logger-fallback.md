# Card 2258

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: core
Owned Files:
- core/http_handler.go
- core/lifecycle.go
- core/lifecycle_test.go
Depends On: 2257

Goal:
Use the canonical logger fallback in lifecycle paths that may run on partially initialized apps.

Scope:
- Use `App.Logger()` when constructing the connection tracker.
- Use `App.Logger()` when logging shutdown errors.
- Add focused coverage proving shutdown errors do not panic with a nil app logger.

Non-goals:
- Do not change logger ownership.
- Do not start or stop logger lifecycles from core.
- Do not add new public APIs.

Files:
- `core/http_handler.go`
- `core/lifecycle.go`
- `core/lifecycle_test.go`

Tests:
- `go test -timeout 20s ./core/...`
- `go vet ./core/...`

Docs Sync:
- None expected; this uses existing fallback behavior.

Done Definition:
- Lifecycle error logging never dereferences a nil logger interface.
- Existing lifecycle tests continue to pass.

Outcome:
- Server preparation and shutdown error logging now use `App.Logger()` for the discard fallback.
- Added coverage for connection tracker fallback logger wiring and nil-logger shutdown error paths.
