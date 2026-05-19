# Card 1583

Milestone: M-018
Recipe: specs/change-recipes/add-http-endpoint.yaml
Priority: P3
State: done
Primary Module: reference/with-tenant-admin
Owned Files:
- `reference/with-tenant-admin/internal/usage/handler.go`
- `reference/with-tenant-admin/internal/usage/handler_test.go`
- `reference/with-tenant-admin/internal/usage/store.go`
- `reference/with-tenant-admin/README.md`

Goal:
- Implement usage recording and reporting handlers for reference/with-tenant-admin,
  and write the reference/with-tenant-admin/README.md with run instructions and
  adapter replacement guidance.

Scope:
- Create internal/usage/store.go:
  - UsageRecord struct: TenantID, Resource string; Count int64; RecordedAt time.Time.
  - InMemoryUsageStore: thread-safe slice of UsageRecord per tenant.
  - Record(tenantID, resource string, count int64) error.
  - Report(tenantID string) ([]UsageRecord, error).
- Create internal/usage/handler.go:
  - POST /admin/usage/:tenantID — RecordUsage: decodes RecordUsageRequest
    (resource string, count int64); stores record; returns 202.
  - GET /admin/usage/:tenantID — GetUsageReport: returns 200 with all records
    for the tenant or 404 if tenant unknown.
  - All handlers behind RequireAdminToken middleware.
- Write internal/usage/handler_test.go covering:
  - RecordUsage stores a record; GetUsageReport returns it.
  - GetUsageReport for unknown tenant returns 404.
  - RecordUsage with zero count returns 400.
  - Multiple records for same tenant accumulate correctly.
- Write reference/with-tenant-admin/README.md with:
  - Admin API surface: endpoints, request/response shapes, auth header.
  - How to run: `ADMIN_TOKEN=secret go run ./...` with example curl commands.
  - How to replace in-memory adapters with real persistence (implement the
    store interfaces and inject via constructor).
  - Fail-closed auth pattern explanation.

Non-goals:
- Do not add billing or payment logic.
- Do not persist usage state beyond the in-memory store.

Files:
- `reference/with-tenant-admin/internal/usage/handler.go`
- `reference/with-tenant-admin/internal/usage/handler_test.go`
- `reference/with-tenant-admin/internal/usage/store.go`
- `reference/with-tenant-admin/README.md`

Tests:
- `go test -timeout 30s ./reference/with-tenant-admin/...`
- `go vet ./reference/with-tenant-admin/...`
- `go build ./reference/with-tenant-admin/...`
- `go run ./internal/checks/reference-layout`

Docs Sync:
- reference/with-tenant-admin/README.md is written in this card.

Done Definition:
- All four usage handler test cases pass.
- RecordUsage with zero count returns 400.
- reference/with-tenant-admin/README.md has run instructions with curl examples.
- `go build ./reference/with-tenant-admin/...` exits 0.
- `go run ./internal/checks/reference-layout` exits 0.

Outcome:
- Implemented in-memory usage recording and report handlers, wired
  `/admin/usage/:tenantID` POST/GET behind the admin token middleware, and
  added focused handler tests.
- Wrote `reference/with-tenant-admin/README.md` with run instructions, curl
  examples, adapter replacement guidance, and the fail-closed auth pattern.
- Validated with `go test -timeout 30s ./internal/usage/...`,
  `go test -timeout 30s ./...`, `go vet ./...`, and `go build ./...` from
  `reference/with-tenant-admin`.
- Ran `go run ./internal/checks/reference-layout`, touched-file `gofmt -l`,
  and `git diff --check` from the repository root.
