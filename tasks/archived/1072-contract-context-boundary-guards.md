# Card 1072

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/context_core.go
- contract/context_test.go
- contract/trace.go
- contract/trace_test.go
- docs/modules/contract/README.md
Depends On:
- 0746

Goal:
Guard `RequestContext` and `TraceContext` against expansion into policy, runtime state, or mutable request bags.

Scope:
- Add tests for RequestContext copy ownership and route-metadata-only semantics.
- Add tests for TraceContext baggage/parent-span defensive copy behavior and non-validation policy.
- Clarify boundaries in docs without adding observability or routing policy.

Non-goals:
- Do not add trace propagation, baggage limits, or sampling policy.
- Do not add auth, tenant, session, or feature fields to RequestContext.
- Do not change context key strategy.

Files:
- contract/context_core.go
- contract/context_test.go
- contract/trace.go
- contract/trace_test.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
- go run ./internal/checks/dependency-rules

Docs Sync:
- Update contract docs only if boundary wording needs tightening.

Done Definition:
- Context carrier copy semantics and boundary expectations are executable.
- No policy or runtime-state ownership moves into `contract`.
- Contract tests, vet, and dependency rules pass.

Outcome:
- Tightened `RequestContext` comments to describe router-owned metadata instead of a broad request bag.
- Added field-shape guard tests for `RequestContext` and `TraceContext`.
- Added trace carrier policy coverage showing baggage and invalid identifiers remain caller-provided carrier data, while validity is reported by helpers.
- Clarified context carrier expansion boundaries in docs.

Validation:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
- go run ./internal/checks/dependency-rules
