# Card 0379

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: reference/workerfleet/internal/app
Owned Files:
- `reference/workerfleet/internal/app/config.go`
- `reference/workerfleet/internal/app/bootstrap.go`
- `reference/workerfleet/internal/app/service.go`
- `reference/workerfleet/README.md`
- `reference/workerfleet/docs/storage.md`
Depends On:
- `tasks/cards/done/0377-workerfleet-mongo-snapshot-and-active-task-store.md`
- `tasks/cards/done/0378-workerfleet-mongo-history-and-alert-store.md`
Blocked By:
- workerfleet MongoDB design review approval

Goal:
- Add runtime configuration and dependency wiring so workerfleet can switch between the existing in-memory backend and the new MongoDB backend.
- Keep the workerfleet HTTP and domain surfaces unchanged while enabling durable persistence in Mongo mode.

Scope:
- Add workerfleet-local config for store backend selection, Mongo connection parameters, timeouts, and retention days.
- Wire Mongo client creation, index bootstrap, and repository construction in app-local bootstrap code.
- Preserve the memory backend as the default or explicit fallback for tests and local development.

Non-goals:
- Do not add new business endpoints.
- Do not change worker status rules or alert evaluation policy.
- Do not introduce heartbeat raw sample storage.

Files:
- `reference/workerfleet/internal/app/config.go`
- `reference/workerfleet/internal/app/bootstrap.go`
- `reference/workerfleet/internal/app/service.go`
- `reference/workerfleet/README.md`
- `reference/workerfleet/docs/storage.md`

Tests:
- `cd reference/workerfleet && go test ./internal/app/...`
- `cd reference/workerfleet && go test ./...`
- Focused tests for config parsing, backend selection, missing Mongo config failures, and successful memory fallback.

Docs Sync:
- Update `reference/workerfleet/README.md` with backend configuration examples.
- Update `reference/workerfleet/docs/storage.md` with startup and index-bootstrap behavior.

Done Definition:
- Workerfleet can boot against either memory or MongoDB via explicit configuration.
- Mongo mode initializes the client, ensures indexes, and wires repositories without changing handlers.
- Memory mode remains available and existing tests still pass.
- Configuration failure paths fail closed with clear startup errors.

Outcome:
- Added workerfleet-local config parsing for memory and Mongo backends, Mongo timeouts, max pool size, and retention days.
- Added app bootstrap wiring for memory and Mongo modes; Mongo mode opens the client, pings the primary, ensures indexes, and wires repositories before exposing handlers.
- Preserved memory as the default backend and added fail-closed validation for missing Mongo URI/database.
- Updated README and storage docs with backend configuration and startup behavior.
- Validation run:
  - `cd reference/workerfleet && go test ./internal/app/...`
  - `cd reference/workerfleet && go test ./internal/platform/store/mongo/...`
  - `cd reference/workerfleet && go test ./...`
  - `cd reference/workerfleet && go build ./...`
  - `cd reference/workerfleet && go vet ./...`
