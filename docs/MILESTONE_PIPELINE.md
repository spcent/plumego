# Milestone Pipeline

Lightweight execution contract for Plumego milestone work.

This pipeline keeps one milestone equal to one PR, but makes the internal
handoffs explicit so parallel work stays reviewable and verifiable.

## Model

```text
Milestone Spec
  -> Plan
  -> Cards
  -> Verify
  -> PR
  -> Human Review
```

## Artifact Rules

| Stage | Purpose | Owner | Output |
|------|---------|-------|--------|
| Milestone | Define scope and hard constraints | Human | `tasks/milestones/active/M-XXX.md` |
| Plan | Turn scope into ordered execution units | Planner | `tasks/milestones/PLAN_TEMPLATE.md` shape |
| Cards | Small reversible work units | Decomposer / Workers | `tasks/cards/TEMPLATE.md` shape |
| Verify | Collect evidence and final verdict | Verifier | `tasks/milestones/VERIFY_TEMPLATE.md` shape |
| PR | Reviewer-facing package | PR Packager | `docs/github-workflows/milestone-pr-template.md` |

## Core Rules

- `1 milestone = 1 PR`
- `1 card = 1 primary module`
- `1 worker = 1 card ownership set`
- parallel execution is allowed only when card-owned files do not overlap
- every milestone must have one final verify report before PR packaging
- milestone-level gates do not replace card-level quick validation

## Stage Contracts

### 1. Milestone

Milestone is the only human-authored scope boundary.

Use [`tasks/milestones/TEMPLATE.md`](../tasks/milestones/TEMPLATE.md) and keep these fields authoritative:

- `M-XXX: <Title>`
- `Branch`
- `Depends on`
- `Parallel OK`
- `Goal`
- `Architecture Decisions`
- `Context — Read Before Touching Code`
- `Affected Modules`
- `Tasks`
- `Acceptance Criteria`
- `Out of Scope`
- `Open Questions -> Codex`
- `Done Definition`

Milestone should answer:

- what must be true when done
- which modules are in scope
- what must not be changed
- which gates decide completion

Milestone should not answer:

- exact card splits
- card ordering inside a parallel phase
- file-level worker ownership

### 2. Plan

Plan is the execution manifest produced from the milestone before worker code changes start.

Recommended file name:

- `tasks/milestones/M-XXX.plan.md`

Required fields:

- `Milestone`
- `Objective`
- `Constraints`
- `Affected Modules`
- `Phase Map`
- `Card Inventory`
- `Dependency Edges`
- `Parallel Groups`
- `Risk Register`
- `Verification Strategy`
- `Exit Condition`

Planner rules:

- do not widen milestone scope
- split work into cards that fit one focused pass
- assign exact file ownership whenever possible
- move shared-risk work earlier

### 3. Cards

Cards are the only worker-executable units.

Use [`tasks/cards/TEMPLATE.md`](../tasks/cards/TEMPLATE.md).

Required fields:

- `Card`
- `Milestone`
- `Priority`
- `State`
- `Primary Module`
- `Owned Files`
- `Depends On`
- `Goal`
- `Scope`
- `Non-goals`
- `Files`
- `Tests`
- `Docs Sync`
- `Done Definition`
- `Outcome`

Card rules:

- one primary module
- up to 5 files by default
- quick validation only; repo-wide gates belong to verify
- if exported symbols change, follow `AGENTS.md §7.1`

### 4. Verify

Verify is the milestone-level evidence bundle.

Recommended file name:

- `tasks/milestones/M-XXX.verify.md`

Required fields:

- `Milestone`
- `Branch`
- `Verified Cards`
- `Scope Check`
- `Ownership Check`
- `Symbol Completeness Check`
- `Module Test Summary`
- `Boundary Check Summary`
- `Repo Gate Summary`
- `Open Issues`
- `Final Verdict`

Verifier rules:

- verify card outcomes against the original milestone scope
- fail if out-of-scope files were touched without explicit justification
- fail if any required gate output is missing
- fail if a removed or renamed exported symbol still has residual callers

### 5. PR

PR packaging remains reviewer-facing and should stay terse.

Use [`docs/github-workflows/milestone-pr-template.md`](./github-workflows/milestone-pr-template.md).

Required inputs to the packager:

- milestone goal and architecture decisions
- plan summary
- card outcomes
- verify verdict
- exact gate output

Packager rules:

- summarize decisions, not a file dump
- surface deviations explicitly
- include every changed file in scope-boundary review
- paste gate output verbatim

## Recommended Flow

1. Human writes milestone spec.
2. Planner emits one plan file.
3. Decomposer emits ordered cards from that plan.
4. Workers execute cards, respecting `Owned Files`.
5. Verifier runs module checks, boundary checks, then repo-wide gates.
6. PR packager fills the PR body from milestone, plan, cards, and verify evidence.

## Minimal Templates

### Milestone Template Fields

```md
# M-XXX: <Title>
- Branch:
- Depends on:
- Parallel OK:

## Goal
## Architecture Decisions
## Context — Read Before Touching Code
## Affected Modules
## Tasks
## Acceptance Criteria
## Out of Scope
## Open Questions -> Codex
## Done Definition
```

### Plan Template Fields

```md
# Plan for M-XXX: <Title>

Milestone:
Objective:
Constraints:
Affected Modules:

## Phase Map
## Card Inventory
## Dependency Edges
## Parallel Groups
## Risk Register
## Verification Strategy
## Exit Condition
```

### Card Template Fields

```md
# Card C-XXX

Milestone:
Priority:
State:
Primary Module:
Owned Files:
Depends On:

Goal:
Scope:
Non-goals:
Files:
Tests:
Docs Sync:
Done Definition:
Outcome:
```

### Verify Template Fields

```md
# Verify M-XXX: <Title>

Milestone:
Branch:
Verified Cards:

## Scope Check
## Ownership Check
## Symbol Completeness Check
## Module Test Summary
## Boundary Check Summary
## Repo Gate Summary
## Open Issues
## Final Verdict
```

### PR Template Fields

```md
## milestone(M-XXX): <Title>
### Goal
### Approach
### Architecture Decisions — Implemented As Specified
### Scope Boundary Check
### Open Questions Resolved
### Quality Gate Results
### Reviewer Checklist
```
