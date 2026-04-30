# Extension Beta Evidence Template

Use this template when preparing an `x/*` module or sub-surface for beta
promotion. A completed evidence document does not grant beta status by itself;
the module manifest and evidence ledger must be updated in the same reviewed
promotion card after all blockers are cleared.

## Candidate

- Module:
- Surface or subpackage:
- Owner:
- Current status:
- Candidate status: `beta`
- Evidence state: incomplete

## Public API Surface

- Snapshot baseline:
- Snapshot candidate:
- Comparison command:

```bash
go run ./internal/checks/extension-release-evidence \
  -module ./x/<family>/... \
  -base <older-minor-release-ref> \
  -head <newer-minor-release-ref> \
  -out-dir docs/extension-evidence/snapshots/<candidate>
```

## Behavior Coverage

| Surface | Evidence | Status |
| --- | --- | --- |
| Happy path |  | missing |
| Negative path |  | missing |
| Boundary checks |  | missing |
| Example or reference wiring |  | missing |

## Release Evidence

- Older release ref:
- Newer release ref:
- API changed between refs: unknown
- Gate command record:

## Owner Sign-Off

Owner:

Decision:

Date:

Statement:

```text
I confirm that this candidate meets the beta criteria in
docs/EXTENSION_STABILITY_POLICY.md and accept the beta compatibility
obligations for the documented public surface.
```

## Blockers

- `release_history_missing`
- `api_snapshot_missing`
- `owner_signoff_missing`

## Promotion Decision

Do not promote while any blocker remains.
