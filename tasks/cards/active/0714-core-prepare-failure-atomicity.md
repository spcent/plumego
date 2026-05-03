# Card 0714

Milestone: M-002
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: active
Primary Module: core
Owned Files: core/http_handler.go, core/lifecycle_test.go, docs/modules/core/README.md
Depends On: 0713-public-compatibility-inventory-decisions

Goal:
Make `Prepare` failure behavior explicit and non-destructive for server configuration errors.

Scope:
Validate server-only preparation inputs before freezing handler/router/middleware state.
Add regression coverage proving TLS configuration failures do not prepare or freeze the app.
Document the resulting failure behavior in the core module guide.

Non-goals:
Do not redesign `Prepare`, `Server`, or `Shutdown`.
Do not add public APIs or dependencies.
Do not change router matching behavior.

Files:
core/http_handler.go
core/lifecycle_test.go
docs/modules/core/README.md

Tests:
go test -timeout 20s ./core/...
go test -race -timeout 60s ./core/...
go run ./internal/checks/dependency-rules

Docs Sync:
Update `docs/modules/core/README.md` if runtime behavior changes.

Done Definition:
TLS missing-file and load errors return before handler preparation freezes mutation.
Successful `Prepare` behavior remains unchanged.
Core tests and dependency check pass.

Outcome:
