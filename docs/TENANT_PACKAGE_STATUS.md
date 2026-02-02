# Tenant Package Status - v1.0

## Current State

The `tenant/` package provides multi-tenancy infrastructure but is **not yet fully integrated** into the plumego core API.

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

### ❌ Missing

1. **No `WithTenant*` options in `core/options.go`**
   - Cannot easily configure tenant support at app level
   - Users must manually create and register `TenantConfigComponent`

2. **No tests** - 0% coverage
   - No `*_test.go` files
   - Untested code

3. **No middleware**
   - No HTTP middleware to extract tenant ID from headers/JWT
   - No automatic quota enforcement middleware

4. **No examples**
   - No example showing how to use multi-tenancy
   - No documentation on integration patterns

5. **Not exported in `plumego.go`**
   - Types not re-exported at top level
   - Users must import `github.com/spcent/plumego/tenant` directly

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
//   - HTTP middleware for tenant extraction
//   - Automatic quota enforcement
//   - Comprehensive test coverage
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
1. Add tests (QuotaEnforcer, PolicyConfig, etc.)
2. Create middleware for tenant extraction
3. Add `core.WithTenant()` option
4. Write examples and documentation
5. Export types in `plumego.go`

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
// it is not yet feature-complete and lacks production-ready middleware
// integration.
//
// Current limitations:
//   - No HTTP middleware for automatic tenant extraction
//   - No integration with core.App options
//   - Limited test coverage
//   - No comprehensive examples
//
// For production multi-tenancy, we recommend:
//   - Implementing custom tenant extraction middleware
//   - Using this package's types as a foundation
//   - Thoroughly testing your integration
//
// Planned for v1.1:
//   - Full middleware integration
//   - Complete test suite
//   - Production-ready examples
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

- [ ] HTTP middleware: `tenant.ExtractFromHeader()`, `tenant.ExtractFromJWT()`
- [ ] Quota enforcement middleware: `tenant.EnforceQuota()`
- [ ] Core integration: `core.WithTenant(manager)`
- [ ] Complete test suite (80%+ coverage)
- [ ] Example: Multi-tenant SaaS application
- [ ] Database sharding integration
- [ ] Tenant isolation verification tools
- [ ] API stability guarantees

## Conclusion

For v1.0: **Keep as experimental**, clearly documented.
For v1.1: **Complete and stabilize** based on user feedback.
