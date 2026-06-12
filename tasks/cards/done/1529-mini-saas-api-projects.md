# Card 1529

Milestone: M-025
Recipe: (none — projects capability)
Context Package: use-cases/mini-saas-api
Priority: P1
State: done
Primary Module: use-cases/mini-saas-api
Owned Files: internal/handler/{projects,idempotency}.go, internal/platform/idemstore, route wiring
Depends On: 1527

## Goal

Projects CRUD works through an x/rest ResourceController with tenant-isolation
404s, member/admin role split, plan limits, and Idempotency-Key replay
protection (stable store/idempotency contract) on every mutating route.

## Scope

x/rest controller in internal/handler; in-memory HashAwareStore implementation
of store/idempotency in internal/platform/idemstore; Idempotent wrapper;
route wiring for /api/v1/projects*.

## Non-goals

No batch endpoints (Base controller answers 501), no metrics (1530),
no x/data/idempotency (stable root contract preferred).

## Files

internal/handler/projects.go
internal/handler/idempotency.go
internal/platform/idemstore/idemstore.go
internal/app/routes.go
internal/app/app.go

## Acceptance Tests

internal/app/project_flow_test.go: TestAcceptanceProjectCRUD
internal/app/project_flow_test.go: TestAcceptanceIdempotentCreate
internal/app/project_flow_test.go: TestProjectCrossTenantIsolation

## Tests

Member-cannot-delete 403; plan limit 429; idempotency payload mismatch 400;
keys tenant-scoped (no cross-tenant replay); no header → no dedup.

## Docs Sync

AGENTS.md (x/rest allowed in handler for controllers).

## Validation

cd use-cases/mini-saas-api && go vet ./... && go test -race -timeout 180s ./...
gofmt -l use-cases/mini-saas-api

## Done Definition

- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated.

## Outcome

ProjectsController embeds rest.BaseResourceController (x/rest beta) and
overrides Index/Show/Create/Update/Delete; every service call passes the
caller's tenant ID. Idempotent wrapper buffers responses, stores them via the
stable store/idempotency contract (in-memory HashAwareStore in
internal/platform/idemstore), replays on same key+payload, 400s on payload
mismatch, 409s while in flight, and releases the key on 5xx. Keys scoped
tenant|method|path|client-key. Delete requires admin role.
