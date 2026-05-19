# Plan for M-005: v1 Release Execution

Milestone: `M-005`
Objective: Close the verifiable path from current pre-v1 state through
`v1.0.0-rc.1` to either final `v1.0.0` or a bounded rc blocker list.
Constraints: no feature work, no unapproved stable public API changes, no
experimental extension promotion without complete evidence, and no release
claims without command output or tag evidence.
Affected Modules: release, stable roots, `cmd/plumego`, extension evidence.

## Phase Map

- Phase 1: Reconcile milestone and card control plane.
- Phase 2: Create and verify `v1.0.0-rc.1`.
- Phase 3: Record final stable-root, CLI, onboarding, and extension evidence.
- Phase 4: Decide final `v1.0.0` or create blocker cards for the next rc.

## Card Inventory

| Card | Goal | Primary Module | Owned Files | Depends On | Quick Gates |
|------|------|----------------|-------------|------------|-------------|
| 1429 | Add verify artifacts and reconcile milestone state | release | `tasks/milestones/*` | none | `agent-workflow`, `module-manifests`, `git status` |
| 1430 | Create and verify `v1.0.0-rc.1` | release | `docs/release/v1.0.0-rc.1.md`, card 1375 | 1429 | `make gates`, `extension-beta-evidence`, `git tag -l` |
| 1431 | Record stable-root final freeze evidence | stable roots | `docs/stable-api/README.md`, release notes | 1430 | stable-root race tests, vet, deprecation strict |
| 1432 | Record CLI and onboarding smoke evidence | `cmd/plumego` | CLI docs, getting-started docs, release notes | 1431 | `cmd/plumego` tests, scaffold dry-runs |
| 1433 | Record extension maturity boundary evidence | extension evidence | maturity docs, evidence ledger, release notes | 1432 | `extension-maturity`, `extension-beta-evidence` |
| 1434 | Make final v1 decision | release | final release notes, active cards | 1433 | release checklist, full gates, git status |

## Dependency Edges

- `1429 -> 1430`
- `1430 -> 1431`
- `1431 -> 1432`
- `1432 -> 1433`
- `1433 -> 1434`

## Parallel Groups

- None. Release execution is intentionally sequential.

## Risk Register

- Risk: release evidence is incomplete or local-only.
  Mitigation: require rc tag, local gates, GitHub gates, and release notes.
- Risk: experimental extensions are advertised as v1-stable.
  Mitigation: preserve blockers in the evidence ledger and maturity dashboard.
- Risk: active tasks drift from executable reality.
  Mitigation: reconcile verify artifacts and active queue before tagging.
- Risk: a gate failure encourages broad fixes.
  Mitigation: create one bounded blocker card per module and stop release flow.

## Verification Strategy

- Card-level checks: each card lists no more than three quick gates.
- Milestone-level checks: run the pre-v1 release checklist before final decision.
- Release evidence: record command output summaries, tag refs, and remote gate
  status in release notes or verify artifacts.

## Exit Condition

- all planned cards completed or explicitly superseded
- verify report shows pass or a documented no-go
- final `v1.0.0` is tagged only after rc observation closes cleanly
