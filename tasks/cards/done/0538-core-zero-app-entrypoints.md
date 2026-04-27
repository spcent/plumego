# Card 0538

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: core
Owned Files:
- core/app.go
- core/app_helpers.go
- core/http_handler.go
- core/lifecycle.go
- core/app_test.go
Depends On: 2247

Goal:
Make zero-value `App` public entrypoints fail predictably instead of exposing nil internals or misleading lifecycle errors.

Scope:
- Return a discard logger from zero-value `App.Logger`.
- Return wrapped core errors for zero-value route/middleware/server preparation paths.
- Avoid nil config dereferences in `Prepare`.
- Add focused regression coverage.

Non-goals:
- Do not make zero-value `App` a recommended construction path.
- Do not change `New(DefaultConfig(), ...)` behavior.
- Do not add new public APIs.

Files:
- `core/app.go`
- `core/app_helpers.go`
- `core/http_handler.go`
- `core/lifecycle.go`
- `core/app_test.go`

Tests:
- `go test -timeout 20s ./core/...`
- `go vet ./core/...`

Docs Sync:
- None expected; this hardens public entrypoints.

Done Definition:
- Zero-value `App` public entrypoints covered by this card do not panic.
- Existing core lifecycle tests continue to pass.

Outcome:
- Zero-value `App.Logger` now returns the same discard logger fallback used for nil logger dependencies.
- Zero-value mutation and preparation paths now return wrapped `app not initialized` core errors.
- `Prepare` no longer dereferences a nil `App.config`.
