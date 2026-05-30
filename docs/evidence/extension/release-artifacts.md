# Extension Release Evidence Artifacts

This note defines how checked-in release evidence artifacts are recorded for
`x/*` `beta` candidates.

Stable-root API baseline artifacts live separately under
`docs/evidence/stable-api/snapshots/`. Use `docs/release/POST_V1_EVIDENCE.md` as the
first release evidence read after `v1.0.0`.

## Current Release Ref State

As of M-009, the working copy sees these local and remote release tags:

```bash
git tag --sort=version:refname
git ls-remote --tags origin
```

Known release refs:

| Tag | Target kind | Evidence role |
| --- | --- | --- |
| `v0.2.0` | lightweight tag | historical pre-v1 tag |
| `v1.0.0-rc.1` | annotated tag | release-candidate evidence |
| `v1.0.0` | annotated tag | first post-v1 baseline evidence point |
| `v1.1.0` | annotated tag | second post-v1 release evidence point |

Use the release tag target commit, not the annotated tag object, when recording
commit-hash release refs in `specs/extension-beta-evidence.yaml`. The ledger may
also record a release tag string such as `v1.1.0` when the tag resolves locally.

Do not use branch heads, arbitrary commits, or `HEAD` as substitutes for release
evidence. A single release ref is an intake artifact only; it does not satisfy
the two-release promotion rule by itself.

## Artifact Rules

- `release_refs` entries in `specs/extension-beta-evidence.yaml` must resolve
  to git commits.
- `api_snapshots` entries must be checked-in files under
  `docs/evidence/extension/snapshots/`.
- A current-head snapshot can be useful as a baseline artifact, but it does not
  satisfy the two-release requirement by itself.
- Do not append a v1 baseline snapshot to the ledger if doing so would clear
  `api_snapshot_missing` before release-backed snapshot evidence is complete.
  In that case, record the artifact in the evidence doc and keep the blocker.
- Promotion remains blocked until the ledger has two release refs, matching API
  snapshots, and owner sign-off.

Validate the ledger with:

```bash
go run ./internal/checks/extension-beta-evidence
```

## Release Evidence Gap Map

`go run ./internal/checks/extension-beta-evidence` currently validates the
blockers below. The rows map each candidate surface to its recorded evidence,
open blockers, and module-owned follow-up work without changing maturity
status.

| Candidate | Recorded evidence | Open blockers | Follow-up |
| --- | --- | --- | --- |
| `x/rest` | beta evidence complete | none | none |
| `x/websocket` | beta evidence complete | none | none |
| `x/observability` | beta evidence complete | none | none |
| `x/gateway` | beta evidence complete | none | none |
| `x/frontend` | beta evidence complete | none | none |
| `x/tenant` | beta evidence complete | none | none |
| `x/ai` stable-tier subpackages | beta evidence complete for provider, session, streaming, and tool | none | root `x/ai` remains experimental |
| `x/data` selected surfaces | beta evidence complete for `x/data/file` and `x/data/idempotency` | none for selected surfaces | root `x/data` remains experimental |
| `x/gateway/discovery` core/static surface | evidence doc, first `v1.0.0` release ref, v1 baseline intake snapshot | second release ref, complete release-backed snapshots, owner sign-off | `tasks/cards/blocked/1372-x-discovery-surface-beta-evidence-closure.md` |
| `x/messaging` app-facing service | beta evidence complete | none for app-facing service | subordinate primitives remain experimental |
