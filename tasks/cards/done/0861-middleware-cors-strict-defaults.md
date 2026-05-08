# Card 0861

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P1
State: done
Primary Module: middleware
Owned Files:
- middleware/cors/cors.go
- middleware/cors/cors_test.go
- docs/modules/middleware/README.md
- docs/stable-api/snapshots/middleware-head.snapshot
Depends On:
- 0728-middleware-coalesce-stability-contract

Goal:
Provide an explicit strict CORS default option path while preserving existing
zero-value compatibility.

Scope:
- Keep zero-value `CORSOptions` wildcard behavior for compatibility.
- Add a named strict default option constructor for production callers.
- Document that production stacks should configure explicit origins or start
  from strict defaults.
- Add regression tests for strict defaults and existing wildcard compatibility.

Non-goals:
- Do not change zero-value behavior in this card.
- Do not add origin pattern matching.
- Do not synthesize explicit CORS denial responses.

Files:
- middleware/cors/cors.go
- middleware/cors/cors_test.go
- docs/modules/middleware/README.md
- docs/stable-api/snapshots/middleware-head.snapshot

Tests:
- go test -timeout 20s ./middleware/cors
- go test -timeout 20s ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md
- docs/stable-api/snapshots/middleware-head.snapshot

Done Definition:
- Strict default CORS options are available by name.
- Compatibility wildcard defaults remain tested.
- CORS package and middleware-wide tests pass.

Outcome:
- Added `cors.StrictDefaultOptions(origins...)` as a production-oriented default
  option path requiring explicit origins.
- Preserved zero-value wildcard origin compatibility.
- Documented strict defaults and explicit-origin guidance.
- Updated the middleware stable API snapshot for the new exported function.
- Added regression coverage for strict allowed/disallowed origins and missing
  origin panic behavior.

Validation:
- `go test -timeout 20s ./middleware/cors`
- `go test -timeout 20s ./middleware/...`
