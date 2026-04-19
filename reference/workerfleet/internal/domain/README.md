# Workerfleet Domain

`reference/workerfleet/internal/domain` owns the app-local worker monitoring model.

This package defines:

- worker identity and runtime state
- pod state merged from Kubernetes
- active task state for a single worker
- worker status evaluation rules
- domain events for worker, task, and pod transitions
- alert type constants used by later cards

Key rules:

- `active_tasks` is a full-set replacement from the worker, not a patch stream.
- transport handlers and storage adapters must reuse these status strings and event names.
- online/offline state is not derived from heartbeat freshness alone.
- a worker with active tasks and `accepting_tasks=false` is still online, but busy.
- a worker with `accepting_tasks=false` and no active tasks is degraded.
