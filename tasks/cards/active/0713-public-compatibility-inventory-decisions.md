# Card 0713

Milestone: v1
Recipe: specs/change-recipes/symbol-change.yaml
Priority: P1
State: active
Primary Module: specs
Owned Files:
- specs/deprecation-inventory.yaml
- docs/release/v1.0.0-rc.1.md
Depends On:
- none

Goal:
Decide the v1 keep/remove outcome for every `decision_required` entry in `specs/deprecation-inventory.yaml` before enabling strict deprecation-inventory gating.

Scope:
- Review all `decision_required` entries in `specs/deprecation-inventory.yaml`.
- For each entry, mark the final outcome as `keep`, `remove`, or `removed`.
- Add migration notes for removed public aliases or unsupported surfaces.
- Move entries that stay experimental into release notes with explicit non-stable wording.

Non-goals:
- Do not implement MQTT or AMQP bridges as part of this card.
- Do not promote experimental `x/*` packages to stable status.
- Do not change stable public API without a symbol-change migration note.

Files:
- specs/deprecation-inventory.yaml
- docs/release/v1.0.0-rc.1.md
- affected package files for entries marked `remove`
- affected package tests for entries marked `remove`

Tests:
- go run ./internal/checks/deprecation-inventory -strict
- go run ./internal/checks/dependency-rules
- make gates

Docs Sync:
- Update `docs/release/v1.0.0-rc.1.md` for public compatibility decisions.
- Update module README files only when an implemented public behavior changes.

Done Definition:
- No `decision_required` entries remain unless they are explicitly scoped to experimental `x/*` and documented in release notes.
- `go run ./internal/checks/deprecation-inventory -strict` passes.
- Removed symbols have no remaining call sites.
- Release notes describe retained experimental compatibility surfaces.

Outcome:
- Pending.
