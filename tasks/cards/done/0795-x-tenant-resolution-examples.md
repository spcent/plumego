# Card 0795

Priority: P1
State: done
Primary Module: x/tenant
Owned Files:
- `x/tenant/resolve/example_test.go`
- `x/tenant/transport/example_test.go`
- `x/tenant/resolve/doc.go`
- `x/tenant/transport/doc.go`
- `docs/modules/x-tenant/README.md`
Depends On:
- `0794-x-tenant-store-db-query-scope-clarity.md`

Goal:
- Add production-oriented runnable examples for tenant resolution flows beyond the current header-only description.

Scope:
- Add offline examples that demonstrate principal-first resolution, custom extractor resolution, and transport helper usage without introducing a new app scaffold.
- Tighten package docs for `x/tenant/resolve` and `x/tenant/transport` so they describe the example-backed canonical flows.
- Update the tenant module primer only enough to point readers at the new example entrypoints and the explicit resolution order.

Non-goals:
- Do not change stable middleware behavior.
- Do not add full end-to-end reference applications.
- Do not broaden into quota, rate-limit, or store adapter behavior in this card.

Files:
- `x/tenant/resolve/example_test.go`
- `x/tenant/transport/example_test.go`
- `x/tenant/resolve/doc.go`
- `x/tenant/transport/doc.go`
- `docs/modules/x-tenant/README.md`

Tests:
- `go test -timeout 20s ./x/tenant/resolve ./x/tenant/transport`
- `go test -race -timeout 60s ./x/tenant/resolve ./x/tenant/transport`

Docs Sync:
- Keep `docs/modules/x-tenant/README.md`, `x/tenant/resolve/doc.go`, and `x/tenant/transport/doc.go` aligned on explicit resolution order and transport helper usage.

Done Definition:
- Tenant resolution now has runnable examples that cover more than the header-only path.
- Package docs and the tenant primer point at the same example-backed flows.
- Examples stay offline and transport-explicit.

Outcome:
- Added runnable offline examples for `x/tenant/resolve` and `x/tenant/transport` that demonstrate principal-first resolution, custom extractor resolution, and canonical tenant transport headers.
- Tightened `x/tenant/resolve` and `x/tenant/transport` package docs to describe the same explicit transport-level responsibilities the examples exercise.
- Updated the tenant module primer to point readers at the new example-backed resolution order instead of only the header-default path.

Validation Run:
- `go test -timeout 20s ./x/tenant/resolve ./x/tenant/transport`
- `go test -race -timeout 60s ./x/tenant/resolve ./x/tenant/transport`
