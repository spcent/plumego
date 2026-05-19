# Card 1192

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P2
State: done
Primary Module: middleware
Owned Files:
- middleware/internal/transport/http.go
- middleware/internal/transport/response_buffer.go
- middleware/internal/transport/transport_test.go
- middleware/coalesce/coalesce.go
- docs/modules/middleware/README.md
Depends On:
- 0756-middleware-debug-streaming-optional-interfaces

Goal:
Separate header overlay and full replacement semantics so buffered replay paths
do not accidentally preserve stale destination headers.

Scope:
- Clarify `CopyHeaders` as overlay semantics.
- Add a `ReplaceHeaders` helper for full destination replacement.
- Use replacement semantics for buffered/replayed responses that own the full
  response header set.
- Add internal transport tests for both overlay and replacement behavior.
- Document the two helper contracts.

Non-goals:
- Do not expose middleware/internal helpers outside the module.
- Do not change lower-level `internal/httputil` header helpers.
- Do not rewrite all local header-copy call sites unless they own a full
  buffered response.

Files:
- middleware/internal/transport/http.go
- middleware/internal/transport/response_buffer.go
- middleware/internal/transport/transport_test.go
- middleware/coalesce/coalesce.go
- docs/modules/middleware/README.md

Tests:
- go test -timeout 20s ./middleware/internal/transport ./middleware/coalesce
- go test -timeout 20s ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Overlay and replacement semantics have distinct helpers and tests.
- Buffered replay paths use replacement where they own the response headers.
- Middleware-wide tests pass.

Outcome:
- Clarified `CopyHeaders` as overlay semantics.
- Added `ReplaceHeaders` for full destination replacement.
- Switched complete buffered replay paths to `ReplaceHeaders`.
- Added internal transport tests for overlay, replacement, cloning, nil source,
  and stale destination header removal.
- Documented both helper contracts.

Validation:
- go test -timeout 20s ./middleware/internal/transport ./middleware/coalesce
- go test -timeout 20s ./middleware/...
