# Card 0648

Priority: P0

Goal:
- Align the control plane with the existing file capability split and remove the remaining empty protocol residue.

Scope:
- extension taxonomy for `x/fileapi`
- architecture and repo metadata alignment
- cleanup rule for empty misleading directories

Non-goals:
- Do not redesign the current file capability split.
- Do not add package-level manifests yet.

Files:
- `specs/repo.yaml`
- `specs/extension-taxonomy.yaml`
- `specs/task-routing.yaml`
- `docs/architecture/AGENT_FIRST_REPO_BLUEPRINT.md`
- `docs/modules/x-fileapi/README.md`

Tests:
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/reference-layout`

Docs Sync:
- Keep the blueprint, extension taxonomy, and new primer aligned on where file API work starts.

Done Definition:
- `x/fileapi` is declared in the control plane and no longer exists only as code.
- File capability discovery is consistent across blueprint, specs, and module primer.
- The remaining empty protocol residue is either removed or made mechanically invalid.

Outcome:
- Registered `x/fileapi` in repo and extension taxonomy metadata.
- Added a dedicated `file_api` task-family entrypoint and a module primer under `docs/modules/x-fileapi/README.md`.
- Added `doc_paths` to `x/fileapi/module.yaml`.
- Removed the empty `contract/protocol` residue directory.

Validation Run:
- `go run ./internal/checks/module-manifests`
- `go run ./internal/checks/agent-workflow`
- `go run ./internal/checks/reference-layout`
