# Card 1581

Milestone: M-018
Recipe: specs/change-recipes/add-http-endpoint.yaml
Priority: P3
State: done
Primary Module: reference/with-tenant-admin
Owned Files:
- `reference/with-tenant-admin/internal/tenant/admin/store.go`
- `reference/with-tenant-admin/internal/tenant/admin/handler.go`
- `reference/with-tenant-admin/internal/tenant/admin/handler_test.go`

Goal:
- Implement tenant lifecycle management handlers for reference/with-tenant-admin:
  CreateTenant, GetTenant, SuspendTenant, and DeleteTenant, backed by an
  in-memory store.

Scope:
- Create internal/tenant/admin/store.go:
  - TenantRecord struct: ID, Name, Status (active|suspended), CreatedAt, SuspendedAt.
  - InMemoryStore: thread-safe map of tenant ID → TenantRecord.
  - Methods: Create, Get, Suspend, Delete — return ErrNotFound for unknown IDs.
- Create internal/tenant/admin/handler.go:
  - POST /admin/tenants — CreateTenant: decodes CreateTenantRequest (Name),
    generates UUID ID, stores record, returns 201 with contract.WriteResponse.
  - GET /admin/tenants/:id — GetTenant: returns 200 or 404 via contract.WriteError.
  - POST /admin/tenants/:id/suspend — SuspendTenant: returns 200 or 404.
  - DELETE /admin/tenants/:id — DeleteTenant: returns 204 or 404.
  - All handlers behind RequireAdminToken middleware (from card 1580).
- Write internal/tenant/admin/handler_test.go covering:
  - CreateTenant with valid name returns 201 with ID.
  - GetTenant for existing ID returns 200 with record.
  - GetTenant for unknown ID returns 404.
  - SuspendTenant transitions status to suspended.
  - DeleteTenant removes record; subsequent Get returns 404.
  - Unauthenticated request to any handler returns 401.

Non-goals:
- Do not add database persistence (in-memory only).
- Do not implement tenant-specific rate limiting or quota in this card.
- Do not use x/tenant's internal CRUD logic; implement at the application layer.

Files:
- `reference/with-tenant-admin/internal/tenant/admin/store.go`
- `reference/with-tenant-admin/internal/tenant/admin/handler.go`
- `reference/with-tenant-admin/internal/tenant/admin/handler_test.go`

Tests:
- `go test -timeout 30s ./reference/with-tenant-admin/internal/tenant/...`
- `go vet ./reference/with-tenant-admin/...`
- `go build ./reference/with-tenant-admin/...`

Docs Sync:
- none at this card; README.md written in card 1583.

Done Definition:
- All five CRUD handler test cases pass.
- Unauthenticated request returns 401.
- `go build ./reference/with-tenant-admin/...` exits 0.

Outcome:
- Implemented the tenant admin in-memory store, lifecycle handlers, admin-token
  protected routes, and CRUD handler tests.
- Validated with `go test -timeout 30s ./internal/tenant/...`,
  `go vet ./...`, and `go build ./...` from `reference/with-tenant-admin`.
- Ran `go run ./internal/checks/reference-layout`, `gofmt -l` for touched
  files, and `git diff --check` from the repository root.
