# Card 2034

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: reference/workerfleet/internal/platform/metrics
Priority: P0
State: done
Primary Module: reference/workerfleet/internal/platform/metrics
Owned Files:
- `reference/workerfleet/internal/platform/metrics/instrumentation.go`
- `reference/workerfleet/internal/platform/metrics/instrumentation_test.go`
- `reference/workerfleet/internal/app/config.go`
- `reference/workerfleet/internal/app/bootstrap.go`
- `reference/workerfleet/docs/metrics.md`
Depends On:
- `tasks/cards/done/2032-workerfleet-metrics-contract-stabilization.md`

## Goal

Allow workerfleet to disable experimental case and step metric series, especially `exec_plan_id`-bearing labels, through explicit app config.

## Scope

- Add config-controlled gating for experimental metric emission.
- Keep stable metric families always on.
- Update docs to describe the default and production recommendation.

## Non-goals

- Do not remove the experimental metric implementations.
- Do not redesign the stable metric catalog in this card.
- Do not add dashboard JSON or Grafana provisioning files.

## Files

- `reference/workerfleet/internal/platform/metrics/instrumentation.go`
- `reference/workerfleet/internal/platform/metrics/instrumentation_test.go`
- `reference/workerfleet/internal/app/config.go`
- `reference/workerfleet/internal/app/bootstrap.go`
- `reference/workerfleet/docs/metrics.md`

## Acceptance Tests

- `reference/workerfleet/internal/platform/metrics/instrumentation_test.go: TestObserverSkipsExperimentalSeriesWhenDisabled`
- `reference/workerfleet/internal/app/config_test.go: TestLoadConfigParsesExperimentalMetricFlags`
- `reference/workerfleet/internal/app/bootstrap_test.go: TestBootstrapUsesConfiguredMetricOptions`

## Tests

- Keep existing stable-metric assertions green when experimental metrics are disabled.
- Add default-value coverage so local development behavior is explicit.

## Docs Sync

- `reference/workerfleet/docs/metrics.md`
- `reference/workerfleet/README.md`
- `reference/workerfleet/docs/design/technical-design.md`
- `reference/workerfleet/docs/design/technical-design.zh-CN.md`

## Validation

- `cd reference/workerfleet && go test -timeout 20s ./internal/platform/metrics/...`
- `cd reference/workerfleet && go test -timeout 20s ./internal/app/...`
- `cd reference/workerfleet && go test -timeout 20s ./...`

## Done Definition

- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

## Outcome

Added an app-level `WORKERFLEET_EXPERIMENTAL_METRICS_ENABLED` switch and wired it into the workerfleet metrics observer. Stable fleet metrics remain always on, while experimental pod, exec plan, and case-step series now default enabled in `dev`, default disabled in `prod`, and can be overridden explicitly through env configuration.
