# Card 2194: Health Status Test Detail DTO Convergence

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P2
State: active
Primary Module: health
Owned Files:
- `health/status_test.go`
Depends On: none

Goal:
Make the health status state test use a typed detail fixture for its fixed
`enabled` field.

Problem:
`TestHealthStatusStates` uses `map[string]any` and a type assertion for a
single fixed boolean detail.

Scope:
- Replace the fixed map detail fixture with a local typed detail struct.

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
- The status state test no longer type-asserts a fixed detail field from a map.
- The listed validation commands pass.

Outcome:
