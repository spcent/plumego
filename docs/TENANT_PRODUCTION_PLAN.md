# Tenant Package - Production Readiness Plan

> **Target**: Upgrade tenant package from **Experimental** to **Stable** for v1.1

## Executive Summary

The tenant package has a solid foundation with core types, middleware, and basic integration already implemented. This plan outlines the work needed to make it production-ready.

**Current Completeness**: ~60%
**Estimated Effort**: 3-4 days
**Target Release**: v1.1.0

---

## Current State Analysis

### ✅ What's Already Implemented (Good Foundation)

| Component | Status | Quality | Notes |
|-----------|--------|---------|-------|
| **Core Types** | ✅ Complete | Good | Config, QuotaConfig, PolicyConfig |
| **Config Manager** | ✅ Complete | Good | InMemoryConfigManager with thread-safety |
| **Quota Enforcement** | ✅ Complete | Good | Fixed-window rate limiting |
| **Policy Evaluation** | ✅ Complete | Good | Allow-list based access control |
| **Context Integration** | ✅ Complete | Good | TenantIDFromContext, ContextWithTenantID |
| **Hooks System** | ✅ Complete | Good | OnResolve, OnPolicy, OnQuota callbacks |
| **Middleware - Resolver** | ✅ Complete | Good | Extract tenant from header/principal |
| **Middleware - Quota** | ✅ Complete | Good | Enforce quota limits with retry-after |
| **Middleware - Policy** | ✅ Complete | Good | Enforce policy rules |
| **Middleware Tests** | ✅ Basic | Medium | 3 tests covering happy/error paths |
| **Core Component** | ✅ Complete | Good | TenantConfigComponent with health checks |

### ❌ What's Missing (Critical Gaps)

| Gap | Priority | Impact | Effort |
|-----|----------|--------|--------|
| **Tenant package tests** | 🔴 Critical | No coverage for quota/policy/config | 4 hours |
| **Core integration** | 🔴 Critical | No WithTenant*() options | 2 hours |
| **Public exports** | 🔴 Critical | Not in plumego.go | 30 min |
| **Persistent storage** | 🟡 High | Only in-memory config | 4 hours |
| **Database isolation** | 🟡 High | No tenant-aware queries | 4 hours |
| **Example application** | 🟡 High | No reference implementation | 3 hours |
| **Documentation** | 🟡 High | No usage guide | 2 hours |
| **Advanced quota** | 🟢 Medium | Only fixed-window, no sliding | 3 hours |
| **Cache isolation** | 🟢 Medium | No tenant-aware caching | 2 hours |
| **Metrics integration** | 🟢 Medium | No tenant-specific metrics | 2 hours |
| **Admin API** | 🟢 Low | No tenant CRUD endpoints | 4 hours |

**Total Estimated Effort**: ~30 hours (3-4 days)

---

## Detailed Implementation Plan

### Phase 1: Testing & Stability (Priority: Critical) ⏱️ 6 hours

**Goal**: Achieve 80%+ test coverage for tenant package

#### 1.1 Tenant Package Unit Tests

**File**: `tenant/config_test.go`
```go
- TestInMemoryConfigManager_SetGet
- TestInMemoryConfigManager_NotFound
- TestInMemoryConfigManager_Concurrent
- TestInMemoryConfigManager_QuotaProvider
- TestInMemoryConfigManager_PolicyProvider
```

**File**: `tenant/quota_test.go`
```go
- TestInMemoryQuotaManager_Allow
- TestInMemoryQuotaManager_Exceed
- TestInMemoryQuotaManager_WindowReset
- TestInMemoryQuotaManager_Unlimited
- TestInMemoryQuotaManager_TokensOnly
- TestInMemoryQuotaManager_RequestsOnly
- TestInMemoryQuotaManager_RetryAfter
- TestInMemoryQuotaManager_Concurrent
```

**File**: `tenant/policy_test.go`
```go
- TestConfigPolicyEvaluator_Allow
- TestConfigPolicyEvaluator_DenyModel
- TestConfigPolicyEvaluator_DenyTool
- TestConfigPolicyEvaluator_EmptyListAllowsAll
- TestConfigPolicyEvaluator_TenantNotFound
```

**File**: `tenant/context_test.go`
```go
- TestContextWithTenantID
- TestTenantIDFromContext_Missing
- TestRequestWithTenantID
```

**Acceptance Criteria**:
- ✅ All tests pass with `go test ./tenant/...`
- ✅ Race detector clean: `go test -race ./tenant/...`
- ✅ Coverage ≥ 80%: `go test -cover ./tenant/...`

---

#### 1.2 Enhanced Middleware Tests

**File**: `middleware/tenant_middleware_test.go` (expand existing)
```go
// Add comprehensive tests
- TestTenantResolver_FromHeader
- TestTenantResolver_PriorityPrincipalOverHeader
- TestTenantResolver_CustomHeader
- TestTenantResolver_AllowMissing
- TestTenantResolver_CustomOnMissing
- TestTenantResolver_Hooks

- TestTenantQuota_UnlimitedQuota
- TestTenantQuota_TokenEstimator
- TestTenantQuota_CustomOnRejected
- TestTenantQuota_RetryAfterHeader
- TestTenantQuota_Hooks

- TestTenantPolicy_AllowAll
- TestTenantPolicy_CustomHeaders
- TestTenantPolicy_CustomOnDenied
- TestTenantPolicy_Hooks
```

**Acceptance Criteria**:
- ✅ 100% branch coverage for middleware
- ✅ All error paths tested
- ✅ Hook callbacks verified

---

### Phase 2: Core Integration (Priority: Critical) ⏱️ 2.5 hours

**Goal**: Make tenant configuration as easy as other core features

#### 2.1 Add Core Options

**File**: `core/options.go` (additions)
```go
// WithTenantConfigManager registers a tenant config manager component
func WithTenantConfigManager(manager tenant.ConfigManager) Option {
    return func(a *App) {
        component := &TenantConfigComponent{
            Name:    "tenant-config",
            Manager: manager,
        }
        a.AddComponent(component)
    }
}

// WithTenantMiddleware adds tenant resolution, quota, and policy middleware
func WithTenantMiddleware(options TenantMiddlewareOptions) Option {
    return func(a *App) {
        // Add tenant resolver
        a.Use(middleware.TenantResolver(middleware.TenantResolverOptions{
            HeaderName:   options.HeaderName,
            AllowMissing: options.AllowMissing,
            Hooks:        options.Hooks,
        }))

        // Add quota enforcement if configured
        if options.QuotaManager != nil {
            a.Use(middleware.TenantQuota(middleware.TenantQuotaOptions{
                Manager: options.QuotaManager,
                Hooks:   options.Hooks,
            }))
        }

        // Add policy enforcement if configured
        if options.PolicyEvaluator != nil {
            a.Use(middleware.TenantPolicy(middleware.TenantPolicyOptions{
                Evaluator: options.PolicyEvaluator,
                Hooks:     options.Hooks,
            }))
        }
    }
}

type TenantMiddlewareOptions struct {
    HeaderName      string
    AllowMissing    bool
    QuotaManager    tenant.QuotaManager
    PolicyEvaluator tenant.PolicyEvaluator
    Hooks           tenant.Hooks
}
```

**Test File**: `core/tenant_options_test.go`
```go
- TestWithTenantConfigManager
- TestWithTenantMiddleware
- TestTenantMiddleware_Integration
```

**Acceptance Criteria**:
- ✅ Can configure tenant with single option
- ✅ Works with other core options
- ✅ Tests verify integration

---

#### 2.2 Export Types in plumego.go

**File**: `plumego.go` (additions)
```go
// Tenant types
type (
    TenantConfig              = tenant.Config
    TenantQuotaConfig         = tenant.QuotaConfig
    TenantPolicyConfig        = tenant.PolicyConfig
    TenantConfigManager       = tenant.ConfigManager
    TenantQuotaManager        = tenant.QuotaManager
    TenantPolicyEvaluator     = tenant.PolicyEvaluator
    TenantHooks               = tenant.Hooks
)

// Tenant functions
var (
    NewInMemoryTenantConfigManager = tenant.NewInMemoryConfigManager
    NewInMemoryQuotaManager        = tenant.NewInMemoryQuotaManager
    NewConfigPolicyEvaluator       = tenant.NewConfigPolicyEvaluator
    TenantIDFromContext            = tenant.TenantIDFromContext
    ContextWithTenantID            = tenant.ContextWithTenantID
)

// Tenant errors
var (
    ErrTenantNotFound  = tenant.ErrTenantNotFound
    ErrQuotaExceeded   = tenant.ErrQuotaExceeded
    ErrPolicyDenied    = tenant.ErrPolicyDenied
)
```

**Acceptance Criteria**:
- ✅ All public tenant APIs accessible via `plumego.` package
- ✅ GoDoc shows tenant types in main package

---

### Phase 3: Persistent Storage (Priority: High) ⏱️ 4 hours

**Goal**: Support database-backed tenant configuration

#### 3.1 Database Config Manager

**File**: `store/db/tenant_config.go`
```go
// DBTenantConfigManager loads tenant configs from database
type DBTenantConfigManager struct {
    db    *sql.DB
    cache *cache.LRU  // Optional caching layer
}

func NewDBTenantConfigManager(db *sql.DB, options ...DBTenantOption) *DBTenantConfigManager

func (m *DBTenantConfigManager) GetTenantConfig(ctx context.Context, tenantID string) (tenant.Config, error)

func (m *DBTenantConfigManager) SetTenantConfig(ctx context.Context, cfg tenant.Config) error

func (m *DBTenantConfigManager) DeleteTenantConfig(ctx context.Context, tenantID string) error

func (m *DBTenantConfigManager) ListTenants(ctx context.Context, limit, offset int) ([]tenant.Config, error)
```

**Schema**:
```sql
CREATE TABLE tenants (
    id VARCHAR(255) PRIMARY KEY,
    quota_requests_per_minute INT DEFAULT 0,
    quota_tokens_per_minute INT DEFAULT 0,
    allowed_models TEXT,  -- JSON array
    allowed_tools TEXT,   -- JSON array
    metadata TEXT,        -- JSON object
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_tenants_updated_at ON tenants(updated_at);
```

**Test File**: `store/db/tenant_config_test.go`
```go
- TestDBTenantConfigManager_CRUD
- TestDBTenantConfigManager_NotFound
- TestDBTenantConfigManager_Cache
- TestDBTenantConfigManager_Concurrent
```

**Acceptance Criteria**:
- ✅ CRUD operations work
- ✅ Optional LRU caching
- ✅ Migration script provided
- ✅ Tests use in-memory SQLite

---

### Phase 4: Database Isolation (Priority: High) ⏱️ 4 hours

**Goal**: Enable tenant-scoped database queries

#### 4.1 Tenant-Aware DB Queries

**File**: `store/db/tenant_isolation.go`
```go
// TenantDB wraps sql.DB with automatic tenant filtering
type TenantDB struct {
    db           *sql.DB
    tenantColumn string  // Default: "tenant_id"
}

// QueryContext automatically adds WHERE tenant_id = ?
func (tdb *TenantDB) QueryContext(ctx context.Context, tenantID string, query string, args ...interface{}) (*sql.Rows, error)

// ExecContext automatically adds WHERE tenant_id = ?
func (tdb *TenantDB) ExecContext(ctx context.Context, tenantID string, query string, args ...interface{}) (sql.Result, error)

// QueryFromContext extracts tenant from context
func (tdb *TenantDB) QueryFromContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error) {
    tenantID := tenant.TenantIDFromContext(ctx)
    return tdb.QueryContext(ctx, tenantID, query, args...)
}
```

**Helper Functions**:
```go
// AddTenantFilter injects tenant_id filter into SQL
func AddTenantFilter(query string, tenantID string, tenantColumn string) (string, []interface{})

// ValidateQuery checks for missing tenant filter (safety check)
func ValidateQuery(query string, tenantColumn string) error
```

**Test File**: `store/db/tenant_isolation_test.go`
```go
- TestTenantDB_Query
- TestTenantDB_Exec
- TestTenantDB_FromContext
- TestAddTenantFilter
- TestValidateQuery
```

**Acceptance Criteria**:
- ✅ Automatic tenant filtering
- ✅ Opt-out for admin queries
- ✅ Safety checks for accidental cross-tenant queries

---

#### 4.2 Integration with Sharding

**File**: `store/db/sharding/tenant_sharding.go`
```go
// TenantShardRouter routes queries based on tenant ID
type TenantShardRouter struct {
    coordinator *Coordinator
    hashFunc    func(string) uint32
}

func (r *TenantShardRouter) ShardForTenant(tenantID string) (*Shard, error)

func (r *TenantShardRouter) QueryTenant(ctx context.Context, tenantID string, query string, args ...interface{}) (*sql.Rows, error)
```

**Acceptance Criteria**:
- ✅ Consistent hashing by tenant ID
- ✅ Integration with existing sharding package

---

### Phase 5: Example Application (Priority: High) ⏱️ 3 hours

**Goal**: Provide reference implementation for multi-tenant SaaS

#### 5.1 Multi-Tenant API Example

**File**: `examples/multi-tenant-saas/main.go`

**Features**:
- Tenant registration endpoint (`POST /tenants`)
- Tenant configuration management (`GET/PUT /tenants/:id/config`)
- Tenant-scoped resource API (`/api/v1/resources`)
- Admin endpoints (bypass tenant isolation)
- Quota enforcement demonstration
- Policy enforcement demonstration
- Metrics per tenant

**Structure**:
```
examples/multi-tenant-saas/
├── main.go           # Application entry point
├── tenant_handler.go # Tenant CRUD endpoints
├── api_handler.go    # Tenant-scoped API
├── admin_handler.go  # Admin endpoints
├── schema.sql        # Database schema
├── README.md         # Usage guide
└── .env.example      # Configuration template
```

**Code Example**:
```go
func main() {
    // Setup tenant config manager
    db := setupDatabase()
    tenantMgr := db.NewDBTenantConfigManager(db)
    quotaMgr := tenant.NewInMemoryQuotaManager(tenantMgr)
    policyEval := tenant.NewConfigPolicyEvaluator(tenantMgr)

    // Create app with tenant support
    app := plumego.New(
        plumego.WithAddr(":8080"),
        plumego.WithTenantConfigManager(tenantMgr),
        plumego.WithTenantMiddleware(plumego.TenantMiddlewareOptions{
            HeaderName:      "X-Tenant-ID",
            AllowMissing:    false,
            QuotaManager:    quotaMgr,
            PolicyEvaluator: policyEval,
            Hooks: plumego.TenantHooks{
                OnQuota: func(ctx context.Context, info tenant.QuotaDecision) {
                    log.Printf("Quota: tenant=%s allowed=%v", info.TenantID, info.Allowed)
                },
            },
        }),
    )

    // Admin routes (no tenant required)
    admin := app.Group("/admin")
    admin.Use(middleware.BasicAuth(middleware.BasicAuthOptions{...}))
    admin.POST("/tenants", createTenant)
    admin.GET("/tenants/:id", getTenant)
    admin.PUT("/tenants/:id/config", updateTenantConfig)

    // Tenant-scoped API routes
    api := app.Group("/api/v1")
    api.GET("/resources", listResources)      // Filtered by tenant
    api.POST("/resources", createResource)    // Tagged with tenant
    api.GET("/resources/:id", getResource)    // Validated for tenant

    app.Boot()
}
```

**Acceptance Criteria**:
- ✅ Complete working example
- ✅ README with setup instructions
- ✅ Demonstrates all tenant features
- ✅ Tests included

---

### Phase 6: Documentation (Priority: High) ⏱️ 2 hours

**Goal**: Comprehensive usage documentation

#### 6.1 Tenant Package README

**File**: `tenant/README.md`

**Sections**:
1. Overview & Use Cases
2. Core Concepts (Config, Quota, Policy)
3. Quick Start
4. Configuration Options
5. Middleware Integration
6. Storage Backends
7. Best Practices
8. Troubleshooting
9. API Reference

#### 6.2 Architecture Guide

**File**: `docs/MULTI_TENANT_ARCHITECTURE.md`

**Sections**:
1. Multi-Tenancy Patterns (shared DB, separate DB, hybrid)
2. Data Isolation Strategies
3. Quota & Rate Limiting
4. Security Considerations
5. Performance Optimization
6. Monitoring & Observability
7. Migration Strategies

#### 6.3 Update Main Documentation

**Files to Update**:
- `README.md` - Add Multi-Tenancy section
- `README_CN.md` - Add Chinese translation
- `CLAUDE.md` - Add tenant module documentation
- `CHANGELOG.md` - Document v1.1 tenant features

**Acceptance Criteria**:
- ✅ Complete documentation coverage
- ✅ Code examples for common patterns
- ✅ Clear migration path from experimental

---

### Phase 7: Advanced Features (Priority: Medium) ⏱️ 7 hours

**Goal**: Production-grade enhancements

#### 7.1 Sliding Window Quota

**File**: `tenant/quota_sliding.go`
```go
// SlidingWindowQuotaManager implements sliding window rate limiting
type SlidingWindowQuotaManager struct {
    provider QuotaConfigProvider
    storage  QuotaStorage  // Redis or in-memory
}

// Uses sub-windows for smoother quota enforcement
func (m *SlidingWindowQuotaManager) Allow(ctx context.Context, tenantID string, req QuotaRequest) (QuotaResult, error)
```

**Benefits**:
- Smoother rate limiting
- No burst at window boundaries
- Better user experience

---

#### 7.2 Cache Isolation

**File**: `store/cache/tenant_cache.go`
```go
// TenantCache wraps cache with automatic tenant key prefixing
type TenantCache struct {
    cache  Cache
    prefix string
}

func (tc *TenantCache) Get(ctx context.Context, key string) (interface{}, error) {
    tenantID := tenant.TenantIDFromContext(ctx)
    fullKey := fmt.Sprintf("%s:%s:%s", tc.prefix, tenantID, key)
    return tc.cache.Get(ctx, fullKey)
}

func (tc *TenantCache) Set(ctx context.Context, key string, value interface{}, ttl time.Duration) error
```

**Acceptance Criteria**:
- ✅ Automatic key prefixing
- ✅ Tenant isolation guaranteed
- ✅ Works with existing cache implementations

---

#### 7.3 Metrics per Tenant

**File**: `metrics/tenant_metrics.go`
```go
// TenantMetrics tracks metrics per tenant
type TenantMetrics struct {
    requestsTotal     *prometheus.CounterVec    // by tenant_id
    requestDuration   *prometheus.HistogramVec  // by tenant_id
    quotaRejections   *prometheus.CounterVec    // by tenant_id
    policyDenials     *prometheus.CounterVec    // by tenant_id
}

func NewTenantMetrics() *TenantMetrics

func (m *TenantMetrics) RecordRequest(ctx context.Context, duration time.Duration)
```

**Dashboard Queries**:
```promql
# Requests per tenant
sum by (tenant_id) (rate(plumego_tenant_requests_total[5m]))

# P99 latency per tenant
histogram_quantile(0.99, sum by (tenant_id, le) (rate(plumego_tenant_request_duration_seconds_bucket[5m])))

# Quota rejection rate
sum by (tenant_id) (rate(plumego_tenant_quota_rejections_total[5m]))
```

---

#### 7.4 Admin API

**File**: `tenant/admin.go`
```go
// AdminService provides CRUD operations for tenant management
type AdminService struct {
    manager ConfigManager
}

func (s *AdminService) CreateTenant(ctx context.Context, cfg Config) error
func (s *AdminService) GetTenant(ctx context.Context, tenantID string) (Config, error)
func (s *AdminService) UpdateTenant(ctx context.Context, cfg Config) error
func (s *AdminService) DeleteTenant(ctx context.Context, tenantID string) error
func (s *AdminService) ListTenants(ctx context.Context, filter ListFilter) ([]Config, error)
```

**REST Endpoints** (example):
```
POST   /admin/tenants
GET    /admin/tenants/:id
PUT    /admin/tenants/:id
DELETE /admin/tenants/:id
GET    /admin/tenants
PUT    /admin/tenants/:id/quota
PUT    /admin/tenants/:id/policy
```

---

## Implementation Timeline

### Week 1: Core Stability

| Day | Tasks | Hours | Deliverables |
|-----|-------|-------|--------------|
| Day 1 | Phase 1.1: Tenant package tests | 4h | 80%+ coverage |
| Day 1 | Phase 1.2: Enhanced middleware tests | 2h | 100% middleware coverage |
| Day 2 | Phase 2.1: Core integration options | 2h | WithTenant*() functions |
| Day 2 | Phase 2.2: Export types | 0.5h | plumego.go exports |
| Day 2 | Phase 3.1: Database config manager | 3.5h | DB-backed configs |

**End of Day 2**: Core tenant features stable and tested ✅

---

### Week 2: Production Features

| Day | Tasks | Hours | Deliverables |
|-----|-------|-------|--------------|
| Day 3 | Phase 4.1: Database isolation | 4h | Tenant-scoped queries |
| Day 3 | Phase 5.1: Example application (start) | 2h | Basic structure |
| Day 4 | Phase 5.1: Example application (finish) | 2h | Complete example |
| Day 4 | Phase 6: Documentation | 2h | README, guides |
| Day 5 | Phase 7.1: Sliding window quota | 3h | Advanced quota |
| Day 5 | Phase 7.2: Cache isolation | 2h | Tenant-aware cache |

**End of Day 5**: Production-ready tenant package ✅

---

## Testing Strategy

### Unit Tests
- **Target**: 80%+ coverage for `tenant/` package
- **Files**: All `*_test.go` files in `tenant/`
- **Focus**: Edge cases, concurrency, error handling

### Integration Tests
- **Target**: End-to-end tenant workflows
- **File**: `tenant/integration_test.go`
- **Scenarios**:
  - Multi-tenant request routing
  - Quota enforcement across requests
  - Policy evaluation with DB configs
  - Cache isolation verification

### Performance Tests
- **File**: `tenant/benchmark_test.go`
- **Benchmarks**:
  - `BenchmarkQuotaManager_Allow`
  - `BenchmarkPolicyEvaluator_Evaluate`
  - `BenchmarkDBConfigManager_Get` (with/without cache)

**Acceptance Criteria**:
- ✅ Quota check < 1ms (in-memory)
- ✅ Config lookup < 5ms (with cache)
- ✅ No memory leaks in long-running tests

---

## Migration from Experimental

### For Existing Users

**Before (v1.0 - Experimental)**:
```go
// Manual setup required
manager := tenant.NewInMemoryConfigManager()
component := &core.TenantConfigComponent{Manager: manager}
app.AddComponent(component)

// Manual middleware
app.Use(middleware.TenantResolver(...))
app.Use(middleware.TenantQuota(...))
```

**After (v1.1 - Stable)**:
```go
// Simple options-based setup
manager := plumego.NewInMemoryTenantConfigManager()
app := plumego.New(
    plumego.WithTenantConfigManager(manager),
    plumego.WithTenantMiddleware(plumego.TenantMiddlewareOptions{...}),
)
```

**Breaking Changes**: None - fully backward compatible

---

## Success Criteria

### Code Quality
- ✅ 80%+ test coverage for tenant package
- ✅ 100% coverage for critical paths (quota, policy)
- ✅ Zero race conditions (`go test -race`)
- ✅ All linters pass (`go vet`, `staticcheck`)

### Documentation
- ✅ Complete README in `tenant/`
- ✅ Architecture guide
- ✅ Working example application
- ✅ Updated main documentation

### Integration
- ✅ `core.WithTenant*()` options
- ✅ Types exported in `plumego.go`
- ✅ Works with existing middleware
- ✅ Database isolation helpers

### Production Readiness
- ✅ Persistent config storage (DB)
- ✅ Performance benchmarks published
- ✅ Metrics integration
- ✅ Admin API for tenant management

### Community
- ✅ Clear migration guide
- ✅ No breaking changes from experimental
- ✅ Responsive to feedback

---

## Risk Assessment

| Risk | Probability | Impact | Mitigation |
|------|-------------|--------|------------|
| **API changes during development** | Medium | High | Freeze API early, get feedback |
| **Performance regressions** | Low | Medium | Benchmark before/after |
| **Database migration issues** | Medium | High | Provide migration scripts, test thoroughly |
| **Backward compatibility break** | Low | Critical | Maintain experimental API, deprecate gradually |
| **Insufficient testing** | Low | High | Enforce 80%+ coverage requirement |

---

## Post-Launch

### v1.1.0 Release
- Mark tenant package as **Stable**
- Announce in CHANGELOG
- Blog post about multi-tenancy features
- Community outreach

### v1.2+ Roadmap
- Redis-backed quota manager
- Distributed quota enforcement
- Tenant usage analytics
- Tenant onboarding workflows
- Advanced policy rules (RBAC, ABAC)
- Multi-region tenant isolation

---

## Appendix: Feature Comparison

### Plumego v1.0 (Experimental) vs v1.1 (Stable)

| Feature | v1.0 | v1.1 |
|---------|------|------|
| In-memory config | ✅ | ✅ |
| Database config | ❌ | ✅ |
| Quota enforcement | ✅ | ✅ |
| Sliding window quota | ❌ | ✅ |
| Policy evaluation | ✅ | ✅ |
| Middleware | ✅ | ✅ (improved) |
| Core integration | ❌ | ✅ |
| Public exports | ❌ | ✅ |
| Database isolation | ❌ | ✅ |
| Cache isolation | ❌ | ✅ |
| Metrics | ❌ | ✅ |
| Tests | Basic | Comprehensive |
| Documentation | Minimal | Complete |
| Examples | ❌ | ✅ |
| Admin API | ❌ | ✅ |

---

## Conclusion

The tenant package has a strong foundation but needs focused effort on:
1. **Testing** (critical for stability)
2. **Integration** (ease of use)
3. **Storage** (production persistence)
4. **Documentation** (adoption)

With 3-4 days of focused development, we can deliver a production-ready multi-tenancy solution that competes with dedicated SaaS frameworks while maintaining plumego's lightweight philosophy.

**Recommendation**: Proceed with this plan for v1.1.0 release.
