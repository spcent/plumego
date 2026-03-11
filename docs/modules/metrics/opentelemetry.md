# OpenTelemetry Tracing

> **Package**: `github.com/spcent/plumego/metrics` | **Integration**: lightweight tracing hooks

## Overview

The current `metrics` package exposes tracing via `metrics.NewOpenTelemetryTracer(name)`. This is a lightweight tracer implementation used by Plumego's observability middleware and core observability component.

This is not an OTLP exporter package and there is no `metrics/opentelemetry.New(...)` constructor in the current codebase.

## Basic Wiring

```go
tracer := metrics.NewOpenTelemetryTracer("plumego-api")
app := core.New(core.WithTracer(tracer))
```

You can then use middleware or components that read the app tracer.

## Direct Usage

```go
ctx, span := tracer.Start(ctx, req)
defer span.End(200, nil)
```

## Introspection

```go
stats := tracer.GetSpanStats()
spans := tracer.Spans()
```

These APIs are useful for tests, local debugging, and lightweight in-process tracing.

## When to Use It

Use `NewOpenTelemetryTracer(...)` when you need:

- request trace IDs and span timing
- middleware-level tracing hooks
- stdlib-only tracing without adding external exporters

If you need full OTLP export pipelines, that should live outside the main stdlib-only module boundary.
