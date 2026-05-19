# Card 1050

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P2
State: done
Primary Module: middleware
Owned Files:
- middleware/conformance/response_writer_contract_test.go
- docs/modules/middleware/README.md
- docs/stable-api/snapshots/middleware-head.snapshot
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
- docs/stable-api/snapshots/middleware-head.snapshot

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
- Added a shared optional-interface matrix conformance test covering `Unwrap`,
  `Flush`, and `Hijack` exposure for accesslog, bodylimit, coalesce,
  compression, debug, httpmetrics, timeout, and tracing.
- Documented that each matrix claim is backed by a positive or negative shared
  conformance case.
- Refreshed the middleware stable API snapshot for the observability internal
  finalizer helpers added by this hardening sequence.

Validation:
- go test -timeout 20s ./middleware/conformance
- go test -timeout 20s ./middleware/...
