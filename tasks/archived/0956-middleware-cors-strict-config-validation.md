# Card 0956

Milestone:
Recipe: specs/change-recipes/middleware.yaml
Priority: P2
State: done
Primary Module: middleware
Owned Files:
- middleware/cors/cors.go
- middleware/cors/cors_test.go
- docs/modules/middleware/README.md
Depends On:
- 0736-middleware-coalesce-empty-key

Goal:
Make strict CORS defaults reject empty-origin configurations consistently.

Scope:
- Add an error-returning strict default helper for callers that do not want
  panic-on-invalid configuration.
- Trim and filter blank origins before constructing strict options.
- Keep existing `StrictDefaultOptions` as the panic-on-error convenience helper.

Non-goals:
- Do not change zero-value wildcard CORS behavior.
- Do not change disallowed-origin fall-through behavior.
- Do not remove existing exported helpers.

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
- Blank strict origins are filtered and all-blank input is rejected.
- `StrictDefaultOptionsE` gives non-panicking validation.
- Existing strict helper preserves panic-on-invalid behavior.

Outcome:
- Added `StrictDefaultOptionsE` and `ErrStrictDefaultOriginsRequired` for
  non-panicking strict CORS config validation.
- `StrictDefaultOptions` now trims and filters blank origins before applying
  defaults, while preserving panic-on-invalid behavior.
- Documented the strict helper validation contract.

Validation:
- `go test -timeout 20s ./middleware/cors`
- `go test -timeout 20s ./middleware/...`
