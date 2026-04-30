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
  structured `contract` errors, including canonical status, code, type,
  category, and resource/method details.
- Query parsing and pagination boundary behavior are covered, including invalid
  page input, max page-size clamping, unknown filters, and unknown sort fields.
- Runnable offline example coverage exists in `x/rest/example_test.go`.

## Beta Sample Matrix

| Surface | Evidence | Status |
| --- | --- | --- |
| Route registration | `x/rest/routes_test.go`, `x/rest/entrypoints_test.go` | covered |
| Resource spec defaults | `x/rest/spec_test.go`, `x/rest/spec_apply_test.go` | covered |
| Repository-backed controller | `x/rest/resource_db_test.go` | covered |
| Error semantics | `TestBaseResourceControllerUsesContractNotImplementedError` | covered |
| Pagination and query parsing | `x/rest/entrypoints_test.go` | covered |
| Offline usage example | `x/rest/example_test.go` | covered |

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

## Candidate Promotion Checklist

When release refs exist, complete this sequence in one promotion card:

1. Compare the two refs with `internal/checks/extension-release-evidence`.
2. Check in the generated release snapshots under
   `docs/extension-evidence/snapshots/x-rest/`.
3. Update `specs/extension-beta-evidence.yaml` with both release refs and
   snapshot paths.
4. Replace blocker entries only for evidence that is actually complete.
5. Record `platform-api` owner sign-off below.
6. Only then update `x/rest/module.yaml` status from `experimental` to `beta`.

Use `docs/extension-evidence/BETA_EVIDENCE_TEMPLATE.md` for any future
extension candidate evidence document.

## Owner Sign-Off

Missing. The `platform-api` owner must confirm the beta criteria before any
`module.yaml` status change.

Required statement:

```text
I confirm that x/rest meets the beta criteria in
docs/EXTENSION_STABILITY_POLICY.md and accept the beta compatibility
obligations for the documented x/rest public surface.
```

## Blockers

- `release_history_missing`
- `api_snapshot_missing`
- `owner_signoff_missing`

## Promotion Decision

Do not promote yet. `x/rest` remains `experimental`.
