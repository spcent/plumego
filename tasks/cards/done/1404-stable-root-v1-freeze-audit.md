# Card 1404

Milestone: v1-package-cleanup
Recipe: specs/change-recipes/stable-root-boundary-review.yaml
Priority: P0
State: active
Primary Module: stable-roots
Owned Files:
- docs/release/v1.0.0-rc.1.md
- specs/deprecation-inventory.yaml
- tasks/cards/active/README.md
Depends On:
- 1403

Goal:
- Confirm the stable root packages remain frozen for v1 and identify only non-breaking cleanup follow-ups.

Scope:
- Review `core`, `router`, `contract`, `middleware`, `security`, `store`, `health`, `log`, and `metrics` for lingering compatibility markers, deprecated comments, or cleanup candidates.
- Verify no stable root imports `x/*`.
- Record any cleanup as future focused cards instead of changing stable API shape.
- Preserve existing GA status and stable API snapshots.

Non-goals:
- Do not remove exported stable symbols.
- Do not change stable response/error envelope behavior.
- Do not rewrite stable package internals opportunistically.

Files:
- docs/release/v1.0.0-rc.1.md
- specs/deprecation-inventory.yaml
- tasks/cards/active/README.md

Tests:
- go run ./internal/checks/dependency-rules
- go run ./internal/checks/deprecation-inventory -strict
- bash scripts/check-stable-api-snapshots.sh

Docs Sync:
- Update release notes only with actual audit evidence or new blockers.

Done Definition:
- Stable roots have no unregistered compatibility or deprecated markers.
- Stable API snapshots still match.
- Any remaining stable cleanup is represented as explicit future cards.

Outcome:
- Completed on May 15, 2026.
- Confirmed stable-root production code has no `x/*` imports via `go run ./internal/checks/dependency-rules`.
- Confirmed retained compatibility, deprecated, alias, TODO, and unsupported-operation markers remain registered via `go run ./internal/checks/deprecation-inventory -strict`; no new stable-root inventory entries were required.
- Confirmed checked-in stable API snapshots still match with `bash scripts/check-stable-api-snapshots.sh`.
- Recorded the re-audit evidence in `docs/release/v1.0.0-rc.1.md`.
