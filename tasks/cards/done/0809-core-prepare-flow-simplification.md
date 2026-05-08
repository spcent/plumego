# Card 0809

Milestone: M-002
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: core
Owned Files: core/http_handler.go, core/lifecycle_test.go
Depends On: 0723-stable-api-snapshot-unexported-fields

Goal:
Simplify the `Prepare` implementation while preserving the stable lifecycle contract.

Scope:
Factor server preparation into small helpers for config snapshot, handler preparation, and server installation.
Avoid duplicate TLS certificate loading when concurrent `Prepare` calls race.
Preserve non-destructive validation failures and idempotent prepared-server behavior.

Non-goals:
Do not change public APIs.
Do not add dependencies.
Do not change shutdown behavior.

Files:
core/http_handler.go
core/lifecycle_test.go

Tests:
go test -timeout 20s ./core/...
go test -race -timeout 60s ./core/...
go run ./internal/checks/dependency-rules

Docs Sync:
None expected because behavior should remain unchanged.

Done Definition:
`Prepare` has smaller reviewable helper steps.
Concurrent `Prepare` does not duplicate TLS loading after one server is already installed.
Core tests and dependency check pass.

Outcome:
Completed.

Changes:

- Added a dedicated server preparation mutex to serialize `Prepare` server construction.
- Split server preparation into config snapshot, handler snapshot, and server installation helpers.
- Added concurrent `Prepare` coverage proving all callers observe the same prepared server.

Validation:

- `go test -timeout 20s ./core/...` passed.
- `go test -race -timeout 60s ./core/...` passed.
- `go run ./internal/checks/dependency-rules` passed.
