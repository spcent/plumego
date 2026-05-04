# Card 0725

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
- 0724-middleware-coalesce-hook-count-contract

Goal:
Normalize wildcard CORS requested-header handling to match the non-wildcard path.

Scope:
- Trim and validate `Access-Control-Request-Headers` before echoing them when `AllowedHeaders` contains `*`.
- Preserve caller-visible header names after trimming.
- Keep blank requested-header lists as fall-through.
- Add regression tests for whitespace, blank lists, and wildcard echo behavior.

Non-goals:
- Do not change CORS origin defaults.
- Do not add explicit CORS denial responses.
- Do not add origin pattern matching.

Files:
- middleware/cors/cors.go
- middleware/cors/cors_test.go
- docs/modules/middleware/README.md

Tests:
- go test ./middleware/cors
- go test ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md if wildcard header behavior wording changes.

Done Definition:
- Wildcard requested headers use the same trimming/blank validation behavior as explicit allowed headers.
- CORS package and middleware-wide tests pass.

Outcome:

