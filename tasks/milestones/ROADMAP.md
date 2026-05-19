# Milestone Roadmap

Pipeline view of all milestones — sequencing, dependencies, and status.  
Update this file when you add, complete, or reorder milestones.

## Status Key

| Symbol | Meaning |
|--------|---------|
| `[ ]`  | Not started |
| `[→]`  | In execution (agent running) |
| `[PR]` | PR open, awaiting review |
| `[✓]`  | Merged |
| `[✗]`  | Cancelled / superseded |

---

## Pipeline

<!--
  Format per row:
  | M-NNN | Title | Status | Depends On | Notes |

  Add new milestones at the bottom. Do not reorder completed milestones.
-->

| Milestone | Title | Status | Depends On | Notes |
|-----------|-------|--------|------------|-------|
| M-001 | v1 trust baseline | [✓] | — | Verify artifact created; archived |
| M-002 | Stable roots freeze and reliability | [✓] | M-001 | Verify artifact created; final release gates moved to M-005 |
| M-003 | Extension evidence pipeline | [✓] | M-002 | Verify artifact created; remaining blockers are explicit |
| M-004 | Stable root cleanup freeze | [✓] | M-002 | Outcome recorded and archived |
| M-005 | v1 release execution | [✓] | M-001, M-002, M-003, M-004 | `v1.0.0` release evidence recorded |
| M-006 | v1.0.1 maintenance lane | [✓] | M-005 | Post-v1 maintenance evidence recorded |
| M-007 | extension v1 baseline evidence intake | [✓] | M-006 | First `v1.0.0` release-ref intake recorded without promoting extensions |
| M-008 | v1.1.0 Release Execution | [✓] | M-007 | Archived with release evidence and beta second-release refs |
| M-009 | Beta Promotions Round 1 | [→] | M-008 | Still blocked on x/gateway/discovery core-static evidence |
| M-010 | Beta Promotions Round 2 | [✓] | M-009 | Archived with promotion evidence |
| M-011 | Benchmark Suite Publication | [✓] | M-008 | Archived; router and middleware benchmarks live in benchmark/ |
| M-012 | Input Validation Bridge | [✓] | M-008 | Archived; x/validate plus optional go-playground adapter |
| M-013 | Migration Guides | [✓] | M-008 | Archived; docs/migration guides and adoption links present |
| M-014 | OpenAPI 3.1 Generation | [✓] | M-009 | Archived; x/openapi generator and plumego generate spec CLI command |
| M-015 | Database Adapters | [✓] | M-009 | Archived; dependency-free x/data adapter contracts and migrate helpers |
| M-016 | Event-Driven Reference Architecture | [ ] | M-010 | Parallel; reference/with-events using x/messaging primitives |
| M-017 | gRPC Support | [✓] | M-014 | Archived; dependency-free x/rpc wrappers plus reference/with-rpc concrete gRPC wiring |
| M-018 | Multi-Tenant Admin Reference | [✓] | M-010 | Archived; reference/with-tenant-admin using x/tenant primitives |
| M-019 | Module Ecosystem Foundation | [✓] | M-017 | Archived; community extension schema, plumego add command, authoring guide |
| M-020 | Community Discovery & Documentation Overhaul | [✓] | M-008 | Archived; README overhaul, stdlib comparison table, stable root pkg.go.dev doc comments |
| M-021 | Agent-First Ecosystem Visibility | [✓] | M-019 | Archived; docs/AGENT_FIRST.md, quality-gates workflow updates, README agent-first section |

---

## Dependency Graph

<!--
  ASCII DAG. Redraw when dependencies change.
  Example:
    M-001 ──► M-003 ──► M-005
    M-002 ──┘
    M-004 (independent)

  Keep it simple. If the graph becomes complex, split milestones.
-->

```
M-001 ──► M-002 ──► M-003 ──┐
             └────► M-004 ──┼──► M-005
                              │
M-001 ────────────────────────┘
M-005 ──► M-006 ──► M-007 ──► M-008 ──► M-009 ──► M-010 ──► M-016
                               │         │                    │
                               │         ├──► M-014 ──► M-017 ──► M-019 ──► M-021
                               │         │
                               │         └──► M-015
                               │
                               ├──► M-011   M-009 ──► M-015
                               ├──► M-012   M-010 ──► M-018
                               ├──► M-013
                               └──► M-020
```

---

## Guidelines

### Sequencing Rules

- A milestone may not start agent execution until all its `Depends On`
  milestones are **merged** to `main`.
- Milestones marked `Parallel OK: yes` with no dependencies can run
  concurrently as long as their **Affected Modules** do not overlap.
- If two parallel milestones accidentally touch the same file, run them
  sequentially instead and update this roadmap.

### Adding a Milestone

1. Copy `TEMPLATE.md` → `active/M-NNN-short-name/M-NNN.md`.
2. Fill all sections.
3. Add a row to the Pipeline table above.
4. If it depends on another milestone, add the edge to the Dependency Graph.
5. Run `make check-spec M=active/M-NNN-short-name` to validate the spec before committing.

### Completing a Milestone

1. PR merged to `main`.
2. Move `active/M-NNN-short-name/` → `done/M-NNN-short-name/`.
3. Add `## Outcome` to the done file (PR number + gate output summary).
4. Update the Pipeline table row: `[ ]` → `[✓]`.
5. If dependents exist, update their status to `[ ]` (now unblocked).
