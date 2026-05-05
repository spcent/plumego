# Card 0751

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: active
Primary Module: contract
Owned Files:
- contract/trace_test.go
- docs/modules/contract/README.md
Depends On:
- 0750

Goal:
Make trace carrier validity expectations explicit for callers that read `TraceContext` from context.

Scope:
- Add tests showing `TraceContextFromContext` can return invalid caller-provided carrier data and callers must use `Valid()`.
- Document the recommended read pattern.
- Preserve the no-policy carrier behavior.

Non-goals:
- Do not validate or reject trace context in `WithTraceContext`.
- Do not add propagation, sampling, or baggage size policy.
- Do not move observability behavior into `contract`.

Files:
- contract/trace_test.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- Update trace context guidance.

Done Definition:
- Caller validity checks are covered and documented.
- Contract tests and vet pass.

Outcome:

