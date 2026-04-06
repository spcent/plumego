# Card 0809

Milestone: contract cleanup
Priority: P2
State: done
Primary Module: contract
Owned Files:
- `contract/trace.go`
Depends On: —

Goal:
- Remove the `Sampled bool` field from `TraceContext` and derive the sampled state
  solely from `Flags & TraceFlagsSampled`, eliminating the dual-write requirement.

Problem:
`TraceContext` carries the sampled bit in two places:

```go
type TraceContext struct {
    ...
    Flags   TraceFlags `json:"flags"`
    Sampled bool       `json:"sampled"`
}
```

Both fields represent the same information: `Sampled == (Flags & TraceFlagsSampled != 0)`.
Every writer must keep them in sync manually. The tracer does this explicitly:

```go
// x/observability/tracer/tracer.go:429
spanCtx.Sampled = t.sampler.ShouldSample(traceID)
if spanCtx.Sampled {
    spanCtx.Flags |= contract.TraceFlagsSampled  // must remember to also set this
}
```

If either update is forgotten, the fields silently disagree. The W3C trace context
spec uses `flags` as the canonical source; `Sampled` is a convenience duplication.

Fix:
- Remove `Sampled bool` from `TraceContext`.
- Add a method `func (tc TraceContext) IsSampled() bool` that returns
  `tc.Flags&TraceFlagsSampled != 0`.
- Update all callers that read `.Sampled` to use `.IsSampled()`.
- Update all callers that write `.Sampled = true` to set
  `tc.Flags |= TraceFlagsSampled` instead (they typically already do both).

Scope:
- Change `TraceContext` struct in `contract/trace.go`.
- Add `IsSampled()` method.
- Update `x/observability/tracer/tracer.go` (2 write sites, 2 read sites).
- Update `contract/trace_test.go` (sets `Sampled: true` in struct literal).

Non-goals:
- No change to `TraceFlags` type or `TraceFlagsSampled` constant.
- No change to `Flags` field itself.
- No change to trace propagation logic.

Files:
- `contract/trace.go`
- `contract/trace_test.go`
- `x/observability/tracer/tracer.go`

Tests:
- `go build ./...`
- `go test -timeout 20s ./contract/... ./x/observability/...`
- `go vet ./...`
- After change, `grep -rn '\.Sampled\b' . --include='*.go'` must return only
  the `IsSampled()` method body.

Docs Sync: —

Done Definition:
- `TraceContext.Sampled` field does not exist.
- `TraceContext.IsSampled()` method returns the correct value from `Flags`.
- No caller sets `.Sampled` directly.
- All tests pass.

Outcome:
