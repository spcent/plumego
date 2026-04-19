# Card 2008

Milestone: —
Recipe: specs/change-recipes/add-http-endpoint.yaml
Priority: P1
State: done
Primary Module: reference/workerfleet/internal/handler
Owned Files:
- `reference/workerfleet/internal/handler/fleet_summary.go`
- `reference/workerfleet/internal/handler/worker_detail.go`
- `reference/workerfleet/internal/handler/task_lookup.go`
- `reference/workerfleet/internal/handler/alert_query.go`
- `reference/workerfleet/internal/handler/query_test.go`
Depends On:
- `tasks/cards/active/2002-workerfleet-api-contract-and-routes.md`
- `tasks/cards/active/2003-workerfleet-store-and-retention.md`
- `tasks/cards/active/2006-workerfleet-alert-engine.md`

Goal:
- Implement the read-side query endpoints needed for operators to inspect fleet health, worker details, active tasks, task history, and alert state.

Scope:
- Fleet summary, worker list/detail, task reverse lookup, and alert query handlers.
- Filter, sort, pagination, and summary aggregation behavior for the monitoring service.
- Read-only diagnostics surfaced through the approved query API.

Non-goals:
- Do not build a web UI.
- Do not add write-side admin actions such as replay or mutating control endpoints.
- Do not add protected ops surfaces beyond read-only monitoring queries unless separately approved.

Files:
- `reference/workerfleet/internal/handler/fleet_summary.go`
- `reference/workerfleet/internal/handler/worker_detail.go`
- `reference/workerfleet/internal/handler/task_lookup.go`
- `reference/workerfleet/internal/handler/alert_query.go`
- `reference/workerfleet/internal/handler/query_test.go`

Tests:
- `go test ./reference/workerfleet/internal/handler/...`
- Coverage for filters by status, namespace, node, and task type, plus pagination and empty-state behavior.

Docs Sync:
- Update `reference/workerfleet/docs/api.md` with read-side query examples and field definitions.

Done Definition:
- Operators can query fleet summary, worker details, task lookup, and alert state from the persisted model.
- Read-side endpoints remain transport-only and delegate aggregation to app-local services/store adapters.
- Query semantics match the API contract established in card 2002.

Outcome:
- The read-side query endpoints defined in card 2002 are now backed by the app-local service and in-memory store.
- Added app-level tests covering worker filters, pagination, empty alert results, and task-history fallback lookup.
- Query semantics remain transport-only in handlers and aggregate from persisted state in the app service/store layer.
- Validation run: `go test ./reference/workerfleet/internal/...`
