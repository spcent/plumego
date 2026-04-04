# Milestone Specs

This directory is the human input surface for Codex-driven autonomous evolution.

## Model

```
Human writes Milestone Spec  →  codex --yolo  →  Milestone PR  →  Human review
```

One spec file = one coherent feature or refactor scope = one PR.  
Codex executes; you review the result at the PR boundary only.

## Directory Layout

```
tasks/milestones/
├── README.md         ← this file
├── TEMPLATE.md       ← copy this for every new milestone
├── active/           ← specs ready for Codex execution
│   └── M-001.md
└── done/             ← completed milestones with outcome notes
    └── M-000.md
```

## How to Invoke Codex

```bash
# Single milestone
make milestone M=active/M-001

# Or directly
codex --yolo "$(cat tasks/milestones/active/M-001.md)"
```

Codex reads `AGENTS.md` automatically on startup for behavior boundaries.  
The milestone spec is the task prompt — it provides goal, scope, ordered steps,
and acceptance criteria.

## How to Write an Effective Spec

A good spec has five parts:

1. **Goal** — one sentence stating what should be true when done.
2. **Context** — which files Codex must read before touching anything.
3. **Tasks** — ordered, atomic steps small enough to verify individually.
4. **Acceptance Criteria** — the exact commands that must pass (maps to quality gates).
5. **Out of Scope** — hard stops to prevent scope creep.

Keep specs under ~80 lines. If it grows longer, split it into two milestones.

## Milestone Lifecycle

| State  | Location              | Meaning                          |
|--------|-----------------------|----------------------------------|
| active | `milestones/active/`  | Ready to run; spec is complete   |
| done   | `milestones/done/`    | Executed; PR merged or closed    |

Move a file from `active/` to `done/` after the PR is merged.
Add an `## Outcome` section to the done file with the PR number and
which validation commands actually ran.

## Numbering

Format: `M-NNN.md` (e.g., `M-001.md`, `M-042.md`).  
Pick the next available integer. No gaps required.

## What Makes a Good Milestone Boundary

- One primary module touched (occasionally two related ones).
- Can be reviewed end-to-end in a single PR pass.
- All acceptance criteria are mechanical (commands, not opinions).
- Does not depend on another in-flight milestone.
