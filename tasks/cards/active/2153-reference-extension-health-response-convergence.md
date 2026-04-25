# Card 2153: Reference Extension Health Response Convergence

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: active
Primary Module: reference
Owned Files:
- `reference/with-gateway/internal/app/routes.go`
- `reference/with-webhook/internal/app/routes.go`
- `reference/with-websocket/internal/app/routes.go`
- `reference/with-messaging/internal/app/routes.go`
- `reference/with-messaging/internal/app/routes_test.go`
Depends On: none

Goal:
Keep the extension demo health endpoints aligned with the typed response DTO
style used by the canonical reference handlers.

Problem:
After the standard-service and messaging handler response cleanup, the extension
demo app route files still define `/healthz` responses with inline
`map[string]any` literals. These demos are copied by agents and users as
examples, so their route-level health handlers should demonstrate explicit
response structs rather than ad hoc maps.

Scope:
- Replace extension demo `/healthz` success maps with local DTO structs.
- Keep route paths, HTTP status codes, and response field names stable.
- Add focused coverage for the with-messaging app health response shape as the
  representative route-level demo pattern.

Non-goals:
- Do not refactor app bootstrap or extension registration.
- Do not change proxy, webhook, websocket, or broker behavior.
- Do not change workerfleet reference code.
- Do not add dependencies.

Files:
- `reference/with-gateway/internal/app/routes.go`
- `reference/with-webhook/internal/app/routes.go`
- `reference/with-websocket/internal/app/routes.go`
- `reference/with-messaging/internal/app/routes.go`
- `reference/with-messaging/internal/app/routes_test.go`

Tests:
- `go test -race -timeout 60s ./reference/with-gateway/... ./reference/with-webhook/... ./reference/with-websocket/... ./reference/with-messaging/...`
- `go test -timeout 20s ./reference/with-gateway/... ./reference/with-webhook/... ./reference/with-websocket/... ./reference/with-messaging/...`
- `go vet ./reference/with-gateway/... ./reference/with-webhook/... ./reference/with-websocket/... ./reference/with-messaging/...`

Docs Sync:
No docs update is required because route behavior and public response fields do
not change.

Done Definition:
- Extension demo health responses use typed DTO structs rather than inline maps.
- The representative with-messaging route-level health response shape is covered
  by a focused test.
- The listed validation commands pass.

Outcome:
