# tracing — Distributed Tracing Middleware

Propagates trace/span context through HTTP requests.

## Import

```go
import "github.com/spcent/plumego/middleware/tracing"
```

## Overview

`tracing.Middleware` injects span context into outgoing requests and extracts parent spans from incoming headers. It delegates the actual trace transport (OpenTelemetry, Zipkin, Jaeger, etc.) to a caller-provided `Tracer`.

## Exports

| Symbol | Description |
|--------|-------------|
| `Middleware(tracer Tracer)` | Constructor |
| `Tracer` | Interface: `Start(ctx, name) context.Context` and `Finish(ctx)` |
| `TraceSpan` | Minimal span handle abstraction |

## Usage

```go
app.Use(tracing.Middleware(myOTelTracer))
```

## Notes

- This is a thin transport adapter. Trace sampling, export, and backend wiring belong in `x/observability`.
- Standard propagation headers (`traceparent`, `x-b3-*`) are read/written by the `Tracer` implementation, not this middleware.
- Place tracing early in the middleware chain to capture request latency from the first byte.
