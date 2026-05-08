# Card 0748

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P2
State: done
Primary Module: middleware
Owned Files:
- middleware/cors/cors.go
- middleware/cors/cors_test.go
- docs/modules/middleware/README.md
- docs/stable-api/snapshots/middleware-head.snapshot
Depends On:
- 0747-middleware-ratelimit-lifecycle

Goal:
Make `cors.StrictDefaultOptions` actually reject wildcard origin
configuration.

Scope:
- Reject `"*"` from `StrictDefaultOptionsE`.
- Preserve `StrictDefaultOptions` panic-on-invalid behavior.
- Add tests for wildcard rejection and existing valid strict origins.
- Update middleware docs.

Non-goals:
- Do not change zero-value `cors.Middleware(CORSOptions{})` wildcard default.
- Do not synthesize denial responses for disallowed CORS requests.

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

Done Definition:
- Strict helper rejects blank-only and wildcard-only origin lists.
- Tests cover E and panic variants.
- Middleware-wide tests pass.

Outcome:
- Added `ErrStrictDefaultWildcardOrigin` and made
  `StrictDefaultOptionsE("*")` reject wildcard origin configuration.
- Preserved `StrictDefaultOptions(...)` panic behavior for invalid strict
  origins.
- Added wildcard rejection tests for both strict helper variants.
- Updated CORS docs and the middleware API snapshot.

Validation:
- go test -timeout 20s ./middleware/cors
- go test -timeout 20s ./middleware/...
