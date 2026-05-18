# Card 1580

Milestone: M-018
Recipe: specs/change-recipes/add-package.yaml
Priority: P3
State: active
Primary Module: reference/with-tenant-admin
Owned Files:
- `reference/with-tenant-admin/main.go`
- `reference/with-tenant-admin/go.mod`
- `reference/with-tenant-admin/internal/config/config.go`
- `reference/with-tenant-admin/internal/app/app.go`
- `reference/with-tenant-admin/internal/app/routes.go`
- `reference/with-tenant-admin/internal/auth/auth.go`

Goal:
- Scaffold the reference/with-tenant-admin application with core.App, admin
  auth middleware, and the route group structure that subsequent cards
  (1581–1583) will populate.

Scope:
- Create reference/with-tenant-admin/go.mod with module
  github.com/spcent/plumego/reference/with-tenant-admin; require stable roots
  and x/tenant.
- Create main.go following reference/standard-service pattern.
- Create internal/config/config.go: Config struct with Addr, AdminToken,
  LogLevel string.
- Create internal/app/app.go: App struct holding core.App, Logger, and
  x/tenant service instances; constructor New(cfg Config, deps Deps).
- Create internal/app/routes.go: RegisterRoutes function wiring the admin
  auth middleware and three route groups: /admin/tenants, /admin/quota,
  /admin/usage — populated by later cards.
- Create internal/auth/auth.go: RequireAdminToken middleware reading
  AdminToken from config; returns 401 with contract.WriteError if header
  is missing or mismatched; uses timing-safe comparison.
- Confirm `go build ./...` succeeds with empty route handlers.

Non-goals:
- Do not implement route handlers in this card (those are cards 1581–1583).
- Do not add OAuth or JWT auth; use a static admin token for simplicity.
- Do not add a database in this card.

Files:
- `reference/with-tenant-admin/main.go`
- `reference/with-tenant-admin/go.mod`
- `reference/with-tenant-admin/internal/config/config.go`
- `reference/with-tenant-admin/internal/app/app.go`
- `reference/with-tenant-admin/internal/app/routes.go`
- `reference/with-tenant-admin/internal/auth/auth.go`

Tests:
- `go build ./reference/with-tenant-admin/...`
- `go vet ./reference/with-tenant-admin/...`
- `go run ./internal/checks/reference-layout`

Docs Sync:
- none at this card; README.md written in card 1583.

Done Definition:
- `go build ./reference/with-tenant-admin/...` exits 0.
- RequireAdminToken returns 401 on missing header and timing-safe comparison
  is used.
- Route groups /admin/tenants, /admin/quota, /admin/usage are registered.
- `go run ./internal/checks/reference-layout` exits 0.

Outcome:
-
