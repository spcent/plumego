# Card 0370

Milestone: —
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P1
State: done
Primary Module: reference/workerfleet/internal/platform/kube
Owned Files:
- `reference/workerfleet/internal/platform/kube/discovery.go`
- `reference/workerfleet/internal/platform/kube/watch.go`
- `reference/workerfleet/internal/platform/kube/mapper.go`
- `reference/workerfleet/internal/domain/pod_reconcile.go`
- `reference/workerfleet/internal/platform/kube/watch_test.go`
Depends On:
- `tasks/cards/done/0366-workerfleet-domain-status-model.md`
- `tasks/cards/done/0368-workerfleet-store-and-retention.md`
- `tasks/cards/done/0369-workerfleet-ingest-and-reconcile.md`

Goal:
- Add Kubernetes inventory sync for a single cluster so pod lifecycle state and worker-reported state can be reconciled.

Scope:
- Namespace/label-based pod discovery and watch.
- Mapping from pod/container state into the workerfleet domain `PodSnapshot`.
- Pod disappearance, restart, and non-running state reconciliation against worker snapshots.

Non-goals:
- Do not implement multi-cluster support.
- Do not add business policy into Plumego stable modules.
- Do not infer current tasks from Kubernetes metadata.

Files:
- `reference/workerfleet/internal/platform/kube/discovery.go`
- `reference/workerfleet/internal/platform/kube/watch.go`
- `reference/workerfleet/internal/platform/kube/mapper.go`
- `reference/workerfleet/internal/domain/pod_reconcile.go`
- `reference/workerfleet/internal/platform/kube/watch_test.go`

Tests:
- `go test ./reference/workerfleet/internal/platform/kube/...`
- Mapping tests for pod phases, restart count changes, deletion handling, and pod-without-worker cases.

Docs Sync:
- Document required namespace, label-selector, and RBAC assumptions in `reference/workerfleet/README.md`.

Done Definition:
- The service can maintain pod-level truth for the monitored single cluster.
- Pod state is merged into worker snapshots through a domain reconciliation path.
- Pod restart and disappearance events are available to later alert rules.

Outcome:
- Added a stdlib-only Kubernetes client for pod listing and watch streaming without introducing `client-go`.
- Added pod-to-worker mapping, single-cluster inventory sync, and pod reconciliation helpers.
- Added focused tests for pod mapping, list/watch behavior, and `SyncOnce` snapshot updates.
- Documented the single-cluster and RBAC assumptions in `reference/workerfleet/README.md`.
- Validation run: `go test ./reference/workerfleet/internal/platform/kube/... ./reference/workerfleet/internal/...`
