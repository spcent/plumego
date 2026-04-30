# Extension Release Evidence Artifacts

This note defines how checked-in release evidence artifacts are recorded for
`x/*` beta candidates.

Stable-root API baseline artifacts live separately under
`docs/stable-api/snapshots/`. Use `docs/release/PRE_V1_RELEASE_CHECKLIST.md`
before tagging a pre-v1 release candidate.

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

## Current Gap Map

`go run ./internal/checks/extension-beta-evidence` currently reports the
following blockers. The rows below map each blocker to module-owned follow-up
work without changing maturity status.

| Candidate | Existing evidence | Missing work | Follow-up |
| --- | --- | --- | --- |
| `x/rest` | evidence doc, current-head snapshot | release refs, release snapshots, owner sign-off | `tasks/cards/active/0723-x-rest-beta-evidence-closure.md` |
| `x/websocket` | evidence doc, current-head snapshot | release refs, release snapshots, owner sign-off | `tasks/cards/active/0724-x-websocket-beta-evidence-closure.md` |
| `x/tenant` | evidence doc, current-head snapshot | release refs, release snapshots, owner sign-off | `tasks/cards/active/0725-x-tenant-beta-evidence-closure.md` |
| `x/observability` | evidence doc, current-head snapshot | release refs, release snapshots, owner sign-off | `tasks/cards/active/0726-x-observability-beta-evidence-closure.md` |
| `x/gateway` | evidence doc, current-head snapshot | release refs, release snapshots, owner sign-off | `tasks/cards/active/0727-x-gateway-beta-evidence-closure.md` |
| `x/ai` stable-tier subpackages | evidence docs, current-head snapshots | release refs, release snapshots, owner sign-off | `tasks/cards/active/0728-x-ai-stable-tier-beta-evidence-closure.md` |
| `x/data` selected surfaces | evidence doc | current-head snapshots, release refs, release snapshots, owner sign-off | `tasks/cards/active/0729-x-data-surface-beta-evidence-closure.md` |
| `x/discovery` core/static surface | evidence doc | current-head snapshot, release refs, release snapshots, owner sign-off | `tasks/cards/active/0730-x-discovery-surface-beta-evidence-closure.md` |
| `x/messaging` app-facing service | evidence doc | current-head snapshot, release refs, release snapshots, owner sign-off | `tasks/cards/active/0731-x-messaging-service-beta-evidence-closure.md` |
