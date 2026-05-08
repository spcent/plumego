# Card 0719

Milestone: M-002
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: core
Owned Files: core/config.go, core/http_handler.go, core/lifecycle_test.go, docs/modules/core/README.md
Depends On: 0718-core-connection-tracker-semantics

Goal:
Define and enforce the `core.AppConfig` runtime validity contract before stable freeze.

Scope:
Add an internal validation path for server preparation.
Reject invalid server addresses and negative HTTP server durations before handler/router freeze.
Keep zero timeout values valid because they preserve standard `http.Server` semantics.
Document that callers should start from `DefaultConfig()` and that invalid config fails before mutation freeze.

Non-goals:
Do not add exported validation APIs.
Do not change request body or concurrency limit ownership.
Do not change router behavior.

Files:
core/config.go
core/http_handler.go
core/lifecycle_test.go
docs/modules/core/README.md

Tests:
go test -timeout 20s ./core/...
go test -race -timeout 60s ./core/...
go run ./internal/checks/dependency-rules

Docs Sync:
Update `docs/modules/core/README.md` for the config validity contract.

Done Definition:
Invalid core server config returns a clear `Prepare` error before freezing route/middleware mutation.
Valid zero timeout fields remain accepted.
Core tests and dependency check pass.

Outcome:
Completed.

Changes:

- Added internal server config validation before TLS loading and handler/router freeze.
- Rejected empty server addresses, negative server timeout values, and negative max header bytes.
- Preserved zero timeout and zero max-header semantics.
- Documented `DefaultConfig()` as the supported baseline and invalid config behavior.

Validation:

- `go test -timeout 20s ./core/...` passed.
- `go test -race -timeout 60s ./core/...` passed.
- `go run ./internal/checks/dependency-rules` passed.
