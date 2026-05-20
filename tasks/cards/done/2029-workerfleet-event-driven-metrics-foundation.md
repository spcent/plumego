# Card 2029

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: reference/workerfleet/internal/platform/metrics
Priority: P1
State: done
Primary Module: reference/workerfleet/internal/platform/metrics
Owned Files:
- `reference/workerfleet/internal/platform/metrics/model.go`
- `reference/workerfleet/internal/platform/metrics/instrumentation.go`
- `reference/workerfleet/internal/platform/metrics/instrumentation_test.go`
- `reference/workerfleet/internal/domain/ingest_service.go`
- `reference/workerfleet/internal/domain/reconcile_service.go`
Depends On:
- `tasks/cards/done/2025-workerfleet-case-step-prometheus-metrics.md`

## Goal

Move workerfleet counters and histograms toward explicit domain-event instrumentation while keeping gauges snapshot-based, so metric semantics remain stable as the domain grows.

## Scope

- Separate state gauges from event-derived counters and histograms.
- Use task, step, and alert domain events as the preferred source for throughput and duration metrics.
- Preserve current metric names unless a bug fix absolutely requires a contract change.

## Non-goals

- Do not redesign the Prometheus collector implementation.
- Do not add per-case or per-worker labels.
- Do not change Grafana dashboards beyond metric semantics that actually change.

## Files

- `reference/workerfleet/internal/platform/metrics/model.go`
- `reference/workerfleet/internal/platform/metrics/instrumentation.go`
- `reference/workerfleet/internal/platform/metrics/instrumentation_test.go`
- `reference/workerfleet/internal/domain/ingest_service.go`
- `reference/workerfleet/internal/domain/reconcile_service.go`

## Acceptance Tests

- `reference/workerfleet/internal/platform/metrics/instrumentation_test.go: TestObserverRecordsFinishedTaskMetrics`
- `reference/workerfleet/internal/platform/metrics/instrumentation_test.go: TestObserverRecordsStepCompletionAndDurationMetrics`
- `reference/workerfleet/internal/platform/metrics/instrumentation_test.go: TestObserverRecordsAlertMetrics`

## Tests

- Add a regression test that unchanged snapshots do not double-count event-derived counters.
- Keep forbidden-label tests green.

## Docs Sync

- `reference/workerfleet/docs/metrics.md`
- `reference/workerfleet/docs/grafana.md`

## Validation

- `cd reference/workerfleet && go test ./internal/platform/metrics/...`
- `cd reference/workerfleet && go test ./internal/domain/...`
- `gofmt -l reference/workerfleet/internal/platform/metrics reference/workerfleet/internal/domain`

## Done Definition

- [ ] Acceptance Tests pass.
- [ ] All Validation commands exit 0.
- [ ] gofmt -l . produces no output.
- [ ] Docs Sync targets updated (if applicable).

## Outcome

- Split workerfleet metric instrumentation into snapshot-derived gauges and domain-event-derived counters/histograms while preserving existing metric names.
- Wired ingest heartbeat processing to publish domain events into metrics so task start, finish, phase transition, and case-step completion metrics no longer depend on snapshot diff guesses.
- Added regression coverage that unchanged snapshots do not double-count event-derived counters and updated metrics/Grafana docs to describe the new semantics.

Validation Run:

- `cd reference/workerfleet && env GOCACHE=.tmp-gocache go test ./internal/platform/metrics/...`
- `cd reference/workerfleet && env GOCACHE=.tmp-gocache go test ./internal/domain/...`
- `gofmt -l reference/workerfleet/internal/platform/metrics reference/workerfleet/internal/domain`
