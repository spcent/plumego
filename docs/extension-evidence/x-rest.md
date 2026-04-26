# x/rest Beta Evidence

Module: `x/rest`

Owner: `platform-api`

Current status: `experimental`

Candidate status: `beta`

Evidence state: incomplete

## Current Coverage

- CRUD route registration surface is covered, including all canonical resource
  routes and selective route options.
- Default controller negative paths cover not-implemented methods with
  structured `contract` errors.
- Query parsing and pagination boundary behavior are covered, including invalid
  page input, max page-size clamping, unknown filters, and unknown sort fields.
- Runnable offline example coverage exists in `x/rest/example_test.go`.

## Primer And Boundary State

- Primer: `docs/modules/x-rest/README.md`
- Manifest: `x/rest/module.yaml`
- Boundary state: documented and aligned with the current transport-only
  resource-controller role.

## Required Release Evidence

Missing. Promotion requires two consecutive minor release refs with no exported
`x/rest` API changes.

Release refs:

- none recorded

## API Snapshot Evidence

One current-head baseline snapshot is recorded. It is useful for comparing the
candidate surface during development, but it is not release evidence and does
not clear `api_snapshot_missing` by itself.

Generate a fresh snapshot with:

```bash
go run ./internal/checks/extension-api-snapshot -module ./x/rest/... -out /tmp/plumego-x-rest-api.snapshot
```

Snapshot refs:

- `docs/extension-evidence/snapshots/first-batch/x-rest-head.snapshot`

## Release Comparison Workflow

Use the release-aware evidence tool when two concrete release refs are
available:

```bash
go run ./internal/checks/extension-release-evidence \
  -module ./x/rest/... \
  -base <older-minor-release-ref> \
  -head <newer-minor-release-ref> \
  -out-dir /tmp/plumego-x-rest-release-evidence
```

Do not clear `release_history_missing` or `api_snapshot_missing` until the
recorded refs and snapshot files come from real releases.

## Owner Sign-Off

Missing. The `platform-api` owner must confirm the beta criteria before any
`module.yaml` status change.

## Blockers

- `release_history_missing`
- `api_snapshot_missing`
- `owner_signoff_missing`

## Promotion Decision

Do not promote yet. `x/rest` remains `experimental`.
