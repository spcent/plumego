# Card 0718

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P1
State: active
Primary Module: middleware
Owned Files:
- middleware/concurrencylimit/concurrency_limit.go
- middleware/concurrencylimit/concurrency_limit_test.go
- middleware/bodylimit/body_limit.go
- middleware/bodylimit/body_limit_test.go
- docs/modules/middleware/README.md
Depends On:
- 0717-middleware-observability-panic-path

Goal:
Make middleware constructor/default patterns clearer without breaking existing callers.

Scope:
- Add non-breaking Config-style constructors for bare-argument middleware where appropriate.
- Keep existing exported functions as compatibility entrypoints.
- Document the stable constructor convention for new middleware.
- Add tests proving old and new entrypoints produce equivalent behavior.

Non-goals:
- Do not remove or rename existing exported functions.
- Do not introduce deprecated wrappers.
- Do not change behavior beyond input normalization needed by the new config path.

Files:
- middleware/concurrencylimit/concurrency_limit.go
- middleware/concurrencylimit/concurrency_limit_test.go
- middleware/bodylimit/body_limit.go
- middleware/bodylimit/body_limit_test.go
- docs/modules/middleware/README.md

Tests:
- go test ./middleware/concurrencylimit ./middleware/bodylimit
- go test ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Existing APIs continue to compile.
- Config-style entrypoints are available for stable documentation.
- Constructor conventions are documented and tested.

Outcome:

