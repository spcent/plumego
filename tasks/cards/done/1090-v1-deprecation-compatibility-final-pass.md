# Card 1090

Milestone: v1
Recipe: specs/change-recipes/api-cleanup.yaml
Priority: P1
State: done
Primary Module: specs
Owned Files:
- `specs/deprecation-inventory.yaml`
- `internal/checks/deprecation-inventory/`
- `docs/release/v1.0.0-rc.1.md`
- `docs/CANONICAL_STYLE_GUIDE.md`
- `tasks/cards/active/README.md`
Depends On: 0747

Goal:
- Finish the v1 deprecation and compatibility inventory so deprecated, duplicate, and agent-confusing code paths are either removed or explicitly retained.

Problem:
Agent-friendly v1 maintenance requires a single obvious path for responses, errors, middleware, routing, and public compatibility. Ambiguous aliases and stale wrappers are expensive unless the repository records why they remain.

Scope:
- Keep `specs/deprecation-inventory.yaml` in strict mode.
- Confirm every entry has a v1 decision, owner, rationale, and release impact.
- Remove or card-split entries whose rationale no longer justifies retention.
- Ensure the checker fails on undecided entries and reports retained compatibility clearly.
- Update style guidance if the final pass identifies a canonical path that agents should prefer.

Non-goals:
- Do not remove exported symbols in this card unless all callers are enumerated and migrated in the same change.
- Do not change experimental maturity status.
- Do not create broad cleanup work without a bounded follow-up card.

Files:
- `specs/deprecation-inventory.yaml`
- `internal/checks/deprecation-inventory/`
- `docs/release/v1.0.0-rc.1.md`
- `docs/CANONICAL_STYLE_GUIDE.md`
- `tasks/cards/active/README.md`

Tests:
- `go run ./internal/checks/deprecation-inventory -strict`
- `go test ./internal/checks/deprecation-inventory/...`
- `go test ./...`

Docs Sync:
- Required for any retained compatibility decision or canonical-path clarification.

Done Definition:
- The deprecation inventory has no undecided v1 items.
- Strict mode blocks new ambiguous compatibility entries.
- Retained aliases are documented as compatibility, not preferred APIs.
- Any larger removal is represented by a dedicated follow-up card with symbol-change validation.

Outcome:
- Strengthened `internal/checks/deprecation-inventory` so strict mode rejects entries missing category, status, owner, decision, or replacement metadata.
- Kept `specs/deprecation-inventory.yaml` in strict mode with no undecided v1 items.
- Updated the report output to include category, owner, and replacement for retained compatibility entries.
- Added checker tests for incomplete and complete retained inventory entries.
- Updated canonical style and release notes with the final inventory metadata rule.
- Validation passed:
  - `go run ./internal/checks/deprecation-inventory -strict`
  - `go test ./internal/checks/deprecation-inventory/...`
  - `go test ./...`
