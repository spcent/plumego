# Card 2026

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: reference/workerfleet/internal/app
Priority: P0
State: done
Primary Module: reference/workerfleet/internal/app
Owned Files:
- `reference/workerfleet/internal/app/bootstrap.go`
- `reference/workerfleet/internal/app/app.go`
- `reference/workerfleet/internal/app/types.go`
- `reference/workerfleet/internal/app/runtime_loops.go`
- `reference/workerfleet/internal/app/alert_loop.go`
Depends On:
- `tasks/cards/done/2025-workerfleet-case-step-prometheus-metrics.md`

## Goal

Split the current workerfleet `Runtime` into a lifecycle shell plus explicit ingest, query, loop, and alert runner components so bootstrap wiring stays maintainable as the app grows.

## Scope

- Refactor `internal/app` so `Runtime` no longer acts as the default holder for every business dependency.
- Introduce explicit component boundaries for query service, loop runner, and alert runner.
- Keep the HTTP API, config surface, and metrics names behaviorally unchanged.

## Non-goals

- Do not change workerfleet HTTP routes.
- Do not change MongoDB schema or query results.
- Do not add new runtime loop behavior such as timeout, retry, or leasing in this card.

## Files

- `reference/workerfleet/internal/app/bootstrap.go`
- `reference/workerfleet/internal/app/app.go`
- `reference/workerfleet/internal/app/types.go`
- `reference/workerfleet/internal/app/runtime_loops.go`
- `reference/workerfleet/internal/app/alert_loop.go`

## Acceptance Tests

- `reference/workerfleet/internal/app/bootstrap_test.go: TestBootstrapBuildsSplitRuntimeShell`
- `reference/workerfleet/internal/app/runtime_loops_test.go: TestLoopRunnerUsesInjectedRuntimeDependencies`
- `reference/workerfleet/internal/app/alert_loop_test.go: TestAlertRunnerUsesInjectedDependencies`

## Tests

- Keep existing bootstrap, loop, and alert-loop tests green after the refactor.
- Add a regression test that `Runtime.Ready` still checks the configured store.

## Docs Sync

- `reference/workerfleet/docs/design/technical-design.md`
- `reference/workerfleet/docs/design/technical-design.zh-CN.md`

## Validation

- `cd reference/workerfleet && go test ./internal/app/...`
- `cd reference/workerfleet && go vet ./internal/app/...`
- `gofmt -l reference/workerfleet/internal/app`

## Done Definition

- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l . produces no output.
- [ ] Docs Sync targets updated (if applicable).

## Outcome

- Refactored `internal/app` so `Runtime` now acts as a lifecycle shell while explicit `LoopRunner` and `AlertRunner` own periodic loop behavior and injected dependencies.
- Kept the HTTP-facing `Service`, ready check, metrics handler, and route wiring behaviorally unchanged while moving internal dependency ownership under `runtimeShell`.
- Added regression coverage for bootstrap shell wiring, ready checks, loop dependency injection, and alert dependency injection.

Validation Run:

- `cd reference/workerfleet && env GOCACHE=/Users/bingrong.yan/projects/go/plumego/.tmp-gocache go test ./internal/app/...`
- `cd reference/workerfleet && env GOCACHE=/Users/bingrong.yan/projects/go/plumego/.tmp-gocache go vet ./internal/app/...`
- `cd reference/workerfleet && env GOCACHE=/Users/bingrong.yan/projects/go/plumego/.tmp-gocache go test ./...`
