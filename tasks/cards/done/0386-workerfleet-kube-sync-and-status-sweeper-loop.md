# Card 0386

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P0
State: done
Primary Module: reference/workerfleet/internal/app
Owned Files:
- `reference/workerfleet/internal/app/config.go`
- `reference/workerfleet/internal/app/bootstrap.go`
- `reference/workerfleet/internal/app/runtime_loops.go`
- `reference/workerfleet/internal/app/runtime_loops_test.go`
- `reference/workerfleet/main.go`
Depends On:
- `tasks/cards/done/0385-workerfleet-runtime-health-and-config.md`
Blocked By: —

Goal:
- Wire Kubernetes inventory sync and worker status sweeping into the running workerfleet process.
- Make pod phase, pod disappearance, restart count, and heartbeat expiry visible without requiring new worker heartbeats.

Scope:
- Add optional Kubernetes inventory sync loop with interval, namespace, label selector, worker container, and bearer token config.
- Add status sweeper loop that periodically re-evaluates persisted worker snapshots against `StatusPolicy`.
- Record loop duration and errors through existing metrics where available.
- Stop loops cleanly on context cancellation during shutdown.

Non-goals:
- Do not add alert evaluation or notifier delivery loops.
- Do not implement Kubernetes deployment manifests.
- Do not change Mongo persistence semantics.

Files:
- `reference/workerfleet/internal/app/config.go`
- `reference/workerfleet/internal/app/bootstrap.go`
- `reference/workerfleet/internal/app/runtime_loops.go`
- `reference/workerfleet/internal/app/runtime_loops_test.go`
- `reference/workerfleet/main.go`

Tests:
- `cd reference/workerfleet && go test ./internal/app/...`
- `cd reference/workerfleet && go test ./internal/platform/kube/...`
- `cd reference/workerfleet && go test ./...`

Docs Sync:
- Update `reference/workerfleet/README.md` with Kubernetes sync and status sweep configuration.

Done Definition:
- Kube sync loop can be enabled or disabled by config.
- Status sweeper updates stale workers even when no heartbeat arrives.
- Loops stop on shutdown without goroutine leaks in tests.

Outcome:
- Added optional Kubernetes inventory sync runtime loop.
- Added status sweeper loop and deterministic sweep method.
- Added Kubernetes runtime config fields.
- Updated README with Kubernetes sync configuration.
