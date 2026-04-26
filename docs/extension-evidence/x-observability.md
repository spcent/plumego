# x/observability Beta Evidence

Module: `x/observability`

Owner: `observability`

Current status: `experimental`

Candidate status: `beta`

Evidence state: incomplete

## Current Coverage

- Prometheus collector and exporter coverage includes handler output format,
  content type, empty collector behavior, label escaping, memory bounding, and
  concurrency.
- Tracer coverage includes span lifecycle, status classification, attributes,
  parent trace propagation, stats, and bounded test cleanup behavior.
- Configuration coverage includes metrics and tracing enabled paths, namespace,
  service name, max series, custom paths, and mutable callback invocation.
- Supporting packages such as `recordbuffer`, `windowmetrics`, `testmetrics`,
  `testlog`, `tracer`, `featuremetrics`, and `dbinsights` have dedicated tests.

## Primer And Boundary State

- Primer: `docs/modules/x-observability/README.md`
- Manifest: `x/observability/module.yaml`
- Boundary state: documented and aligned with keeping request-scoped transport
  observability in stable middleware while exporter and adapter wiring stays in
  `x/observability`.

## Required Release Evidence

Missing. Promotion requires two consecutive minor release refs with no exported
`x/observability/*` API changes.

Release refs:

- none recorded

## API Snapshot Evidence

Missing. Generate snapshots with:

```bash
go run ./internal/checks/extension-api-snapshot -module ./x/observability/... -out /tmp/plumego-x-observability-api.snapshot
```

Snapshot refs:

- none recorded

## Owner Sign-Off

Missing. The `observability` owner must confirm the beta criteria before any
`module.yaml` status change.

## Blockers

- `release_history_missing`
- `api_snapshot_missing`
- `owner_signoff_missing`

## Promotion Decision

Do not promote yet. `x/observability` remains `experimental`.
