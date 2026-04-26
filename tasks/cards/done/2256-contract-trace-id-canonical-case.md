# Card 2256

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: contract
Owned Files:
- contract/trace.go
- contract/trace_test.go
Depends On: 2255

Goal:
Canonicalize parsed trace and span identifiers to lowercase hexadecimal.

Scope:
- Accept uppercase hex input while returning lowercase `TraceID` and `SpanID`.
- Keep length, hex, and all-zero validation unchanged.
- Add focused coverage for uppercase inputs.

Non-goals:
- Do not add tracing infrastructure.
- Do not parse or inject HTTP trace headers.
- Do not change `TraceContext` storage behavior.

Files:
- `contract/trace.go`
- `contract/trace_test.go`

Tests:
- `go test -timeout 20s ./contract/...`
- `go vet ./contract/...`

Docs Sync:
- None expected; this normalizes parser output.

Done Definition:
- Parsed IDs have stable lowercase representation.
- Existing trace validation tests continue to pass.

Outcome:
- `ParseTraceID` and `ParseSpanID` now return lowercase canonical IDs.
- Added coverage for uppercase trace and span inputs.
