# Card 2292

Milestone:
Recipe: specs/change-recipes/add-http-endpoint.yaml
Priority: P1
State: done
Primary Module: reference
Owned Files:
- reference/production-service/README.md
- reference/production-service/internal/app/routes.go
- reference/production-service/internal/app/app.go
- reference/production-service/internal/config/config.go
- tasks/cards/active/README.md
Depends On: 2281, 2288

Goal:
Deepen `reference/production-service` into a real production vertical slice
with explicit auth, tenant, storage, and ops boundaries while preserving visible
wiring.

Scope:
- Add a minimal protected API route that demonstrates auth and tenant-aware
  request context.
- Keep storage app-local and standard-library-only.
- Document production route exposure and configuration.

Non-goals:
- Do not add external databases or new dependencies.
- Do not mount `x/devtools`.
- Do not turn the reference into a hidden bundle.

Files:
- `reference/production-service/README.md`
- `reference/production-service/internal/app/routes.go`
- `reference/production-service/internal/app/app.go`
- `reference/production-service/internal/config/config.go`
- `tasks/cards/active/README.md`

Tests:
- `go test -timeout 20s ./reference/production-service/...`
- `go run ./internal/checks/reference-layout`
- `scripts/check-spec tasks/cards/done/2292-production-service-auth-tenant-storage-slice.md`

Docs Sync:
- Required because the production reference behavior changes.

Done Definition:
- Production reference shows a small protected application workflow, not only
  health and metrics.
- Security and tenant ownership remain explicit in app wiring.

Outcome:
- Added protected `GET /api/profile` to `reference/production-service` using
  `APP_API_TOKEN` bearer auth and `x/tenant/resolve` with `X-Tenant-ID`.
- Added an app-local standard-library in-memory tenant profile store to make the
  production reference a small vertical slice instead of only health and ops
  routes.
- Documented protected API route exposure, fail-closed token behavior, tenant
  header usage, and app-local storage replacement guidance.

Validations:
- `go test -timeout 20s ./reference/production-service/...`
- `go run ./internal/checks/reference-layout`
- `scripts/check-spec tasks/cards/done/2292-production-service-auth-tenant-storage-slice.md`
