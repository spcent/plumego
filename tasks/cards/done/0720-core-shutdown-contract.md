# Card 0720

Milestone: M-002
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: core
Owned Files: core/lifecycle.go, core/lifecycle_test.go, docs/modules/core/README.md
Depends On: 0719-core-config-validation-contract

Goal:
Make `Shutdown` behavior explicit for unprepared apps, repeated calls, and active connection drain logging.

Scope:
Treat shutdown before explicit server preparation as a clear lifecycle error.
Keep shutdown after preparation idempotent with respect to repeated calls.
Ensure active-connection drain logging is started at most once per prepared server.
Document the resulting shutdown contract.

Non-goals:
Do not own process signal handling in `core`.
Do not drive logger lifecycle.
Do not change `http.Server.Shutdown` semantics.

Files:
core/lifecycle.go
core/lifecycle_test.go
docs/modules/core/README.md

Tests:
go test -timeout 20s ./core/...
go test -race -timeout 60s ./core/...
go run ./internal/checks/dependency-rules

Docs Sync:
Update `docs/modules/core/README.md` with shutdown-before-prepare and repeated-shutdown behavior.

Done Definition:
`Shutdown` before `Prepare` returns a clear core lifecycle error.
Repeated `Shutdown` calls do not start duplicate drain loops.
Core tests and dependency check pass.

Outcome:
Completed.

Changes:

- `Shutdown` now returns `core shutdown_app: server not prepared` before explicit server preparation.
- Active connection drain logging starts at most once per prepared server.
- Repeated `Shutdown` calls remain accepted for prepared servers.
- Documented the shutdown contract in the core module guide.

Validation:

- `go test -timeout 20s ./core/...` passed.
- `go test -race -timeout 60s ./core/...` passed.
- `go run ./internal/checks/dependency-rules` passed.
