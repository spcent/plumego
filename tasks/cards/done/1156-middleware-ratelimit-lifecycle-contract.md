# Card 1156

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P1
State: done
Primary Module: middleware
Owned Files:
- middleware/ratelimit/abuse_guard.go
- middleware/ratelimit/abuse_guard_test.go
- docs/modules/middleware/README.md
Depends On:
- 0753-middleware-coalesce-unreplayable-flush-hijack

Goal:
Make the ratelimit lifecycle contract unambiguous so production callers use
the managed constructor when middleware owns limiter resources.

Scope:
- Update package and constructor comments so `NewAbuseGuard` is the production
  lifecycle entrypoint.
- Mark `AbuseGuard` as the compatibility convenience constructor and document
  that it cannot be stopped when it creates a limiter internally.
- Add or adjust tests that exercise `NewAbuseGuard(...).Middleware()` as the
  production path.
- Keep existing public symbols compatible.

Non-goals:
- Do not remove or rename `AbuseGuard`.
- Do not change limiter algorithm behavior.
- Do not add application lifecycle abstractions to middleware.

Files:
- middleware/ratelimit/abuse_guard.go
- middleware/ratelimit/abuse_guard_test.go
- docs/modules/middleware/README.md

Tests:
- go test -timeout 20s ./middleware/ratelimit
- go test -timeout 20s ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md

Done Definition:
- Public docs and code comments no longer call `AbuseGuard` the production
  canonical lifecycle entrypoint.
- `NewAbuseGuard` production usage is covered by tests and docs.
- Middleware-wide tests pass.

Outcome:
- Updated package and constructor comments so `NewAbuseGuard` is the production
  lifecycle entrypoint when middleware owns limiter resources.
- Clarified that `AbuseGuard` remains a compatibility constructor and cannot
  stop an internally created limiter.
- Added production-path coverage for `NewAbuseGuard(...).Middleware()` blocking
  and idempotent `Stop`.
- Updated middleware docs to match the lifecycle contract.

Validation:
- go test -timeout 20s ./middleware/ratelimit
- go test -timeout 20s ./middleware/...
