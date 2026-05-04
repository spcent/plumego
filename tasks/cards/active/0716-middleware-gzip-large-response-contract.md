# Card 0716

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P0
State: active
Primary Module: middleware
Owned Files:
- middleware/compression/gzip.go
- middleware/compression/gzip_test.go
- docs/modules/middleware/README.md
Depends On:
- 0715-middleware-timeout-buffer-contract

Goal:
Align gzip large-response behavior, documentation, and tests.

Scope:
- Clarify that MaxBufferBytes only applies before compression has started.
- Add tests for large eligible responses and large responses bypassed before gzip starts.
- Avoid per-request repeated compressed type allocation where practical.
- Ensure gzip close/write error handling is as deterministic as net/http permits.

Non-goals:
- Do not redesign streaming compression.
- Do not add third-party compression packages.
- Do not change the gzip public constructor.

Files:
- middleware/compression/gzip.go
- middleware/compression/gzip_test.go
- docs/modules/middleware/README.md

Tests:
- go test ./middleware/compression
- go test ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Gzip docs no longer promise behavior the code does not provide.
- Large response edge cases have regression tests.
- Middleware tests pass.

Outcome:

