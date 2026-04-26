# Extension Release Evidence Artifacts

This note defines how checked-in release evidence artifacts are recorded for
`x/*` beta candidates.

## Current Release Ref State

As of this card, the repository has no local or remote release tags visible to
the working copy:

```bash
git tag --sort=version:refname
git ls-remote --tags origin
```

Do not use branch heads, arbitrary commits, or `HEAD` as substitutes for
two-minor-release evidence. Until real release refs exist, candidates must keep
`release_history_missing`.

## Artifact Rules

- `release_refs` entries in `specs/extension-beta-evidence.yaml` must resolve
  to git commits.
- `api_snapshots` entries must be checked-in files under
  `docs/extension-evidence/snapshots/`.
- A current-head snapshot can be useful as a baseline artifact, but it does not
  satisfy the two-release requirement by itself.
- Promotion remains blocked until the ledger has two release refs, matching API
  snapshots, and owner sign-off.

Validate the ledger with:

```bash
go run ./internal/checks/extension-beta-evidence
```
