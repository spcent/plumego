# Card 0749

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/freeze_test.go
- docs/modules/contract/README.md
Depends On:
- 0748

Goal:
Freeze `WriteResponse` and `WriteJSON` status semantics so callers understand the boundary between raw JSON writing, success envelopes, and error envelopes.

Scope:
- Add regression coverage for valid caller-selected non-2xx statuses.
- Document that `WriteResponse` writes the success envelope for any valid HTTP status, while errors must use `WriteError`.
- Preserve existing behavior for health/readiness handlers that choose non-2xx statuses with structured bodies.

Non-goals:
- Do not restrict `WriteResponse` to 2xx.
- Do not change response envelope shape.
- Do not migrate existing health handlers.

Files:
- contract/freeze_test.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- Update frozen behavior and stable semantics docs.

Done Definition:
- Status handling is covered for valid non-2xx response writes.
- Docs explain that `WriteError` remains the canonical error path.
- Contract tests and vet pass.

Outcome:
- Added freeze coverage showing `WriteResponse` preserves valid non-2xx statuses with the success envelope.
- Added freeze coverage showing `WriteJSON` preserves valid redirect statuses for raw JSON payloads.
- Documented that non-2xx structured health/readiness bodies may use `WriteResponse`, while error payloads must use `WriteError`.

Validation:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
