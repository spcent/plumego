# Card 0731

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P2
State: active
Primary Module: middleware
Owned Files:
- middleware/conformance/response_writer_contract_test.go
- middleware/conformance/helpers_test.go
- docs/modules/middleware/README.md
Depends On:
- 0730-middleware-requestid-public-contract

Goal:
Add a shared response-writer conformance suite for high-risk middleware wrapper
behavior.

Scope:
- Add conformance tests for panic propagation, flush forwarding, hijack boundary,
  post-timeout write behavior, and partial response handling where applicable.
- Reuse existing middleware packages rather than adding new production wrapper
  abstractions.
- Keep the compatibility matrix aligned with tested behavior.

Non-goals:
- Do not rewrite every response writer wrapper.
- Do not expose internal transport test helpers publicly.
- Do not add non-standard-library dependencies.

Files:
- middleware/conformance/response_writer_contract_test.go
- middleware/conformance/helpers_test.go
- docs/modules/middleware/README.md

Tests:
- go test -timeout 20s ./middleware/conformance
- go test -timeout 20s ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Shared conformance tests cover the listed response-writer edge cases.
- Middleware-wide tests pass.
- Matrix wording remains consistent with tested behavior.

Outcome:
