# Card 2022

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: active
Primary Module: reference/workerfleet/internal/app
Owned Files:
- `reference/workerfleet/internal/app/config.go`
- `reference/workerfleet/internal/app/bootstrap.go`
- `reference/workerfleet/internal/app/alert_loop.go`
- `reference/workerfleet/internal/app/alert_loop_test.go`
- `reference/workerfleet/main.go`
Depends On:
- `tasks/cards/active/2020-workerfleet-runtime-health-and-config.md`
Blocked By:
- runtime health/config card

Goal:
- Wire alert evaluation and Feishu/Webhook notification dispatch into the running workerfleet process.
- Make firing and resolved alert records actively delivered instead of only persisted/queryable.

Scope:
- Add alert evaluation interval and notifier delivery timeout config.
- Add Feishu webhook URL and generic webhook URL/header config.
- Evaluate domain alerts periodically from persisted worker snapshots.
- Dispatch newly emitted firing and resolved alerts to configured sinks.
- Ensure notifier errors are logged and counted without exposing secrets.

Non-goals:
- Do not implement durable notification retry queues.
- Do not add exactly-once notification semantics.
- Do not change alert dedupe keys or domain alert rules.

Files:
- `reference/workerfleet/internal/app/config.go`
- `reference/workerfleet/internal/app/bootstrap.go`
- `reference/workerfleet/internal/app/alert_loop.go`
- `reference/workerfleet/internal/app/alert_loop_test.go`
- `reference/workerfleet/main.go`

Tests:
- `cd reference/workerfleet && go test ./internal/app/...`
- `cd reference/workerfleet && go test ./internal/platform/notifier/...`
- `cd reference/workerfleet && go test ./...`

Docs Sync:
- Update `reference/workerfleet/docs/alerts.md` and `reference/workerfleet/docs/notifiers.md`.

Done Definition:
- Alert evaluation loop emits and persists firing/resolved records.
- Configured Feishu and Webhook sinks receive emitted alerts.
- Delivery failures do not crash the service and do not log secrets.

