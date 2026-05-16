# Extension Release Evidence Artifacts

This note defines how checked-in release evidence artifacts are recorded for
`x/*` `beta` candidates.

Stable-root API baseline artifacts live separately under
`docs/stable-api/snapshots/`. Use `docs/release/POST_V1_EVIDENCE.md` as the
first release evidence read after `v1.0.0`.

## Current Release Ref State

As of M-007, the working copy sees these local and remote release tags:

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

Use the `v1.0.0` tag target commit, not the annotated tag object, when recording
release refs in `specs/extension-beta-evidence.yaml`.

Do not use branch heads, arbitrary commits, or `HEAD` as substitutes for release
evidence. A single `v1.0.0` release ref is an intake artifact only; it does not
satisfy the two-release promotion rule by itself.

## Artifact Rules

- `release_refs` entries in `specs/extension-beta-evidence.yaml` must resolve
  to git commits.
- `api_snapshots` entries must be checked-in files under
  `docs/extension-evidence/snapshots/`.
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
| `x/frontend` | evidence doc, current-head snapshot | second-release evidence, release-backed snapshots, owner sign-off | future frontend-owned card |
| `x/tenant` | evidence doc, current-head snapshot | release refs, release snapshots, owner sign-off | `tasks/cards/active/1445-x-tenant-v1-baseline-evidence.md`, then `tasks/cards/active/1367-x-tenant-beta-evidence-closure.md` |
| `x/ai` stable-tier subpackages | evidence docs, current-head snapshots | release refs, release snapshots, owner sign-off | `tasks/cards/active/1446-x-ai-stable-tier-v1-baseline-evidence.md`, then `tasks/cards/active/1370-x-ai-stable-tier-beta-evidence-closure.md` |
| `x/data` selected surfaces | evidence doc | snapshots, release refs, release snapshots, owner sign-off | `tasks/cards/active/1447-data-discovery-messaging-v1-baseline-gap-index.md`, then `tasks/cards/active/1371-x-data-surface-beta-evidence-closure.md` |
| `x/discovery` core/static surface | evidence doc | snapshot, release refs, release snapshots, owner sign-off | `tasks/cards/active/1447-data-discovery-messaging-v1-baseline-gap-index.md`, then `tasks/cards/active/1372-x-discovery-surface-beta-evidence-closure.md` |
| `x/messaging` app-facing service | evidence doc | snapshot, release refs, release snapshots, owner sign-off | `tasks/cards/active/1447-data-discovery-messaging-v1-baseline-gap-index.md`, then `tasks/cards/active/1373-x-messaging-service-beta-evidence-closure.md` |
