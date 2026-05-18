# Card 1501

Milestone: M-008
Recipe: specs/change-recipes/update-docs.yaml
Priority: P0
State: done
Primary Module: release
Owned Files:
- `docs/release/v1.1.0.md`

Goal:
- Complete docs/release/v1.1.0.md with full release notes following the structure of docs/release/v1.0.0.md.
- Record the go/no-go decision with rationale.

Scope:
- Write the following sections in docs/release/v1.1.0.md: Summary, What's New, Stable API Surface,
  Extension Status, Known Issues, Upgrade Path, and Go/No-Go Decision.
- Reference the gate output appended by card 1500.
- List all extension surfaces that gained their first v1.0.0 release-ref evidence in M-007.

Non-goals:
- Do not invent features or changes that are not in the repository.
- Do not modify any Go source files.
- Do not publish these notes externally; this is an internal evidence artifact.

Files:
- `docs/release/v1.1.0.md`
- `docs/release/v1.0.0.md` (read-only reference for structure)

Tests:
- File exists and is non-empty after this card.
- Go/No-Go decision section is present and explicit.

Docs Sync:
- docs/release/v1.1.0.md is the primary artifact completed by this card.

Done Definition:
- docs/release/v1.1.0.md contains all required sections from the v1.0.0.md template.
- Go/No-Go decision is recorded as an explicit line: "Decision: GO" or "Decision: NO-GO".
- No placeholder text remains in the document.

Outcome:
- Done. docs/release/v1.1.0.md now contains the required release-note
  sections, references the card 1500 gate output, lists the M-007 first
  release-ref evidence surfaces, and records `Decision: GO`.
