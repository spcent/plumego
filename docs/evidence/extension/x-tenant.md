# x/tenant Beta Evidence

Module: `x/tenant`

Owner: `multitenancy`

Current status: `beta`

Candidate status: `beta`

Evidence state: complete

## Current Coverage

- Resolution examples cover principal-first and custom extractor flows.
- End-to-end middleware coverage exercises resolve, policy, quota, and
  rate-limit behavior in sequence.
- Negative paths cover missing tenant identity, policy deny, quota exhaustion,
  rate limiting, and tenant isolation.
- Tenant-aware store/db coverage documents and tests the supported fail-closed
  query-scoping subset.

## Primer And Boundary State

- Primer: `docs/modules/x/tenant/README.md`
- Manifest: `x/tenant/module.yaml`
- Boundary state: documented and aligned with keeping tenant semantics out of
  stable `middleware` and stable `store`.

## Required Release Evidence

Recorded. Promotion evidence uses two consecutive minor release refs with no
exported `x/tenant/*` API changes.

Release refs:

- `6a99c5e0bc61c12378bcdab5a6a7c4d756b9fa96` (`v1.0.0` tag target)
- `v1.1.0`

## API Snapshot Evidence

Release-backed API snapshots are recorded for the promotion pair below. The
current-head baseline snapshot remains useful during development, but the
release-backed comparison is the promotion evidence.

Generate a fresh snapshot with:

```bash
go run ./internal/checks/extension-api-snapshot -module ./x/tenant/... -out /tmp/plumego-x-tenant-api.snapshot
```

Snapshot refs:

- `docs/evidence/extension/snapshots/first-batch/x-tenant-head.snapshot`
- `docs/evidence/extension/snapshots/x-tenant/base.snapshot`
- `docs/evidence/extension/snapshots/x-tenant/head.snapshot`

v1 baseline intake artifacts:

- `docs/evidence/extension/snapshots/v1-baseline/x-tenant/base.snapshot`
- `docs/evidence/extension/snapshots/v1-baseline/x-tenant/head.snapshot`

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

Release refs: `v1.0.0`, `v1.1.0`

API snapshot comparison:

- Base: `docs/evidence/extension/snapshots/x-tenant/base.snapshot`
- Head: `docs/evidence/extension/snapshots/x-tenant/head.snapshot`
- Result: **API unchanged** across both refs

## Owner Sign-Off

Signed off by `multitenancy` for v1.1.0:

> I confirm that `x/tenant` meets the beta criteria in
> docs/reference/extension-stability-policy.md and accept the beta compatibility
> obligations for the documented public surface.

## Blockers

None. All promotion blockers cleared.

## Promotion Posture

Promoted to `beta` at v1.1.0. API unchanged across `v1.0.0` to `v1.1.0`.
