# Card 1514

Milestone: M-010
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P1
State: done
Primary Module: x/frontend
Owned Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-frontend.md`

Goal:
- Evaluate x/frontend for beta promotion: generate release-to-release API
  snapshots between v1.0.0 and v1.1.0 and compare exported symbols.
- If the API is stable, record evidence and owner sign-off and promote to beta.
- If the API has changed, record an explicit blocker in the evidence ledger and
  leave status as experimental.

Scope:
- Run go run ./internal/checks/extension-release-evidence for x/frontend to
  generate release-to-release snapshots.
- Compare exported symbols between v1.0.0 snapshot and v1.1.0 snapshot.
- If stable: set second_release_ref, add api_snapshot_ref, record owner_signoff
  in specs/extension-beta-evidence.yaml; update x/frontend/module.yaml status
  to beta.
- If unstable: add api_changed blocker entry in specs/extension-beta-evidence.yaml
  with specific symbol diff; leave module.yaml status as experimental.
- Update docs/extension-evidence/x-frontend.md with evaluation findings.

Non-goals:
- Do not change x/frontend runtime behavior.
- Do not infer stability from prose; use snapshot comparison only.
- Do not promote if any exported symbol was added, removed, or changed.

Files:
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/x-frontend.md`
- `x/frontend/module.yaml` (only if promoting to beta)

Tests:
- `go run ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-maturity`

Docs Sync:
- docs/extension-evidence/x-frontend.md is the primary artifact of this card.

Done Definition:
- x/frontend is either promoted to beta with complete evidence, or has an
  explicit api_changed blocker recorded in specs/extension-beta-evidence.yaml.
- `go run ./internal/checks/extension-beta-evidence` exits 0.
- The evaluation result (promote or block) is not ambiguous.

Outcome:
- Done. `extension-release-evidence` for `./x/frontend` across `v1.0.0` to
  `v1.1.0` reported API unchanged. Recorded release-backed snapshots, release
  refs, and `frontend` sign-off; promoted `x/frontend` to beta and synced the
  evidence, module primer, maturity dashboard, deprecation table, and release
  artifacts note.
