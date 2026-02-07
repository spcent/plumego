# Tenant Package v1.1 - Implementation Checklist

> **Sprint Goal**: Upgrade tenant package from Experimental → Stable
> **Target**: v1.1.0 Release
> **Estimated**: 30 hours / 3-4 days

---

## 🎯 Phase 1: Testing & Stability (6 hours) - CRITICAL

### Day 1 Morning: Tenant Package Unit Tests (4 hours)

- [ ] **tenant/config_test.go** (1 hour)
  - [ ] `TestInMemoryConfigManager_SetGet` - Basic CRUD
  - [ ] `TestInMemoryConfigManager_NotFound` - Error handling
  - [ ] `TestInMemoryConfigManager_Concurrent` - Race conditions
  - [ ] `TestInMemoryConfigManager_UpdatedAt` - Timestamp tracking
  - [ ] `TestInMemoryConfigManager_QuotaProvider` - Interface impl
  - [ ] `TestInMemoryConfigManager_PolicyProvider` - Interface impl
  - [ ] Verify: `go test -race -cover ./tenant/config_test.go`

- [ ] **tenant/quota_test.go** (2 hours)
  - [ ] `TestInMemoryQuotaManager_Allow` - Basic allow
  - [ ] `TestInMemoryQuotaManager_Exceed_Requests` - Request limit
  - [ ] `TestInMemoryQuotaManager_Exceed_Tokens` - Token limit
  - [ ] `TestInMemoryQuotaManager_WindowReset` - Fixed window reset
  - [ ] `TestInMemoryQuotaManager_Unlimited` - Zero limit behavior
  - [ ] `TestInMemoryQuotaManager_TokensOnly` - Tokens without requests
  - [ ] `TestInMemoryQuotaManager_RequestsOnly` - Requests without tokens
  - [ ] `TestInMemoryQuotaManager_RetryAfter` - Correct retry calculation
  - [ ] `TestInMemoryQuotaManager_Concurrent` - Thread safety
  - [ ] `TestInMemoryQuotaManager_TenantNotFound` - Missing tenant
  - [ ] Verify: `go test -race -cover ./tenant/quota_test.go`

- [ ] **tenant/policy_test.go** (30 min)
  - [ ] `TestConfigPolicyEvaluator_Allow` - Allow all
  - [ ] `TestConfigPolicyEvaluator_DenyModel` - Block model
  - [ ] `TestConfigPolicyEvaluator_DenyTool` - Block tool
  - [ ] `TestConfigPolicyEvaluator_EmptyListAllowsAll` - Empty = allow
  - [ ] `TestConfigPolicyEvaluator_TenantNotFound` - Missing tenant
  - [ ] Verify: `go test -race -cover ./tenant/policy_test.go`

- [ ] **tenant/context_test.go** (30 min)
  - [ ] `TestContextWithTenantID` - Set tenant
  - [ ] `TestTenantIDFromContext` - Get tenant
  - [ ] `TestTenantIDFromContext_Missing` - Nil context
  - [ ] `TestRequestWithTenantID` - HTTP request wrapping
  - [ ] Verify: `go test -cover ./tenant/context_test.go`

**Checkpoint**: ✅ `go test -race -cover ./tenant/... | grep "coverage: [8-9][0-9]%"`

---

### Day 1 Afternoon: Enhanced Middleware Tests (2 hours)

- [ ] **middleware/tenant_middleware_test.go** - Expand existing tests

  **Resolver Tests** (45 min):
  - [ ] `TestTenantResolver_FromHeader` - Header extraction
  - [ ] `TestTenantResolver_PriorityPrincipalOverHeader` - Priority order
  - [ ] `TestTenantResolver_CustomHeader` - Custom header name
  - [ ] `TestTenantResolver_AllowMissing` - Optional tenant
  - [ ] `TestTenantResolver_CustomOnMissing` - Custom error handler
  - [ ] `TestTenantResolver_Hooks` - Callback invocation

  **Quota Tests** (45 min):
  - [ ] `TestTenantQuota_UnlimitedQuota` - No limits
  - [ ] `TestTenantQuota_TokenEstimator` - Custom token calculation
  - [ ] `TestTenantQuota_CustomOnRejected` - Custom rejection handler
  - [ ] `TestTenantQuota_RetryAfterHeader` - HTTP header correctness
  - [ ] `TestTenantQuota_Hooks` - Callback invocation
  - [ ] `TestTenantQuota_NilManager` - Graceful degradation

  **Policy Tests** (30 min):
  - [ ] `TestTenantPolicy_AllowAll` - Empty policy allows all
  - [ ] `TestTenantPolicy_CustomHeaders` - Custom model/tool headers
  - [ ] `TestTenantPolicy_CustomOnDenied` - Custom denial handler
  - [ ] `TestTenantPolicy_Hooks` - Callback invocation
  - [ ] `TestTenantPolicy_NilEvaluator` - Graceful degradation

**Checkpoint**: ✅ `go test -v ./middleware/*tenant*_test.go` - All pass

---

## 🔌 Phase 2: Core Integration (2.5 hours) - CRITICAL

### Day 2 Morning: Core Options (2 hours)

- [ ] **core/options.go** - Add tenant options
  - [ ] Implement `WithTenantConfigManager(manager ConfigManager) Option`
  - [ ] Implement `TenantMiddlewareOptions` struct
  - [ ] Implement `WithTenantMiddleware(options TenantMiddlewareOptions) Option`
  - [ ] Handle nil manager gracefully
  - [ ] Integrate with existing middleware chain

- [ ] **core/tenant_options_test.go** - Test new options
  - [ ] `TestWithTenantConfigManager` - Component registration
  - [ ] `TestWithTenantMiddleware_Full` - All features enabled
  - [ ] `TestWithTenantMiddleware_QuotaOnly` - Just quota
  - [ ] `TestWithTenantMiddleware_PolicyOnly` - Just policy
  - [ ] `TestTenantOptions_Integration` - End-to-end test

**Checkpoint**: ✅ Example usage compiles:
```go
app := core.New(
    core.WithTenantConfigManager(manager),
    core.WithTenantMiddleware(core.TenantMiddlewareOptions{...}),
)
```

---

### Day 2 Morning: Public Exports (30 min)

- [ ] **plumego.go** - Export tenant types
  - [ ] Add type aliases for tenant.Config, QuotaConfig, PolicyConfig
  - [ ] Add type aliases for interfaces (ConfigManager, QuotaManager, PolicyEvaluator)
  - [ ] Add function aliases (NewInMemory*, TenantIDFromContext)
  - [ ] Add error variables (ErrTenantNotFound, ErrQuotaExceeded, ErrPolicyDenied)
  - [ ] Verify GoDoc shows tenant types

**Checkpoint**: ✅ `go doc plumego | grep -i tenant` shows exports

---

## 💾 Phase 3: Persistent Storage (4 hours) - HIGH PRIORITY

### Day 2 Afternoon: Database Config Manager (4 hours)

- [ ] **store/db/tenant_config.go** - DB implementation (2.5 hours)
  - [ ] Define `DBTenantConfigManager` struct
  - [ ] Implement `NewDBTenantConfigManager(db *sql.DB, options...)`
  - [ ] Implement `GetTenantConfig(ctx, tenantID) (Config, error)`
  - [ ] Implement `SetTenantConfig(ctx, cfg Config) error`
  - [ ] Implement `DeleteTenantConfig(ctx, tenantID) error`
  - [ ] Implement `ListTenants(ctx, limit, offset) ([]Config, error)`
  - [ ] Add optional LRU caching layer
  - [ ] Implement `QuotaConfig()` and `PolicyConfig()` interfaces

- [ ] **store/db/migrations/001_create_tenants.sql** - Schema (30 min)
  ```sql
  CREATE TABLE tenants (
      id VARCHAR(255) PRIMARY KEY,
      quota_requests_per_minute INT DEFAULT 0,
      quota_tokens_per_minute INT DEFAULT 0,
      allowed_models TEXT,
      allowed_tools TEXT,
      metadata TEXT,
      created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
      updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
  );
  ```
  - [ ] Add indexes
  - [ ] Add migration down script

- [ ] **store/db/tenant_config_test.go** - Tests (1 hour)
  - [ ] `TestDBTenantConfigManager_Create`
  - [ ] `TestDBTenantConfigManager_Get`
  - [ ] `TestDBTenantConfigManager_Update`
  - [ ] `TestDBTenantConfigManager_Delete`
  - [ ] `TestDBTenantConfigManager_List`
  - [ ] `TestDBTenantConfigManager_NotFound`
  - [ ] `TestDBTenantConfigManager_Cache`
  - [ ] `TestDBTenantConfigManager_Concurrent`
  - [ ] Use in-memory SQLite for tests

**Checkpoint**: ✅ All DB tests pass with SQLite

---

## 🔐 Phase 4: Database Isolation (4 hours) - HIGH PRIORITY

### Day 3 Morning: Tenant-Aware Queries (4 hours)

- [ ] **store/db/tenant_isolation.go** - Isolation helpers (2 hours)
  - [ ] Define `TenantDB` struct
  - [ ] Implement `QueryContext(ctx, tenantID, query, args...) (*Rows, error)`
  - [ ] Implement `ExecContext(ctx, tenantID, query, args...) (Result, error)`
  - [ ] Implement `QueryFromContext(ctx, query, args...) (*Rows, error)`
  - [ ] Implement `AddTenantFilter(query, tenantID, column) (string, []any)`
  - [ ] Implement `ValidateQuery(query, column) error` - safety check
  - [ ] Support opt-out for admin queries

- [ ] **store/db/tenant_isolation_test.go** - Tests (1 hour)
  - [ ] `TestTenantDB_Query` - SELECT with filter
  - [ ] `TestTenantDB_Exec` - UPDATE/DELETE with filter
  - [ ] `TestTenantDB_FromContext` - Extract from context
  - [ ] `TestAddTenantFilter` - SQL injection safety
  - [ ] `TestValidateQuery` - Detect missing filters

- [ ] **store/db/sharding/tenant_sharding.go** - Sharding integration (1 hour)
  - [ ] Define `TenantShardRouter` struct
  - [ ] Implement `ShardForTenant(tenantID) (*Shard, error)`
  - [ ] Implement `QueryTenant(ctx, tenantID, query, args...) (*Rows, error)`
  - [ ] Use consistent hashing by tenant ID

**Checkpoint**: ✅ Tenant queries never leak across tenants (integration test)

---

## 📚 Phase 5: Example Application (3 hours) - HIGH PRIORITY

### Day 3 Afternoon & Day 4 Morning: Multi-Tenant SaaS Example (3 hours)

- [ ] **examples/multi-tenant-saas/** - Create directory structure
  - [ ] Create `main.go` - Application entry point
  - [ ] Create `tenant_handler.go` - Admin tenant CRUD
  - [ ] Create `api_handler.go` - Tenant-scoped API
  - [ ] Create `admin_handler.go` - Admin endpoints
  - [ ] Create `schema.sql` - Database schema
  - [ ] Create `README.md` - Setup guide
  - [ ] Create `.env.example` - Configuration template

- [ ] **main.go** - Core application (1 hour)
  - [ ] Setup database connection
  - [ ] Initialize tenant config manager (DB-backed)
  - [ ] Initialize quota manager
  - [ ] Initialize policy evaluator
  - [ ] Configure app with `WithTenant*()` options
  - [ ] Add hooks for logging
  - [ ] Register all routes

- [ ] **tenant_handler.go** - Admin endpoints (30 min)
  - [ ] `POST /admin/tenants` - Create tenant
  - [ ] `GET /admin/tenants/:id` - Get tenant
  - [ ] `PUT /admin/tenants/:id` - Update tenant
  - [ ] `DELETE /admin/tenants/:id` - Delete tenant
  - [ ] `GET /admin/tenants` - List tenants (paginated)
  - [ ] `PUT /admin/tenants/:id/quota` - Update quota
  - [ ] `PUT /admin/tenants/:id/policy` - Update policy

- [ ] **api_handler.go** - Tenant API (30 min)
  - [ ] `GET /api/v1/resources` - List (filtered by tenant)
  - [ ] `POST /api/v1/resources` - Create (tagged with tenant)
  - [ ] `GET /api/v1/resources/:id` - Get (validated for tenant)
  - [ ] `PUT /api/v1/resources/:id` - Update (validated)
  - [ ] `DELETE /api/v1/resources/:id` - Delete (validated)

- [ ] **README.md** - Documentation (30 min)
  - [ ] Overview
  - [ ] Setup instructions
  - [ ] Environment variables
  - [ ] API documentation
  - [ ] Example requests (curl)
  - [ ] Testing guide

- [ ] **Testing** (30 min)
  - [ ] Manual test: Create tenant
  - [ ] Manual test: API calls with tenant header
  - [ ] Manual test: Quota enforcement
  - [ ] Manual test: Policy enforcement
  - [ ] Verify cross-tenant isolation

**Checkpoint**: ✅ Full example runs and demonstrates all features

---

## 📖 Phase 6: Documentation (2 hours) - HIGH PRIORITY

### Day 4 Afternoon: Comprehensive Documentation (2 hours)

- [ ] **tenant/README.md** - Package documentation (1 hour)
  - [ ] Overview & Use Cases
  - [ ] Core Concepts
  - [ ] Quick Start
  - [ ] Configuration Options
  - [ ] Middleware Integration
  - [ ] Storage Backends
  - [ ] Best Practices
  - [ ] Troubleshooting
  - [ ] API Reference

- [ ] **docs/MULTI_TENANT_ARCHITECTURE.md** - Architecture guide (30 min)
  - [ ] Multi-Tenancy Patterns
  - [ ] Data Isolation Strategies
  - [ ] Quota & Rate Limiting
  - [ ] Security Considerations
  - [ ] Performance Optimization
  - [ ] Monitoring & Observability
  - [ ] Migration Strategies

- [ ] **Update existing docs** (30 min)
  - [ ] README.md - Add Multi-Tenancy section
  - [ ] README_CN.md - Add Chinese translation
  - [ ] CLAUDE.md - Add tenant module to structure
  - [ ] CLAUDE.md - Update stability matrix
  - [ ] CHANGELOG.md - Document v1.1 changes

**Checkpoint**: ✅ All documentation complete and reviewed

---

## ⚡ Phase 7: Advanced Features (7 hours) - MEDIUM PRIORITY

### Day 5: Advanced Enhancements

- [ ] **tenant/quota_sliding.go** - Sliding window (3 hours)
  - [ ] Define `SlidingWindowQuotaManager` struct
  - [ ] Implement sub-window tracking
  - [ ] Implement `Allow()` with weighted calculation
  - [ ] Add tests for smooth rate limiting
  - [ ] Benchmark vs fixed window

- [ ] **store/cache/tenant_cache.go** - Cache isolation (2 hours)
  - [ ] Define `TenantCache` wrapper
  - [ ] Implement `Get(ctx, key)` with auto-prefixing
  - [ ] Implement `Set(ctx, key, value, ttl)` with auto-prefixing
  - [ ] Implement `Delete(ctx, key)` with auto-prefixing
  - [ ] Add tests for isolation
  - [ ] Integration test with existing cache

- [ ] **metrics/tenant_metrics.go** - Metrics (2 hours)
  - [ ] Define `TenantMetrics` struct
  - [ ] Add Prometheus counters/histograms by tenant_id
  - [ ] Implement `RecordRequest(ctx, duration)`
  - [ ] Implement `RecordQuotaRejection(ctx)`
  - [ ] Implement `RecordPolicyDenial(ctx)`
  - [ ] Add example Grafana dashboard queries

- [ ] **tenant/admin.go** - Admin service (OPTIONAL)
  - [ ] Define `AdminService` struct
  - [ ] Implement CRUD operations
  - [ ] Add validation logic
  - [ ] Add audit logging

---

## ✅ Final Verification (Before Release)

### Code Quality
- [ ] Run `go test -race -cover ./...` - All pass, coverage ≥ 80%
- [ ] Run `go vet ./...` - No warnings
- [ ] Run `gofmt -w .` - All formatted
- [ ] Run `staticcheck ./...` - No issues (if available)
- [ ] Check for TODO comments - None blocking

### Documentation
- [ ] README.md complete and accurate
- [ ] tenant/README.md complete
- [ ] Example application documented
- [ ] CHANGELOG.md updated
- [ ] Migration guide written

### Integration
- [ ] Core options work
- [ ] Exports in plumego.go verified
- [ ] Example application runs
- [ ] All middleware functional
- [ ] Database migrations tested

### Performance
- [ ] Quota check < 1ms (in-memory)
- [ ] Config lookup < 5ms (with cache)
- [ ] No memory leaks (long-running test)
- [ ] Benchmarks documented

---

## 📦 Release Checklist

- [ ] Update CHANGELOG.md with v1.1.0 notes
- [ ] Update README.md badge (v1.0.0-rc.1 → v1.1.0)
- [ ] Update README_CN.md badge
- [ ] Update CLAUDE.md version header
- [ ] Mark tenant package as **Stable** in docs
- [ ] Tag release: `git tag v1.1.0`
- [ ] Push tag: `git push origin v1.1.0`
- [ ] Create GitHub release with notes
- [ ] Announce in community channels

---

## Progress Tracking

**Current Status**: Planning Complete ✅

### Quick Stats
- Total Tasks: ~120
- Completed: 0
- In Progress: 0
- Remaining: ~120
- Estimated Hours: 30h
- Estimated Days: 3-4

### Phase Status
- ⏸️  Phase 1: Testing & Stability (0%)
- ⏸️  Phase 2: Core Integration (0%)
- ⏸️  Phase 3: Persistent Storage (0%)
- ⏸️  Phase 4: Database Isolation (0%)
- ⏸️  Phase 5: Example Application (0%)
- ⏸️  Phase 6: Documentation (0%)
- ⏸️  Phase 7: Advanced Features (0%)

---

## Notes & Decisions

### Design Decisions
- Use fixed-window for initial quota (simple, predictable)
- DB schema supports both PostgreSQL and MySQL
- Sliding window is optional advanced feature
- Cache isolation via key prefixing (simple, effective)

### Trade-offs
- Fixed window: Simpler but allows bursts at boundaries
- In-memory default: Fast but not distributed
- Hooks: Flexible but require user implementation

### Future Enhancements (v1.2+)
- Redis-backed quota manager (distributed)
- Advanced RBAC/ABAC policies
- Tenant usage analytics dashboard
- Automated tenant onboarding workflows
- Multi-region tenant routing

---

**Last Updated**: 2026-02-03
**Status**: Ready to begin implementation
**Next Action**: Start Phase 1 - Testing & Stability
