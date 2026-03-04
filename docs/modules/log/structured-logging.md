# Structured Logging

> **Package**: `github.com/spcent/plumego/log` | **Best Practices**

## Overview

Use `log.StructuredLogger` + `log.Fields` to emit stable, machine-readable events.

Benefits:
- Queryable fields (`status`, `duration_ms`, `tenant_id`)
- Consistent correlation (`trace_id`, `request_id`)
- Backend-agnostic API (`gLogger` and `JSONLogger` both satisfy the interface)

## Recommended Field Conventions

Use `snake_case` consistently.

Common fields:
- `trace_id`
- `request_id`
- `tenant_id`
- `method`
- `path`
- `status`
- `duration_ms`
- `error`

## Core Patterns

### 1) Request-scoped logger

```go
func handle(logger log.StructuredLogger, w http.ResponseWriter, r *http.Request) {
    traceID := contract.TraceIDFromContext(r.Context())

    reqLogger := logger.WithFields(log.Fields{
        "trace_id":   traceID,
        "request_id": traceID,
        "method":     r.Method,
        "path":       r.URL.Path,
    })

    reqLogger.InfoCtx(r.Context(), "request started", nil)

    // business logic...

    reqLogger.InfoCtx(r.Context(), "request completed", log.Fields{
        "status":      200,
        "duration_ms": 12,
    })
}
```

### 2) Error logging with context

```go
if err != nil {
    logger.ErrorCtx(ctx, "database query failed", log.Fields{
        "error":      err.Error(),
        "query_name": "ListUsers",
        "tenant_id":  tenantID,
    })
}
```

### 3) Reuse common fields via `WithFields`

```go
base := logger.WithFields(log.Fields{
    "service": "billing",
    "module":  "invoice",
})

base.Info("job started", log.Fields{"job_id": jobID})
base.Warn("retry scheduled", log.Fields{"retry_in_ms": 500})
```

## Trace Correlation

For HTTP apps, prefer middleware wiring:
- `observability.RequestID()` injects request ID into context and headers
- `observability.Logging(...)` emits request completion logs with trace metadata

For non-HTTP flows, add trace metadata explicitly:

```go
ctx := log.WithTraceID(context.Background(), log.NewTraceID())
logger.InfoCtx(ctx, "async task accepted", log.Fields{"task": "reindex"})
```

## Performance Notes

- Use `WithFields` for stable dimensions (service/module/tenant), and pass per-event fields only when needed.
- Avoid constructing expensive debug payloads unless `log.V(1)` is true.
- Prefer bounded payload fields (IDs, counts, durations) over raw large blobs.
