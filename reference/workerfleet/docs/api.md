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
      "exec_plan_id": "plan-20260419-001",
      "task_type": "simulation",
      "phase": "running",
      "phase_name": "running",
      "current_step": {
        "step": "simulate",
        "step_name": "simulation",
        "status": "running",
        "started_at": "2026-04-19T10:00:30Z",
        "updated_at": "2026-04-19T10:01:00Z",
        "attempt": 1,
        "error_class": ""
      }
    },
    {
      "task_id": "task-2",
      "exec_plan_id": "plan-20260419-001",
      "task_type": "simulation",
      "phase": "preparing",
      "phase_name": "warming",
      "current_step": {
        "step": "download_bundle",
        "step_name": "download bundle",
        "status": "running",
        "started_at": "2026-04-19T10:00:55Z",
        "updated_at": "2026-04-19T10:01:00Z",
        "attempt": 1
      }
    }
  ]
}
```

Notes:

- `active_tasks` is a full-set replacement, not a delta stream.
- `task_id` identifies one case. `exec_plan_id` groups cases under one simulation execution plan.
- `current_step` reports the case step currently being processed. Omit it for legacy workers that cannot report step state.
- `current_step.step` should come from a controlled worker-side enum such as `prepare_env`, `download_bundle`, `download_data`, `convert_data`, `simulate`, `upload_result`, or `cleanup_env`.
- `current_step.status` should be one of `running`, `succeeded`, `failed`, `canceled`, or `skipped`; unknown values are normalized to `unknown`.
- `current_step.error_class` must be low-cardinality, for example `object_store_timeout` or `dependency_missing`, not a raw exception message.
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
