# Card 5407: Health Readiness Convergence

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: x/ops
Owned Files:
- `x/ops/healthhttp/manager.go`
- `x/ops/healthhttp/readiness.go`
- `x/ops/healthhttp/manager_test.go`
- `x/ops/healthhttp/handlers_test.go`
Depends On: 5406

Goal:
Converge x/ops readiness behavior on the stable `health.HealthState.IsReady`
predicate and fill the missing initial readiness timestamp.

Scope:
- Use `HealthState.IsReady` when deriving readiness from aggregate health.
- Ensure a newly created health manager exposes a timestamped initial
  `ReadinessStatus`.
- Add focused tests for initial readiness timestamps and degraded aggregate
  readiness.

Non-goals:
- Do not move x/ops HTTP handlers or manager orchestration into stable
  `health`.
- Do not change HTTP response envelopes or status mapping beyond the readiness
  predicate convergence.
- Do not alter health state values.

Files:
- `x/ops/healthhttp/manager.go`
- `x/ops/healthhttp/readiness.go`
- `x/ops/healthhttp/manager_test.go`
- `x/ops/healthhttp/handlers_test.go`

Tests:
- `go test -race -timeout 60s ./x/ops/...`
- `go test -timeout 20s ./x/ops/...`
- `go vet ./x/ops/...`

Docs Sync:
No docs change required; this aligns implementation with the documented
`HealthState.IsReady` contract.

Done Definition:
- Readiness derived from aggregate health calls the stable readiness predicate.
- Initial manager readiness includes a non-zero timestamp.
- The listed validation commands pass.

Outcome:
- Added an initial readiness timestamp when constructing a health HTTP manager.
- Changed readiness derivation from aggregate health to call
  `HealthState.IsReady`.
- Added coverage for timestamped initial readiness and degraded aggregate
  readiness.
- Validation run:
  - `go test -race -timeout 60s ./x/ops/...`
  - `go test -timeout 20s ./x/ops/...`
  - `go vet ./x/ops/...`
