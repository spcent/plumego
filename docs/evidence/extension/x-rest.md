# x/rest Beta Evidence

Module: `x/rest`

Owner: `platform-api`

Current status: `beta`

Evidence state: complete

## Current Coverage

- CRUD route registration surface is covered, including all canonical resource
  routes and selective route options.
- Default controller negative paths cover not-implemented methods with
  structured `contract` errors, including canonical status, code, type,
  category, and resource/method details.
- Query parsing and pagination boundary behavior are covered, including invalid
  page input, max page-size clamping, unknown filters, and unknown sort fields.
- Runnable offline example coverage exists in `x/rest/example_test.go`.

## Primer And Boundary State

- Primer: `docs/modules/x/rest/README.md`
- Manifest: `x/rest/module.yaml`
- Boundary state: documented and aligned with the current transport-only
  resource-controller role.

## Beta Sample Matrix

| Surface | Evidence | Status |
| --- | --- | --- |
| Route registration | `x/rest/routes_test.go`, `x/rest/entrypoints_test.go` | covered |
| Resource spec defaults | `x/rest/spec_test.go`, `x/rest/spec_apply_test.go` | covered |
| Repository-backed controller | `x/rest/resource_db_test.go` | covered |
| Error semantics | `TestBaseResourceControllerUsesContractNotImplementedError` | covered |
| Pagination and query parsing | `x/rest/entrypoints_test.go` | covered |
| Offline usage example | `x/rest/example_test.go` | covered |

## Required Release Evidence

Recorded. This promotion record uses two consecutive minor release refs with no
exported `x/rest` API changes.

Release refs:

- `d2c25c3`
- `ec70358`

## API Snapshot Evidence

Release-backed API snapshots are recorded for the promotion pair below. A
current-head baseline snapshot can still be useful during development, but it
does not replace the release-backed comparison.

Generate a fresh snapshot with:

```bash
go run ./internal/checks/extension-api-snapshot -module ./x/rest/... -out /tmp/plumego-x-rest-api.snapshot
```

Snapshot refs:

- `docs/evidence/extension/snapshots/x-rest/base.snapshot`
- `docs/evidence/extension/snapshots/x-rest/head.snapshot`

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

## Release Evidence

Release refs: `d2c25c3`, `ec70358`

API snapshot comparison:

- Base: `docs/evidence/extension/snapshots/x-rest/base.snapshot`
- Head: `docs/evidence/extension/snapshots/x-rest/head.snapshot`
- Result: **API unchanged** across both refs

## Owner Sign-Off

Signed off by `platform-api` at v0.2.0:

> I confirm that x/rest meets the beta criteria in
> docs/reference/extension-stability-policy.md and accept the beta compatibility
> obligations for the documented x/rest public surface.

## Historical Promotion Workflow

This is the workflow that was required before the promotion was recorded:

1. Compare the two refs with `internal/checks/extension-release-evidence`.
2. Check in the generated release snapshots under
   `docs/evidence/extension/snapshots/x-rest/`.
3. Update `specs/extension-beta-evidence.yaml` with both release refs and
   snapshot paths.
4. Replace blocker entries only for evidence that is actually complete.
5. Record `platform-api` owner sign-off above.
6. Only then update `x/rest/module.yaml` status from `experimental` to `beta`.

Use `docs/evidence/extension/BETA_EVIDENCE_TEMPLATE.md` for any future
extension candidate evidence document.

## Blockers

None. All promotion blockers cleared.

## Promotion Posture

Promoted to `beta` at v0.2.0. API stable across d2c25c3–ec70358.
