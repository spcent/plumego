# Card 2123: Workerfleet Handler Error Contract And Placeholder Cleanup

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: high
State: active
Primary Module: reference/workerfleet
Owned Files:
- reference/workerfleet/internal/handler/worker_register.go
- reference/workerfleet/internal/handler/worker_query.go
- reference/workerfleet/internal/handler/health.go
- reference/workerfleet/internal/handler/query_test.go
- reference/workerfleet/internal/handler/health_test.go
Depends On: none

Goal:
- Make Workerfleet reference handlers use safe, stable HTTP error responses.
- Remove raw service/readiness error text and lowercase placeholder codes from handler responses.

Scope:
- Audit worker register, query, and health handlers for `Message(err.Error())` and lowercase not-implemented codes.
- Keep reference app route wiring, service interfaces, Mongo/memory stores, and domain behavior unchanged.
- Add or update handler tests for not-found, conflict, not-implemented/not-configured, readiness failure, and query placeholder responses.
- Keep this work inside the Workerfleet reference application.

Non-goals:
- Do not change main-module stable roots or `x/*` packages.
- Do not implement missing Workerfleet platform features as part of this card.
- Do not add new dependencies to the reference app.

Files:
- `reference/workerfleet/internal/handler/worker_register.go`: normalize service and placeholder error responses.
- `reference/workerfleet/internal/handler/worker_query.go`: normalize query placeholder/error responses if present.
- `reference/workerfleet/internal/handler/health.go`: replace raw readiness error messages with safe response text.
- `reference/workerfleet/internal/handler/query_test.go`: cover query error contracts.
- `reference/workerfleet/internal/handler/health_test.go`: cover readiness error contracts.

Tests:
- `cd reference/workerfleet && go test ./internal/handler/...`
- `cd reference/workerfleet && go test ./...`
- `cd reference/workerfleet && go vet ./...`

Docs Sync:
- Not required unless `reference/workerfleet/README.md` documents exact handler error payloads.

Done Definition:
- Workerfleet handler responses in scope no longer expose raw service or readiness error strings.
- Placeholder and not-configured responses use explicit uppercase stable codes.
- Tests cover the normalized error contracts.
- The three listed validation commands pass from the `reference/workerfleet` module.

Outcome:

