# Card 5502: Health Test Consolidation

Milestone: none
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: health
Owned Files:
- `health/health_test.go`
- `health/status_test.go`
Depends On: 5501

Goal:
Remove overlapping thin model-field tests now that public DTO shape and examples
are covered by focused contract tests.

Scope:
- Keep `health/health_test.go` focused on `HealthState` behavior.
- Remove trivial `ComponentHealth`, `ReadinessStatus`, and `HealthStatus`
  field-assignment tests that duplicate JSON contract and example coverage.

Non-goals:
- Do not reduce JSON contract coverage.
- Do not change runtime behavior.
- Do not change public APIs.

Files:
- `health/health_test.go`
- `health/status_test.go`

Tests:
- `go test -race -timeout 60s ./health/...`
- `go test -timeout 20s ./health/...`
- `go vet ./health/...`

Docs Sync:
No docs change required; this is test hygiene.

Done Definition:
- `health` tests no longer repeat simple struct field assignment checks.
- State behavior, JSON contracts, and examples remain covered.
- The listed validation commands pass.

Outcome:
- Removed trivial `ComponentHealth`, `ReadinessStatus`, and `HealthStatus`
  field assignment tests.
- Deleted `health/status_test.go`; the remaining tests now focus on
  `HealthState` behavior, JSON contracts, and executable examples.
- Validation run:
  - `go test -race -timeout 60s ./health/...`
  - `go test -timeout 20s ./health/...`
  - `go vet ./health/...`
