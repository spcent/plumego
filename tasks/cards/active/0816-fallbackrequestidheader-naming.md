# Card 0816

Milestone: contract cleanup
Priority: P3
State: active
Primary Module: contract
Owned Files:
- `contract/observability_policy.go`
Depends On: —

Goal:
- Rename `FallbackRequestIDHeader` to `LegacyTraceIDHeader` to accurately
  describe what the constant is rather than what it is used for.

Problem:
`observability_policy.go` exports two header constants:

```go
RequestIDHeader         = "X-Request-ID"   // canonical request-id header
FallbackRequestIDHeader = "X-Trace-ID"     // legacy fallback
```

The name `FallbackRequestIDHeader` says "fallback request ID" but the value
`"X-Trace-ID"` is a trace ID header — a different concept. The comment says it
is a "legacy request id header", which explains the intent but not the name.

The mismatch matters because:
1. A reader who sees `FallbackRequestIDHeader` expects its value to look like a
   request-ID header (similar format to `"X-Request-ID"`), not a trace header.
2. `RequestIDFromRequest` in the same file reads this header as a fallback for
   resolving a request ID from incoming traffic. The fact that a trace-ID header
   is used for this purpose is a legacy quirk that the name should communicate
   directly rather than hide behind "FallbackRequestID".

Proposed rename: `LegacyTraceIDHeader`
- "Legacy" signals that this is kept for backwards compatibility, not a first-class
  API choice.
- "TraceID" accurately names the concept the header carries.

Scope:
- Rename the constant in `contract/observability_policy.go`.
- Update the two callers in `middleware/requestid/request_id.go`.
- The string value `"X-Trace-ID"` does NOT change.

Non-goals:
- No change to `RequestIDHeader`.
- No change to the `RequestIDFromRequest` resolution logic.
- No removal of the fallback behaviour.

Files:
- `contract/observability_policy.go`
- `middleware/requestid/request_id.go`

Tests:
- `go build ./...`
- `go test -timeout 20s ./contract/... ./middleware/...`
- After rename, `grep -rn 'FallbackRequestIDHeader' . --include='*.go'` must
  return empty.

Docs Sync: —

Done Definition:
- `FallbackRequestIDHeader` does not exist.
- `LegacyTraceIDHeader` is the constant name for `"X-Trace-ID"`.
- All tests pass.

Outcome:
