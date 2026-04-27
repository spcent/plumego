# Card 5501: Health State Known Validation

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: health
Owned Files:
- `health/core.go`
- `health/health_test.go`
- `health/example_test.go`
- `docs/modules/health/README.md`
- `x/ops/healthhttp/history.go`
Depends On: none

Goal:
Make known-state validation a single health-owned behavior instead of
duplicating the allowed `HealthState` set in extension code.

Scope:
- Add a non-breaking `HealthState.IsKnown` method.
- Cover known and unknown state behavior in health tests and examples.
- Replace the duplicate `x/ops/healthhttp` private state validator.
- Update the health module docs to mention the known-state predicate.

Non-goals:
- Do not change readiness semantics.
- Do not change JSON payload shape.
- Do not add new health manager, HTTP, or orchestration behavior.

Files:
- `health/core.go`
- `health/health_test.go`
- `health/example_test.go`
- `docs/modules/health/README.md`
- `x/ops/healthhttp/history.go`

Tests:
- `rg -n "func isValidHealthState" x/ops/healthhttp health`
- `go test -race -timeout 60s ./health/...`
- `go test -timeout 20s ./health/...`
- `go test -timeout 20s ./x/ops/...`
- `go vet ./health/...`

Docs Sync:
Update `docs/modules/health/README.md` because this adds a stable public helper.

Done Definition:
- `HealthState.IsKnown` is the single known-state predicate for health states.
- `x/ops/healthhttp` no longer duplicates the status switch.
- The listed validation commands pass.

Outcome:
- Added `HealthState.IsKnown` as the health-owned known-state predicate.
- Added tests and executable example coverage for known and unknown states.
- Replaced `x/ops/healthhttp` history query validation with `state.IsKnown()`
  and removed its duplicate private switch.
- Updated health module docs for the new stable helper.
- Validation run:
  - `rg -n "func isValidHealthState" x/ops/healthhttp health` returned no matches
  - `go test -race -timeout 60s ./health/...`
  - `go test -timeout 20s ./health/...`
  - `go test -timeout 20s ./x/ops/...`
  - `go vet ./health/...`
