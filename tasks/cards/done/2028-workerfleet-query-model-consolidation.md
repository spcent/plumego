# Card 2028

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: reference/workerfleet/internal/handler
Priority: P1
State: done
Primary Module: reference/workerfleet/internal/handler
Owned Files:
- `reference/workerfleet/internal/handler/worker_query.go`
- `reference/workerfleet/internal/handler/worker_register.go`
- `reference/workerfleet/internal/app/service.go`
- `reference/workerfleet/internal/app/service_test.go`
- `reference/workerfleet/internal/handler/query_test.go`
Depends On:
- `tasks/cards/active/2026-workerfleet-runtime-shell-split.md`

## Goal

Define one stable read-model surface for worker detail, task detail, case timeline, and exec-plan drilldown so handler and app layers stop growing parallel DTO families.

## Scope

- Consolidate workerfleet query DTO ownership into one layer.
- Remove duplicated read-model structs and translation drift between handler and app.
- Keep route paths and JSON response shapes backward compatible.

## Non-goals

- Do not change storage interfaces or MongoDB document shape.
- Do not introduce frontend-specific response variants.
- Do not add new business fields beyond the already implemented API contract.

## Files

- `reference/workerfleet/internal/handler/worker_query.go`
- `reference/workerfleet/internal/handler/worker_register.go`
- `reference/workerfleet/internal/app/service.go`
- `reference/workerfleet/internal/app/service_test.go`
- `reference/workerfleet/internal/handler/query_test.go`

## Acceptance Tests

- `reference/workerfleet/internal/handler/query_test.go: TestGetCaseTimelineSuccess`
- `reference/workerfleet/internal/handler/query_test.go: TestListExecPlanCasesParsesFilters`
- `reference/workerfleet/internal/app/service_test.go: TestServiceGetTaskFallsBackToHistory`

## Tests

- Add a regression test that empty `current_step` is omitted from JSON responses.
- Keep query pagination tests green after DTO consolidation.

## Docs Sync

- `reference/workerfleet/docs/api.md`

## Validation

- `cd reference/workerfleet && go test ./internal/handler/...`
- `cd reference/workerfleet && go test ./internal/app/...`
- `gofmt -l reference/workerfleet/internal/handler reference/workerfleet/internal/app`

## Done Definition

- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l . produces no output.
- [ ] Docs Sync targets updated (if applicable).

## Outcome

- Promoted workerfleet query read models in `internal/app` into the single JSON-ready response contract for worker detail, task detail, case timeline, exec-plan drilldown, fleet summary, and alert list responses.
- Removed duplicate handler-side output DTOs and conversion helpers so query handlers now write app read models directly without translation drift.
- Added regression coverage that empty `current_step` is omitted from JSON responses and kept timeline/drilldown parsing behavior unchanged.

Validation Run:

- `cd reference/workerfleet && env GOCACHE=.tmp-gocache go test ./internal/handler/...`
- `cd reference/workerfleet && env GOCACHE=.tmp-gocache go test ./internal/app/...`
- `cd reference/workerfleet && env GOCACHE=
