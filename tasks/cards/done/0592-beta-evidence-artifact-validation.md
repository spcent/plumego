# Card 0592

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P1
State: done
Primary Module: internal/checks
Owned Files:
- internal/checks/extension-beta-evidence/main.go
- internal/checks/extension-beta-evidence/README.md
- docs/extension-evidence/release-artifacts.md
- specs/extension-beta-evidence.yaml
- tasks/cards/active/README.md
Depends On: 2296

Goal:
Make beta evidence harder to fake by validating listed release refs and
snapshot artifacts instead of only counting ledger entries.

Scope:
- Validate that non-empty `release_refs` resolve as git refs.
- Validate that non-empty `api_snapshots` point at checked-in files.
- Document current release-ref availability and artifact rules.

Non-goals:
- Do not promote any module.
- Do not invent release refs when no release tags exist.
- Do not require two refs before real release history exists.

Files:
- `internal/checks/extension-beta-evidence/main.go`
- `internal/checks/extension-beta-evidence/README.md`
- `docs/extension-evidence/release-artifacts.md`
- `specs/extension-beta-evidence.yaml`
- `tasks/cards/active/README.md`

Tests:
- `go test ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-beta-evidence`
- `scripts/check-spec tasks/cards/done/0592-beta-evidence-artifact-validation.md`

Docs Sync:
- Required because beta evidence artifact semantics change.

Done Definition:
- The beta evidence checker validates refs and artifact paths for any listed
  evidence while preserving existing blockers for incomplete candidates.

Outcome:
- Extended `extension-beta-evidence` so non-empty `release_refs` must resolve
  as git commits and non-empty `api_snapshots` must be checked-in files under
  `docs/extension-evidence/snapshots/`.
- Documented that the current repository has no visible local or remote release
  tags, so branch heads or arbitrary commits must not be used as release
  evidence substitutes.
- Added the snapshot artifact root to the evidence ledger metadata while
  preserving existing incomplete-evidence blockers.

Validations:
- `go test ./internal/checks/extension-beta-evidence`
- `go run ./internal/checks/extension-beta-evidence`
- `scripts/check-spec tasks/cards/done/0592-beta-evidence-artifact-validation.md`
