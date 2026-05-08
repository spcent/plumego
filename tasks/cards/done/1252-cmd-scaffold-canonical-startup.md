# Card 1252

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P0
State: done
Primary Module: cmd/plumego scaffold
Owned Files: cmd/plumego/internal/scaffold/scaffold.go, cmd/plumego/internal/scaffold/scaffold_test.go, cmd/plumego/commands/project_smoke_test.go
Depends On: 0762

Goal:
Harden generated canonical app startup and middleware error handling.

Scope:
- Make generated `app.New` return middleware registration errors.
- Make generated `App.Start` treat `http.ErrServerClosed` as normal shutdown.
- Preserve generated file set and route behavior.
- Add tests that generated app content contains canonical error handling.

Non-goals:
- Do not change scaffold template names.
- Do not add graceful signal handling to generated apps in this card.
- Do not change generated dependencies.

Files:
- `cmd/plumego/internal/scaffold/scaffold.go`
- `cmd/plumego/internal/scaffold/scaffold_test.go`
- `cmd/plumego/commands/project_smoke_test.go`

Tests:
- `go test ./internal/scaffold ./commands`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- Not required unless README startup snippets change.

Done Definition:
- Generated middleware registration failures are returned.
- Generated `Start` handles normal server shutdown distinctly from failures.

Outcome:
- Generated canonical `app.New` now returns middleware registration errors.
- Generated canonical `App.Start` now treats `http.ErrServerClosed` as normal
  shutdown and returns shutdown errors when no earlier server error occurred.
- Added scaffold template tests for the canonical startup/error-handling shape.

Validation:
- `go test ./internal/scaffold ./commands`
- `go test ./...`
- `go vet ./...`
