# Card 1388

Milestone: workerfleet-hardening
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P0
State: done
Primary Module: reference/workerfleet/internal/platform/store
Owned Files:
- reference/workerfleet/internal/platform/store/interfaces.go
- reference/workerfleet/internal/platform/store/memory/store.go
- reference/workerfleet/internal/platform/store/mongo/worker_snapshot_store.go
- reference/workerfleet/internal/platform/store/mongo/active_task_store.go
- reference/workerfleet/internal/app/service.go
Depends On:
- 1387

Goal:
- Propagate request cancellation and deadlines through workerfleet read/query store paths.

Scope:
- Add `context.Context` to app-local query store methods used by HTTP reads.
- Migrate `ListWorkerSnapshots`, `GetWorkerSnapshot`, `ListCurrentWorkerSnapshots`, `FleetCounts`, `ActiveTasks`, `GetTask`, `TaskHistory`, `LatestTask`, and alert list/read methods where they are part of query paths.
- Update memory and Mongo query implementations to honor caller context.
- Use caller context as the parent for Mongo operation timeouts instead of `context.Background()`.
- Re-run the symbol search protocol for each changed interface method.

Non-goals:
- Do not migrate write/ingest store methods in this card unless required for compilation.
- Do not change response envelopes or query parameters.
- Do not add pagination or indexing changes.

Files:
- reference/workerfleet/internal/platform/store/interfaces.go
- reference/workerfleet/internal/platform/store/memory/store.go
- reference/workerfleet/internal/platform/store/mongo/worker_snapshot_store.go
- reference/workerfleet/internal/platform/store/mongo/active_task_store.go
- reference/workerfleet/internal/app/service.go

Tests:
- cd reference/workerfleet && go test -timeout 20s ./internal/platform/store/...
- cd reference/workerfleet && go test -timeout 20s ./internal/app/...
- cd reference/workerfleet && go test -timeout 20s ./...

Docs Sync:
- Update storage docs only if the public readiness or request-cancellation behavior is documented.

Done Definition:
- HTTP query service methods pass `r.Context()` through to store calls.
- Mongo query operations derive timeouts from the caller context.
- Context cancellation has focused tests on at least one Mongo or fake-store query path.
- Old query method signatures have no remaining callers.

Outcome:
- Added `context.Context` to workerfleet read/query store interfaces and migrated app service, readiness, runtime sweep, Kubernetes inventory sync, and alert evaluation read paths.
- Mongo query operations now derive operation timeouts from the caller context.
- Memory query methods return cancellation errors before taking locks.
- Added a focused canceled-context memory store test.
- Validation run:
  - `cd reference/workerfleet && go test -timeout 20s ./internal/platform/store/...`
  - `cd reference/workerfleet && go test -timeout 20s ./internal/app/...`
  - `cd reference/workerfleet && go test -timeout 20s ./internal/domain/...`
  - `cd reference/workerfleet && go test -timeout 20s ./...`
