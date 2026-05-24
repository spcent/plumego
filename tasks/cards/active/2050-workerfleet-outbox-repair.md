# Card 2050

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: implementation
Priority: P0
State: active
Primary Module: reference/workerfleet
Owned Files:
- reference/workerfleet/internal/app/alert_loop.go
- reference/workerfleet/internal/app/alert_loop_test.go
- reference/workerfleet/docs/alerts.md
- reference/workerfleet/docs/notifiers.md
Depends On: 2049

Goal:
Ensure persisted alert records cannot permanently miss notification outbox jobs after a transient enqueue failure.

Scope:
Add an app-local outbox repair pass that idempotently enqueues missing per-sink jobs from persisted alert records before delivery, and document the repair semantics.

Non-goals:
- Do not introduce external queues.
- Do not claim exactly-once delivery.
- Do not add Mongo transactions in this card.

Files:
- reference/workerfleet/internal/app/alert_loop.go
- reference/workerfleet/internal/app/alert_loop_test.go
- reference/workerfleet/docs/alerts.md
- reference/workerfleet/docs/notifiers.md

Acceptance Tests:
- reference/workerfleet/internal/app/alert_loop_test.go: TestDeliverNotificationOutboxRepairsMissingJobs

Tests:
- Existing alert loop tests.
- Verify repair remains idempotent when jobs already exist.

Docs Sync:
- reference/workerfleet/docs/alerts.md
- reference/workerfleet/docs/notifiers.md

Validation:
- cd reference/workerfleet && go test -timeout 30s ./internal/app
- cd reference/workerfleet && go vet ./internal/app
- gofmt -l reference/workerfleet/internal/app

Done Definition:
- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l . produces no output.
- [ ] Docs Sync targets updated (if applicable).

Outcome:
