# Card 2200: Reference Workerfleet Metrics Output Helper

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P1
State: active
Primary Module: reference/workerfleet
Owned Files:
- `reference/workerfleet/internal/platform/metrics/prometheus_test.go`
Depends On: none

Goal:
Name the repeated Prometheus text containment assertions in the workerfleet
metrics tests.

Problem:
`prometheus_test.go` repeats inline `strings.Contains` checks for expected
metric families, samples, and handler output. The failure shape is duplicated,
and adding future metrics assertions requires retyping the same loop.

Scope:
- Add a local helper that asserts metrics text contains one or more expected
  fragments.
- Use it for collector text assertions and handler body assertions.

Non-goals:
- Do not change metrics collection, formatting, or public APIs.
- Do not add dependencies.

Files:
- `reference/workerfleet/internal/platform/metrics/prometheus_test.go`

Tests:
- from `reference/workerfleet`: `go test -race -timeout 60s ./internal/platform/metrics/...`
- from `reference/workerfleet`: `go test -timeout 20s ./internal/platform/metrics/...`
- from `reference/workerfleet`: `go vet ./internal/platform/metrics/...`

Docs Sync:
No docs change required; this is test cleanup only.

Done Definition:
- Metrics output assertions use a named helper.
- The listed validation commands pass.
