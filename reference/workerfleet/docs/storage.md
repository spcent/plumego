# Workerfleet Storage

The initial workerfleet storage layer is app-local and in-memory.

Current state:

- `worker_snapshot` equivalent state is stored as a per-worker snapshot.
- `worker_active_task` equivalent state is stored as a per-worker active-task set plus a task index.

History and retention:

- task history records are retained for seven days
- worker event records are retained for seven days
- alert records are retained for seven days
- current worker snapshots and active-task sets are not pruned by retention

This layout is intentionally app-local to `reference/workerfleet` and does not widen Plumego stable store packages.
