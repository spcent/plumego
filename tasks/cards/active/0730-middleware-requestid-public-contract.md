# Card 0730

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P2
State: active
Primary Module: middleware
Owned Files:
- middleware/requestid/generator.go
- middleware/requestid/requestid.go
- middleware/requestid/requestid_test.go
- docs/modules/middleware/README.md
Depends On:
- 0729-middleware-cors-strict-defaults

Goal:
Make request-id security semantics explicit and continuously tested.

Scope:
- Document that request IDs are correlation identifiers, not secrets, tokens, or
  authorization material.
- Document that the default generated ID encodes timestamp components and can be
  decoded with `DecodeRequestID`.
- Add tests that show generated IDs are decodable and preserve the correlation
  contract.
- Keep package-local default generator behavior compatible.

Non-goals:
- Do not remove `DecodeRequestID`.
- Do not change the ID format.
- Do not replace the package default generator.

Files:
- middleware/requestid/generator.go
- middleware/requestid/requestid.go
- middleware/requestid/requestid_test.go
- docs/modules/middleware/README.md

Tests:
- go test -timeout 20s ./middleware/requestid
- go test -timeout 20s ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Security semantics are visible in package and module docs.
- Tests cover default ID decodability and request correlation use.
- Requestid package and middleware-wide tests pass.

Outcome:
