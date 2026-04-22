# Card 2023

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: active
Primary Module: reference/workerfleet/internal/domain
Owned Files:
- `reference/workerfleet/internal/domain/task.go`
- `reference/workerfleet/internal/domain/rules.go`
- `reference/workerfleet/internal/domain/reconcile_test.go`
- `reference/workerfleet/internal/handler/worker_heartbeat.go`
- `reference/workerfleet/docs/api.md`
Depends On:
- `tasks/cards/done/2019-workerfleet-service-entrypoint.md`
Blocked By: —

Goal:
- Add explicit `exec_plan_id` and current case step reporting to the worker heartbeat contract.
- Make the domain able to detect case step transitions and step completion inputs without adding high-cardinality Prometheus labels.

Scope:
- Extend active task and task report models with `ExecPlanID` and `CurrentStep`.
- Define controlled step status/result fields and low-cardinality `error_class`.
- Extend heartbeat JSON decode/encode structures.
- Add domain reconciliation events for step transition and step completion.
- Preserve full active-task set replacement semantics.

Non-goals:
- Do not add Mongo persistence for step history.
- Do not add Prometheus case/step metrics in this card.
- Do not expose case timeline query APIs yet.

Files:
- `reference/workerfleet/internal/domain/task.go`
- `reference/workerfleet/internal/domain/rules.go`
- `reference/workerfleet/internal/domain/reconcile_test.go`
- `reference/workerfleet/internal/handler/worker_heartbeat.go`
- `reference/workerfleet/docs/api.md`

Tests:
- `cd reference/workerfleet && go test ./internal/domain/...`
- `cd reference/workerfleet && go test ./internal/handler/...`
- `cd reference/workerfleet && go test ./...`

Docs Sync:
- Update `reference/workerfleet/docs/api.md` heartbeat examples with `exec_plan_id` and `current_step`.

Done Definition:
- Worker heartbeats can report current step per active case.
- Domain reconciliation emits deterministic step transition/completion events.
- Existing worker heartbeat behavior remains backward compatible when `current_step` is omitted.

