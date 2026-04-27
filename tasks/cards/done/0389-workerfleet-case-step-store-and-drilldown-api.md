# Card 0389

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: reference/workerfleet/internal/platform/store
Owned Files:
- `reference/workerfleet/internal/platform/store/interfaces.go`
- `reference/workerfleet/internal/platform/store/types.go`
- `reference/workerfleet/internal/platform/store/retention.go`
- `reference/workerfleet/internal/platform/store/memory/store.go`
- `reference/workerfleet/internal/platform/store/memory/store_test.go`
- `reference/workerfleet/internal/app/service.go`
- `reference/workerfleet/internal/app/service_test.go`
- `reference/workerfleet/docs/storage.md`
Depends On:
- `tasks/cards/done/0388-workerfleet-case-step-domain-and-heartbeat-contract.md`
Blocked By: —

Goal:
- Persist and expose case step history for drilldown after Grafana identifies an abnormal node, pod, exec plan, or step.
- Keep Prometheus aggregate-only while Mongo/API own case-level detail.

Scope:
- Add app-local store interfaces and types for case step history.
- Persist step history in the memory backend first.
- Add service-level query methods for case timeline and exec-plan case drilldown.
- Document Mongo persistence as a follow-up if needed after interface shape is stable.

Non-goals:
- Do not add Prometheus metrics.
- Do not implement Mongo case step history in this card.
- Do not add frontend UI.

Files:
- `reference/workerfleet/internal/platform/store/interfaces.go`
- `reference/workerfleet/internal/platform/store/types.go`
- `reference/workerfleet/internal/platform/store/retention.go`
- `reference/workerfleet/internal/platform/store/memory/store.go`
- `reference/workerfleet/internal/platform/store/memory/store_test.go`
- `reference/workerfleet/internal/app/service.go`
- `reference/workerfleet/internal/app/service_test.go`
- `reference/workerfleet/docs/storage.md`

Tests:
- `cd reference/workerfleet && go test ./internal/platform/store/memory/...`
- `cd reference/workerfleet && go test ./internal/app/...`
- `cd reference/workerfleet && go test ./...`

Docs Sync:
- Update `reference/workerfleet/docs/storage.md` with case step history storage responsibilities.

Done Definition:
- Case step history has an explicit store interface and memory implementation.
- Service layer can return a case timeline from stored step history.
- Prometheus remains free of `case_id` and `task_id` labels.

Outcome:
- Added optional `CaseStepHistoryStore` interfaces/types without extending `QueryStore`, so Mongo is not forced to implement case step history in this card.
- Implemented memory case step history persistence, filtering, retention, and worker step event materialization.
- Added app service timeline and exec-plan drilldown query methods over case step history.
- Updated storage documentation with Prometheus aggregate-only and Mongo follow-up responsibilities.
- Validation passed:
  - `cd reference/workerfleet && go test ./internal/platform/store/memory/...`
  - `cd reference/workerfleet && go test ./internal/app/...`
  - `cd reference/workerfleet && go test ./...`
