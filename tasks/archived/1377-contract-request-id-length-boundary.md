# Card 1377

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: done
Primary Module: contract
Owned Files:
- contract/request_id.go
- contract/request_id_test.go
- docs/modules/contract/README.md
Depends On:
- 0760

Goal:
Prevent oversized request ids from being echoed into contract JSON responses.

Scope:
- Add a contract-level maximum accepted request id length.
- Apply it in `WithRequestID` and `ErrorBuilder.RequestID` through existing normalization.
- Add tests for accepted boundary, rejection, and response echo behavior.
- Document that generation policy remains middleware-owned.

Non-goals:
- Do not introduce request id generation in `contract`.
- Do not change header name or middleware ownership.
- Do not add dependencies.

Files:
- contract/request_id.go
- contract/request_id_test.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- Update request id stable semantics.

Done Definition:
- Oversized request ids are not stored or echoed by contract helpers.
- Middleware ownership remains documented.
- Target checks pass.

Outcome:
- Confirmed contract request ids are capped at 128 bytes after trimming and
  reject empty, control-character, and oversized values.
- Confirmed oversized request ids are not echoed by success or error response
  helpers.
- No runtime change was required; the current implementation and tests already
  satisfy the card.

Validation:
- go test -timeout 20s ./contract/...
- go vet ./contract/...
