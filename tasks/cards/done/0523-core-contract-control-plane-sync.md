# Card 0523

Milestone:
Recipe: specs/change-recipes/fix-bug.yaml
Priority: P2
State: done
Primary Module: contract
Owned Files:
- contract/module.yaml
- core/module.yaml
- docs/modules/contract/README.md
- docs/modules/core/README.md
Depends On: 2232

Goal:
Synchronize the core and contract control-plane docs with the current public surfaces and file layout.

Scope:
- Fix stale file references in module primers.
- Expand manifest public entrypoints where current documented usage already depends on them.
- Keep docs focused on implemented behavior only.

Non-goals:
- Do not change Go code.
- Do not add new public APIs.
- Do not rewrite broader roadmap or architecture docs.

Files:
- `contract/module.yaml`
- `core/module.yaml`
- `docs/modules/contract/README.md`
- `docs/modules/core/README.md`

Tests:
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/reference-layout`

Docs Sync:
- This card is the docs/control-plane sync.

Done Definition:
- Contract primer points to existing files.
- Core and contract manifests expose the stable entrypoints shown in the module primers.
- Control-plane checks pass.

Outcome:
- Synced core and contract manifest entrypoint lists with the stable surfaces described in their module primers.
- Fixed the stale `contract/error.go` primer reference and added the current binding, context, request-id, and trace carrier first-read files.
- Validation run: `go run ./internal/checks/module-manifests`; `go run ./internal/checks/agent-workflow`; `go run ./internal/checks/reference-layout`.
