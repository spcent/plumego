# Plan for M-018: Multi-Tenant Admin Reference

Milestone: `M-018`
Objective: Ship reference/with-tenant-admin demonstrating tenant lifecycle
management, quota administration, and usage recording using x/tenant primitives,
with all admin endpoints enforcing fail-closed auth and a README explaining
how to swap in-memory adapters for real persistence.
Constraints: self-contained Go module importing only stable roots and x/tenant,
handler shape func(http.ResponseWriter, *http.Request) throughout, no payment
or billing logic, no CRUD additions to x/tenant stable contracts, all tests
pass without external infrastructure.
Affected Modules: reference/with-tenant-admin.

## Phase Map

- Phase 1: Orient — read x/tenant sub-package APIs (resolve, policy, quota)
  and reference/with-tenant/ canonical structure before scaffolding.
- Phase 2: Implement (parallel) — scaffold the app and implement tenant admin,
  quota admin, and usage handlers concurrently.
- Phase 3: Documentation — write README.md covering API surface, run instructions,
  adapter replacement guidance, and fail-closed auth pattern.
- Phase 4: Validate and Ship — run acceptance criteria, commit.

## Card Inventory

| Card | Goal | Primary Module | Owned Files | Depends On | Quick Gates |
|------|------|----------------|-------------|------------|-------------|
| 1580 | Scaffold reference/with-tenant-admin app skeleton with auth middleware | reference/with-tenant-admin | `reference/with-tenant-admin/main.go`, `reference/with-tenant-admin/go.mod`, `reference/with-tenant-admin/internal/config/config.go`, `reference/with-tenant-admin/internal/app/app.go` | M-010 | `go build ./reference/with-tenant-admin/...` |
| 1581 | Implement internal/tenant/admin/ with tenant CRUD handlers | reference/with-tenant-admin | `reference/with-tenant-admin/internal/tenant/admin/handler.go`, `reference/with-tenant-admin/internal/tenant/admin/store.go`, `reference/with-tenant-admin/internal/tenant/admin/handler_test.go` | 1580 | `go test ./reference/with-tenant-admin/internal/tenant/...` |
| 1582 | Implement internal/quota/admin/ with quota management handlers | reference/with-tenant-admin | `reference/with-tenant-admin/internal/quota/admin/handler.go`, `reference/with-tenant-admin/internal/quota/admin/handler_test.go` | 1580 | `go test ./reference/with-tenant-admin/internal/quota/...` |
| 1583 | Implement internal/usage/ with usage recording and report handlers | reference/with-tenant-admin | `reference/with-tenant-admin/internal/usage/handler.go`, `reference/with-tenant-admin/internal/usage/adapter.go`, `reference/with-tenant-admin/internal/usage/handler_test.go` | 1580 | `go test ./reference/with-tenant-admin/internal/usage/...` |

## Dependency Edges

- `1580 -> 1581`
- `1580 -> 1582`
- `1580 -> 1583`

## Parallel Groups

- Group A: card 1580 — must complete first; provides the scaffold and auth middleware
  all other cards depend on.
- Group B (parallel after A): cards 1581, 1582, 1583 — independent internal packages,
  no file overlap.
- Group C (sequential after B): README.md after all four packages exist and are tested.

## Risk Register

- Risk: x/tenant quota contracts do not expose the set-quota and reset-quota operations
  the admin API needs.
  Mitigation: card 1580 reads x/tenant/quota/ before scaffolding; if the contract is
  insufficient, record a blocker — do not add quota CRUD to x/tenant in this milestone.
- Risk: fail-closed auth check is accidentally optional (missing early return on auth failure).
  Mitigation: card 1581 test suite includes a request without an Authorization header
  that must return 401; handler_test.go is reviewed for this case before done.

## Verification Strategy

- Card-level checks: each handler card runs `go test` with both success and auth-failure
  cases after writing the handler.
- Build check: `go build ./reference/with-tenant-admin/...` after all four cards.
- Auth check: every handler test exercises the missing/invalid auth path and asserts 401.
- Reference-layout check: `go run ./internal/checks/reference-layout` after all cards.

## Exit Condition

- all four implementation cards completed with tests
- all admin endpoints fail closed on missing/invalid auth header (asserted in tests)
- reference builds and tests pass without external infrastructure
- README.md has run instructions and adapter replacement guidance
- verify report shows pass
- milestone acceptance criteria ready for PR packaging
