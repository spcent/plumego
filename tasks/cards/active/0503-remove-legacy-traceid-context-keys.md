# Card 0503

Priority: P0

Goal:
- Remove the two legacy context-key types for TraceID/RequestContext and
  consolidate all trace propagation on the single modern unexported key.

Problem:
- Three parallel mechanisms exist for storing/reading the trace ID in context:
  1. `TraceIDKey{}` (exported, old) → value is `string`
  2. `TraceContextKey{}` (exported, old) → value is `RequestContext`
  3. `traceContextKeyVar` (unexported, new) → value is `*TraceContext`
- `TraceIDFromContext()` and `RequestContextFrom()` contain three-level
  fallback chains to bridge all three, making them fragile and hard to follow.
- Callers in middleware packages use different keys inconsistently.

Scope:
- Identify every setter and getter for `TraceIDKey` and `TraceContextKey`
  across the codebase.
- Migrate them to write/read via the unexported `traceContextKey` path
  (`TraceContextFromContext` / `WithTraceContext`).
- Remove exported `TraceIDKey` and `TraceContextKey` types once no call sites
  remain.
- Simplify `TraceIDFromContext()` to a single lookup with no fallback.
- Update `RequestContextFrom()` to delegate to `TraceContextFromContext()`
  rather than maintain its own key chain.

Non-goals:
- Do not change `TraceContext` or `Span` data structures.
- Do not touch `trace.go` sampling or collection logic.

Files:
- `contract/context_core.go`
- `contract/trace.go`
- `contract/observability_policy.go`
- `middleware/tracing/` (likely setter)
- `middleware/accesslog/` (likely getter)
- Any other middleware using `TraceIDKey` or `TraceContextKey`

Tests:
- `go test ./contract/... ./middleware/...`
- `go vet ./...`
- `go build ./...`

Done Definition:
- `TraceIDKey` and `TraceContextKey` exported types are removed.
- `TraceIDFromContext()` has no fallback chain; one lookup only.
- All middleware reads and writes trace info through `TraceContextFromContext` /
  `WithTraceContext`.
- All tests pass.
