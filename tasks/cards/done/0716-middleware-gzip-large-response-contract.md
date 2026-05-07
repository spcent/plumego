# Card 0716

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P0
State: done
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
- Clarified gzip docs and comments so `MaxBufferBytes` is scoped to the pre-compression buffer.
- Added regression coverage for large eligible compressed responses, pre-compression bypass, and post-start large writes.
- Fixed gzip writer behavior so responses cannot switch back to uncompressed output after gzip headers have been flushed.
- Promoted compressed content-type prefixes to a package-level table.
- Validation:
  - `go test ./middleware/compression`
  - `go test ./middleware/...`
