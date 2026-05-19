# Card 0753

Milestone: cmd stable hardening
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P0
State: done
Primary Module: cmd/plumego
Owned Files: cmd/plumego/commands/dev.go, cmd/plumego/commands/dev_test.go, cmd/plumego/internal/devserver/dashboard.go, cmd/plumego/internal/devserver/dashboard_info_test.go, cmd/plumego/DEV_SERVER.md
Depends On: 0716

Goal:
Reduce dev dashboard exposure before stable by guarding dangerous local tooling
endpoints.

Scope:
- Enforce safe local defaults for dashboard binding.
- Add a clear opt-in or token guard for non-loopback dashboard use.
- Protect build/restart/stop/API-test/pprof/config-edit surfaces.
- Add focused handler tests for unauthorized access.

Non-goals:
- Do not build a production auth system.
- Do not change core middleware APIs.
- Do not remove dashboard features.

Files:
- `cmd/plumego/commands/dev.go`
- `cmd/plumego/commands/dev_test.go`
- `cmd/plumego/internal/devserver/dashboard.go`
- `cmd/plumego/internal/devserver/dashboard_info_test.go`
- `cmd/plumego/DEV_SERVER.md`

Tests:
- `go test ./commands ./internal/devserver`
- `go build .`

Docs Sync:
- `cmd/plumego/DEV_SERVER.md` if flags or auth expectations change.

Done Definition:
- Dangerous dashboard routes are not open by accident.
- Loopback development remains ergonomic.
- Tests cover denied and allowed paths.

Outcome:
- Added `--dashboard-token` and plumbed it into dev dashboard configuration.
- Rejected non-loopback dashboard binding unless a dashboard token is configured.
- Added timing-safe token checks for dashboard action endpoints including build,
  restart, stop, config edit, API test, metrics clear, and raw pprof.
- Documented the local-first dashboard safety behavior.
- Added regression tests for remote binding rejection and token enforcement.
- Validation Run:
  - `go test ./commands ./internal/devserver`
  - `go build .`
