# Card 1452

Milestone: M-003
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P1
State: done
Primary Module: control-plane
Owned Files:
- `tasks/cards/active/1452-post-v1-maturity-roadmap.md`
Depends On:

Goal:
- Record the post-v1 maturity roadmap as executable task cards before changing implementation or release posture.

Scope:
- Keep the roadmap split into small reversible cards.
- Execute current, unblocked work first: CLI gate stability and release posture sync.
- Leave extension beta closure work blocked until real release refs, release-backed snapshots, and owner sign-off exist.

Non-goals:
- Do not promote any `x/*` module from this planning card.
- Do not use `HEAD` as release-history evidence.
- Do not change runtime behavior from this planning card.

Files:
- `tasks/cards/active/1452-post-v1-maturity-roadmap.md`
- `tasks/cards/active/1453-cli-gate-timeout-stability.md`
- `tasks/cards/active/1454-release-posture-sync.md`
- `tasks/cards/blocked/1367-x-tenant-beta-evidence-closure.md`
- `tasks/cards/blocked/1370-x-ai-stable-tier-beta-evidence-closure.md`
- `tasks/cards/blocked/1371-x-data-surface-beta-evidence-closure.md`
- `tasks/cards/blocked/1372-x-discovery-surface-beta-evidence-closure.md`
- `tasks/cards/blocked/1373-x-messaging-service-beta-evidence-closure.md`

Tests:
- `scripts/check-spec tasks/cards/active/1452-post-v1-maturity-roadmap.md`
- `scripts/check-spec tasks/cards/active/1453-cli-gate-timeout-stability.md`
- `scripts/check-spec tasks/cards/active/1454-release-posture-sync.md`

Docs Sync:
- None. This card records execution order only.

Done Definition:
- The roadmap exists in `tasks/cards/` with unblocked work in `active/` and evidence-gated promotion work explicitly blocked.

Outcome:
- Added the post-v1 maturity roadmap to task cards, with unblocked work in
  `tasks/cards/active/` and release-evidence-dependent beta closure work
  remaining in `tasks/cards/blocked/`.
