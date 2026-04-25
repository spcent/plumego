# Card 2144: x/ops Admin HTTP Safety and Response Convergence

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: active
Primary Module: x/ops
Owned Files:
- `x/ops/ops.go`
- `x/ops/ops_test.go`
- `docs/modules/x-ops/README.md`
Depends On: none

Goal:
Converge `x/ops` admin HTTP handlers around safe error logging, typed response
payloads, and one local validation helper for required query parameters.

Problem:
`x/ops/ops.go` is a protected admin surface, but several handlers still build
success payloads with ad hoc `map[string]any` values and repeat required-query
validation for `queue`, `message_id`, `provider`, and `tenant_id`. More
importantly, `writeHookError` currently logs `err.Error()` from caller-provided
hooks. Hook errors can contain queue names, tenant identifiers, provider
messages, or credentials from downstream systems, so this violates the repo rule
that secrets and sensitive values must not be logged by shared transport code.

Scope:
- Replace repeated required-query validation with one private helper that
  writes the existing structured validation error shape.
- Replace admin success maps with small local response DTO structs where the
  endpoint shape is stable (`summary`, `queues`, `replay`, `receipt`,
  `channels`, `quota`).
- Change hook-error logging to record safe metadata only (`code`, `path`,
  error type/category) without logging raw `err.Error()`.
- Add focused tests for required query validation, DTO response shape, and
  secret-safe hook error logging.

Non-goals:
- Do not change route paths or auth behavior.
- Do not add tenant policy logic to `x/ops`.
- Do not change hook interfaces or stable root APIs.
- Do not introduce a new response envelope.

Files:
- `x/ops/ops.go`
- `x/ops/ops_test.go`
- `docs/modules/x-ops/README.md`

Tests:
- `go test -race -timeout 60s ./x/ops/...`
- `go test -timeout 20s ./x/ops/...`
- `go vet ./x/ops/...`

Docs Sync:
Update `docs/modules/x-ops/README.md` only to document the safe hook-error
logging contract if it is not already explicit.

Done Definition:
- `writeHookError` no longer logs raw hook error text.
- Required-query failures share one private implementation path.
- Admin success responses no longer use one-off maps where local DTOs are
  clearer.
- The listed validation commands pass.

Outcome:
