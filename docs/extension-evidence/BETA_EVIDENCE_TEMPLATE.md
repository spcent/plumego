# Extension Beta Evidence Template

Use this template when preparing an `x/*` module or sub-surface for beta
promotion. A completed evidence document does not grant beta status by itself;
the `specs/extension-beta-evidence.yaml` ledger and the module manifest must be
updated in the same reviewed promotion card after all blockers are cleared.

See `docs/EXTENSION_STABILITY_POLICY.md` for promotion criteria and
`docs/release/PROMOTION_CARD_TEMPLATE.md` for the promotion PR template.

## Candidate

- Module:
- Surface or subpackage (leave blank for full-module candidates):
- Owner:
- Current status: `experimental`
- Candidate status: `beta`
- Evidence state: incomplete

## Public API Surface

- Baseline snapshot: `docs/extension-evidence/snapshots/<candidate>/base.snapshot`
- Candidate snapshot: `docs/extension-evidence/snapshots/<candidate>/head.snapshot`

Generate release-backed snapshots (requires two concrete git tags):

```bash
go run ./internal/checks/extension-release-evidence \
  -module ./x/<family>/... \
  -base <older-minor-release-ref> \
  -head <newer-minor-release-ref> \
  -out-dir docs/extension-evidence/snapshots/<candidate>
```

Do not clear `api_snapshot_missing` from a head-only snapshot. The snapshots
must come from the two release refs recorded under `release_refs` in
`specs/extension-beta-evidence.yaml`.

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
- Comparison output: (paste `module ... api unchanged/changed` line)

## Owner Sign-Off

Owner:

Statement (required verbatim):

> I confirm that `x/<family>` meets the beta criteria in
> docs/EXTENSION_STABILITY_POLICY.md and accept the beta compatibility
> obligations for the documented `x/<family>` public surface.

Date:

## Updating the Evidence Ledger

After completing the steps above, update `specs/extension-beta-evidence.yaml`:

```yaml
- module: x/<family>
  owner: <owner>
  current_status: beta           # changed from experimental
  evidence_doc: docs/extension-evidence/x-<family>.md
  release_refs:
    - <older-ref>
    - <newer-ref>
  api_snapshots:
    - docs/extension-evidence/snapshots/<candidate>/base.snapshot
    - docs/extension-evidence/snapshots/<candidate>/head.snapshot
  owner_signoff: signed          # required to clear owner_signoff_missing
  blockers: []                   # all three blockers must be gone
```

`owner_signoff` must be the literal string `signed`, `complete`, or `approved`
for the `extension-beta-evidence` check to accept it.

## Promotion Checklist

Complete in this order within a single promotion card:

1. Run `go run ./internal/checks/extension-release-evidence` for both refs.
2. Check in generated snapshots under `docs/extension-evidence/snapshots/<candidate>/`.
3. Update `specs/extension-beta-evidence.yaml` with refs, snapshots, and `owner_signoff: signed`.
4. Confirm blockers list is empty (`blockers: []`).
5. Record owner sign-off statement in this evidence file.
6. Update `x/<family>/module.yaml` status from `experimental` to `beta`.
7. Update `docs/EXTENSION_MATURITY.md` dashboard row status and evidence column.
8. Update `README.md` and `README_CN.md` support matrix rows.
9. Run `make gates` — all checks must pass.
10. Run `cd website && pnpm sync && pnpm build` — website build must pass.

## Blockers

- `release_history_missing`
- `api_snapshot_missing`
- `owner_signoff_missing`

## Promotion Posture

Do not promote while any blocker remains. Do not change `module.yaml` status
in the same commit that creates incomplete evidence.
