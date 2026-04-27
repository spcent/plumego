# Card 0527

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/trace.go
- contract/trace_test.go
Depends On: 2236

Goal:
Reject all-zero trace and span identifiers so malformed tracing carriers fail closed.

Scope:
- Make `ParseTraceID` reject the all-zero 16-byte trace id.
- Make `ParseSpanID` reject the all-zero 8-byte span id.
- Add tests for parse helpers and validity helpers.

Non-goals:
- Do not add tracing infrastructure to `contract`.
- Do not implement traceparent header parsing.
- Do not change trace carrier type names.

Files:
- `contract/trace.go`
- `contract/trace_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this hardens existing trace ID validation.

Done Definition:
- All-zero trace and span ids fail parsing and `IsValid*` checks.
- Existing tracing tests continue to pass.

Outcome:
- Rejected all-zero trace and span identifiers in parse helpers.
- Covered all-zero IDs through both parse and validity helper tests.
- Validation run: `go test -timeout 20s ./contract/...`; `go vet ./contract/...`.
