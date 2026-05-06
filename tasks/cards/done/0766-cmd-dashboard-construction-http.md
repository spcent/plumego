# Card 0766

Milestone: cmd stable hardening
Recipe: specs/change-recipes/refactor-small.yaml
Priority: P1
State: done
Primary Module: cmd/plumego dashboard internals
Owned Files: cmd/plumego/internal/devserver/dashboard.go, cmd/plumego/internal/devserver/analyzer.go, cmd/plumego/internal/devserver/dashboard_info_test.go, cmd/plumego/internal/devserver/analyzer_test.go
Depends On: 0765

Goal:
Reduce dashboard composition complexity and standardize outbound HTTP safety.

Scope:
- Split `NewDashboard` into small helpers for config validation, app construction, service wiring, and routes.
- Add bounded response reads for analyzer routes/config/health/metrics calls.
- Keep dashboard API behavior and route URLs unchanged.

Non-goals:
- Do not change WebSocket protocol behavior.
- Do not change dashboard UI assets.
- Do not redesign analyzer response types.

Files:
- `cmd/plumego/internal/devserver/dashboard.go`
- `cmd/plumego/internal/devserver/analyzer.go`
- `cmd/plumego/internal/devserver/dashboard_info_test.go`
- `cmd/plumego/internal/devserver/analyzer_test.go`

Tests:
- `go test ./internal/devserver`
- `go test ./...`
- `go vet ./...`

Docs Sync:
- Not required.

Done Definition:
- Dashboard constructor responsibilities are visibly separated.
- Analyzer HTTP reads are timeout- and size-bounded consistently.

Outcome:
- Split dashboard construction into validation, app middleware setup, service wiring, and route registration helpers.
- Centralized analyzer HTTP clients, timeouts, URL joining, and response body limits.
- Added bounded response tests for analyzer routes, config, health, and metrics reads, plus config status handling.

Validation:
- `go test ./internal/devserver`
- `go test ./...`
- `go vet ./...`
