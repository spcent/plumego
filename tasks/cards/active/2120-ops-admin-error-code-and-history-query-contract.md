# Card 2120: Ops Admin Error Code And History Query Contract

Milestone: none
Recipe: specs/change-recipes/http-endpoint-bugfix.yaml
Priority: high
State: active
Primary Module: x/ops
Owned Files:
- x/ops/ops.go
- x/ops/ops_test.go
- x/ops/healthhttp/history.go
- x/ops/healthhttp/history_test.go
- docs/modules/x-ops/README.md
Depends On: none

Goal:
- Normalize protected ops endpoint error codes and health-history query errors.
- Remove lowercase ad hoc not-configured codes and direct invalid query error messages from public responses.

Scope:
- Audit `x/ops/ops.go` not-configured and not-implemented admin helper responses.
- Audit `x/ops/healthhttp/history.go` invalid query parameter response construction.
- Preserve explicit auth gating and fail-closed behavior.
- Add focused tests for queue/receipt not-configured responses and invalid health-history query parameters.

Non-goals:
- Do not change queue replay, queue stats, receipt lifecycle, health manager, or metrics semantics.
- Do not move ops endpoints into stable `health`, `metrics`, `core`, or `middleware`.
- Do not add tenant-aware policy to `x/ops`.

Files:
- `x/ops/ops.go`: replace lowercase ad hoc codes with explicit stable ops codes.
- `x/ops/ops_test.go`: cover protected admin not-configured errors.
- `x/ops/healthhttp/history.go`: use safe invalid-query messages while retaining useful parameter details.
- `x/ops/healthhttp/history_test.go`: cover invalid query error responses.
- `docs/modules/x-ops/README.md`: document ops error-code expectations if public behavior changes.

Tests:
- `go test -race -timeout 60s ./x/ops/...`
- `go test -timeout 20s ./x/ops/...`
- `go vet ./x/ops/...`

Docs Sync:
- Required if public ops error codes or health-history response examples change.

Done Definition:
- Protected ops errors in scope use explicit uppercase stable codes.
- Health-history invalid query responses do not expose raw parse or validation errors as response messages.
- Auth behavior remains fail-closed and unchanged.
- The three listed validation commands pass.

Outcome:

