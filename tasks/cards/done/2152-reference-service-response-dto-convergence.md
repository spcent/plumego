# Card 2152: Reference Service Response DTO Convergence

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: reference
Owned Files:
- `reference/standard-service/internal/handler/health.go`
- `reference/standard-service/internal/handler/api.go`
- `reference/standard-service/internal/handler/handler_test.go`
- `reference/with-messaging/internal/handler/messaging.go`
- `reference/with-messaging/internal/handler/messaging_test.go`
Depends On: none

Goal:
Keep the canonical reference service and the messaging demo aligned with the
repository's typed response DTO style.

Problem:
The reference handlers still build several public success responses with
`map[string]any` literals. Because the reference application is the canonical
layout for users and agents, it should demonstrate explicit response structs
instead of ad hoc maps. The messaging demo also returns a simple publish result
map where a DTO is clearer.

Scope:
- Replace standard-service health, hello, greet, and status success maps with
  local DTO structs.
- Replace the with-messaging publish success map with a local DTO.
- Add focused handler tests that decode the contract envelope and assert the
  typed fields.
- Keep route paths, HTTP statuses, and response field names stable.

Non-goals:
- Do not refactor application bootstrap or route registration.
- Do not change workerfleet reference handlers.
- Do not change extension runtime behavior.
- Do not add dependencies.

Files:
- `reference/standard-service/internal/handler/health.go`
- `reference/standard-service/internal/handler/api.go`
- `reference/standard-service/internal/handler/handler_test.go`
- `reference/with-messaging/internal/handler/messaging.go`
- `reference/with-messaging/internal/handler/messaging_test.go`

Tests:
- `go test -race -timeout 60s ./reference/standard-service/... ./reference/with-messaging/...`
- `go test -timeout 20s ./reference/standard-service/... ./reference/with-messaging/...`
- `go vet ./reference/standard-service/... ./reference/with-messaging/...`

Docs Sync:
No docs update is required unless route behavior or public response fields
change.

Done Definition:
- Touched reference success responses use local DTO structs rather than ad hoc
  maps.
- Tests cover the standard-service and messaging demo response shapes.
- The listed validation commands pass.

Outcome:
- Replaced standard-service health, hello, greet, and status success maps with
  local DTO structs.
- Replaced the with-messaging publish success map with a local DTO.
- Added focused handler tests that decode the contract envelope and assert the
  typed response fields.
- Kept route paths, HTTP statuses, and response field names stable.

Validation:
- `go test -race -timeout 60s ./reference/standard-service/... ./reference/with-messaging/...`
- `go test -timeout 20s ./reference/standard-service/... ./reference/with-messaging/...`
- `go vet ./reference/standard-service/... ./reference/with-messaging/...`
