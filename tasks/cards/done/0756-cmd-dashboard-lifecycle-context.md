# Card 0756

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P0
State: done
Primary Module: cmd/plumego devserver lifecycle
Owned Files: cmd/plumego/internal/devserver/dashboard.go, cmd/plumego/internal/devserver/dashboard_info_test.go
Depends On: 0755

Goal:
Make dashboard lifecycle cleanup and request cancellation stable.

Scope:
- Ensure `Dashboard.Start` cleans up listener, server goroutine, subscriptions, and state on any mid-start failure.
- Use request-derived contexts with timeout for dashboard restart/build actions where applicable.
- Extract small helpers from `NewDashboard` only where it improves lifecycle clarity.
- Add regression tests for request context use and start failure cleanup if feasible.

Non-goals:
- Do not change dashboard route URLs.
- Do not redesign dashboard API payloads.
- Do not change WebSocket protocol behavior.

Files:
- `cmd/plumego/internal/devserver/dashboard.go`
- `cmd/plumego/internal/devserver/dashboard_info_test.go`

Tests:
- `go test ./internal/devserver`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- Not required.

Done Definition:
- Start failure paths leave no live server/subscription state.
- Restart actions do not use unbounded `context.Background()`.
- Tests cover the lifecycle edge cases that can be simulated.

Outcome:
- `Dashboard.Start` now uses a deferred cleanup path after the listener/server
  are registered, so subscription failures shut down the HTTP server, cancel
  subscriptions, wait for the serve goroutine, and clear lifecycle state.
- Dashboard restart actions now derive a bounded action context from
  `r.Context()` instead of using `context.Background()`.
- Added regression coverage for subscription failure cleanup and cancelled
  restart requests.

Validation:
- `go test ./internal/devserver`
- `go test ./...`
- `go vet ./...`
