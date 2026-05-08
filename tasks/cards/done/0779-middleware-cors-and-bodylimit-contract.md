# Card 0779

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P1
State: done
Primary Module: middleware
Owned Files:
- middleware/cors/cors.go
- middleware/cors/cors_test.go
- middleware/bodylimit/body_limit.go
- middleware/bodylimit/body_limit_test.go
- docs/modules/middleware/README.md
Depends On:
- 0719-middleware-debug-stable-boundary

Goal:
Lock down the less obvious CORS and body-limit stable contracts.

Scope:
- Document CORS fall-through behavior for disallowed origin, method, and headers.
- Add regression tests for wildcard credentials, disallowed preflight, and blank request headers.
- Document bodylimit terminal error behavior and post-overrun write handling.
- Add regression tests for downstream writes after an overrun response.

Non-goals:
- Do not change CORS default permissiveness in this card.
- Do not replace bodylimit with http.MaxBytesReader.
- Do not introduce policy engines.

Files:
- middleware/cors/cors.go
- middleware/cors/cors_test.go
- middleware/bodylimit/body_limit.go
- middleware/bodylimit/body_limit_test.go
- docs/modules/middleware/README.md

Tests:
- go test ./middleware/cors ./middleware/bodylimit
- go test ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- CORS fall-through and credentials behavior are documented and tested.
- Bodylimit overrun behavior is documented and tested.
- Middleware tests pass.

Outcome:
- Documented CORS fall-through behavior for disallowed origins, methods, headers, and blank requested-header lists.
- Documented wildcard-origin credential behavior.
- Documented bodylimit terminal overrun behavior and suppressed downstream writes.
- Added CORS regression tests for blank requested headers and credential wildcard headers.
- Added bodylimit regression coverage for post-overrun writes reporting consumed bytes without mutating the 413 response.
- Validation:
  - `go test ./middleware/cors ./middleware/bodylimit`
  - `go test ./middleware/...`
