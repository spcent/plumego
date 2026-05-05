# Card 0751

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P2
State: done
Primary Module: middleware
Owned Files:
- middleware/conformance/response_writer_contract_test.go
- docs/modules/middleware/README.md
Depends On:
- 0750-middleware-coalesce-header-method-normalization

Goal:
Include recovery in the documented and tested optional response-writer
interface matrix.

Scope:
- Add recovery to shared `Unwrap`, `Flush`, and `Hijack` conformance matrix.
- Add recovery row to middleware response writer compatibility docs.
- Keep recovery-specific panic behavior tests in the recovery package.

Non-goals:
- Do not change recovery response writer behavior unless the matrix exposes an
  inconsistency.
- Do not expand the matrix beyond documented stable middleware wrappers.

Files:
- middleware/conformance/response_writer_contract_test.go
- docs/modules/middleware/README.md

Tests:
- go test -timeout 20s ./middleware/conformance
- go test -timeout 20s ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Recovery matrix claims are covered by shared positive conformance tests.
- Documentation and tests agree.
- Middleware-wide tests pass.

Outcome:
- Added recovery to the shared optional response-writer interface conformance
  matrix for `Unwrap`, `Flush`, and `Hijack`.
- Added recovery to the documented response writer compatibility table.

Validation:
- go test -timeout 20s ./middleware/conformance
- go test -timeout 20s ./middleware/...
