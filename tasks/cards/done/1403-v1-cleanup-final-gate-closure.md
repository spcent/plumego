# Card 1403

Milestone: v1-cleanup-phase-5
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P0
State: done
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
- Ran the full v1 cleanup final gate with sandbox-safe cache paths:
  - `GOCACHE=/private/tmp/plumego-gocache GOMODCACHE=/private/tmp/plumego-gomodcache make gates`
- Re-ran release-relevant extension checks:
  - `go run ./internal/checks/extension-beta-evidence`
  - `go run ./internal/checks/extension-maturity`
- `make gates` passed, including control-plane checks, deprecation inventory, stable API snapshots, doc snippets, `go vet ./...`, race tests, normal tests, stable-root coverage, `cmd/plumego` submodule vet/tests, and website sync/check/build.
- Stable-root coverage reported by the gate was 87.2%.
- Extension beta evidence remains unchanged: `x/gateway`, `x/observability`, `x/rest`, and `x/websocket` are beta; remaining candidates still report their recorded blockers.
- Updated `docs/release/v1.0.0-rc.1.md` with the May 14, 2026 v1 cleanup gate evidence.
