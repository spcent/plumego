# Card 2194: Health Status Test Detail Assertion Convergence

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P2
State: active
Primary Module: health
Owned Files:
- `health/status_test.go`
Depends On: none

Goal:
Make the health status state test use a named helper for its fixed `enabled`
detail assertion.

Problem:
`TestHealthStatusStates` uses an inline string key and type assertion for a
single fixed boolean detail. `HealthStatus.Details` is intentionally a generic
extension map, but the test can keep the fixed detail assertion clearer.

Scope:
- Add a local helper for asserting boolean detail fields.
- Name the fixed `enabled` detail key.

Non-goals:
- Do not change health models or public APIs.
- Do not add dependencies.

Files:
- `health/status_test.go`

Tests:
- `go test -race -timeout 60s ./health/...`
- `go test -timeout 20s ./health/...`
- `go vet ./health/...`

Docs Sync:
No docs change required; this is test cleanup.

Done Definition:
- The status state test no longer repeats inline detail key/type assertion
  logic.
- The listed validation commands pass.

Outcome:
