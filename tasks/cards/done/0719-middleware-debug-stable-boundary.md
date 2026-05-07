# Card 0719

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P1
State: done
Primary Module: middleware
Owned Files:
- middleware/debug/debug_errors.go
- middleware/debug/debug_errors_test.go
- docs/modules/middleware/README.md
Depends On:
- 0718-middleware-constructor-default-convergence

Goal:
Make debug middleware safe and explicit enough for stable package inclusion.

Scope:
- Add a maximum captured response body limit with a conservative default.
- Preserve the existing preview truncation behavior.
- Add tests for cap enforcement, streaming skip behavior, and production guidance.
- Clarify docs that this middleware is development-only unless explicitly configured.

Non-goals:
- Do not make debug middleware production-default.
- Do not support websocket, hijack, or streaming capture.
- Do not change the error response envelope.

Files:
- middleware/debug/debug_errors.go
- middleware/debug/debug_errors_test.go
- docs/modules/middleware/README.md

Tests:
- go test ./middleware/debug
- go test ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Debug capture is bounded.
- Streaming and production boundaries are documented.
- Focused and middleware-wide tests pass.

Outcome:
- Added bounded debug response capture through `DebugErrorConfig.MaxBodyBytes`.
- Changed debug capture overflow to pass through the original response instead of truncating it.
- Preserved response preview truncation when replacement is still safe.
- Documented development-only guidance and streaming/capture boundaries.
- Validation:
  - `go test ./middleware/debug`
  - `go test ./middleware/...`
