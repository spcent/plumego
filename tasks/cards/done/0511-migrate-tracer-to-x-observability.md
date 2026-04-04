# Card 0511

Priority: P1

Goal:
- Move the full Tracer/Span/Collector/Sampler subsystem out of the stable
  `contract` package and into `x/observability`, so `contract` contains only
  the lightweight transport primitives that belong there.

Problem:
- `contract/trace.go` is 976 lines and includes: `Tracer`, `Span`, `Trace`,
  `TraceCollector`, `TraceSampler`, span lifecycle management, batch collection,
  and sampling strategies.
- This is a fully-featured distributed tracing library living in a stable
  transport-primitive package. It violates the boundary rule added to
  `specs/dependency-rules.yaml` and `CANONICAL_STYLE_GUIDE.md §16.1`.
- `contract` should export only the context keys and ID accessors needed by
  middleware: `TraceID`, `SpanID`, `TraceContext`, `WithTraceIDString`,
  `TraceIDFromContext`, `ContextWithTraceContext`, `TraceContextFromContext`.

Scope:
- Create `x/observability/tracer/` (or extend existing `x/observability`)
  to host `Tracer`, `Span`, `TraceCollector`, `TraceSampler`, and related types.
- Keep in `contract/trace.go` (stripped):
  - `TraceID`, `SpanID` types
  - `TraceContext` struct
  - `WithTraceIDString`, `TraceIDFromContext`
  - `WithTraceContext`, `TraceContextFromContext` (canonical With/From names)
- Update all callers of the moved types to import from `x/observability/tracer`.
- Delete the full instrumentation code from `contract/trace.go`.

Non-goals:
- Do not change the HTTP wire representation of trace IDs.
- Do not change `middleware/tracing` or `middleware/accesslog` context-key usage.

Files:
- `contract/trace.go` (strip to primitives)
- `x/observability/tracer/` (new or extended)
- Any file importing the moved types

Tests:
- `go test ./contract/...`
- `go test ./x/observability/...`
- `go vet ./...`

Done Definition:
- `contract/trace.go` ≤ 80 lines, contains only primitive types and accessors.
- All Tracer/Span logic lives in `x/observability`.
- All tests pass.
