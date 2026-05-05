# Card 0745

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P2
State: active
Primary Module: middleware
Owned Files:
- middleware/conformance/response_writer_contract_test.go
- docs/modules/middleware/README.md
Depends On:
- 0744-middleware-cors-option-normalization

Goal:
Make the shared response-writer conformance suite cover the documented optional
interface matrix.

Scope:
- Add shared tests for `Unwrap`, `Flush`, and `Hijack` exposure for matrix
  packages where applicable.
- Cover both positive and negative matrix claims.
- Keep package-specific behavior tests in their owning packages.

Non-goals:
- Do not expose internal transport helpers publicly.
- Do not rewrite response writer implementations.
- Do not change documented streaming guidance except to align with tests.

Files:
- middleware/conformance/response_writer_contract_test.go
- docs/modules/middleware/README.md

Tests:
- go test -timeout 20s ./middleware/conformance
- go test -timeout 20s ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Matrix claims have shared conformance coverage.
- Existing response writer edge-case tests still pass.
- Middleware-wide tests pass.

Outcome:

