# Card 1403

Milestone: v1-cleanup-phase-5
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P0
State: active
Primary Module: release
Owned Files:
- docs/release/v1.0.0-rc.1.md
- tasks/cards/active/README.md
Depends On:
- 1402

Goal:
- Close the v1 cleanup sequence with full validation evidence before final release gating.

Scope:
- Run the repo-wide gates with sandbox-safe Go cache paths.
- Re-run release-relevant extension maturity and beta evidence checks.
- Confirm CLI submodule and website checks are either covered by gates or recorded separately.
- Record any remaining blockers, caveats, or required follow-up cards in the release notes or task queue.

Non-goals:
- Do not tag `v1.0.0` in this card.
- Do not change API shape during final validation.
- Do not widen cleanup scope if a gate failure reveals unrelated debt; open a new focused card instead.

Files:
- docs/release/v1.0.0-rc.1.md
- tasks/cards/active/README.md

Tests:
- GOCACHE=/private/tmp/plumego-gocache GOMODCACHE=/private/tmp/plumego-gomodcache make gates
- go run ./internal/checks/extension-beta-evidence
- go run ./internal/checks/extension-maturity

Docs Sync:
- Update release notes only with actual validation evidence and unresolved blockers.

Done Definition:
- `make gates` passes with sandbox-safe cache paths.
- Extension maturity and beta evidence checks pass.
- Any remaining release blockers are recorded as concrete follow-up cards.

Outcome:

