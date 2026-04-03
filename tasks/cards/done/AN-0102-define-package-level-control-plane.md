# Card AN-0102

Priority: P1

Goal:
- Define the first package-level metadata spec for ambiguity hotspots.

Scope:
- package-level metadata format
- discoverability rules
- first hotspot package list

Non-goals:
- Do not annotate the entire repository.
- Do not replace module manifests with package-level metadata.

Files:
- `specs/package-index.yaml`
- `specs/agent-entrypoints.yaml`
- `docs/PRODUCT_PRD_AGENT_NATIVE.md`
- `docs/ROADMAP_AGENT_NATIVE.md`

Tests:
- `go run ./internal/checks/agent-workflow`

Docs Sync:
- Keep package-level control plane intent aligned between the PRD, roadmap, and specs.

Done Definition:
- Plumego has a defined package-level metadata contract for hotspot packages.
- The repository exposes one discoverable place to find package-level entrypoint guidance.
- The first hotspot package list is explicit and bounded.

Outcome:
- Added `specs/package-index.yaml` as the first machine-readable hotspot package discovery surface.
- Registered the first bounded hotspot package set:
  - `x/fileapi`
  - `x/data/file`
  - `x/rest`
  - `x/gateway`
  - `x/tenant/core`
  - `x/tenant/store/db`
  - `middleware/httpmetrics`
  - `middleware/requestid`
  - `contract`
- Updated `specs/agent-entrypoints.yaml` so module-level task entrypoints explicitly point to `specs/package-index.yaml` when package choice needs refinement.
- Updated the PRD and agent-native roadmap to treat the package index as the package-level control-plane surface.

Validation Run:
- `go run ./internal/checks/agent-workflow`
