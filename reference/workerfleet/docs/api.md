# Workerfleet API

This document defines the MVP HTTP contract for the workerfleet monitoring service.

Base path: `/v1`

## `POST /v1/workers/register`

Registers or refreshes worker identity.

Request body:

```json
{
  "worker_id": "worker-1",
  "namespace": "sim",
  "pod_name": "worker-1",
  "pod_uid": "pod-uid-1",
  "node_name": "node-a",
  "container_name": "worker",
  "image": "worker:1.0.0",
  "version": "1.0.0",
  "observed_at": "2026-04-19T10:00:00Z"
}
```

Success response:

```json
{
  "data": {
    "worker_id": "worker-1",
    "status": "unknown",
    "registered_at": "2026-04-19T10:00:00Z"
  }
}
```

## `POST /v1/workers/heartbeat`

Reports worker liveness, readiness, and the full active-task set.

Request body:

```json
{
  "worker_id": "worker-1",
  "process_alive": true,
  "accepting_tasks": false,
  "observed_at": "2026-04-19T10:01:00Z",
  "last_error": "",
  "active_tasks": [
    {
      "task_id": "task-1",
      "task_type": "simulation",
      "phase": "running",
      "phase_name": "running"
    },
    {
      "task_id": "task-2",
      "task_type": "simulation",
      "phase": "preparing",
      "phase_name": "warming"
    }
  ]
}
```

Notes:

- `active_tasks` is a full-set replacement, not a delta stream.
- `process_alive` and `accepting_tasks` are distinct.
- `accepting_tasks=false` with active tasks means the worker is busy, not necessarily offline.

Success response:

```json
{
  "data": {
    "worker_id": "worker-1",
    "status": "online",
    "status_reason": "busy",
    "observed_at": "2026-04-19T10:01:00Z",
    "active_task_count": 2
  }
}
```

## `GET /v1/workers`

Lists workers with optional filters.

Query parameters:

- `status`
- `namespace`
- `node_name`
- `task_type`
- `accepting_tasks`
- `page`
- `page_size`

Success response:

```json
{
  "data": {
    "items": [
      {
        "worker_id": "worker-1",
        "namespace": "sim",
        "pod_name": "worker-1",
        "node_name": "node-a",
        "status": "online",
        "status_reason": "busy",
        "process_alive": true,
        "accepting_tasks": false,
        "last_seen_at": "2026-04-19T10:01:00Z",
        "active_task_count": 2,
        "active_tasks": [
          {
            "task_id": "task-1",
            "task_type": "simulation",
            "phase": "running",
            "phase_name": "running"
          }
        ]
      }
    ],
    "page": 1,
    "page_size": 50,
    "total": 1
  }
}
```

## `GET /v1/workers/{worker_id}`

Returns one worker with current active tasks and readiness state.

## `GET /v1/tasks/{task_id}`

Returns the current or most recent task view for a task ID.

Example success payload:

```json
{
  "data": {
    "task_id": "task-1",
    "worker_id": "worker-1",
    "task_type": "simulation",
    "phase": "running",
    "phase_name": "running",
    "status": "active",
    "started_at": "2026-04-19T10:00:30Z",
    "updated_at": "2026-04-19T10:01:00Z"
  }
}
```

## `GET /v1/fleet/summary`

Returns fleet totals for workers and active tasks.

## `GET /v1/alerts`

Lists alert records.

Query parameters:

- `worker_id`
- `alert_type`
- `status`
- `page`
- `page_size`

## Error model

Errors use the standard `contract.WriteError` envelope.

- Invalid JSON: `INVALID_JSON`
- Invalid query parameter: `INVALID_QUERY`
- Missing required field: `REQUIRED_FIELD_MISSING`
- Resource not found: `RESOURCE_NOT_FOUND`
