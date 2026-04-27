# Card 0625: Health State Helper Cleanup

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: health
Owned Files:
- `health/core.go`
- `health/health_test.go`
Depends On: none

Goal:
Remove redundant private health-state helpers and keep tests focused on the
stable public readiness behavior.

Scope:
- Inline the trivial `HealthState.IsReady` implementation.
- Remove the private state validator that has no production caller.
- Drop tests that only exercise private implementation details.

Non-goals:
- Do not add, remove, or rename exported symbols.
- Do not change readiness semantics.
- Do not touch `x/ops/healthhttp`.

Files:
- `health/core.go`
- `health/health_test.go`

Tests:
- `go test -race -timeout 60s ./health/...`
- `go test -timeout 20s ./health/...`
- `go vet ./health/...`

Docs Sync:
No docs change required; public behavior is unchanged.

Done Definition:
- `HealthState.IsReady` remains the single readiness predicate in `health`.
- No private helper remains solely for tests.
- The listed validation commands pass.

Outcome:
- Removed the private `HealthState.isReady` wrapper and inlined the readiness
  predicate in `HealthState.IsReady`.
- Removed the private health-state validator because it had no production
  callers and was only tested as an implementation detail.
- Converted `TestHealthState_IsReady` into named subtests covering the public
  readiness contract.
- Validation run:
  - `go test -race -timeout 60s ./health/...`
  - `go test -timeout 20s ./health/...`
  - `go vet ./health/...`
