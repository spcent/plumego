# x/frontend Beta Evidence

Module: `x/frontend`

Owner: `frontend`

Current status: `experimental`

Candidate status: `beta`

Evidence state: incomplete

## Current Coverage

- Mount construction coverage includes directory-backed mounts,
  caller-provided `http.FileSystem` mounts, explicit `Mount` registration, nil
  registrar/filesystem handling, and missing directory/index startup failures.
- Path safety coverage includes traversal, encoded traversal, backslash
  traversal, dotted filenames, unsafe backend-open prevention, and directory
  symlink escape rejection.
- Response semantics coverage includes navigation-only SPA fallback, missing
  asset 404 behavior, HEAD, method restrictions, cache-control split, custom
  pages, MIME overrides, unsafe custom header rejection, and custom page cache
  isolation.
- Precompressed response coverage includes `.br` and `.gz` selection, quality
  ordering, wildcard handling, invalid quality values, `identity` refusal,
  orphan variant rejection, directory variant plans, and lazy `RegisterFS`
  probing.
- Basic benchmarks cover normal asset serving and precompressed asset serving.

## Primer And Boundary State

- Primer: `docs/modules/x-frontend/README.md`
- Manifest: `x/frontend/module.yaml`
- Boundary state: documented and aligned with keeping frontend asset policy in
  `x/frontend`, while stable `router` keeps only primitive static file mounts.

## Required Release Evidence

Missing. Promotion requires two consecutive minor release refs with no exported
`x/frontend` API changes.

Release refs:

- none recorded

## API Snapshot Evidence

One current-head baseline snapshot is recorded. It is useful for comparing the
candidate surface during development, but it is not release evidence and does
not clear `api_snapshot_missing` by itself.

Generate a fresh snapshot with:

```bash
go run ./internal/checks/extension-api-snapshot -module ./x/frontend -out /tmp/plumego-x-frontend-api.snapshot
```

Snapshot refs:

- `docs/extension-evidence/snapshots/first-batch/x-frontend-head.snapshot`

## Release Comparison Workflow

Use the release-aware evidence tool when two concrete release refs are
available:

```bash
go run ./internal/checks/extension-release-evidence \
  -module ./x/frontend \
  -base <older-minor-release-ref> \
  -head <newer-minor-release-ref> \
  -out-dir /tmp/plumego-x-frontend-release-evidence
```

Do not clear `release_history_missing` or `api_snapshot_missing` until the
recorded refs and snapshot files come from real releases.

## Owner Sign-Off

Missing. The `frontend` owner must confirm the beta criteria before any
`module.yaml` status change.

## Blockers

- `release_history_missing`
- `api_snapshot_missing`
- `owner_signoff_missing`

## Promotion Decision

Do not promote yet. `x/frontend` remains `experimental`.
