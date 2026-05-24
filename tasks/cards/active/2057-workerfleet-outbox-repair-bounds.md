# Card 2057

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: implementation
Priority: P1
State: active
Primary Module: reference/workerfleet
Owned Files:
- reference/workerfleet/internal/app/alert_loop.go
- reference/workerfleet/internal/app/alert_loop_test.go
- reference/workerfleet/internal/platform/store/interfaces.go
- reference/workerfleet/internal/platform/store/memory/notification_outbox.go
- reference/workerfleet/internal/platform/store/mongo/notification_outbox_store.go
Depends On: 2056

Goal:
Keep notification outbox repair from scanning and re-enqueueing the full retained alert history on every delivery tick.

Scope:
Introduce an app-local bounded repair path that only considers alert records still missing notification jobs for configured sinks, or otherwise records enough store state to avoid repeated full-history repair work. Preserve the idempotent repair guarantee from card 2050.

Non-goals:
- Do not introduce external queues.
- Do not add Mongo transactions.
- Do not claim exactly-once delivery.

Files:
- reference/workerfleet/internal/app/alert_loop.go
- reference/workerfleet/internal/app/alert_loop_test.go
- reference/workerfleet/internal/platform/store/interfaces.go
- reference/workerfleet/internal/platform/store/memory/notification_outbox.go
- reference/workerfleet/internal/platform/store/mongo/notification_outbox_store.go

Acceptance Tests:
- reference/workerfleet/internal/app/alert_loop_test.go: TestDeliverNotificationOutboxRepairsOnlyMissingJobs

Tests:
- Existing alert loop, memory outbox, and Mongo outbox tests.

Docs Sync:

Validation:
- cd reference/workerfleet && go test -timeout 30s ./internal/app ./internal/platform/store/memory ./internal/platform/store/mongo
- cd reference/workerfleet && go vet ./internal/app ./internal/platform/store/memory ./internal/platform/store/mongo
- gofmt -l reference/workerfleet/internal/app reference/workerfleet/internal/platform/store

Done Definition:
- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l . produces no output.
- [ ] Docs Sync targets updated (if applicable).

Outcome:
