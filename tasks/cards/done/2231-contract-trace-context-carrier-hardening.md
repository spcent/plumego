# Card 2231

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: contract
Owned Files:
- contract/trace.go
- contract/trace_test.go
Depends On: 2230

Goal:
Keep trace context carrier values deterministic and immutable from caller-owned map or pointer mutation.

Scope:
- Copy baggage maps when storing and returning `TraceContext`.
- Keep `TraceContextFromContext` signature compatible while returning a defensive copy.
- Make invalid span-id updates nil-context safe.
- Add regression tests for mutation isolation and nil-context behavior.

Non-goals:
- Do not add tracing infrastructure to `contract`.
- Do not implement HTTP baggage extraction or injection.
- Do not change exported trace carrier type names.

Files:
- `contract/trace.go`
- `contract/trace_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this preserves the documented carrier boundary.

Done Definition:
- Caller mutation of the original or returned baggage map cannot mutate the context-stored carrier.
- `WithSpanIDString(nil, invalid)` returns a non-nil context.
- Existing tracing context tests continue to pass.

Outcome:
- Added defensive copying for stored and returned `TraceContext` values, including baggage and parent span id.
- Made invalid span-id updates on nil contexts return a non-nil base context without setting malformed trace state.
- Validation run: `go test -timeout 20s ./contract/...`; `go vet ./contract/...`.
