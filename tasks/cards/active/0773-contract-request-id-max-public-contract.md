# Card 0773

Milestone: v1
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P1
State: active
Primary Module: contract
Owned Files:
- contract/request_id.go
- contract/request_id_test.go
- docs/modules/contract/README.md
Depends On:
- 0772

Goal:
Make the request-id length cap an explicit public contract.

Scope:
- Document the exact 128-byte request-id acceptance limit in contract docs.
- Keep the limit implementation stable and covered by tests.
- Avoid adding new public API unless the existing docs/tests cannot express the contract clearly.

Non-goals:
- Do not add request-id generation policy to `contract`.
- Do not change the accepted character set.
- Do not change middleware ownership.

Files:
- contract/request_id.go
- contract/request_id_test.go
- docs/modules/contract/README.md

Tests:
- go test -timeout 20s ./contract/...
- go vet ./contract/...

Docs Sync:
- Update request-id stable semantics.

Done Definition:
- Stable docs say the exact request-id limit is 128 bytes after trimming.
- Tests still cover boundary acceptance and rejection.
- Target checks pass.

Outcome:

