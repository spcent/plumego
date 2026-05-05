# Card 0735

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P2
State: done
Primary Module: middleware
Owned Files:
- middleware/compression/gzip.go
- middleware/compression/gzip_test.go
- middleware/conformance/response_writer_contract_test.go
- docs/modules/middleware/README.md
Depends On:
- 0734-middleware-concurrency-context-cancel

Goal:
Make gzip behavior deterministic when a handler flushes before writing a body.

Scope:
- Define early `Flush()` as committing headers and forcing pass-through for
  later writes.
- Prevent gzip from attempting to add compression headers after an early flush.
- Add focused package and shared conformance tests.

Non-goals:
- Do not rewrite gzip buffering.
- Do not add streaming compression for unknown content types.
- Do not change normal gzip response behavior.

Files:
- middleware/compression/gzip.go
- middleware/compression/gzip_test.go
- middleware/conformance/response_writer_contract_test.go
- docs/modules/middleware/README.md

Tests:
- go test -timeout 20s ./middleware/compression
- go test -timeout 20s ./middleware/conformance
- go test -timeout 20s ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- `Flush()` before first `Write()` produces a valid uncompressed response.
- Existing gzip compression and panic finalization tests still pass.
- Shared conformance includes the edge case.

Outcome:
- Defined early `Flush()` as committing headers and forcing later writes through
  an uncompressed pass-through path.
- Added package and shared conformance coverage for `Flush()` before first
  `Write()`.
- Documented the gzip early-flush contract.

Validation:
- `go test -timeout 20s ./middleware/compression`
- `go test -timeout 20s ./middleware/conformance`
- `go test -timeout 20s ./middleware/...`
