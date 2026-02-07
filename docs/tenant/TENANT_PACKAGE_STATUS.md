# Tenant Package Status - v1.0

## Current State

The `tenant/` package provides multi-tenancy infrastructure and is integrated via core options, but remains **experimental**.

### ✅ Implemented

1. **Core Types** (`tenant/config.go`)
   - `Config` - Tenant configuration structure
   - `ConfigManager` interface
   - `InMemoryConfigManager` - In-memory implementation
   - Quota and Policy config providers

2. **Quota Management** (`tenant/quota.go`)
   - `QuotaConfig` - Resource limits per tenant
   - `QuotaEnforcer` - Runtime quota enforcement
   - Metrics: requests/sec, concurrent connections, storage

3. **Policy Management** (`tenant/policy.go`)
   - `PolicyConfig` - Feature flags and permissions
   - Policy evaluation logic

4. **Context Integration** (`tenant/context.go`)
   - `FromContext()` - Extract tenant ID from request context
   - `WithTenantID()` - Inject tenant ID into context

5. **Lifecycle Hooks** (`tenant/hooks.go`)
   - `TenantHook` interface
   - Before/After request hooks for tenant-aware processing

6. **Core Component** (`core/tenant_component.go`)
   - `TenantConfigComponent` - Pluggable component
   - Health check support

7. **Core Options** (`core/options.go`)
   - `WithTenantConfigManager`
   - `WithTenantMiddleware`
   - Tests in `core/tenant_options_test.go`

8. **HTTP Middleware** (`middleware/tenant_*.go`)
   - `TenantResolver`, `TenantPolicy`, `TenantQuota`
   - Basic middleware tests in `middleware/tenant_middleware_test.go`

9. **Database Integration** (`store/db/*`)
   - `NewDBTenantConfigManager`, `NewTenantDB`
   - SQL query isolation helpers

10. **Package Tests** (`tenant/*_test.go`)
   - Config, quota, policy, context coverage

### ❌ Missing

1. **No dedicated tenant guide**
   - Only a README section; no package-level guide or cookbook

2. **Limited end-to-end integration coverage**
   - Core + middleware + DB wiring lacks full E2E tests

3. **No full example application**
   - Quick snippets exist, but no runnable multi-tenant sample app

## Recommendation for v1.0

### Option A: Mark as Experimental 🟡 **RECOMMENDED**

**Action:**
1. Add package comment marking it as experimental
2. Keep in codebase but document limitations
3. No breaking API guarantees until v1.1

**Changes needed:**
```go
// Package tenant provides multi-tenancy support (EXPERIMENTAL).
//
// This package is under active development and its API may change
// in future versions. Use at your own risk for production systems.
//
// Missing features:
//   - Dedicated tenant guide and cookbook
//   - End-to-end integration tests
//   - Production-ready example application
//
// For stable multi-tenancy, consider implementing your own solution
// using this package as a reference.
package tenant
```

**Pros:**
- Keeps code for future development
- Users can experiment with it
- No breaking changes needed
- Shows project roadmap

**Cons:**
- Might confuse users about stability
- Takes up space in codebase

---

### Option B: Remove for v1.0, Add in v1.1 🔴

**Action:**
1. Move to a `feature/tenant` branch
2. Remove from main/v1.0
3. Complete implementation for v1.1

**Pros:**
- Clean v1.0 surface area
- No half-finished features
- Time to properly test and document

**Cons:**
- Loses existing work temporarily
- Users who want multi-tenancy must wait

---

### Option C: Complete for v1.0 🟢

**Action:**
1. Expand integration tests (core + middleware + DB)
2. Write a dedicated tenant guide and cookbook
3. Provide a runnable multi-tenant example app
4. Evaluate moving from experimental to stable

**Estimated effort:** 2-3 days

**Pros:**
- Full-featured v1.0
- Competitive advantage
- Shows maturity

**Cons:**
- Delays v1.0 release
- Risk of bugs in untested code

---

## Decision: Option A (Experimental) ✅

**Rationale:**
- v1.0 focuses on stability of core features
- Multi-tenancy is advanced use case
- Better to ship stable core than half-finished advanced feature
- Can iterate quickly in v1.1 based on feedback

## Implementation for v1.0

### 1. Add Experimental Warning

Add to `/home/user/plumego/tenant/config.go`:

```go
// Package tenant provides multi-tenancy infrastructure (EXPERIMENTAL).
//
// ⚠️  EXPERIMENTAL: This package's API may change in minor versions.
//
// The tenant package enables multi-tenancy support with per-tenant
// configuration, quota enforcement, and policy management. However,
// it is not yet feature-complete for production use.
//
// Current limitations:
//   - Experimental API surface (may change in minor versions)
//   - Limited integration tests across storage backends
//   - No dedicated tenant guide beyond the main README
//
// For production multi-tenancy, we recommend:
//   - Wiring tenant middleware explicitly
//   - Validating isolation and quota behavior in your environment
//   - Treating the API as experimental until stabilized
//
// Planned for v1.1:
//   - Integration test suite across storage backends
//   - Dedicated tenant guide and example app
//   - Stability guarantees
package tenant
```

### 2. Document Usage

Create `/home/user/plumego/tenant/README.md` with:
- Current capabilities
- Known limitations
- Basic usage example
- Roadmap for v1.1

### 3. Update Main README

Add note in main README:
```markdown
## Multi-Tenancy (Experimental)

Plumego includes experimental multi-tenancy support in the `tenant/` package.
This feature is under active development. See `tenant/README.md` for details.
```

## Future Roadmap (v1.1+)

- [ ] Dedicated tenant guide and cookbook
- [ ] End-to-end integration tests (core + middleware + DB, target 80%+ coverage)
- [ ] Middleware usage guide and recommended chain
- [ ] Example: Multi-tenant SaaS application
- [ ] Database sharding integration
- [ ] Tenant isolation verification tools
- [ ] API stability guarantees

## Conclusion

For v1.0: **Keep as experimental**, clearly documented. ✅ **DONE**
For v1.1: **Complete and stabilize** - See implementation plan below.

---

## v1.1 Implementation Plan

**Status**: Planning Complete ✅
**Target Release**: v1.1.0
**Estimated Effort**: 30 hours (3-4 days)

### Detailed Planning Documents

1. **TENANT_PRODUCTION_PLAN.md** - Comprehensive architecture and feature design
2. **TENANT_IMPLEMENTATION_CHECKLIST.md** - Day-by-day task breakdown

### Quick Summary

**What's Already Done** (~60%):
- ✅ Core types (Config, Quota, Policy)
- ✅ In-memory implementations
- ✅ Core options (`WithTenantConfigManager`, `WithTenantMiddleware`)
- ✅ Middleware (Resolver, Quota, Policy)
- ✅ DB-backed config manager + TenantDB isolation helpers
- ✅ Context integration + hooks
- ✅ Basic unit tests

**What's Missing** (~40%):
- ❌ End-to-end integration tests (core + middleware + DB)
- ❌ Dedicated tenant guide and cookbook
- ❌ Runnable example application
- ❌ Stability guarantees / graduation criteria

### Implementation Phases

1. **Phase 1**: Integration tests & coverage
2. **Phase 2**: Documentation (guide + cookbook)
3. **Phase 3**: Example application
4. **Phase 4**: Advanced features (sharding, isolation tooling)
5. **Phase 5**: Stability review and graduation criteria

**Next Action**: Define integration test matrix (core + middleware + DB)
