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

One current-head baseline snapshot is recorded. It is useful for comparing the
candidate surface during development, but it is not release evidence and does
not clear `api_snapshot_missing` by itself.

Generate a fresh snapshot with:

```bash
go run ./internal/checks/extension-api-snapshot -module ./x/observability/... -out /tmp/plumego-x-observability-api.snapshot
```

Snapshot refs:

- `docs/extension-evidence/snapshots/first-batch/x-observability-head.snapshot`

## Release Comparison Workflow

Use the release-aware evidence tool when two concrete release refs are
available:

```bash
go run ./internal/checks/extension-release-evidence \
  -module ./x/observability/... \
  -base <older-minor-release-ref> \
  -head <newer-minor-release-ref> \
  -out-dir /tmp/plumego-x-observability-release-evidence
```

Do not clear `release_history_missing` or `api_snapshot_missing` until the
recorded refs and snapshot files come from real releases.

## Owner Sign-Off

Missing. The `observability` owner must confirm the beta criteria before any
`module.yaml` status change.

## Blockers

- `release_history_missing`
- `api_snapshot_missing`
- `owner_signoff_missing`

## Promotion Decision

Do not promote yet. `x/observability` remains `experimental`.
