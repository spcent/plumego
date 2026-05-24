# Card 2051

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: implementation
Priority: P0
State: done
Primary Module: reference/workerfleet
Owned Files:
- reference/workerfleet/internal/platform/store/mongo/notification_outbox_store.go
- reference/workerfleet/internal/platform/store/mongo/notification_outbox_store_test.go
Depends On: 2050

Goal:
Make Mongo notification outbox claims compare-and-swap on the exact due condition so expired processing jobs cannot be double-claimed.

Scope:
Tighten Mongo claim filters for pending and expired-processing jobs and add a deterministic unit-style backend test or integration-safe test covering stale claim races.

Non-goals:
- Do not add a new queue backend.
- Do not add distributed transactions.
- Do not change notification job public fields.

Files:
- reference/workerfleet/internal/platform/store/mongo/notification_outbox_store.go
- reference/workerfleet/internal/platform/store/mongo/notification_outbox_store_test.go

Acceptance Tests:
- reference/workerfleet/internal/platform/store/mongo/notification_outbox_store_test.go: TestNotificationOutboxClaimUsesDueConditionCAS

Tests:
- Existing Mongo outbox tests.

Docs Sync:

Validation:
- cd reference/workerfleet && go test -timeout 30s ./internal/platform/store/mongo
- cd reference/workerfleet && go vet ./internal/platform/store/mongo
- gofmt -l reference/workerfleet/internal/platform/store/mongo

Done Definition:
- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

Outcome:
- Tightened Mongo notification job claim updates to include the exact due condition for both pending and expired processing jobs.
- Added `TestNotificationOutboxClaimUsesDueConditionCAS` to lock the CAS filter shape without requiring a live MongoDB.
- Validation:
  - `cd reference/workerfleet && go test -timeout 30s ./internal/platform/store/mongo`
  - `cd reference/workerfleet && go vet ./internal/platform/store/mongo`
  - `gofmt -l reference/workerfleet/internal/platform/store/mongo`
