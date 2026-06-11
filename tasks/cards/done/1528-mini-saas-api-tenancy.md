# Card 1528

Milestone: M-025
Recipe: (none — tenancy capability)
Context Package: use-cases/mini-saas-api
Priority: P1
State: done
Primary Module: use-cases/mini-saas-api
Owned Files: internal/handler/{tenant,members}.go, internal/app/{routes,tenantlimits}.go, login-by-slug additions
Depends On: 1527

## Goal

The tenant admin surface works end to end with the x/tenant chain enforced:
workspace read/update, member add/promote/remove with the last-owner guard,
login-by-workspace-slug, and per-tenant rate limit + request quota.

## Scope

Tenant/member handlers, x/tenant wiring (resolve → ratelimit → quota) confined
to internal/app, uniform per-tenant limit providers, config knobs
APP_TENANT_RPS / APP_TENANT_BURST / APP_TENANT_QUOTA_PER_MINUTE.

## Non-goals

No projects CRUD (1529), no metrics/audit endpoint (1530), no per-tenant
custom limit storage, no x/tenant/policy or session packages.

## Files

internal/handler/tenant.go
internal/handler/members.go
internal/app/routes.go
internal/app/tenantlimits.go
internal/config/config.go

## Acceptance Tests

internal/app/tenant_flow_test.go: TestAcceptanceTenantAdminFlow
internal/app/tenant_flow_test.go: TestTenantLimitsArePerTenant

## Tests

Rate limit 429 after burst; quota 429 after window; member 403 on admin
actions; unknown email 404; last-owner removal 409; cross-tenant isolation.

## Docs Sync

env.example (tenant limit knobs).

## Validation

cd use-cases/mini-saas-api && go vet ./... && go test -race -timeout 120s ./...
gofmt -l use-cases/mini-saas-api

## Done Definition

- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated.

## Outcome

x/tenant beta chain wired per-route in internal/app only: resolve lifts
authn.Principal.TenantID into tenant context; TokenBucketRateLimiter and
FixedWindowQuotaManager enforce per-tenant state over uniform config
providers. Tenant admin endpoints with RequireRole(admin); login accepts
workspace_slug for multi-workspace users; member management re-derives roles
on token refresh. All flows covered by httptest acceptance tests.
