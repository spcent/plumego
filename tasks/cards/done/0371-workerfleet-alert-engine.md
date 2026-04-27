# Card 0371

Milestone: —
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P1
State: done
Primary Module: reference/workerfleet/internal/domain
Owned Files:
- `reference/workerfleet/internal/domain/alert_rules.go`
- `reference/workerfleet/internal/domain/alert_engine.go`
- `reference/workerfleet/internal/domain/alert_dedupe.go`
- `reference/workerfleet/internal/domain/alert_engine_test.go`
- `reference/workerfleet/docs/alerts.md`
Depends On:
- `tasks/cards/done/0366-workerfleet-domain-status-model.md`
- `tasks/cards/done/0368-workerfleet-store-and-retention.md`
- `tasks/cards/done/0369-workerfleet-ingest-and-reconcile.md`
- `tasks/cards/done/0370-workerfleet-kube-inventory-sync.md`

Goal:
- Implement the first alert rules for worker offline, worker degraded, worker not accepting tasks, stage stuck, pod restart burst, pod missing, and task conflict.

Scope:
- Alert rule evaluation from current worker snapshots, active tasks, pod state, and recent history.
- Dedupe-key generation and alert state transitions for firing and resolved events.
- Alert persistence model shared with later notifier delivery.

Non-goals:
- Do not send notifications yet.
- Do not add business-specific alert routing policies.
- Do not make alert rules configurable beyond the agreed MVP thresholds.

Files:
- `reference/workerfleet/internal/domain/alert_rules.go`
- `reference/workerfleet/internal/domain/alert_engine.go`
- `reference/workerfleet/internal/domain/alert_dedupe.go`
- `reference/workerfleet/internal/domain/alert_engine_test.go`
- `reference/workerfleet/docs/alerts.md`

Tests:
- `go test ./reference/workerfleet/internal/domain/...`
- Coverage for fire, dedupe, and resolve behavior across offline, degraded, stuck, pod restart burst, and conflict conditions.

Docs Sync:
- Document MVP alert types, thresholds, and dedupe keys in `reference/workerfleet/docs/alerts.md`.

Done Definition:
- Alert rules evaluate from persisted domain state, not from ad hoc handler conditions.
- Each MVP alert type has deterministic fire and resolve behavior with tests.
- Alert records are ready for a separate notifier card to consume.

Outcome:
- Added domain alert evaluation, dedupe, and resolve logic for the MVP workerfleet alert set.
- Added alert record ownership to the domain model and aligned the in-memory store with current snapshot and alert queries.
- Added focused tests for firing, dedupe, recovery, and task-conflict alerts.
- Documented alert types, state model, and dedupe keys in `reference/workerfleet/docs/alerts.md`.
- Validation run: `go test ./reference/workerfleet/internal/domain/... ./reference/workerfleet/internal/...`
