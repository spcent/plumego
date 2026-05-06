# Card 0756

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P2
State: done
Primary Module: middleware
Owned Files:
- middleware/debug/debug_errors.go
- middleware/debug/debug_errors_test.go
- middleware/conformance/response_writer_contract_test.go
- docs/modules/middleware/README.md
Depends On:
- 0755-middleware-gzip-delayed-writeheader-decision

Goal:
Make debug error wrapping explicit and safe around streaming and optional
response-writer interfaces.

Scope:
- Preserve pass-through behavior when a debug-wrapped handler flushes or
  hijacks the response.
- Expose/forward `Unwrap`, `Flush`, and `Hijack` consistently for the debug
  recorder when the underlying writer supports them.
- Add regression tests for flush and hijack pass-through behavior.
- Update the conformance matrix and docs.

Non-goals:
- Do not make debug middleware production-oriented.
- Do not replace structured production error handling.
- Do not capture streaming bodies after flush/hijack.

Files:
- middleware/debug/debug_errors.go
- middleware/debug/debug_errors_test.go
- middleware/conformance/response_writer_contract_test.go
- docs/modules/middleware/README.md

Tests:
- go test -timeout 20s ./middleware/debug ./middleware/conformance
- go test -timeout 20s ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Debug wrapper forwards flush/hijack without buffering stale debug output.
- Conformance matrix matches implementation.
- Middleware-wide tests pass.

Outcome:
- Added `Unwrap`, `Flush`, and `Hijack` support to the debug recorder.
- `Flush` now commits buffered content and switches debug replacement off.
- `Hijack` now delegates to the underlying writer and disables debug
  replacement.
- Added debug regression tests for flush and hijack pass-through.
- Updated the shared response-writer conformance matrix and docs.

Validation:
- go test -timeout 20s ./middleware/debug ./middleware/conformance
- go test -timeout 20s ./middleware/...
