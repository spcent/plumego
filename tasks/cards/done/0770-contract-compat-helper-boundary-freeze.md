# Card 0770

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P2
State: done
Primary Module: contract
Owned Files:
- contract/context_bind.go
- contract/validation.go
- contract/trace.go
- docs/modules/contract/README.md
Depends On:
- 0769

Goal:
Freeze the compatibility-helper boundaries for `BindJSON`, `BindQuery`, `ValidateStruct`, and `TraceContext` without expanding `contract` ownership.

Scope:
- Clarify code comments for `BindJSON` body retention, `BindQuery` reflection limits, `ValidateStruct` scope, and `TraceContext.Valid` requirements.
- Document that these helpers are compatibility surface, not preferred expansion points.
- Add or adjust focused tests only if comments reveal an unguarded behavior.

Non-goals:
- Do not rename `EnableBodyCache`.
- Do not add validator rules or trace propagation policy.
- Do not migrate external callers in this card.

Files:
- contract/context_bind.go
- contract/validation.go
- contract/trace.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- Update compatibility-helper boundary notes.

Done Definition:
- Compatibility-helper limits are clear at the code and docs level.
- No new validation, tracing, or binding framework ownership is added to `contract`.
- Target checks pass.

Outcome:
- Clarified `BindJSON` body retention semantics, including the exact meaning of `EnableBodyCache=false`.
- Clarified `BindQuery` reflection support limits and non-support for nested/map request graphs.
- Clarified `ValidateStruct` rule scope, zero-value `required`, and unsupported `min`/`max` no-op compatibility behavior.
- Clarified that `TraceContext` readers must call `Valid()` before treating a carrier as propagation-ready.
- Added a contract docs section for compatibility-helper boundaries.

Validation:
- go test -timeout 60s ./contract/...
- go vet ./contract/...
