# Card 0929

Priority: P0
State: active
Primary Module: contract
Owned Files:
- `contract/auth.go`
- `x/tenant/core/context.go`
Depends On:

Goal:
- Establish one canonical mechanism for carrying tenant ID in a request context and remove the competing path.

Problem:
- Tenant ID reaches the request context through two independent, uncoordinated mechanisms:
  1. `contract.Principal.TenantID` (`auth.go:13`): stored via `contract.WithPrincipal(ctx, p)` under `principalContextKey{}`, read back via `contract.PrincipalFromContext(ctx).TenantID`.
  2. `x/tenant/core.ContextWithTenantID(ctx, id)` (`context.go:11`): stored under `tenantIDContextKey{}`, read back via `x/tenant/core.TenantIDFromContext(ctx)`.
- The two mechanisms use different context keys, so they do not share state. A request processed by tenant middleware that calls `ContextWithTenantID` will have the ID in the tenant context key but not in the principal's `TenantID` field, and vice versa.
- Code that needs the tenant ID must pick one source with no canonical guidance. If authentication middleware sets `Principal.TenantID` and tenant middleware also runs, downstream code may read either key and get the same or different values.
- `contract.Principal.TenantID` exists to support SaaS multi-tenancy, but `x/tenant` is the dedicated multi-tenancy subsystem. Carrying the tenant ID in both places duplicates state and creates a correctness risk when they disagree.

Scope:
- Choose one canonical context mechanism. The preferred answer is `x/tenant/core.ContextWithTenantID` / `TenantIDFromContext` because tenant resolution is the responsibility of the `x/tenant` subsystem.
- Remove `Principal.TenantID` field from `contract.Principal`.
- Remove references to `Principal.TenantID` across the codebase; replace reads with `TenantIDFromContext(r.Context())`.
- Document in `contract/auth.go` that tenant identity is carried separately via `x/tenant/core` context accessors, not via `Principal`.
- Document in `x/tenant/core/context.go` that `TenantIDFromContext` is the sole canonical tenant ID accessor.

Non-goals:
- Do not remove `Principal.Subject` or any other field.
- Do not merge `contract` and `x/tenant/core` packages.
- Do not add a re-export alias for `TenantIDFromContext` in `contract`.

Files:
- `contract/auth.go`
- `contract/auth_test.go`
- `x/tenant/core/context.go`
- Any file that reads `Principal.TenantID`.

Tests:
- `go test -timeout 20s ./contract/... ./x/tenant/core/...`
- `go vet ./contract/... ./x/tenant/core/...`
- `go build ./...`

Docs Sync:
- Update `contract` and `x/tenant/core` inline comments to reflect the single canonical tenant ID path.

Done Definition:
- `contract.Principal` has no `TenantID` field.
- No code in the repository reads `Principal.TenantID`.
- `x/tenant/core.TenantIDFromContext` is the only context accessor for tenant ID.
- All tests pass.

Outcome:
- Pending.
