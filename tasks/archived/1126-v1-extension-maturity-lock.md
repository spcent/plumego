# Card 1126

Milestone: v1
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P0
State: done
Primary Module: x
Owned Files:
- `docs/EXTENSION_MATURITY.md`
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/`
- `docs/release/v1.0.0-rc.1.md`
- `tasks/cards/active/README.md`
Depends On: 0746

Goal:
- Lock all `x/*` modules as experimental for v1 unless complete release evidence and owner sign-off already exist.

Problem:
Several extension families have beta-candidate docs and active evidence cards, but the evidence ledger still lacks two real release refs, matching API snapshots, and owner sign-off. v1 must not accidentally promote extensions through wording drift.

Scope:
- Verify `docs/EXTENSION_MATURITY.md` matches `specs/extension-beta-evidence.yaml`.
- Confirm all beta candidates still list real blockers when evidence is incomplete.
- Keep cards `0723` through `0731` and `0738` blocked until release refs and sign-off exist.
- Remove or soften any docs wording that implies `x/*` GA or beta support without evidence.
- Update release notes to state that extension blockers do not block stable-root v1.

Non-goals:
- Do not fabricate release refs from `HEAD`.
- Do not promote any extension maturity state.
- Do not change extension runtime behavior.

Files:
- `docs/EXTENSION_MATURITY.md`
- `specs/extension-beta-evidence.yaml`
- `docs/extension-evidence/`
- `docs/release/v1.0.0-rc.1.md`
- `tasks/cards/active/README.md`

Tests:
- `go run ./internal/checks/extension-maturity`
- `go run ./internal/checks/extension-beta-evidence`
- `go test ./x/...`

Docs Sync:
- Required for extension maturity, release notes, and evidence docs.

Done Definition:
- No `x/*` module is promoted by implication.
- Evidence blockers remain explicit and machine-checkable.
- Release notes clearly separate stable-root v1 readiness from post-v1 extension beta work.
- Blocked evidence cards remain accurate and actionable.

Outcome:
- Verified `docs/EXTENSION_MATURITY.md` and `specs/extension-beta-evidence.yaml` keep all `x/*` modules experimental for v1.
- Added an explicit v1 release lock note to the extension maturity dashboard.
- Confirmed beta-candidate entries still carry release history, API snapshot, and owner sign-off blockers.
- Kept blocked evidence cards `0723` through `0731` and `0738` as post-v1 evidence work.
- Updated release notes to separate stable-root v1 readiness from post-v1 extension beta promotion.
- Validation passed:
  - `go run ./internal/checks/extension-maturity`
  - `go run ./internal/checks/extension-beta-evidence`
  - `go test ./x/...`
