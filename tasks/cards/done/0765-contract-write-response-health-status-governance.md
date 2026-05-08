# Card 0765

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P0
State: done
Primary Module: contract
Owned Files:
- contract/conformance_test.go
- x/ops/healthhttp/helpers.go
- x/ops/healthhttp/handlers.go
- x/ops/healthhttp/readiness.go
- x/ops/healthhttp/metrics.go
Depends On:
- 0764

Goal:
Make health/readiness uses of `contract.WriteResponse` with dynamic non-2xx statuses explicit and machine-governed.

Scope:
- Move health/readiness status-selected success envelopes behind a small `x/ops/healthhttp` helper.
- Extend contract conformance so dynamic `WriteResponse` status expressions are rejected unless an explicit function-level allowlist permits them.
- Keep ordinary external `WriteResponse` calls on known 2xx statuses.

Non-goals:
- Do not change the `WriteResponse` public API.
- Do not change health/readiness response JSON shape or status mapping.
- Do not introduce new error envelopes.

Files:
- contract/conformance_test.go
- x/ops/healthhttp/helpers.go
- x/ops/healthhttp/handlers.go
- x/ops/healthhttp/readiness.go
- x/ops/healthhttp/metrics.go

Tests:
- go test -timeout 20s ./contract/...
- go test -timeout 20s ./x/ops/healthhttp/...
- go vet ./contract/... ./x/ops/healthhttp/...

Docs Sync:
- Not required unless the frozen `WriteResponse` semantics change.

Done Definition:
- Dynamic non-2xx success envelope use is isolated to an explicit health helper.
- New dynamic external `WriteResponse` statuses fail contract conformance unless allowlisted.
- Target checks pass.

Outcome:
- Added `x/ops/healthhttp.writeHealthResponse` as the single explicit helper for health/readiness success envelopes that intentionally carry derived non-2xx statuses.
- Replaced dynamic `contract.WriteResponse` calls in health summary, detailed health, component health, readiness, and health report handlers with the helper.
- Hardened contract conformance so external `WriteResponse` calls must use known 2xx statuses unless an exact function-level dynamic-status allowlist permits the call.

Validation:
- go test -timeout 60s ./contract/...
- go test -timeout 20s ./x/ops/healthhttp/...
- go vet ./contract/... ./x/ops/healthhttp/...
