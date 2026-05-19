# Card 1502

Milestone: M-008
Recipe: specs/change-recipes/analysis-only.yaml
Priority: P0
State: done
Primary Module: release
Owned Files:
- `docs/stable-api/snapshots/`

Goal:
- Compare stable-root API snapshots between v1.0.0 and HEAD to confirm zero exported-symbol changes.
- Append the comparison result to docs/release/v1.1.0.md under an API Snapshot Comparison section.

Scope:
- Run `go run ./internal/checks/extension-release-evidence` or equivalent snapshot comparison tool
  against each stable root (core, router, contract, middleware, security, store, health, log, metrics).
- Record any exported-symbol additions, removals, or signature changes found.
- If zero changes: confirm and proceed.
- If changes found: stop, record findings in docs/release/v1.1.0.md, and do not proceed to card 1503.

Non-goals:
- Do not change any stable-root source files; this is analysis only.
- Do not compare x/* extension packages (those use the extension evidence pipeline).
- Do not generate new snapshots in this card; use existing snapshots in docs/stable-api/snapshots/.

Files:
- `docs/stable-api/snapshots/`
- `docs/release/v1.1.0.md`

Tests:
- `go run ./internal/checks/dependency-rules`
- `go run ./internal/checks/module-manifests`

Docs Sync:
- docs/release/v1.1.0.md updated with API Snapshot Comparison section.

Done Definition:
- Comparison covers all nine stable roots.
- Result is recorded in docs/release/v1.1.0.md: either "zero exported-symbol changes" or a
  specific list of changes with a block recommendation.
- Card does not proceed if unexpected changes are found; a blocker is recorded instead.

Outcome:
- Done. Compared all nine stable roots from `v1.0.0` to `HEAD` with
  `go run ./internal/checks/extension-release-evidence`; every root reported
  `api unchanged` and `snapshots match`. Recorded the result in
  docs/release/v1.1.0.md and ran dependency-rules plus module-manifests.
