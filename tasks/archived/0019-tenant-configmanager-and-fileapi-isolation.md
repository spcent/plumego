# Card 0019

Priority: P1
State: done
Primary Module: x/tenant
Owned Files:
  - x/tenant/core/config.go
  - x/fileapi/handler.go

Depends On: —

Goal:
Two problems both related to tenant-isolation correctness, grouped into one card:

**Problem 1: x/tenant.ConfigManager interface is too broad (core/config.go:52-61)**

`ConfigManager` embeds three independent provider interfaces into one large interface:

```go
type ConfigManager interface {
    QuotaConfigProvider
    PolicyConfigProvider
    RateLimitConfigProvider
    GetTenantConfig(ctx, tenantID) (Config, error)
}
```

Comments acknowledge this is a "convenience" design ("so a single implementation can be passed
directly"), but the cost is:
- Any new provider interface forces all ConfigManager implementors to implement it in lockstep
- Components holding ConfigManager only actually need one of the sub-interfaces, but are forced
  to depend on the entire facade, violating the principle of minimal interface authority
- InMemoryConfigManager is the sole implementation, making it an implicit monolith

Fix: keep ConfigManager as an app-layer facade, but have each subsystem (quota, policy,
ratelimit) depend on its own specific provider interface rather than ConfigManager, decoupling
them.

**Problem 2: x/fileapi handlers missing tenant context validation (handler.go)**

`Upload()` (line 53-100) reads tenantID from context and rejects the request when absent.
However, the following handlers had no tenant check at all:

- `Download()` (line 121): accessed via fileID with no check that the caller belongs to the
  file's tenant
- `GetInfo()` (line 163): same issue
- `Delete()` (line 187): same issue
- `List()` (line 210): filtered via query parameter `tenant_id` — not an authentication check

An attacker who knows any fileID could read or delete files across tenant boundaries.

Additionally, `Download()` line 141:
```go
go h.metadata.UpdateAccessTime(context.Background(), fileID)
```
The fire-and-forget goroutine used `context.Background()`, bypassing request-chain tracing and
preventing graceful exit on server shutdown.

Scope:
- **ConfigManager decoupling**:
  - Change subsystem function signatures to depend on the minimal interface
    (QuotaConfigProvider, PolicyConfigProvider, RateLimitConfigProvider) rather than the full
    ConfigManager
  - The ConfigManager interface itself remains unchanged (app layer can still pass a single
    implementation)
  - Add a doc comment at the top of core/config.go explaining: ConfigManager is an app-layer
    facade; subsystems should depend on their specific sub-interfaces
- **fileapi tenant isolation**:
  - `Download`, `GetInfo`, `Delete`: read tenantID from context (same as Upload); verify
    `FileMeta.TenantID` matches context; return 403 on mismatch
  - `List`: ignore query parameter `tenant_id`; force-read from context (eliminates
    cross-tenant enumeration)
  - `GetURL`: same logic as Download
  - Remove fire-and-forget goroutine; call UpdateAccessTime synchronously within the request
    context (or accept losing the update and flush on handler close); log a warning on failure
    rather than silently discarding

Non-goals:
- Do not change InMemoryConfigManager implementation logic
- Do not add new methods to ConfigManager
- Do not change the file storage backend (x/data/file)
- Do not modify fileapi route registration or URL structure

Files:
  - x/tenant/core/config.go (doc comment added; subsystem interface dependencies updated)
  - x/tenant/core/quota.go, policy.go, ratelimit.go (depend on specific provider interfaces)
  - x/fileapi/handler.go (Download/GetInfo/Delete/List/GetURL gain tenant validation;
    fire-and-forget goroutine removed)
  - x/fileapi/handler_test.go (cross-tenant access rejection test cases added)

Tests:
  - go test ./x/tenant/...
  - go test ./x/fileapi/...

Docs Sync: —

Done Definition:
- `grep -n "go h.metadata.UpdateAccessTime" x/fileapi/handler.go` returns empty
- `Download`, `GetInfo`, `Delete` all contain tenantID validation and return 403 on cross-tenant access
- `List` no longer reads `tenant_id` from query parameter as the filter source
- `go test ./x/fileapi/...` includes cross-tenant 403 tests and passes

Outcome:
- **fileapi security**: `Download`, `GetInfo`, `Delete`, `GetURL` now read `tenantID` from
  context, fetch `fileMeta`, verify `fileMeta.TenantID == tenantID` — return 403 Forbidden
  on mismatch. `List` forces `TenantID` from context instead of query param (prevents
  cross-tenant enumeration). Removed fire-and-forget goroutine in Download;
  UpdateAccessTime is now synchronous within the request context.
- **ConfigManager doc**: Added doc comment to `x/tenant/core/config.go` ConfigManager
  interface explaining it is an app-layer facade; subsystems should depend on their specific
  sub-interface.
- **Tests**: Added cross-tenant 403 tests for Download, GetInfo, Delete, GetURL, List.
  Added TestHandler_List_MissingTenant.
