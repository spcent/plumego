# Card 0817

Milestone: contract cleanup
Priority: P2
State: done
Primary Module: contract
Owned Files:
- `contract/context_core.go`
- `contract/context_test.go`
Depends On: —

Goal:
- Replace the `Ctx.TraceID string` snapshot field with a `Ctx.TraceID() string`
  method that reads the live value from the request context, eliminating the
  risk of stale data.

Problem:
`Ctx` captures the trace ID as a plain string field at construction time:

```go
// context_core.go
traceID := TraceIDFromContext(r.Context())
ctx := &Ctx{
    ...
    TraceID: traceID,   // snapshot — could go stale
    ...
}
```

The trace ID can be updated in the request context after `Ctx` is created — for
example, when a propagation middleware runs after the handler is registered.
Callers who access `ctx.TraceID` (the field) will read the construction-time
snapshot. Callers who use `TraceIDFromContext(ctx.R.Context())` will read the
current value.

Current field callers:

- `x/webhook/in.go` (2 places) — logs `"trace_id": ctx.TraceID`
- `contract/context_test.go` — tests the field

Both would silently log or assert a stale trace ID if the context was updated
after `NewCtx` returned.

Fix: convert `TraceID string` to a method:

```go
// Remove the field from the struct.
// Add:
func (c *Ctx) TraceID() string {
    if c == nil || c.R == nil {
        return ""
    }
    return TraceIDFromContext(c.R.Context())
}
```

Callers change from `ctx.TraceID` → `ctx.TraceID()`. The method always returns
the current value from the request context.

Scope:
- Remove `TraceID string` from the `Ctx` struct.
- Remove the `traceID` local variable and assignment in `newCtxWithConfig`.
- Add `func (c *Ctx) TraceID() string` to `context_core.go`.
- Update `x/webhook/in.go` (2 sites).
- Update `contract/context_test.go`.

Non-goals:
- No change to `TraceIDFromContext`.
- No change to `TraceContext.TraceID` (field of a different struct).
- No change to how the trace context is stored or propagated.

Files:
- `contract/context_core.go`
- `contract/context_test.go`
- `x/webhook/in.go`

Tests:
- `go build ./...`
- `go test -timeout 20s ./contract/... ./x/webhook/...`
- `go vet ./...`

Docs Sync: —

Done Definition:
- `Ctx.TraceID` is a method, not a field.
- `ctx.TraceID()` returns `TraceIDFromContext(ctx.R.Context())`.
- No stale snapshot is stored in the struct.
- All tests pass.

Outcome:
