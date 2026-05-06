# Card 0762

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P0
State: active
Primary Module: cmd/plumego devserver build
Owned Files: cmd/plumego/internal/devserver/builder.go, cmd/plumego/internal/devserver/dashboard.go, cmd/plumego/internal/devserver/api_tester.go, cmd/plumego/internal/devserver/builder_test.go, cmd/plumego/internal/devserver/dashboard_info_test.go
Depends On: 0761

Goal:
Make dashboard build/rebuild/API-test actions respect request cancellation and timeouts.

Scope:
- Change builder build execution to accept `context.Context`.
- Add timeout-aware command execution for default/custom builds.
- Update dashboard build/rebuild/config-save restart paths to use request-derived bounded contexts.
- Change API test execution to accept context from the dashboard request.
- Add regression tests for cancelled build/API-test behavior where feasible.

Non-goals:
- Do not redesign dashboard routes or response payloads.
- Do not change app runner process ownership.
- Do not add a background job queue.

Files:
- `cmd/plumego/internal/devserver/builder.go`
- `cmd/plumego/internal/devserver/dashboard.go`
- `cmd/plumego/internal/devserver/api_tester.go`
- `cmd/plumego/internal/devserver/builder_test.go`
- `cmd/plumego/internal/devserver/dashboard_info_test.go`

Tests:
- `go test ./internal/devserver`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- Not required.

Done Definition:
- `BuildAndRun(ctx)` can cancel the build step.
- Dashboard build/restart/config-save/API-test actions do not use unbounded background contexts.

Outcome:
