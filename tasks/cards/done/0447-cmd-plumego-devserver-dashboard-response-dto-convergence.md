# Card 0447: cmd/plumego Devserver Dashboard Response DTO Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: done
Primary Module: cmd/plumego
Owned Files:
- `cmd/plumego/internal/devserver/dashboard.go`
- `cmd/plumego/internal/devserver/dashboard_info_test.go`
- `cmd/plumego/DEV_SERVER.md`
Depends On: none

Goal:
Converge devserver dashboard JSON success responses on local typed DTO structs
while preserving the existing dashboard API field names and behavior.

Problem:
`cmd/plumego/internal/devserver/dashboard.go` still assembles several internal
dashboard API responses with nested ad hoc maps. The metrics handler also
mutates nested maps through type assertions, which is brittle and inconsistent
with the newer local DTO shape already used for dashboard actions.

Scope:
- Replace dashboard status, health, routes, metrics, pprof types, and pprof
  JSON preview success maps with local DTO structs.
- Keep JSON field names and route paths unchanged.
- Add focused tests that decode representative dashboard success responses into
  typed DTOs.
- Update devserver docs to state that dashboard JSON success payloads use typed
  response DTOs.

Non-goals:
- Do not change dashboard UI assets, route registration, build/restart/stop
  behavior, pprof raw download behavior, or analyzer behavior.
- Do not move devserver code into stable roots or `x/*`.
- Do not add dependencies.

Files:
- `cmd/plumego/internal/devserver/dashboard.go`
- `cmd/plumego/internal/devserver/dashboard_info_test.go`
- `cmd/plumego/DEV_SERVER.md`

Tests:
- `go test -race -timeout 60s ./internal/devserver/...` from `cmd/plumego`
- `go test -timeout 20s ./internal/devserver/...` from `cmd/plumego`
- `go vet ./internal/devserver/...` from `cmd/plumego`

Docs Sync:
Update `cmd/plumego/DEV_SERVER.md` for the dashboard response DTO policy.

Done Definition:
- Dashboard success responses in scope no longer use one-off map payloads.
- Metrics response construction no longer mutates nested maps through type
  assertions.
- Focused devserver tests cover representative typed response shapes.
- The listed validation commands pass.

Outcome:
- Added local typed DTOs for dashboard status, health, routes, metrics, pprof
  types, and pprof preview responses.
- Replaced nested response maps and removed metrics construction through nested
  map type assertions.
- Added focused devserver handler tests that decode typed success payloads from
  the contract envelope.
- Documented the dashboard success DTO policy in `cmd/plumego/DEV_SERVER.md`.

Validation:
- `go test -race -timeout 60s ./internal/devserver/...` from `cmd/plumego`
- `go test -timeout 20s ./internal/devserver/...` from `cmd/plumego`
- `go vet ./internal/devserver/...` from `cmd/plumego`
