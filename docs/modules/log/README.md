# Log Module

> **Package Path**: `github.com/spcent/plumego/log` | **Stability**: High | **Priority**: P1

## Overview

`log/` provides a standard-library-first logging layer with a stable `StructuredLogger` interface.

Key capabilities:
- Structured fields via `log.Fields`
- Context-aware logging with `trace_id` propagation
- Explicit `DEBUG/INFO/WARN/ERROR` levels
- Two built-in implementations:
  - `NewGLogger()` for glog-style text output and file rotation
  - `NewJSONLogger(...)` for JSON output
- Lifecycle hooks (`Start/Stop`) for integration with `core.Boot()`

## Core Interface

```go
type StructuredLogger interface {
    WithFields(fields Fields) StructuredLogger

    Debug(msg string, fields Fields)
    Info(msg string, fields Fields)
    Warn(msg string, fields Fields)
    Error(msg string, fields Fields)

    DebugCtx(ctx context.Context, msg string, fields Fields)
    InfoCtx(ctx context.Context, msg string, fields Fields)
    WarnCtx(ctx context.Context, msg string, fields Fields)
    ErrorCtx(ctx context.Context, msg string, fields Fields)
}
```

## Quick Start

### GLogger (default)

```go
import (
    "context"

    log "github.com/spcent/plumego/log"
)

logger := log.NewGLogger()
logger.Info("server started", log.Fields{"addr": ":8080"})

ctx := log.WithTraceID(context.Background(), log.NewTraceID())
logger.InfoCtx(ctx, "request completed", log.Fields{
    "method": "GET",
    "path":   "/health",
    "status": 200,
})
```

### JSON Logger

```go
import (
    "os"

    log "github.com/spcent/plumego/log"
)

logger := log.NewJSONLogger(log.JSONLoggerConfig{
    Output: os.Stdout,
    Level:  log.INFO,
    Fields: log.Fields{
        "service": "api-gateway",
        "env":     "prod",
    },
    // Optional: make Debug/DebugCtx follow V(1) filtering.
    RespectVerbosity: true,
    // Local verbosity threshold used by JSON logger.
    Verbosity: 1,
})

logger.Warn("downstream latency high", log.Fields{
    "provider":    "billing",
    "latency_ms":  380,
    "request_path": "/api/orders",
})
```

## Context Utilities

Use context helpers only for trace correlation metadata:

```go
ctx := log.WithTraceID(context.Background(), "trace-123")
traceID := log.TraceIDFromContext(ctx)
logger.InfoCtx(ctx, "business event", log.Fields{"trace_id": traceID})
```

## Levels and Verbosity

`Debug` logs emit at `DEBUG` level and are gated by verbosity (`V(1)` by default patterns).

```go
if log.V(1) {
    log.VLog(1, "diagnostic message")
}
```

You can configure global glog behavior without relying on process-wide flag parsing:

```go
if err := log.InitWithConfig(log.InitConfig{
    LogDir:          "/var/log/myapp",
    LogToStderr:     false,
    AlsoLogToStderr: true,
    Verbosity:       1,
    VModule:         "router=2,core=1",
}); err != nil {
    panic(err)
}
```

For request paths, prefer middleware stack:
- Register middleware explicitly via `app.Use(...)` (for example `RequestID + Logging + Recovery`)
- Request/trace IDs are propagated into request context and response headers

## Integration with Core

Use `core.WithLogger(customLogger)` to inject the app logger explicitly.

If the logger implements `log.Lifecycle`, core lifecycle will call `Start()` and `Stop()` automatically.
