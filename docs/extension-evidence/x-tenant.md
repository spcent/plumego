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

Partially recorded. Promotion requires two consecutive minor release refs with
no exported `x/tenant/*` API changes. The `v1.0.0` tag target is the first
post-v1 release-ref intake point only; it does not clear
`release_history_missing` by itself.

Release refs:

- `6a99c5e0bc61c12378bcdab5a6a7c4d756b9fa96` (`v1.0.0` tag target)

## API Snapshot Evidence

One current-head baseline snapshot is recorded. It is useful for comparing the
candidate surface during development, but it is not release evidence and does
not clear `api_snapshot_missing` by itself.

Generate a fresh snapshot with:

```bash
go run ./internal/checks/extension-api-snapshot -module ./x/tenant/... -out /tmp/plumego-x-tenant-api.snapshot
```

Snapshot refs:

- `docs/extension-evidence/snapshots/first-batch/x-tenant-head.snapshot`

v1 baseline intake artifacts:

- `docs/extension-evidence/snapshots/v1-baseline/x-tenant/base.snapshot`
- `docs/extension-evidence/snapshots/v1-baseline/x-tenant/head.snapshot`

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

Do not clear `release_history_missing` or `api_snapshot_missing` until two
release refs and release-backed snapshot evidence are recorded.

## Release Evidence

First release-ref intake recorded.

Release refs: `v1.0.0`

API snapshot comparison: `v1.0.0` to `v1.0.0`, unchanged

## Owner Sign-Off

Missing. The `multitenancy` owner must confirm the beta criteria before any
`module.yaml` status change.

## Blockers

- `release_history_missing`
- `api_snapshot_missing`
- `owner_signoff_missing`

## Promotion Posture

Do not promote yet. `x/tenant` remains `experimental`.
