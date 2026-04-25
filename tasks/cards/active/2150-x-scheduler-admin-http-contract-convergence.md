# Card 2150: x/scheduler Admin HTTP Contract Convergence

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: P1
State: active
Primary Module: x/scheduler
Owned Files:
- `x/scheduler/admin_http.go`
- `x/scheduler/admin_http_test.go`
- `docs/modules/x-scheduler/README.md`
Depends On: none

Goal:
Converge scheduler admin HTTP handlers around typed success responses and
structured validation for malformed boolean query filters.

Problem:
`AdminHandler` already uses `contract.WriteResponse` and `contract.WriteError`,
but several success bodies still use one-off maps (`status`, `cleared`,
`affected`). The job list query parser also silently ignores malformed
`running`, `paused`, and `asc` boolean values. That makes invalid query input
change filtering behavior instead of returning a canonical validation error.

Scope:
- Replace admin action, dead-letter clear/delete, and bulk operation success
  maps with local DTO structs.
- Add a single private path for parsing optional boolean query parameters that
  reports malformed values.
- Make `GET /scheduler/jobs` return a structured validation error for invalid
  boolean query filters.
- Add focused tests for the typed response shapes and invalid boolean query
  handling.
- Update the scheduler module primer with the admin query validation contract.

Non-goals:
- Do not redesign scheduler routing or path layout.
- Do not change job state, retry, DLQ, or group/tag semantics.
- Do not add external dependencies.
- Do not move scheduler behavior into stable roots.

Files:
- `x/scheduler/admin_http.go`
- `x/scheduler/admin_http_test.go`
- `docs/modules/x-scheduler/README.md`

Tests:
- `go test -race -timeout 60s ./x/scheduler/...`
- `go test -timeout 20s ./x/scheduler/...`
- `go vet ./x/scheduler/...`

Docs Sync:
Update `docs/modules/x-scheduler/README.md` to document that malformed boolean
admin query values fail with a structured validation error.

Done Definition:
- Admin success responses no longer use one-off maps for simple action/result
  payloads.
- Invalid `running`, `paused`, or `asc` query values fail through
  `contract.WriteError`.
- Focused tests cover the response DTOs and query validation.
- The listed validation commands pass.

Outcome:
