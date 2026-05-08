# Card 0881

Milestone: cmd stable hardening
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P0
State: done
Primary Module: cmd/plumego/internal/devserver
Owned Files: cmd/plumego/internal/devserver/dashboard.go, cmd/plumego/internal/devserver/interfaces.go, cmd/plumego/internal/devserver/dashboard_info_test.go, cmd/plumego/DEV_SERVER.md
Depends On: 0730

Goal:
Make the development dashboard server lifecycle observable, cancellable, and safe for repeated start/stop.

Scope:
- Replace background `ListenAndServe` plus sleep with explicit listener setup and serve error reporting.
- Track dashboard subscriptions/goroutines with owned cleanup.
- Ensure `Stop` shuts down the dashboard server, WebSocket hub, subscriptions, and runner.
- Add focused tests for bind failure and stop cleanup behavior.

Non-goals:
- Do not redesign dashboard APIs.
- Do not change dashboard auth/CORS policy.
- Do not replace the core app server abstraction.

Files:
- `cmd/plumego/internal/devserver/dashboard.go`
- `cmd/plumego/internal/devserver/interfaces.go`
- `cmd/plumego/internal/devserver/dashboard_info_test.go`
- `cmd/plumego/DEV_SERVER.md`

Tests:
- `go test ./internal/devserver`
- `go test -race ./internal/devserver`
- `go build .`

Docs Sync:
- `cmd/plumego/DEV_SERVER.md`

Done Definition:
- Dashboard bind failures return from `Start`.
- Dashboard stop closes owned background work.
- Race test passes for devserver.

Outcome:
- Replaced detached `ListenAndServe` startup with explicit listener setup and server completion tracking.
- Added dashboard subscription ownership and Stop cleanup for event/lifecycle subscriptions.
- Documented startup failure and shutdown semantics.

Validation:
- `go test ./internal/devserver`
- `go test -race ./internal/devserver`
- `go build .`
