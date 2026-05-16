# Milestone Roadmap

Pipeline view of all milestones — sequencing, dependencies, and status.  
Update this file when you add, complete, or reorder milestones.

## Status Key

| Symbol | Meaning |
|--------|---------|
| `[ ]`  | Not started |
| `[→]`  | In execution (Codex running) |
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
| M-001 | v1 trust baseline | [→] | — | Verify artifact created; archival remains separate |
| M-002 | Stable roots freeze and reliability | [→] | M-001 | Verify artifact created; final release gates move to M-005 |
| M-003 | Extension evidence pipeline | [→] | M-002 | Verify artifact created; remaining blockers are explicit |
| M-004 | Stable root cleanup freeze | [→] | M-002 | Outcome recorded; reconcile in card 1429 |
| M-005 | v1 release execution | [→] | M-001, M-002, M-003, M-004 | Sequential rc.1 to final v1/no-go plan |
| M-006 | v1.0.1 maintenance lane | [→] | M-005 | Post-v1 maintenance cards for generated data, CI warnings, CLI onboarding truth, and release evidence index |

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
M-005 ───────────────────────────► M-006
```

---

## Guidelines

### Sequencing Rules

- A milestone may not start Codex execution until all its `Depends On`
  milestones are **merged** to `main`.
- Milestones marked `Parallel OK: yes` with no dependencies can run
  concurrently as long as their **Affected Modules** do not overlap.
- If two parallel milestones accidentally touch the same file, run them
  sequentially instead and update this roadmap.

### Adding a Milestone

1. Copy `TEMPLATE.md` → `active/M-NNN.md`.
2. Fill all sections.
3. Add a row to the Pipeline table above.
4. If it depends on another milestone, add the edge to the Dependency Graph.
5. Run `make check-spec M=active/M-NNN` to validate the spec before committing.

### Completing a Milestone

1. PR merged to `main`.
2. Move `active/M-NNN.md` → `done/M-NNN.md`.
3. Add `## Outcome` to the done file (PR number + gate output summary).
4. Update the Pipeline table row: `[ ]` → `[✓]`.
5. If dependents exist, update their status to `[ ]` (now unblocked).
