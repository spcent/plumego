# Card 1389

Milestone: workerfleet-hardening
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P0
State: done
Primary Module: reference/workerfleet/internal/domain
Owned Files:
- reference/workerfleet/internal/domain/ingest_service.go
- reference/workerfleet/internal/platform/store/interfaces.go
- reference/workerfleet/internal/platform/store/memory/store.go
- reference/workerfleet/internal/platform/store/mongo/worker_event_store.go
- reference/workerfleet/internal/platform/store/mongo/task_history_store.go
Depends On:
- 1388

Goal:
- Propagate context through worker registration, heartbeat ingest, and write-side persistence.

Scope:
- Add `context.Context` to domain store ports used by `IngestService`.
- Update `Register` and `Heartbeat` to accept context and pass it to snapshot, event, task-history, and active-task writes.
- Update app service callers to pass the request context into ingest.
- Migrate memory and Mongo write implementations to honor caller context.
- Re-run the symbol search protocol for `Register`, `Heartbeat`, and changed store write methods.

Non-goals:
- Do not change worker status rules.
- Do not change JSON request or response fields.
- Do not introduce cross-service transaction semantics.

Files:
- reference/workerfleet/internal/domain/ingest_service.go
- reference/workerfleet/internal/platform/store/interfaces.go
- reference/workerfleet/internal/platform/store/memory/store.go
- reference/workerfleet/internal/platform/store/mongo/worker_event_store.go
- reference/workerfleet/internal/platform/store/mongo/task_history_store.go

Tests:
- cd reference/workerfleet && go test -timeout 20s ./internal/domain/...
- cd reference/workerfleet && go test -timeout 20s ./internal/platform/store/...
- cd reference/workerfleet && go test -timeout 20s ./...

Docs Sync:
- Required only if documented cancellation or timeout behavior changes.

Done Definition:
- Register and heartbeat paths no longer perform persistence through background-only contexts.
- Write-side Mongo operations derive timeouts from caller context.
- Cancellation behavior is covered with focused tests or documented as adapter-limited where a fake cannot observe it.
- Old write method signatures have no remaining callers.

Outcome:
- Added `context.Context` to workerfleet write-side store ports and migrated snapshot upsert, active-task replace, worker event append, task history append, case-step history append, and alert append methods.
- Updated `IngestService.Register` and `IngestService.Heartbeat` to accept caller context and pass it through all read/write persistence calls.
- Updated app service, runtime sweep, Kubernetes sync, and alert engine write callers to pass their active context.
- Added focused canceled-context tests for ingest register and memory write paths.
- Validation run:
  - `cd reference/workerfleet && go test -timeout 20s ./internal/domain/...`
  - `cd reference/workerfleet && go test -timeout 20s ./internal/platform/store/...`
  - `cd reference/workerfleet && go test -timeout 20s ./internal/app/...`
  - `cd reference/workerfleet && go test -timeout 20s ./...`
