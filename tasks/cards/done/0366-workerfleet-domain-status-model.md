# Card 0366

Milestone: —
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P0
State: done
Primary Module: reference/workerfleet/internal/domain
Owned Files:
- `reference/workerfleet/internal/domain/worker.go`
- `reference/workerfleet/internal/domain/task.go`
- `reference/workerfleet/internal/domain/pod.go`
- `reference/workerfleet/internal/domain/status.go`
- `reference/workerfleet/internal/domain/alert.go`
- `reference/workerfleet/internal/domain/rules.go`
- `reference/workerfleet/internal/domain/events.go`
- `reference/workerfleet/internal/domain/README.md`
Depends On:
- None

Goal:
- Define the workerfleet domain model, state machine, event taxonomy, and multi-task reconciliation semantics under `reference/workerfleet/internal/domain/`.
- Fix the canonical meaning of worker online/degraded/offline/unknown so later handlers, storage, and alerting reuse one rule set.

Scope:
- App-local domain types for worker identity, runtime, pod snapshot, active tasks, task history, alert types, and domain events.
- Pure rule functions for worker status evaluation and active-task set reconciliation.
- Documentation of the all-active-tasks payload contract and the review-approved planning assumptions.

Non-goals:
- Do not add HTTP handlers or route wiring.
- Do not add Kubernetes integration.
- Do not add storage implementations or notifier integrations.

Files:
- `reference/workerfleet/internal/domain/worker.go`
- `reference/workerfleet/internal/domain/task.go`
- `reference/workerfleet/internal/domain/pod.go`
- `reference/workerfleet/internal/domain/status.go`
- `reference/workerfleet/internal/domain/alert.go`
- `reference/workerfleet/internal/domain/rules.go`
- `reference/workerfleet/internal/domain/events.go`
- `reference/workerfleet/internal/domain/README.md`

Tests:
- `go test ./reference/workerfleet/internal/domain/...`
- Table-driven tests for status transitions: online, degraded, offline, unknown.
- Task reconciliation tests for multi-task add, phase change, removal, and conflict detection.

Docs Sync:
- Add or update `reference/workerfleet/README.md` only if the final domain contract is exposed to integrators.

Done Definition:
- Domain types, status enums, alert enums, and event names are defined in one place.
- Multi-task active-set semantics are documented as full-set replacement, not incremental patching.
- Status evaluation and task reconciliation are covered by focused tests.
- Later cards can depend on this card without redefining status strings or event names.

Outcome:
- Implemented `reference/workerfleet/internal/domain` with worker, pod, task, alert, and event types.
- Added centralized worker-status evaluation, worker-report merge, pod merge, and full active-task reconciliation helpers.
- Added focused tests for status transitions, task reconciliation, and merge behavior.
- Validation run: `go test ./reference/workerfleet/internal/domain/...`
