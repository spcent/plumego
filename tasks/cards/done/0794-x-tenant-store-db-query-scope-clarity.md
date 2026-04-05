# Card 0794

Priority: P1
State: active
Primary Module: x/tenant
Owned Files:
- `docs/modules/x-tenant/README.md`
- `docs/architecture/X_TENANT_BLUEPRINT.md`
- `x/tenant/store/db/doc.go`
- `x/tenant/store/db/tenant_db_test.go`
Depends On:

Goal:
- Clarify supported query-scoping behavior and sharp edges for the tenant-aware DB adapter without widening stable `store`.

Scope:
- Bring the tenant module primer, tenant architecture blueprint, and `x/tenant/store/db` package docs into alignment around supported query-scoping behavior.
- Add or tighten focused tests in `x/tenant/store/db/tenant_db_test.go` where the documented scoping behavior or limitations are currently under-specified.
- Keep the documented behavior explicit about fail-closed expectations and adapter limits.

Non-goals:
- Do not move tenant-aware behavior into stable `store`.
- Do not add new tenant onboarding or CRUD flows.
- Do not broaden this card into quota or ratelimit coverage.

Files:
- `docs/modules/x-tenant/README.md`
- `docs/architecture/X_TENANT_BLUEPRINT.md`
- `x/tenant/store/db/doc.go`
- `x/tenant/store/db/tenant_db_test.go`

Tests:
- `go test -timeout 20s ./x/tenant/store/db`
- `go test -race -timeout 60s ./x/tenant/store/db`
- `go vet ./x/tenant/...`

Docs Sync:
- Keep the module primer, architecture blueprint, and `x/tenant/store/db` package docs aligned on supported scoping behavior, sharp edges, and stable-root boundary rules.

Done Definition:
- Query-scoping behavior and limitations are documented consistently in one canonical way.
- Focused tests cover the documented adapter behavior and fail-closed expectations.
- No stable-root boundary drift is introduced while clarifying the tenant-aware adapter.

Outcome:
- Expanded `x/tenant/store/db` package docs to describe the supported query-rewrite subset, explicit sharp edges, and fail-closed behavior for missing tenant context or invalid tenant-column configuration.
- Synced the tenant module primer and tenant architecture blueprint with the actual `store/db` adapter behavior and boundary expectations.
- Added focused tests for fail-closed invalid-column configuration and multi-row `INSERT ... VALUES` tenant-argument rewriting.

Validation Run:
- `go test -timeout 20s ./x/tenant/store/db`
- `go test -race -timeout 60s ./x/tenant/store/db`
- `go vet ./x/tenant/...`
