# Card 0735

Priority: P3

Goal:
- Either implement `TraceContext.Baggage` propagation or remove the field and
  document that baggage is out of scope for this package.

Problem:

`trace.go:35`:
```go
type TraceContext struct {
    TraceID      TraceID           `json:"trace_id"`
    SpanID       SpanID            `json:"span_id"`
    ParentSpanID *SpanID           `json:"parent_span_id,omitempty"`
    Baggage      map[string]string `json:"baggage,omitempty"`  // ← never set
    Flags        TraceFlags        `json:"flags"`
    Sampled      bool              `json:"sampled"`
}
```

`Baggage` is part of the W3C Trace Context / OpenTelemetry specification: it
carries key-value pairs that propagate alongside a trace across service
boundaries. However, no code in the `contract` package reads or writes the
`Baggage` field:

- `WithTraceIDString` (trace.go:70) only sets `TraceID`.
- `WithTraceContext` stores the full struct, but callers only populate
  `TraceID` and `SpanID`.
- No middleware in the package extracts `Baggage` from incoming HTTP headers
  (W3C `baggage:` header) or injects it into outgoing requests.

The field exists as a placeholder that cannot be used. A caller who sets
`TraceContext{Baggage: map[string]string{"tenant": "acme"}}` and expects it
to be propagated to downstream services will find it silently dropped.

Two options:

**Option A: Remove `Baggage` from `TraceContext`**
If baggage propagation belongs in `x/observability/tracer` (per the comment
on line 30: "Full tracing infrastructure … lives in x/observability/tracer"),
remove the field here and document that baggage is out of scope for the
`contract` package's minimal trace context.

**Option B: Document as intentionally unimplemented**
Add a comment:
```go
// Baggage carries W3C baggage key-value pairs.
// Extraction from HTTP headers and injection into outgoing requests is not
// implemented in this package; see x/observability/tracer for full support.
Baggage map[string]string `json:"baggage,omitempty"`
```

Option A is cleaner (no dead fields), but Option B preserves forward
compatibility if baggage propagation is added later.

Non-goals:
- Do not implement W3C baggage header parsing in this card.
- Do not change `x/observability/tracer`.

Files:
- `contract/trace.go`

Tests:
- `go build ./...`
- If Option A: verify `grep -rn 'Baggage' . --include='*.go'` has no remaining
  references outside the test itself.

Done Definition:
- `Baggage` is either removed (Option A) or has a doc comment explaining its
  unimplemented status (Option B).
- All tests pass.

Outcome:
- Completed in the 2026-04-05 contract cleanup batch.
- Verified as part of the shared contract/task-card completion pass.

Validation Run:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/reference-layout`
- `go build ./...`
- `go test -timeout 20s ./...`
- `go test -race -timeout 60s ./...`
- `go vet ./...`
