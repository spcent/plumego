# Card 2290

Milestone:
Recipe: specs/change-recipes/new-extension-module.yaml
Priority: P1
State: done
Primary Module: internal/checks
Owned Files:
- internal/checks/extension-beta-evidence/main.go
- internal/checks/extension-beta-evidence/README.md
- specs/checks.yaml
- docs/EXTENSION_STABILITY_POLICY.md
- tasks/cards/active/README.md
Depends On: 2278, 2279, 2280, 2289

Goal:
Close the first loop of the extension beta evidence pipeline with a local check
that reads the evidence ledger, validates candidate consistency, and reports
promotion blockers.

Scope:
- Add an `internal/checks/extension-beta-evidence` command.
- Validate candidate module, owner, current status, evidence doc, release refs,
  API snapshots, and blocker consistency.
- Print a deterministic per-candidate report suitable for local and CI use.
- Add the check to the repo evidence checks.
- Document how this fits with `extension-api-snapshot`.

Non-goals:
- Do not promote any `x/*` module.
- Do not fetch remote releases or create git tags.
- Do not add a YAML dependency.
- Do not mutate `specs/extension-beta-evidence.yaml`.

Files:
- `internal/checks/extension-beta-evidence/main.go`
- `internal/checks/extension-beta-evidence/README.md`
- `specs/checks.yaml`
- `docs/EXTENSION_STABILITY_POLICY.md`
- `tasks/cards/active/README.md`

Tests:
- `go test ./internal/checks/...`
- `go run ./internal/checks/extension-beta-evidence`
- `scripts/check-spec tasks/cards/done/2290-extension-beta-evidence-report-check.md`

Docs Sync:
- Required because promotion workflow commands change.

Done Definition:
- Evidence ledger drift and blocker consistency can be checked locally.
- The check reports all current candidates and fails when ledger data becomes
  internally inconsistent.
- The extension stability policy points to the new report command.

Outcome:
- Added `go run ./internal/checks/extension-beta-evidence` to validate the
  beta evidence ledger against declared extension roots, module manifests,
  evidence docs, release refs, API snapshots, owner sign-off state, and blocker
  consistency.
- Added deterministic per-candidate blocker reporting for current beta
  candidates.
- Added the command to `specs/checks.yaml` evidence checks and documented it in
  the extension stability promotion process.

Validations:
- `go test ./internal/checks/...`
- `go run ./internal/checks/extension-beta-evidence`
- `scripts/check-spec tasks/cards/done/2290-extension-beta-evidence-report-check.md`
