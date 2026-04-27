# Card 0533

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: core
Owned Files:
- core/app.go
- core/app_helpers.go
- core/http_handler.go
- core/middleware.go
- core/routing.go
- core/app_test.go
- core/lifecycle_test.go
Depends On: 2242

Goal:
Make core public `App` entrypoints fail predictably when called on a nil receiver instead of panicking.

Scope:
- Guard nil receivers on route registration, middleware registration, `Server`, `Shutdown`, `ServeHTTP`, `Routes`, and `Logger`.
- Return wrapped core errors where the API already returns errors.
- Write a structured unavailable error from nil `ServeHTTP`.
- Add focused regression coverage.

Non-goals:
- Do not change normal app lifecycle behavior.
- Do not add new public APIs.
- Do not make nil app usage a recommended path in docs.

Files:
- `core/app.go`
- `core/app_helpers.go`
- `core/http_handler.go`
- `core/middleware.go`
- `core/routing.go`
- `core/app_test.go`
- `core/lifecycle_test.go`

Tests:
- `go test -timeout 20s ./core/...`
- `go vet ./core/...`

Docs Sync:
- None expected; this hardens existing public entrypoints.

Done Definition:
- Public nil `*App` calls covered by this card do not panic.
- Normal core tests continue to pass.

Outcome:
- Nil `*App` registration and lifecycle entrypoints now return wrapped core errors.
- Nil `ServeHTTP` writes the canonical structured unavailable response instead of panicking.
- Nil query helpers return inert values: an empty route set, empty URL, and discard logger.
