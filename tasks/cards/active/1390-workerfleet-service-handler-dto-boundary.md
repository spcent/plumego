# Card 1390

Milestone: workerfleet-hardening
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: active
Primary Module: reference/workerfleet/internal/app
Owned Files:
- reference/workerfleet/internal/app/service.go
- reference/workerfleet/internal/app/service_test.go
- reference/workerfleet/internal/handler/worker_register.go
- reference/workerfleet/internal/handler/worker_heartbeat.go
- reference/workerfleet/internal/handler/worker_query.go
Depends On:
- 1389

Goal:
- Remove the reverse dependency where app service methods expose handler DTOs.

Scope:
- Define app-level command/result/view types in `internal/app` or a small app-local boundary file.
- Keep HTTP request/response DTOs in `internal/handler`.
- Move mapping between app results and HTTP JSON DTOs into handlers.
- Keep handler shape as `func(http.ResponseWriter, *http.Request)`.
- Preserve all route paths, status codes, JSON field names, and error envelopes.

Non-goals:
- Do not change domain status rules or store interfaces.
- Do not add a generic mapper package.
- Do not change API docs except to fix accidental drift found during tests.

Files:
- reference/workerfleet/internal/app/service.go
- reference/workerfleet/internal/app/service_test.go
- reference/workerfleet/internal/handler/worker_register.go
- reference/workerfleet/internal/handler/worker_heartbeat.go
- reference/workerfleet/internal/handler/worker_query.go

Tests:
- cd reference/workerfleet && go test -timeout 20s ./internal/app/...
- cd reference/workerfleet && go test -timeout 20s ./internal/handler/...
- cd reference/workerfleet && go test -timeout 20s ./...

Docs Sync:
- Not required unless an API payload changes; payload changes are out of scope.

Done Definition:
- `internal/app` no longer imports `internal/handler`.
- Handler DTOs remain transport-owned.
- Existing handler tests still prove response shape compatibility.
- Target checks pass.

Outcome:
