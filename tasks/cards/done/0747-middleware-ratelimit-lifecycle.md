# Card 0747

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P1
State: done
Primary Module: middleware
Owned Files:
- middleware/ratelimit/abuse_guard.go
- middleware/ratelimit/abuse_guard_test.go
- reference/production-service/internal/app/app.go
- docs/modules/middleware/README.md
- docs/stable-api/snapshots/middleware-head.snapshot

Goal:
Make abuse guard limiter ownership and cleanup explicit enough for stable
production use.

Scope:
- Provide a lifecycle-aware ratelimit constructor or type that exposes Stop for
  middleware-owned limiters.
- Keep the existing `AbuseGuard(AbuseGuardConfig)` API compatible.
- Update the production reference to own and stop the limiter explicitly.
- Document lifecycle ownership rules for injected and middleware-created
  limiters.

Non-goals:
- Do not change `security/abuse.Limiter` internals unless strictly required.
- Do not remove existing `AbuseGuard` compatibility entrypoint.
- Do not introduce app lifecycle service locator behavior.

Files:
- middleware/ratelimit/abuse_guard.go
- middleware/ratelimit/abuse_guard_test.go
- reference/production-service/internal/app/app.go
- docs/modules/middleware/README.md
- docs/stable-api/snapshots/middleware-head.snapshot

Tests:
- go test -timeout 20s ./middleware/ratelimit
- go test -timeout 20s ./reference/production-service/...
- go test -timeout 20s ./middleware/...

Docs Sync:
- docs/modules/middleware/README.md
- reference/production-service/README.md if behavior text is affected

Done Definition:
- Production wiring has an explicit limiter Stop path.
- Middleware-owned limiter lifecycle is test-covered.
- Existing AbuseGuard callers remain source-compatible.

Outcome:
- Added `ratelimit.NewAbuseGuard(...)` and `AbuseGuardMiddleware` with
  `Middleware()` and idempotent `Stop()` for middleware-owned limiter cleanup.
- Kept `ratelimit.AbuseGuard(...)` source-compatible as the legacy middleware
  constructor.
- Updated the production reference to own the abuse guard and stop it when the
  service exits.
- Documented limiter ownership rules and refreshed the middleware API snapshot.

Validation:
- go test -timeout 20s ./middleware/ratelimit
- go test -timeout 20s ./reference/production-service/...
- go test -timeout 20s ./middleware/...
