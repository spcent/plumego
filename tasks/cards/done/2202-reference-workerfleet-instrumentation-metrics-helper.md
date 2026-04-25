# Card 2202: Reference Workerfleet Instrumentation Metrics Helper

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P1
State: done
Primary Module: reference/workerfleet
Owned Files:
- `reference/workerfleet/internal/platform/metrics/instrumentation_test.go`
Depends On: none

Goal:
Consolidate repeated instrumentation metrics text assertions in workerfleet.

Problem:
`instrumentation_test.go` contains many repeated loops that assert Prometheus
text contains expected samples, plus a separate loop for forbidden labels. The
assertion shape is duplicated across observer scenarios and obscures the actual
metric expectations.

Scope:
- Add local helpers for expected and forbidden metrics text fragments.
- Use them across observer instrumentation tests.

Non-goals:
- Do not change observer behavior, metric names, labels, or public APIs.
- Do not add dependencies.

Files:
- `reference/workerfleet/internal/platform/metrics/instrumentation_test.go`

Tests:
- from `reference/workerfleet`: `go test -race -timeout 60s ./internal/platform/metrics/...`
- from `reference/workerfleet`: `go test -timeout 20s ./internal/platform/metrics/...`
- from `reference/workerfleet`: `go vet ./internal/platform/metrics/...`

Docs Sync:
No docs change required; this is test cleanup only.

Done Definition:
- Instrumentation metrics assertions use named helpers.
- The listed validation commands pass.

Outcome:
- Added `assertMetricsTextContains` and `assertMetricsTextOmits`.
- Validation passed for workerfleet metrics race tests, normal tests, and vet.
