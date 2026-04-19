# Card 2004

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: active
Primary Module: reference/workerfleet/internal/domain
Owned Files:
- `reference/workerfleet/internal/domain/ingest_service.go`
- `reference/workerfleet/internal/domain/reconcile_service.go`
- `reference/workerfleet/internal/handler/worker_register.go`
- `reference/workerfleet/internal/handler/worker_heartbeat.go`
- `reference/workerfleet/internal/domain/ingest_service_test.go`
Depends On:
- `tasks/cards/active/2001-workerfleet-domain-status-model.md`
- `tasks/cards/active/2002-workerfleet-api-contract-and-routes.md`
- `tasks/cards/active/2003-workerfleet-store-and-retention.md`

Goal:
- Implement the worker register and heartbeat ingest path so each report updates worker status, active tasks, and emitted domain events consistently.

Scope:
- Domain services that merge worker reports into snapshots and reconcile full active-task sets.
- Register and heartbeat handlers wired to the domain services.
- Event emission for worker transitions, task starts, phase changes, task finishes, and state conflicts.

Non-goals:
- Do not add Kubernetes inventory sync.
- Do not add alert notifier delivery.
- Do not add UI endpoints beyond the already approved ingest/query API.

Files:
- `reference/workerfleet/internal/domain/ingest_service.go`
- `reference/workerfleet/internal/domain/reconcile_service.go`
- `reference/workerfleet/internal/handler/worker_register.go`
- `reference/workerfleet/internal/handler/worker_heartbeat.go`
- `reference/workerfleet/internal/domain/ingest_service_test.go`

Tests:
- `go test ./reference/workerfleet/internal/domain/...`
- `go test ./reference/workerfleet/internal/handler/...`
- Coverage for first registration, repeated heartbeat, multi-task add/remove, phase changes, stale-to-offline transitions, and accepting-tasks changes.

Docs Sync:
- Update `reference/workerfleet/docs/api.md` examples if ingest payloads or semantics shift during implementation.

Done Definition:
- Register and heartbeat requests update current snapshots and active-task sets correctly.
- Domain events are emitted from the centralized reconciliation path, not ad hoc in handlers.
- Worker readiness and task-state changes are queryable from persisted state without replaying raw request logs.

Outcome:
- Pending execution.
