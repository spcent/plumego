# Card 1582

Milestone: M-018
Recipe: specs/change-recipes/add-http-endpoint.yaml
Priority: P3
State: active
Primary Module: reference/with-tenant-admin
Owned Files:
- `reference/with-tenant-admin/internal/quota/admin/handler.go`
- `reference/with-tenant-admin/internal/quota/admin/handler_test.go`

Goal:
- Implement quota administration handlers for reference/with-tenant-admin using
  x/tenant/quota contracts: GetQuota, SetQuota, and ResetQuota.

Scope:
- Create internal/quota/admin/handler.go:
  - GET /admin/quota/:tenantID — GetQuota: reads current quota and usage from
    x/tenant/quota; returns 200 with QuotaResponse (limit, used, remaining) or
    404 if tenant unknown.
  - PUT /admin/quota/:tenantID — SetQuota: decodes SetQuotaRequest (limit int64),
    updates quota via x/tenant/quota.Manager.SetLimit; returns 200.
  - POST /admin/quota/:tenantID/reset — ResetQuota: resets used counter to zero
    via x/tenant/quota.Manager.Reset; returns 200.
  - All handlers behind RequireAdminToken middleware.
- Write internal/quota/admin/handler_test.go covering:
  - GetQuota for existing tenant returns limit and remaining correctly.
  - GetQuota for unknown tenant returns 404.
  - SetQuota updates limit; subsequent GetQuota reflects new limit.
  - ResetQuota sets used to 0; subsequent GetQuota shows full remaining.
  - Unauthenticated request returns 401.
  - SetQuota with negative limit returns 400.

Non-goals:
- Do not implement quota enforcement in this card (that is x/tenant/quota's job).
- Do not persist quota state beyond the in-memory x/tenant/quota store.
- Do not add per-resource quota granularity.

Files:
- `reference/with-tenant-admin/internal/quota/admin/handler.go`
- `reference/with-tenant-admin/internal/quota/admin/handler_test.go`

Tests:
- `go test -timeout 30s ./reference/with-tenant-admin/internal/quota/...`
- `go vet ./reference/with-tenant-admin/...`
- `go build ./reference/with-tenant-admin/...`

Docs Sync:
- none at this card; README.md written in card 1583.

Done Definition:
- All six quota handler test cases pass.
- SetQuota with negative limit returns 400.
- `go build ./reference/with-tenant-admin/...` exits 0.

Outcome:
-
