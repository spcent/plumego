# x/tenant Beta Evidence

Module: `x/tenant`

Owner: `multitenancy`

Current status: `experimental`

Candidate status: `beta`

Evidence state: incomplete

## Current Coverage

- Resolution examples cover principal-first and custom extractor flows.
- End-to-end middleware coverage exercises resolve, policy, quota, and
  rate-limit behavior in sequence.
- Negative paths cover missing tenant identity, policy deny, quota exhaustion,
  rate limiting, and tenant isolation.
- Tenant-aware store/db coverage documents and tests the supported fail-closed
  query-scoping subset.

## Primer And Boundary State

- Primer: `docs/modules/x-tenant/README.md`
- Manifest: `x/tenant/module.yaml`
- Boundary state: documented and aligned with keeping tenant semantics out of
  stable `middleware` and stable `store`.

## Required Release Evidence

Missing. Promotion requires two consecutive minor release refs with no exported
`x/tenant/*` API changes.

Release refs:

- none recorded

## API Snapshot Evidence

Missing. Generate snapshots with:

```bash
go run ./internal/checks/extension-api-snapshot -module ./x/tenant/... -out /tmp/plumego-x-tenant-api.snapshot
```

Snapshot refs:

- none recorded

## Release Comparison Workflow

Use the release-aware evidence tool when two concrete release refs are
available:

```bash
go run ./internal/checks/extension-release-evidence \
  -module ./x/tenant/... \
  -base <older-minor-release-ref> \
  -head <newer-minor-release-ref> \
  -out-dir /tmp/plumego-x-tenant-release-evidence
```

Do not clear `release_history_missing` or `api_snapshot_missing` until the
recorded refs and snapshot files come from real releases.

## Owner Sign-Off

Missing. The `multitenancy` owner must confirm the beta criteria before any
`module.yaml` status change.

## Blockers

- `release_history_missing`
- `api_snapshot_missing`
- `owner_signoff_missing`

## Promotion Decision

Do not promote yet. `x/tenant` remains `experimental`.
