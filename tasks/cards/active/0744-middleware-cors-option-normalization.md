# Card 0744

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P2
State: active
Primary Module: middleware
Owned Files:
- middleware/cors/cors.go
- middleware/cors/cors_test.go
- docs/modules/middleware/README.md
Depends On:
- 0743-middleware-ratelimit-blank-key

Goal:
Normalize raw CORS option inputs so common whitespace mistakes do not silently
disable configured origins, methods, or headers.

Scope:
- Trim and filter blank origins, methods, allowed headers, and exposed headers
  during CORS option defaulting.
- Preserve wildcard behavior and explicit fall-through behavior.
- Add tests for trimmed origin/method/header matching.

Non-goals:
- Do not change zero-value wildcard default.
- Do not synthesize CORS denial responses.
- Do not add route-aware CORS policy.

Files:
- middleware/cors/cors.go
- middleware/cors/cors_test.go
- docs/modules/middleware/README.md

Tests:
- go test -timeout 20s ./middleware/cors
- go test -timeout 20s ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Raw CORS options are normalized consistently.
- Strict helper behavior remains unchanged.
- Targeted and middleware-wide tests pass.

Outcome:

