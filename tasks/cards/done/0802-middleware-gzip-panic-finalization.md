# Card 0802

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P1
State: done
Primary Module: middleware
Owned Files:
- middleware/compression/gzip.go
- middleware/compression/gzip_test.go
Depends On:
- 0722-middleware-timeout-panic-propagation

Goal:
Finalize gzip responses deterministically during panic unwinding.

Scope:
- Ensure gzip writer close/finalization runs in a defer after wrapping downstream handlers.
- Preserve panic propagation to outer recovery or the server.
- Add regression tests for panic before gzip starts and after gzip starts.

Non-goals:
- Do not recover or rewrite panic responses in gzip middleware.
- Do not redesign compression buffering.
- Do not change public API.

Files:
- middleware/compression/gzip.go
- middleware/compression/gzip_test.go

Tests:
- go test ./middleware/compression
- go test ./middleware/...

Docs Sync:
- None unless compatibility wording needs clarification.

Done Definition:
- Gzip finalization runs on normal return and panic return.
- Panic still propagates.
- Compression package and middleware-wide tests pass.

Outcome:
- Moved gzip finalization into a defer that preserves panic propagation.
- Avoided flushing an unstarted buffered response during panic unwinding so outer recovery can still write the canonical error.
- Finalizes already-started gzip streams during panic unwinding.
- Added regression tests for panic before compression starts and after compression starts.
- Validation:
  - `go test ./middleware/compression`
  - `go test ./middleware/...`
