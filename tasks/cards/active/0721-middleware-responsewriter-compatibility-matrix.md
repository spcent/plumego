# Card 0721

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P2
State: active
Primary Module: middleware
Owned Files:
- docs/modules/middleware/README.md
- middleware/internal/transport/http.go
- middleware/internal/transport/http_test.go
- internal/httputil/http_response.go
- internal/httputil/http_response_test.go
Depends On:
- 0720-middleware-cors-and-bodylimit-contract

Goal:
Make response writer compatibility and header-copy semantics discoverable and consistent.

Scope:
- Add a middleware compatibility matrix for buffering, streaming, flush, hijack, and websocket/SSE support.
- Clarify header-copy semantics between middleware transport helpers and internal httputil helpers.
- Add or update tests that pin replace-vs-append header copy behavior.

Non-goals:
- Do not rewrite every response writer wrapper.
- Do not add new public interfaces.
- Do not change behavior without a focused regression test.

Files:
- docs/modules/middleware/README.md
- middleware/internal/transport/http.go
- middleware/internal/transport/http_test.go
- internal/httputil/http_response.go
- internal/httputil/http_response_test.go

Tests:
- go test ./middleware/internal/transport ./internal/httputil
- go test ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Stable users can identify which middleware are compatible with streaming and optional response writer interfaces.
- Header-copy behavior is tested and documented.
- Middleware tests pass.

Outcome:

