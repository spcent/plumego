# Card 2027

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: reference/workerfleet/internal/app
Priority: P0
State: done
Primary Module: reference/workerfleet/internal/app
Owned Files:
- `reference/workerfleet/internal/app/runtime_loops.go`
- `reference/workerfleet/internal/app/alert_loop.go`
- `reference/workerfleet/internal/app/types.go`
- `reference/workerfleet/internal/app/runtime_loops_test.go`
- `reference/workerfleet/internal/app/config.go`
Depends On:
- `tasks/cards/active/2026-workerfleet-runtime-shell-split.md`

## Goal

Add loop-level timeout, non-reentry, and retry-backoff guardrails so kube sync, status sweep, and alert evaluation remain predictable under slow or failing downstream dependencies.

## Scope

- Add explicit per-loop execution settings and defaults.
- Prevent overlapping executions of the same loop within one process.
- Add bounded retry/backoff behavior and observable result classification.
- Leave a lease-ready seam for future multi-replica coordination without implementing real leader election yet.

## Non-goals

- Do not add multi-cluster logic.
- Do not implement distributed leader election.
- Do not redesign metrics semantics outside loop runtime/result observability.

## Files

- `reference/workerfleet/internal/app/runtime_loops.go`
- `reference/workerfleet/internal/app/alert_loop.go`
- `reference/workerfleet/internal/app/types.go`
- `reference/workerfleet/internal/app/runtime_loops_test.go`
- `reference/workerfleet/internal/app/config.go`

## Acceptance Tests

- `reference/workerfleet/internal/app/runtime_loops_test.go: TestLoopRunnerPreventsOverlappingExecutions`
- `reference/workerfleet/internal/app/runtime_loops_test.go: TestLoopRunnerCancelsSlowIterationOnTimeout`
- `reference/workerfleet/internal/app/runtime_loops_test.go: TestLoopRunnerBacksOffAfterFailure`

## Tests

- Cover kube sync, status sweep, and alert evaluation loop settings independently.
- Add a regression test that disabled loops still never start.

## Docs Sync

- `reference/workerfleet/README.md`
- `reference/workerfleet/docs/design/technical-design.md`
- `reference/workerfleet/docs/design/technical-design.zh-CN.md`

## Validation

- `cd reference/workerfleet && go test ./internal/app/...`
- `cd reference/workerfleet && go test ./...`
- `gofmt -l reference/workerfleet/internal/app`

## Done Definition

- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l . produces no output.
- [ ] Docs Sync targets updated (if applicable).

## Outcome

- Added a shared guarded loop scheduler with same-process non-overlap, per-iteration timeout, bounded failure backoff, and a no-op lease seam for future multi-replica ownership.
- Switched kube sync, status sweep, and alert evaluation to explicit loop settings so each runtime loop now runs through the same guardrail path.
- Added acceptance coverage for overlap prevention, timeout cancellation, failure backoff, and per-loop setting independence.

Validation Run:

- `cd reference/workerfleet && env GOCACHE=/Users/bingrong.yan/projects/go/plumego/.tmp-gocache go test ./internal/app/...`
- `cd reference/workerfleet && env GOCACHE=/Users/bingrong.yan/projects/go/plumego/.tmp-gocache go test ./...`
- `gofmt -l reference/workerfleet/internal/app`
