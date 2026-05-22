# Card 2048

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: implementation
Priority: P1
State: active
Primary Module: reference/workerfleet
Owned Files:
- reference/workerfleet/internal/app
- reference/workerfleet/internal/platform/store
- reference/workerfleet/internal/platform/store/mongo
- reference/workerfleet/internal/platform/notifier
- reference/workerfleet/docs/alerts.md
Depends On: 2047

Goal:
Make workerfleet notification delivery durable and retryable through an app-local outbox.

Scope:
Persist per-sink notification jobs, add a delivery loop with retry/backoff and low-cardinality error classification, and keep alert evaluation independent from direct sink availability.

Non-goals:
- Do not claim exactly-once delivery.
- Do not add external queues or third-party dependencies.
- Do not split workerfleet into separate API/worker binaries in this card.

Files:
- reference/workerfleet/internal/app/alert_loop.go
- reference/workerfleet/internal/platform/store/interfaces.go
- reference/workerfleet/internal/platform/store/mongo/notification_outbox_store.go
- reference/workerfleet/internal/platform/notifier/dispatcher.go
- reference/workerfleet/docs/alerts.md

Acceptance Tests:
- reference/workerfleet/internal/platform/store/mongo/notification_outbox_store_test.go: TestNotificationOutboxEnqueueClaimRetryDeliver
- reference/workerfleet/internal/app/alert_loop_test.go: TestEvaluateAndNotifyAlertsEnqueuesOutbox

Tests:
- Retry classification for transient and permanent notification failures.
- Idempotent enqueue by alert ID and sink type.

Docs Sync:
- reference/workerfleet/docs/alerts.md
- reference/workerfleet/docs/notifiers.md
- reference/workerfleet/docs/design/technical-design.md

Validation:
- cd reference/workerfleet && go test -timeout 30s ./internal/app ./internal/platform/store/mongo ./internal/platform/notifier
- cd reference/workerfleet && go vet ./internal/app ./internal/platform/store/mongo ./internal/platform/notifier
- gofmt -l reference/workerfleet/internal/app reference/workerfleet/internal/platform

Done Definition:
- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l . produces no output.
- [ ] Docs Sync targets updated (if applicable).

Outcome:
