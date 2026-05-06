# Card 0755

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P2
State: active
Primary Module: middleware
Owned Files:
- middleware/compression/gzip.go
- middleware/compression/gzip_test.go
- docs/modules/middleware/README.md
Depends On:
- 0754-middleware-ratelimit-lifecycle-contract

Goal:
Make gzip compression behavior consistent for handlers that call
`WriteHeader(200)` before writing a body.

Scope:
- Delay the compression decision when `WriteHeader` is called before a
  detectable response content type exists.
- Preserve explicit no-compression decisions for error, streaming,
  pre-encoded, and binary responses.
- Add regression coverage for `WriteHeader(200)` followed by text body without
  pre-setting `Content-Type`.
- Document the delayed decision behavior.

Non-goals:
- Do not compress error responses.
- Do not change websocket/SSE bypass behavior.
- Do not add a new compression threshold or algorithm.

Files:
- middleware/compression/gzip.go
- middleware/compression/gzip_test.go
- docs/modules/middleware/README.md

Tests:
- go test -timeout 20s ./middleware/compression
- go test -timeout 20s ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Bare `WriteHeader(200)` no longer permanently disables gzip before the first
  body write.
- Existing skip cases remain covered.
- Middleware-wide tests pass.

Outcome:
