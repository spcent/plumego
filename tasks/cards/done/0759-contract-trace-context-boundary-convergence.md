# Card 0759

Priority: P2
State: done
Primary Module: contract
Owned Files:
- `contract/trace.go`
- `contract/observability_policy.go`
- `contract/errors.go`
- `log/context.go`
- `middleware/requestid/request_id.go`
- `middleware/tracing/tracing.go`
- `middleware/accesslog/accesslog.go`
- `middleware/internal/observability/helpers.go`
- `contract/trace_test.go`

Goal:
- Remove duplicate trace/request-id storage conventions so transport code has
  one canonical source of trace identity in request context.

Problem:
- `contract` stores structured trace data through `TraceContext` and exposes
  `TraceIDFromContext` / `WithTraceIDString`.
- `log` also keeps a separate trace-id-only context slot through
  `log.WithTraceID` / `log.TraceIDFromContext`.
- Request-id and tracing middleware currently write both representations, and
  helper code reads from both, which means the same logical identifier is stored
  twice with different contracts.
- This creates avoidable drift and makes trace propagation rules harder to
  reason about.

Scope:
- Choose one canonical request-context trace source.
- Reduce cross-package bridging to a single boundary adaptation instead of
  double-writing trace ids through middleware stacks.
- Keep structured trace context support in `contract` explicit and stable.

Non-goals:
- Do not add full tracing infrastructure to `contract`.
- Do not redesign W3C propagation behavior in this card.
- Do not change request-id header names or generation policy.

Files:
- `contract/trace.go`
- `contract/observability_policy.go`
- `contract/errors.go`
- `log/context.go`
- `middleware/requestid/request_id.go`
- `middleware/tracing/tracing.go`
- `middleware/accesslog/accesslog.go`
- `middleware/internal/observability/helpers.go`
- `contract/trace_test.go`

Tests:
- Add or update tests so middleware and logging helpers read trace identity from
  the single canonical source.
- `go test -race -timeout 60s ./contract/...`
- `go test -timeout 20s ./log/... ./middleware/requestid/... ./middleware/tracing/... ./middleware/accesslog/...`

Docs Sync:
- Update docs if the canonical trace/request-id context contract changes for
  middleware authors.

Done Definition:
- Request-context trace identity has one canonical storage contract.
- Middleware no longer double-writes trace ids into separate context systems.
- Logging helpers consume the same canonical trace source.

Outcome:
- `contract.TraceContext` is now the only request-context trace source.
- Removed `log.WithTraceID` / `log.TraceIDFromContext` and deleted their tests.
- Logging helpers now read trace identity from `contract.TraceContext`.
- Request-id, tracing, and access-log middleware no longer double-write trace ids.
- Removed `contract -> log` imports by dropping `Ctx.Logger` and simplifying `AdaptCtxHandler(...)`.
