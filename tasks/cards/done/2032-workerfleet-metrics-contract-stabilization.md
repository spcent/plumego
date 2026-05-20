# Card 2032

Milestone: —
Recipe: specs/change-recipes/fix-bug.yaml
Context Package: reference/workerfleet/internal/platform/metrics
Priority: P2
State: done
Primary Module: reference/workerfleet/internal/platform/metrics
Owned Files:
- `reference/workerfleet/internal/platform/metrics/model.go`
- `reference/workerfleet/internal/platform/metrics/model_test.go`
- `reference/workerfleet/docs/metrics.md`
- `reference/workerfleet/docs/grafana.md`
- `reference/workerfleet/docs/case-step-metrics.md`
Depends On:
- `tasks/cards/active/2029-workerfleet-event-driven-metrics-foundation.md`

## Goal

Document and test a stable workerfleet metric contract so Grafana panels and alert rules can distinguish frozen metric surfaces from still-experimental ones.

## Scope

- Classify implemented workerfleet metrics as stable or experimental.
- Freeze allowed label sets in tests and docs.
- Align Grafana examples with the stabilized metric catalog only.

## Non-goals

- Do not change collector internals in this card.
- Do not introduce dashboard JSON exports.
- Do not add new metric families beyond contract classification.

## Files

- `reference/workerfleet/internal/platform/metrics/model.go`
- `reference/workerfleet/internal/platform/metrics/model_test.go`
- `reference/workerfleet/docs/metrics.md`
- `reference/workerfleet/docs/grafana.md`
- `reference/workerfleet/docs/case-step-metrics.md`

## Acceptance Tests

- `reference/workerfleet/internal/platform/metrics/model_test.go: TestCatalogNamesAreStableAndNamespaced`
- `reference/workerfleet/internal/platform/metrics/model_test.go: TestCatalogRejectsForbiddenLabels`
- `reference/workerfleet/internal/platform/metrics/model_test.go: TestStableMetricCatalogFreezesLabelContract`

## Tests

- Add a regression test that experimental metrics are explicitly flagged and cannot silently reuse forbidden labels.

## Docs Sync

- `reference/workerfleet/docs/metrics.md`
- `reference/workerfleet/docs/grafana.md`
- `reference/workerfleet/docs/case-step-metrics.md`

## Validation

- `cd reference/workerfleet && go test ./internal/platform/metrics/...`
- `cd reference/workerfleet && go test ./...`
- `gofmt -l reference/workerfleet/internal/platform/metrics`

## Done Definition

- [x] Acceptance Tests pass.
- [x] All Validation commands exit 0.
- [x] gofmt -l . produces no output.
- [x] Docs Sync targets updated (if applicable).

## Outcome

Added an explicit workerfleet metric catalog contract with stable versus experimental classification in code. Metric specs now carry stability metadata, validation covers namespacing, duplicate labels, and forbidden labels, stable metric label sets are frozen in tests, and docs now distinguish stabilized dashboard and alert metrics from experimental pod, exec-plan, and case-step drilldown series.
