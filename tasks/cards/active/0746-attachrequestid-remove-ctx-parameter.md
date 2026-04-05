# Card 0746

Priority: P2
State: active
Primary Module: contract
Owned Files: contract/observability_policy.go, middleware/requestid/request_id.go

Goal:
- Remove the `ctx context.Context` parameter from `AttachRequestID`, derive
  from `r.Context()` internally, and update the only call site.
  This eliminates a footgun where passing a context not derived from
  `r.Context()` silently drops all existing request-scoped values.

Problem:

`observability_policy.go:68-81`:
```go
func (p ObservabilityPolicy) AttachRequestID(
    ctx context.Context,           // ← caller-supplied context
    w http.ResponseWriter,
    r *http.Request,
    id string,
    includeInRequest bool,
) *http.Request {
    if r == nil {
        return r
    }
    ctx = WithTraceIDString(ctx, id)   // rewrites whatever was passed
    ...
    return r.WithContext(ctx)           // replaces r.Context() entirely
}
```

The function accepts a separate `ctx` argument and returns `r.WithContext(ctx)`.
If the caller passes anything other than a context derived from `r.Context()`,
every value stored in the original request context (authenticated principal,
middleware values, log fields, etc.) is silently discarded. There is no guard,
no panic, and no documentation warning.

In the only call site (`middleware/requestid/request_id.go:71-73`):
```go
ctx := contract.WithTraceIDString(r.Context(), id)   // ← adds contract trace ID
ctx = log.WithTraceID(ctx, id)                        // ← adds log trace ID
r = contract.DefaultObservabilityPolicy.AttachRequestID(ctx, w, r, id, ...)
```

`AttachRequestID` then calls `WithTraceIDString(ctx, id)` internally — the
contract trace ID is applied **twice** to the same context. The result is
identical (same ID, same key), but the duplication means the function's
internal `WithTraceIDString` call is redundant and confusing.

Fix:
1. Remove the `ctx` parameter from `AttachRequestID`. Derive from `r.Context()`
   internally so the function is safe to call unconditionally:
   ```go
   func (p ObservabilityPolicy) AttachRequestID(
       w http.ResponseWriter,
       r *http.Request,
       id string,
       includeInRequest bool,
   ) *http.Request {
       if r == nil {
           return r
       }
       ctx := WithTraceIDString(r.Context(), id)
       ...
       return r.WithContext(ctx)
   }
   ```

2. Update `middleware/requestid/request_id.go` to call the new signature and
   add the log trace ID to the request AFTER `AttachRequestID` returns:
   ```go
   r = contract.DefaultObservabilityPolicy.AttachRequestID(w, r, id, cfg.includeInRequest)
   // add log trace ID to the already-updated request context
   r = r.WithContext(log.WithTraceID(r.Context(), id))
   ```

Non-goals:
- Do not change `WithTraceIDString`, `RequestIDFromRequest`, or
  `MiddlewareLogFields`.
- Do not change the semantics of request ID attachment (headers, response
  header, context key).

Files:
- `contract/observability_policy.go`
- `middleware/requestid/request_id.go`

Tests:
- Verify with `grep -rn "AttachRequestID" . --include="*.go"` that no other
  call site exists.
- Verify the existing `request_id` middleware tests still pass.
- `go test ./contract/...`
- `go test ./middleware/requestid/...`
- `go vet ./...`

Done Definition:
- `AttachRequestID` no longer accepts a `ctx` parameter.
- The middleware test suite passes without changes to test assertions.
- `grep -rn "AttachRequestID"` finds only the definition and one updated call site.
- All tests pass.
