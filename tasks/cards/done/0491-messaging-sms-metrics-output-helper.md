# Card 0491: Messaging SMS Metrics Output Helper

Milestone: none
Recipe: specs/change-recipes/module-cleanup.yaml
Priority: P1
State: done
Primary Module: x/messaging
Owned Files:
- `x/messaging/sms_metrics_test.go`
Depends On: none

Goal:
Consolidate repeated SMS Prometheus exporter body assertions behind a named
test helper.

Problem:
`sms_metrics_test.go` repeats inline `strings.Contains` checks for metrics body
fragments. The checks are behaviorally the same but have slightly different
failure messages, making the test less uniform and harder to extend.

Scope:
- Add a local helper for asserting an exported metrics body contains one or
  more fragments.
- Use it in SMS Prometheus exporter tests.

Non-goals:
- Do not change SMS metrics names, labels, values, or exporter behavior.
- Do not add dependencies.

Files:
- `x/messaging/sms_metrics_test.go`

Tests:
- `go test -race -timeout 60s ./x/messaging/...`
- `go test -timeout 20s ./x/messaging/...`
- `go vet ./x/messaging/...`

Docs Sync:
No docs change required; this is test cleanup only.

Done Definition:
- SMS metrics body assertions use a named helper.
- The listed validation commands pass.

Outcome:
- Added `assertMetricsBodyContains` for SMS Prometheus exporter body checks.
- Validation passed for messaging race tests, normal tests, and vet.
