# Card 0758

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P2
State: active
Primary Module: middleware
Owned Files:
- reference/production-service/internal/app/app.go
- reference/production-service/internal/app/app_test.go
- docs/modules/middleware/README.md
Depends On:
- 0757-middleware-header-copy-semantics

Goal:
Guard the recommended production observability wiring so access logging,
metrics, and tracing are not accidentally double-wired.

Scope:
- Add reference-level test coverage for the recommended middleware composition:
  accesslog is logging-only, while metrics/tracing use standalone middleware.
- Keep application wiring explicit and reviewable.
- Update docs if the reference test makes the recommendation more concrete.

Non-goals:
- Do not add a new stable observability bundle API.
- Do not move exporter setup into stable middleware.
- Do not change metrics or tracing primitives.

Files:
- reference/production-service/internal/app/app.go
- reference/production-service/internal/app/app_test.go
- docs/modules/middleware/README.md

Tests:
- go test -timeout 20s ./reference/production-service/...
- go test -timeout 20s ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Reference app has a regression guard for non-duplicated observability wiring.
- Middleware docs point to explicit standalone composition.
- Reference and middleware tests pass.

Outcome:
