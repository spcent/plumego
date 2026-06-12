# Card 1526

Milestone: M-025
Recipe: (none — new use-case domain layer)
Context Package: use-cases/mini-saas-api
Priority: P1
State: done
Primary Module: use-cases/mini-saas-api
Owned Files: internal/domain/** (new)
Depends On: 1525

## Goal

The domain layer (`access`, `ident`, `user`, `tenantspace`, `project`, `audit`)
exists with in-memory repositories behind interfaces, tenant isolation by
construction, and the last-owner invariant — all covered by unit tests.

## Scope

- internal/domain/access — RBAC role lattice (owner > admin > member)
- internal/domain/ident — crypto/rand 128-bit hex IDs
- internal/domain/user — model, repository, service (register/authenticate, bcrypt via security/password)
- internal/domain/tenantspace — tenant + membership, slug uniqueness, last-owner guard
- internal/domain/project — tenant-scoped CRUD with per-plan count limits
- internal/domain/audit — append-only per-tenant ring recorder

## Non-goals

No HTTP handlers, no JWT, no x/* imports, no changes outside internal/domain.

## Files

internal/domain/user/{user,store,service}.go
internal/domain/tenantspace/{tenantspace,store,service}.go
internal/domain/project/{project,store,service}.go
internal/domain/audit/audit.go
internal/domain/{access/access,ident/ident}.go

## Acceptance Tests

internal/domain/tenantspace/service_test.go: TestLastOwnerCannotBeDemoted
internal/domain/project/service_test.go: TestTenantIsolation
internal/domain/user/service_test.go: TestRegisterAndAuthenticate

## Tests

Role lattice truth table; duplicate email/slug; weak password; invalid role;
plan limit; cross-tenant membership/project isolation; audit ring + ordering.

## Docs Sync

(none — internal layer; ARCHITECTURE.md already describes ownership)

## Validation

cd use-cases/mini-saas-api && go vet ./... && go test -race -timeout 60s ./internal/domain/...
gofmt -l use-cases/mini-saas-api

## Done Definition

- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (n/a).

## Outcome

Six domain packages created with mutex-guarded in-memory stores behind
repository interfaces. Tenant isolation enforced structurally (every repository
method takes tenantID); last-owner invariant in tenantspace.Service; per-plan
project limits in project.Service; uniform ErrInvalidCredentials for unknown
email vs wrong password. 6/6 packages pass go test -race.
