# Card 2205: Pubsub Prometheus Output Helper

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P1
State: active
Primary Module: x/pubsub
Owned Files:
- `x/pubsub/prometheus_test.go`
Depends On: none

Goal:
Consolidate repeated Prometheus output substring assertions in pubsub tests.

Problem:
`prometheus_test.go` repeats many `strings.Contains` checks for metric names,
labels, and Prometheus format markers. The assertions are equivalent but use
different messages, which makes the test file noisy and harder to extend.

Scope:
- Add a local helper for asserting metrics output contains one or more expected
  fragments.
- Use it for straightforward exporter output checks.

Non-goals:
- Do not change Prometheus exporter behavior, metric names, labels, or HTTP
  handling.
- Do not rewrite format-specific tests that intentionally inspect lines.
- Do not add dependencies.

Files:
- `x/pubsub/prometheus_test.go`

Tests:
- `go test -race -timeout 60s ./x/pubsub/...`
- `go test -timeout 20s ./x/pubsub/...`
- `go vet ./x/pubsub/...`

Docs Sync:
No docs change required; this is test cleanup only.

Done Definition:
- Straightforward Prometheus output assertions use a named helper.
- The listed validation commands pass.
