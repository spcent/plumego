# Card 2002

Milestone: —
Recipe: specs/change-recipes/add-http-endpoint.yaml
Priority: P0
State: active
Primary Module: reference/workerfleet/internal/handler
Owned Files:
- `reference/workerfleet/internal/handler/worker_register.go`
- `reference/workerfleet/internal/handler/worker_heartbeat.go`
- `reference/workerfleet/internal/handler/worker_query.go`
- `reference/workerfleet/internal/app/routes.go`
- `reference/workerfleet/docs/api.md`
Depends On:
- `tasks/cards/active/2001-workerfleet-domain-status-model.md`

Goal:
- Define the HTTP contract for worker registration, heartbeat/task sync, fleet summary, worker detail, task lookup, and alert queries.
- Keep route wiring explicit and Plumego-style under `reference/workerfleet/internal/app/routes.go`.

Scope:
- Handler request and response shapes for registration, heartbeat/task sync, fleet summary, worker detail, task lookup, and alert queries.
- Explicit route coverage for `POST /v1/workers/register`, `POST /v1/workers/heartbeat`, `GET /v1/workers`, `GET /v1/workers/{worker_id}`, `GET /v1/tasks/{task_id}`, `GET /v1/fleet/summary`, and `GET /v1/alerts`.
- Query parameter, pagination, and error-code conventions for the monitoring service.

Non-goals:
- Do not add storage adapters.
- Do not implement Kubernetes discovery.
- Do not build a UI.

Files:
- `reference/workerfleet/internal/handler/worker_register.go`
- `reference/workerfleet/internal/handler/worker_heartbeat.go`
- `reference/workerfleet/internal/handler/worker_query.go`
- `reference/workerfleet/internal/app/routes.go`
- `reference/workerfleet/docs/api.md`

Tests:
- `go test ./reference/workerfleet/internal/handler/...`
- Handler contract tests for invalid JSON, missing fields, multi-task payloads, filter parsing, and not-found responses.

Docs Sync:
- Update `reference/workerfleet/README.md` with the approved API endpoints and payload examples.

Done Definition:
- Route paths, methods, payloads, response shapes, pagination, and error codes are stable and documented.
- Handlers use the domain types from card 2001 instead of redefining status/task contracts.
- Route registration is grep-friendly and explicit in one place.

Outcome:
- Pending execution.
