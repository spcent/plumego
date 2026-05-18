# Plan for M-008: v1.1.0 Release Execution

Milestone: `M-008`
Objective: Tag v1.1.0, produce full gate evidence and release notes, update the
second_release_ref in the extension beta evidence ledger, and move all five
blocked beta-evidence closure cards to active so they can proceed to promotion.
Constraints: no stable-root API changes, no new external dependencies, second_release_ref
set only after the tag exists, sequential phases because tag and evidence depend on
prior verification steps.
Affected Modules: release, extension evidence, tasks.

## Phase Map

- Phase 1: Orient — read all context files and confirm gates are green before any edits.
- Phase 2: Pre-Release Verification — run full gate suite, write release notes, compare API snapshots.
- Phase 3: Tag and Evidence — create git tag v1.1.0, update evidence ledger, move blocked cards.
- Phase 4: Validate and Ship — final acceptance run, commit, push, PR.

## Card Inventory

| Card | Goal | Primary Module | Owned Files | Depends On | Quick Gates |
|------|------|----------------|-------------|------------|-------------|
| 1500 | Run full gate suite and record output | release | `docs/release/v1.1.0.md` (gate section) | M-007 | `make gates`, `go vet ./...` |
| 1501 | Write docs/release/v1.1.0.md release notes | release | `docs/release/v1.1.0.md` | 1500 | file exists, no-go decision recorded |
| 1502 | Compare stable-root API snapshots v1.0.0 → HEAD | release | `docs/stable-api/snapshots/` | 1501 | zero exported-symbol delta |
| 1503 | Create annotated git tag v1.1.0 | release | git tag | 1502 | `git tag --list v1.1.0` |
| 1504 | Update beta evidence ledger with second_release_ref v1.1.0; move blocked cards to active | extension evidence | `specs/extension-beta-evidence.yaml`, `tasks/cards/active/1367-x-tenant-beta-evidence-closure.md`, `tasks/cards/active/1370-x-ai-stable-tier-beta-evidence-closure.md`, `tasks/cards/active/1371-x-data-surface-beta-evidence-closure.md`, `tasks/cards/active/1372-x-discovery-surface-beta-evidence-closure.md`, `tasks/cards/active/1373-x-messaging-service-beta-evidence-closure.md` | 1503 | `go run ./internal/checks/extension-beta-evidence` |

## Dependency Edges

- `1500 -> 1501`
- `1501 -> 1502`
- `1502 -> 1503`
- `1503 -> 1504`

## Parallel Groups

- None. All cards are sequential; the tag must exist before evidence is updated,
  and gates must pass before the tag is created.

## Risk Register

- Risk: make gates fails due to a test regression introduced after M-007 merged.
  Mitigation: card 1500 captures the failure; do not proceed to 1501 until gates are green.
- Risk: stable-root snapshot comparison reveals an unintended exported-symbol change.
  Mitigation: card 1502 blocks the tag; resolve the change or document it as intentional
  before proceeding.
- Risk: second_release_ref is set to a commit SHA instead of the annotated tag.
  Mitigation: card 1504 explicitly checks `git tag --list v1.1.0` before editing the ledger.

## Verification Strategy

- Card-level checks: each card lists its own gate commands; run them immediately after
  the card's edits and record the output.
- Milestone-level checks: run the full Acceptance Criteria suite in Phase 4 before
  committing anything.
- Snapshot check: `go run ./internal/checks/extension-beta-evidence` after 1504
  must exit 0 with second_release_ref visible in output.

## Exit Condition

- all five planned cards completed
- git tag v1.1.0 exists in the repository
- specs/extension-beta-evidence.yaml has second_release_ref = v1.1.0 for all five surfaces
- blocked cards 1367, 1370, 1371, 1372, 1373 exist in tasks/cards/active/
- verify report shows pass
- milestone acceptance criteria ready for PR packaging
