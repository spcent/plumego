# Plan for M-024: Stable-Tier Metadata and Control-Plane Clarity

Milestone: `M-024`
Objective: Finish the remaining verified non-runtime audit work after M-022 and M-023 by making stable-tier AI metadata machine-readable and cleaning up the residual docs/specs/tasks ambiguity.
Constraints: depends on M-022 and M-023 because the affected module families overlap; one primary module per card; max 5 files per card; max 3 validation commands per card; no new dependencies.
Affected Modules: `x/ai`, `docs`, `specs`, `middleware`, `tasks`

## Phase Map

- Phase 1: add stable-tier `x/ai` subpackage manifests
- Phase 2: clean docs/specs discoverability drift
- Phase 3: reconcile milestone archive clarity

## Card Inventory

| Card | Goal | Primary Module | Owned Files | Depends On | Quick Gates |
|------|------|----------------|-------------|------------|-------------|
| 2062 | Add machine-readable manifests for `x/ai/provider` and `x/ai/session` so stable-tier evidence stops relying on root-family prose only. | `x/ai` | `x/ai/provider/module.yaml`, `x/ai/session/module.yaml`, `x/ai/module.yaml`, `docs/modules/x-ai/README.md`, `docs/EXTENSION_MATURITY.md` | M-022, M-023 | `module-manifests`, `extension-maturity` |
| 2063 | Add machine-readable manifests for `x/ai/streaming` and `x/ai/tool` and align the stable-tier dashboard wording with the new manifest granularity. | `x/ai` | `x/ai/streaming/module.yaml`, `x/ai/tool/module.yaml`, `x/ai/module.yaml`, `docs/modules/x-ai/README.md`, `docs/EXTENSION_MATURITY.md` | 2062 | `module-manifests`, `extension-maturity` |
| 2064 | Replace the case-only `docs/AGENT_FIRST.md` vs `docs/agent-first.md` split with clearly distinguished internal vs external document naming and references. | `docs` | `AGENTS.md`, `docs/README.md`, `docs/AGENT_FIRST.md`, `docs/agent-first.md`, `README.md` | M-022 | `agent-workflow`, `gofmt -l .` |
| 2065 | Declare `middleware/internal/telemetry` in machine-readable boundary rules and remove the dead `plumego.go` forbidden-path residue. | `specs` | `specs/dependency-rules.yaml`, `middleware/module.yaml`, `docs/modules/middleware/README.md` | M-022 | `dependency-rules`, `module-manifests` |
| 2066 | Make `middleware/conformance` discoverable as a test-only conformance suite instead of an implementation-looking orphan package. | `middleware` | `middleware/conformance/README.md`, `docs/modules/middleware/README.md` | 2065 | `go test ./middleware/...`, `gofmt -l .` |
| 2067 | Reconcile milestone archive truth with the files on disk: explain missing historical artifacts, clarify superseded drafts, and make numbering history explicit. | `tasks` | `tasks/milestones/ROADMAP.md`, `tasks/milestones/STATUS.md`, `tasks/milestones/superseded/README.md`, `tasks/milestones/ARCHIVE_INDEX.md` | M-022 | `agent-workflow`, `gofmt -l .` |

## Dependency Edges

- `2062 -> 2063`
- `2065 -> 2066`
- `M-022 -> 2062`
- `M-022 -> 2064`
- `M-022 -> 2065`
- `M-022 -> 2067`
- `M-023 -> 2062`

## Parallel Groups

- Group A: `2062`, `2064`, `2065`, `2067`
- Group B: `2063`
- Group C: `2066`

## Risk Register

- Risk: subpackage manifests for `x/ai` drift from root-family stability prose.
  Mitigation: keep `x/ai/module.yaml` as the family index and let each stable-tier subpackage manifest own its local public surface and tests.
- Risk: control-plane cleanup rewrites history instead of documenting it.
  Mitigation: add archive notes and explicit missing-artifact truth instead of inventing retroactive milestone directories or changing historical goals.
- Risk: docs/specs cleanup gets mixed back into runtime migration.
  Mitigation: M-023 owns runtime resilience convergence; do not touch `x/ai/resilience` here.

## Finding Disposition

- Verified and queued: `docs/AGENT_FIRST.md` and `docs/agent-first.md` are now distinguished in `docs/README.md`, but the case-only filename split remains an avoidable source of confusion.
- Verified and queued: `middleware/internal/telemetry` is still imported by multiple middleware packages while remaining absent from `specs/dependency-rules.yaml`; the dead `special_rules.forbidden_paths: [plumego.go]` residue also still exists.
- Verified and queued: `middleware/conformance` is still a test-only directory with no local README and no file-level explanation at its own path.
- Verified and queued: the on-disk milestone archive is incomplete relative to `ROADMAP.md` / `STATUS.md`, while `superseded/` retains draft files that require explicit archive framing.
- Deferred here intentionally: `x/ai` dual-stack resilience convergence moved into M-023 so it can land first without waiting on control-plane cleanup.

## Verification Strategy

- Card-level checks: use the quick gates listed per card, plus focused package tests for `x/ai` or `middleware` when behavior changes.
- Milestone-level checks: `dependency-rules`, `agent-workflow`, `module-manifests`, `extension-maturity`, then focused `go test` for `x/ai` and `middleware`.

## Checkpoints

| Phase | Checkpoint Gate | Status |
|-------|-----------------|--------|
| Phase 1 | `go run ./internal/checks/module-manifests && go run ./internal/checks/extension-maturity` | pending |
| Phase 2 | `go run ./internal/checks/dependency-rules && go test -timeout 20s ./middleware/...` | pending |
| Phase 3 | `go run ./internal/checks/agent-workflow` | pending |

## Exit Condition

- all planned cards completed or explicitly superseded
- all phase checkpoints recorded as passed
- verify report shows pass
- milestone acceptance criteria ready for PR packaging
