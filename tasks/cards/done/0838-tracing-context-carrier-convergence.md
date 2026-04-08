# Card 0838

Priority: P1
State: active
Primary Module: contract
Owned Files:
- `contract/trace.go`
- `middleware/tracing/tracing.go`
- `middleware/accesslog/accesslog.go`
- `x/observability/otel.go`
- `x/observability/tracer/tracer.go`

Goal:
- Converge tracing context carriage onto one canonical context carrier across stable middleware and observability extensions.
- Remove duplicate trace-context storage paths so tracing middleware no longer has to translate between parallel context formats.

Problem:
- Stable `contract` already owns `TraceContext`, `WithTraceContext`, and `TraceContextFromContext` as the canonical tracing context primitive.
- `x/observability/tracer` uses `contract.TraceContext` directly, but `x/observability/otel.go` maintains a separate private `internalTraceContext` with its own context key.
- `middleware/tracing` and `middleware/accesslog.Logging` therefore have to reconstruct `contract.TraceContext` after `tracer.Start(...)`, instead of reading one canonical carrier directly.
- This leaves tracing identity split across two context paths and duplicates span-id / trace-id hydration logic in middleware.

Scope:
- Pick one canonical trace context carrier and remove the losing carrier path.
- Update `x/observability/otel` to participate in the same trace-context contract used by stable middleware and `x/observability/tracer`.
- Remove duplicated trace-context reconstruction logic from middleware once a single carrier exists.
- Keep request correlation (`request_id`) distinct from tracing identity.
- Update tests across contract, middleware, and observability packages in the same change.

Non-goals:
- Do not redesign tracer sampling, span collection, or exporter behavior.
- Do not collapse `request_id` into `trace_id`.
- Do not move tracing infrastructure into stable `contract`; only the minimal context carrier belongs there.
- Do not preserve duplicate carrier paths for compatibility.

Files:
- `contract/trace.go`
- `middleware/tracing/tracing.go`
- `middleware/accesslog/accesslog.go`
- `x/observability/otel.go`
- `x/observability/tracer/tracer.go`

Tests:
- `go test -timeout 20s ./contract/... ./middleware/tracing ./middleware/accesslog ./x/observability/...`
- `go test -race -timeout 60s ./contract/... ./middleware/tracing ./middleware/accesslog ./x/observability/...`
- `go vet ./contract/... ./middleware/tracing ./middleware/accesslog ./x/observability/...`

Docs Sync:
- Keep tracing docs and comments aligned on the rule that `contract` owns one minimal trace-context carrier and observability implementations must use it instead of maintaining parallel context formats.

Done Definition:
- There is one canonical trace-context carrier across stable middleware and observability extensions.
- `middleware/tracing` and `middleware/accesslog` no longer rehydrate `contract.TraceContext` from a parallel carrier path.
- Observability implementations compile and test against the converged carrier with no residual duplicate context storage path.
- Code comments and manifests describe the same converged tracing-context ownership.

Outcome:
- Completed.
- Converged tracing context carriage on `contract.TraceContext` as the single canonical context carrier across middleware and observability implementations.
- Removed the duplicate trace-context storage path so middleware reads one carrier directly instead of rehydrating trace/span identity from parallel context state.
