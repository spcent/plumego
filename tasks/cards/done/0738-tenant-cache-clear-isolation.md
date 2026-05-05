# Card 0738

Milestone:
Recipe: specs/change-recipes/store-stable.yaml
Priority: P1
State: done
Primary Module: x/tenant/store/cache
Owned Files:
- x/tenant/store/cache/tenant_cache.go
- x/tenant/store/cache/tenant_cache_test.go
Depends On:

Goal:
Prevent tenant-aware cache Clear from deleting data outside the current tenant.

Scope:
- Make TenantCache.Clear preserve tenant isolation.
- Prefer explicit unsupported behavior unless a tenant-scoped clear capability exists.
- Add regression coverage proving Clear cannot wipe another tenant.

Non-goals:
- Do not change stable store/cache interfaces.
- Do not add scan/prefix support to unrelated cache backends.

Files:
- x/tenant/store/cache/tenant_cache.go
- x/tenant/store/cache/tenant_cache_test.go

Tests:
- go test -timeout 20s ./x/tenant/store/cache
- go test -timeout 20s ./store/cache

Docs Sync:
- None unless public tenant cache behavior documentation exists and needs the unsupported Clear behavior recorded.

Done Definition:
- Clear no longer deletes cross-tenant entries.
- Existing tenant cache operations still scope keys correctly.
- Targeted tests pass.

Outcome:
TenantCache.Clear now validates tenant context and returns ErrCapabilityUnsupported instead of clearing the underlying cache globally. Added regression coverage proving unsupported Clear leaves other tenants' entries intact.

Validation:
- go test -timeout 20s ./x/tenant/store/cache
- go test -timeout 20s ./store/cache
