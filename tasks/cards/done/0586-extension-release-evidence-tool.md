# Card 0586

Milestone:
Recipe: specs/change-recipes/new-extension-module.yaml
Priority: P1
State: done
Primary Module: internal/checks
Owned Files:
- internal/checks/extension-release-evidence/main.go
- internal/checks/extension-release-evidence/README.md
- specs/checks.yaml
- docs/EXTENSION_STABILITY_POLICY.md
- tasks/cards/active/README.md
Depends On: 2279, 2290

Goal:
Add release-aware beta evidence automation that can compare exported extension
APIs across two git refs.

Scope:
- Add `internal/checks/extension-release-evidence`.
- Generate API snapshots for a package pattern at `base` and `head` refs using
  temporary source trees extracted from git refs.
- Compare snapshots through the existing `extension-api-snapshot` semantics.
- Print deterministic report lines and write optional snapshot files.
- Document how promotion cards should use the tool.

Non-goals:
- Do not mutate `specs/extension-beta-evidence.yaml`.
- Do not create tags or infer release refs.
- Do not promote any module.

Files:
- `internal/checks/extension-release-evidence/main.go`
- `internal/checks/extension-release-evidence/README.md`
- `specs/checks.yaml`
- `docs/EXTENSION_STABILITY_POLICY.md`
- `tasks/cards/active/README.md`

Tests:
- `go test ./internal/checks/...`
- `go run ./internal/checks/extension-release-evidence -module ./x/rest/... -base HEAD -head HEAD`
- `scripts/check-spec tasks/cards/done/0586-extension-release-evidence-tool.md`

Docs Sync:
- Required because beta promotion evidence workflow changes.

Done Definition:
- A promotion card can generate and compare exported API snapshots for two refs
  with one local command.
- The command proves equality/difference but does not change ledger state.

Outcome:
- Added `go run ./internal/checks/extension-release-evidence` for comparing
  exported extension APIs across two git refs.
- The tool extracts temporary source trees with `git archive`, generates
  snapshots with `extension-api-snapshot`, compares them, and optionally writes
  snapshot files to `-out-dir`.
- Added the command to evidence checks and documented it in the extension
  stability promotion process.

Validations:
- `go test ./internal/checks/...`
- `go run ./internal/checks/extension-release-evidence -module ./x/rest/... -base HEAD -head HEAD`
- `scripts/check-spec tasks/cards/done/0586-extension-release-evidence-tool.md`
